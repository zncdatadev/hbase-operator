package controller

import (
	"context"
	"slices"

	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/config"
	"github.com/zncdatadev/operator-go/pkg/productlogging"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

// RBAC for the resources the SDK GenericReconciler owns on behalf of an HbaseCluster, plus the
// referenced AuthenticationClass. Regenerate config/rbac/role.yaml with `make manifests`.
//
// +kubebuilder:rbac:groups=hbase.kubedoop.dev,resources=hbaseclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hbase.kubedoop.dev,resources=hbaseclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hbase.kubedoop.dev,resources=hbaseclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps;secrets;services;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=authentication.kubedoop.dev,resources=authenticationclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list;watch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

const (
	metricsPortName  = "metrics"
	uiPortName       = "ui-http"
	hdfsConfigVolume = "hdfs-config"
	prometheusPath   = "/prometheus"
)

// roleContainerPorts reproduces the pre-Gen3 per-role container port sets.
var roleContainerPorts = map[string][]corev1.ContainerPort{
	hbasev1alpha1.MasterRole: {
		{Name: hbasev1alpha1.MasterRole, ContainerPort: 16000, Protocol: corev1.ProtocolTCP},
		{Name: uiPortName, ContainerPort: 16010, Protocol: corev1.ProtocolTCP},
	},
	hbasev1alpha1.RegionServerRole: {
		{Name: hbasev1alpha1.RegionServerRole, ContainerPort: 16020, Protocol: corev1.ProtocolTCP},
		{Name: uiPortName, ContainerPort: 16030, Protocol: corev1.ProtocolTCP},
		{Name: metricsPortName, ContainerPort: 9100, Protocol: corev1.ProtocolTCP},
	},
	hbasev1alpha1.RestServerRole: {
		{Name: RestHTTPPortName, ContainerPort: 8080, Protocol: corev1.ProtocolTCP},
		{Name: uiPortName, ContainerPort: 8085, Protocol: corev1.ProtocolTCP},
		{Name: metricsPortName, ContainerPort: 9100, Protocol: corev1.ProtocolTCP},
	},
}

// roleMetricsPorts is the port each role's native Prometheus endpoint (/prometheus, HBase
// 2.6.1+) serves on. It parametrizes the per-role-group metrics Service.
var roleMetricsPorts = map[string]int32{
	hbasev1alpha1.MasterRole:       16010,
	hbasev1alpha1.RegionServerRole: 16030,
	hbasev1alpha1.RestServerRole:   8085,
}

// probePortNames maps each role to the container port its TCP probes target.
var probePortNames = map[string]string{
	hbasev1alpha1.MasterRole:       hbasev1alpha1.MasterRole,
	hbasev1alpha1.RegionServerRole: hbasev1alpha1.RegionServerRole,
	hbasev1alpha1.RestServerRole:   RestHTTPPortName,
}

// HbaseRoleGroupHandler is HBase's implementation of all three product seams:
//
//   - RoleProvider (DeclareRoles) — what each role IS: ports, primary container name, the
//     entrypoint, the probe set, the log producers and the default scheduling policy.
//   - RoleGroupResolver (ResolveRoleGroup) — what follows from the CR and the live cluster:
//     hbase-site.xml, including the zookeeper keys read from the znode discovery ConfigMap.
//   - RoleGroupHandler (BuildResources) — the few things neither seam can express: the
//     whole-file hbase-env.sh, the HDFS/Kerberos volumes and the Prometheus metrics Service.
//
// It embeds BaseRoleGroupHandler, so the framework owns the role group ConfigMap, the Services,
// the StatefulSet (sidecars and podOverrides included) and the role-level PDB.
type HbaseRoleGroupHandler struct {
	*reconciler.BaseRoleGroupHandler[*hbasev1alpha1.HbaseCluster]
}

var (
	_ reconciler.RoleGroupHandler[*hbasev1alpha1.HbaseCluster] = &HbaseRoleGroupHandler{}
	_ reconciler.RoleProvider[*hbasev1alpha1.HbaseCluster]     = &HbaseRoleGroupHandler{}
)

// NewHbaseRoleGroupHandler creates the handler. It carries only reconcile-invariant
// collaborators; everything a role is made of is declared per reconcile by DeclareRoles, with
// the cr in hand.
func NewHbaseRoleGroupHandler(scheme *runtime.Scheme) *HbaseRoleGroupHandler {
	base := reconciler.NewBaseRoleGroupHandler[*hbasev1alpha1.HbaseCluster](scheme)

	// hbase-site.xml / ssl-*.xml render with the Hadoop-style XML adapter; anything else
	// key-value falls back to properties.
	base.ConfigGenerator = config.NewMultiFormatConfigGenerator()
	base.ConfigGenerator.RegisterDefaultFormats()

	// The role group ConfigMap is mounted at /kubedoop/mount/config/hbase; the entrypoint
	// copies it (plus the HDFS discovery config) into /kubedoop/config.
	base.ConfigMountPath = HbaseMountConfigDir

	return &HbaseRoleGroupHandler{BaseRoleGroupHandler: base}
}

// DeclareRoles implements reconciler.RoleProvider: one statement per role, produced once per
// reconcile with the cr in hand. The entrypoint is declared here rather than patched onto the
// built container, so a user's podOverrides still outrank it.
func (h *HbaseRoleGroupHandler) DeclareRoles(
	_ context.Context, _ client.Client, cr *hbasev1alpha1.HbaseCluster,
) (reconciler.RoleCatalog, error) {
	krb := KerberosConfigFor(cr)

	catalog := make(reconciler.RoleCatalog, len(roleContainerPorts))
	for role, ports := range roleContainerPorts {
		probe := roleProbeHandler(role)

		catalog[role] = reconciler.RoleDeclaration{
			// All three typed role blocks are optional in the HbaseCluster API. Declaring a
			// supported role that this cluster omits must therefore not emit a warning.
			Optional: true,

			// The primary container is named after the role, matching the per-container
			// logging key below.
			MainContainerName: role,
			ContainerPorts:    ports,
			ServicePorts:      servicePortsFrom(ports),

			// HBase has no user-facing `command`, so declaring the entrypoint beats nobody.
			// The script is the last element of Command: RoleDeclaration deliberately has no
			// Args, because args reach the container through cliOverrides, which a user CAN
			// state and which a product-appended list would silently replace.
			Command: slices.Concat(mainContainerCommand, mainContainerArgs(role, krb)),

			// The pre-Gen3 probe set, restated here because the framework generates only a
			// readiness probe and would target ContainerPorts[0] rather than the role's RPC port.
			StartupProbe: &corev1.Probe{
				ProbeHandler:        probe,
				InitialDelaySeconds: 120,
				PeriodSeconds:       5,
				FailureThreshold:    3,
				TimeoutSeconds:      10,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler:        probe,
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				FailureThreshold:    3,
				TimeoutSeconds:      10,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler:     probe,
				PeriodSeconds:    10,
				FailureThreshold: 3,
				TimeoutSeconds:   10,
			},

			LogProducers: []productlogging.ContainerLogging{
				{
					Container: role,
					Framework: productlogging.LoggingFrameworkLog4j,
					Pattern:   ConsoleConversionPattern,
				},
			},

			// The default scheduling policy is a product default for the framework-owned half
			// of the role's config block, folded BENEATH whatever the CR states — so a user's
			// `config.affinity` still wins, without the product patching the built pod.
			ConfigDefaults: defaultAffinityConfig(cr.Name, role),
		}
	}
	return catalog, nil
}

// roleProbeHandler is the TCP probe target of a role: its RPC port, not ContainerPorts[0].
func roleProbeHandler(roleName string) corev1.ProbeHandler {
	return corev1.ProbeHandler{
		TCPSocket: &corev1.TCPSocketAction{Port: intstr.FromString(probePortNames[roleName])},
	}
}

// BuildResources delegates the bulk to the framework, then applies the HBase-specific pieces.
func (h *HbaseRoleGroupHandler) BuildResources(
	ctx context.Context,
	k8sClient client.Client,
	cr *hbasev1alpha1.HbaseCluster,
	buildCtx *reconciler.RoleGroupBuildContext,
) (*reconciler.RoleGroupResources, error) {
	// Front the role's UI with the framework's oauth2-proxy sidecar when OIDC is configured.
	// Registration must precede the base build — injection happens inside it.
	if err := registerOidcSidecar(ctx, k8sClient, cr, buildCtx, uiPortFor(buildCtx.RoleName)); err != nil {
		return nil, err
	}

	krb := KerberosConfigFor(cr)

	// Volume providers are consumed by the base build, i.e. before podOverrides are
	// strategic-merged, which is what keeps a user's overrides outranking the product and what
	// puts these mounts through the framework's own duplicate-mount validation.
	buildCtx.VolumeProviders = append(buildCtx.VolumeProviders, hdfsConfigVolumeProvider(cr))
	if krb != nil {
		buildCtx.VolumeProviders = append(buildCtx.VolumeProviders, krb)
	}

	resources, err := h.BaseRoleGroupHandler.BuildResources(ctx, k8sClient, cr, buildCtx)
	if err != nil {
		return nil, err
	}

	// hbase-env.sh is a whole shell file, not key-value config, so it bypasses the merge
	// pipeline. Writing it only when absent keeps a user-supplied file authoritative.
	if resources.ConfigMap != nil {
		if _, exists := resources.ConfigMap.Data[HbaseEnvFileName]; !exists {
			resources.ConfigMap.Data[HbaseEnvFileName] = hbaseEnvSh(buildCtx.RoleName, krb)
		}
	}

	resources.MetricsService = h.buildMetricsService(buildCtx, resources.ConfigMap.Labels)

	return resources, nil
}

// hdfsConfigVolumeProvider mounts the HDFS discovery ConfigMap beside the role group config; the
// entrypoint copies both into the effective config dir.
func hdfsConfigVolumeProvider(cr *hbasev1alpha1.HbaseCluster) reconciler.VolumeProvider {
	return &configMapVolumeProvider{
		volumeName:    hdfsConfigVolume,
		configMapName: cr.Spec.ClusterConfigSpec.HdfsConfigMapName,
		mountPath:     HdfsMountConfigDir,
	}
}

// configMapVolumeProvider adapts a ConfigMap into the framework's VolumeProvider contract.
type configMapVolumeProvider struct {
	volumeName    string
	configMapName string
	mountPath     string
}

var _ reconciler.VolumeProvider = &configMapVolumeProvider{}

func (p *configMapVolumeProvider) Volumes() []corev1.Volume {
	defaultMode := int32(420)
	return []corev1.Volume{
		{
			Name: p.volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode:          &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{Name: p.configMapName},
				},
			},
		},
	}
}

func (p *configMapVolumeProvider) VolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{Name: p.volumeName, MountPath: p.mountPath},
	}
}

// servicePortsFrom converts a role's container ports into client Service ports, targeting the
// named container port so a port renumber cannot silently detach the Service.
func servicePortsFrom(containerPorts []corev1.ContainerPort) []corev1.ServicePort {
	ports := make([]corev1.ServicePort, 0, len(containerPorts))
	for _, cp := range containerPorts {
		ports = append(ports, corev1.ServicePort{
			Name:       cp.Name,
			Port:       cp.ContainerPort,
			Protocol:   cp.Protocol,
			TargetPort: intstr.FromString(cp.Name),
		})
	}
	return ports
}

// uiPortFor returns the role's ui-http container port (the oauth2-proxy upstream).
func uiPortFor(roleName string) int32 {
	for _, p := range roleContainerPorts[roleName] {
		if p.Name == uiPortName {
			return p.ContainerPort
		}
	}
	return 0
}

// buildMetricsService builds the per-role-group Prometheus metrics Service. HBase 2.6.1+ serves
// the Prometheus text format natively at /prometheus, on the role's UI port. The shape (name,
// headless, scrape annotations, selector) is the framework builder's; the observability e2e suite
// asserts it field by field.
func (h *HbaseRoleGroupHandler) buildMetricsService(
	buildCtx *reconciler.RoleGroupBuildContext,
	baseLabels map[string]string,
) *corev1.Service {
	return builder.NewMetricsServiceBuilder(
		buildCtx.ResourceName,
		buildCtx.ClusterNamespace,
		roleMetricsPorts[buildCtx.RoleName],
		baseLabels,
	).
		WithPath(prometheusPath).
		// HBase exposes /prometheus on the role's UI port. Keep the Service port named
		// "metrics" for discovery, but route it to the container's real ui-http port.
		WithTargetPortName(uiPortName).
		WithSelector(h.SelectorLabels(buildCtx)).
		Build()
}
