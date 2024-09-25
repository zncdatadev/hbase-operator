package common

import (
	"context"
	"path"

	authv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/authentication/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/constants"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/authz"
	auzhz "github.com/zncdatadev/hbase-operator/internal/controller/authz"
)

var (
	roleNameToCommandArg = map[string]string{
		"master":       "master",
		"regionserver": "regionserver",
		"restserver":   "rest",
	}
	HBaseConfigDir      = path.Join(constants.KubedoopConfigDir)
	HDFSConfigDir       = path.Join(constants.KubedoopConfigDir)
	HBaseMountConfigDir = path.Join(constants.KubedoopConfigDirMount, "hbase")
	HDFSMountConfigDir  = path.Join(constants.KubedoopConfigDirMount, "hdfs")
)

var _ builder.StatefulSetBuilder = &StatefulSetBuilder{}

type StatefulSetBuilder struct {
	builder.StatefulSet
	Ports         []corev1.ContainerPort
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
	RoleName      string
	ClusterName   string

	krb5Config *auzhz.HbaseKerberosConfig
}

func NewStatefulSetBuilder(
	client *client.Client,
	name string,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	replicas *int32,
	ports []corev1.ContainerPort,
	image *util.Image,
	krb5SecretClass string,
	tlsSecretClass string,
	options builder.WorkloadOptions,
) *StatefulSetBuilder {

	var krb5Config *auzhz.HbaseKerberosConfig
	if krb5SecretClass != "" && tlsSecretClass != "" {
		krb5Config = auzhz.NewHbaseKerberosConfig(
			client.GetOwnerNamespace(),
			options.ClusterName,
			options.RoleName,
			options.RoleGroupName,
			krb5SecretClass,
			tlsSecretClass,
		)
	}

	return &StatefulSetBuilder{
		StatefulSet:   *builder.NewStatefulSetBuilder(client, name, replicas, image, options),
		Ports:         ports,
		ClusterConfig: clusterConfig,
		RoleName:      options.RoleName,
		ClusterName:   options.ClusterName,
		krb5Config:    krb5Config,
	}

}

func (b *StatefulSetBuilder) getVolumeMounts() []corev1.VolumeMount {
	objs := []corev1.VolumeMount{
		{
			Name:      "hdfs-config",
			MountPath: path.Join(constants.KubedoopConfigDirMount, "hdfs"),
		},
		{
			Name:      "hbase-config",
			MountPath: path.Join(constants.KubedoopConfigDirMount, "hbase"),
		},
	}
	return objs
}

func (b *StatefulSetBuilder) getVolumes() []corev1.Volume {

	defaultMode := int32(420)

	objs := []corev1.Volume{
		{
			Name: "hdfs-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.ClusterConfig.HdfsConfigMapName,
					},
				},
			},
		},
		{
			Name: "hbase-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.GetName(),
					},
				},
			},
		},
	}

	return objs
}

func (b *StatefulSetBuilder) GetMainContainerCommanArgs() []string {
	hbaseSubArg := roleNameToCommandArg[b.RoleName]

	setupKrb5 := ""
	if b.krb5Config != nil {
		setupKrb5 = b.krb5Config.GetContainerCommands()
	}

	arg := `mkdir -p ` + HBaseConfigDir + `
cp ` + path.Join(HBaseMountConfigDir, "*") + ` ` + constants.KubedoopConfigDir + `
cp ` + path.Join(HDFSMountConfigDir, "*") + ` ` + constants.KubedoopConfigDir + `

` + setupKrb5 + `

prepare_signal_handlers()
{
	unset term_child_pid
	unset term_kill_needed
	trap 'handle_term_signal' TERM
}

handle_term_signal()
{
	if [ "${term_child_pid}" ]; then
		kill -TERM "${term_child_pid}" 2>/dev/null
	else
		term_kill_needed="yes"
	fi
}

wait_for_termination()
{
	set +e
	term_child_pid=$1
	if [[ -v term_kill_needed ]]; then
		kill -TERM "${term_child_pid}" 2>/dev/null
	fi
	wait ${term_child_pid} 2>/dev/null
	trap - TERM
	wait ${term_child_pid} 2>/dev/null
	set -e
}

prepare_signal_handlers
bin/hbase ` + hbaseSubArg + ` start &
wait_for_termination $!
`
	intendedString := util.IndentTab4Spaces(arg)
	return []string{intendedString}
}

func (b *StatefulSetBuilder) getEnvVars() []corev1.EnvVar {
	objs := []corev1.EnvVar{
		{
			Name:  "HBASE_CONF_DIR",
			Value: HBaseConfigDir,
		},
		{
			Name:  "HADOOP_CONF_DIR",
			Value: HDFSConfigDir,
		},
	}

	return objs
}

func (b *StatefulSetBuilder) buildContainer() []corev1.Container {
	containers := []corev1.Container{}
	mainContainerBuilder := builder.NewContainer(b.RoleName, b.GetImage()).
		SetCommand([]string{"/bin/bash", "-x", "-euo", "pipefail", "-c"}).
		SetArgs(b.GetMainContainerCommanArgs()).
		AddVolumeMounts(b.getVolumeMounts()).
		AddEnvVars(b.getEnvVars()).
		AddPorts(b.Ports).
		SetStartupProbe(b.GetStartupProbe()).
		SetLivenessProbe(b.GetLivenessProbe()).
		SetReadinessProbe(b.GetReadinessProbe())

	if b.krb5Config != nil {
		mainContainerBuilder.AddEnvVars(b.krb5Config.GetContainerEnvvars())
		mainContainerBuilder.AddVolumeMounts(b.krb5Config.GetVolumeMounts())
	}

	// TODO: add vector container

	containers = append(containers, *mainContainerBuilder.Build())

	return containers
}

func (b *StatefulSetBuilder) GetDefaultAffinityBuilder() *AffinityBuilder {
	affinityLabels := map[string]string{
		constants.LabelKubernetesInstance: b.ClusterName,
		constants.LabelKubernetesName:     "hbase",
	}
	antiAffinityLabels := map[string]string{
		constants.LabelKubernetesInstance:  b.ClusterName,
		constants.LabelKubernetesName:      "hbase",
		constants.LabelKubernetesComponent: b.RoleName,
	}

	affinity := NewAffinityBuilder(
		*NewPodAffinity(affinityLabels, false, false).Weight(20),
		*NewPodAffinity(antiAffinityLabels, false, true).Weight(70),
	)

	return affinity
}

func (b *StatefulSetBuilder) getProbeHandler() *corev1.ProbeHandler {
	// switch-cas
	var prob *corev1.ProbeHandler
	switch b.RoleName {
	case "master":
		prob = &corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(b.RoleName),
			},
		}
	case "regionserver":
		prob = &corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(b.RoleName),
			},
		}
	case "restserver":
		prob = &corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString("rest-http"),
			},
		}
	default:
		return nil
	}

	return prob
}

func (b *StatefulSetBuilder) GetLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        *b.getProbeHandler(),
		InitialDelaySeconds: 10,
		PeriodSeconds:       10,
		FailureThreshold:    3,
		TimeoutSeconds:      10,
	}
}

func (b *StatefulSetBuilder) GetReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:     *b.getProbeHandler(),
		PeriodSeconds:    10,
		FailureThreshold: 3,
		TimeoutSeconds:   10,
	}
}

func (b *StatefulSetBuilder) GetStartupProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        *b.getProbeHandler(),
		InitialDelaySeconds: 120,
		PeriodSeconds:       5,
		FailureThreshold:    3,
		TimeoutSeconds:      10,
	}
}
func (b *StatefulSetBuilder) buildOidcContainer(ctx context.Context) (*corev1.Container, error) {
	if b.ClusterConfig.Authentication == nil || b.ClusterConfig.Authentication.AuthenticationClass == "" {
		return nil, nil
	}

	authClass := &authv1alpha1.AuthenticationClass{}
	if err := b.GetClient().GetWithOwnerNamespace(ctx, b.ClusterConfig.Authentication.AuthenticationClass, authClass); err != nil {
		return nil, err
	}

	if authClass.Spec.AuthenticationProvider.OIDC == nil || b.ClusterConfig.Authentication.Oidc == nil {
		logger.V(5).Info("OIDC not configured", "cluster", b.ClusterName, "role", b.RoleName, "authClass", b.ClusterConfig.Authentication.AuthenticationClass, "oidc", b.ClusterConfig.Authentication.Oidc)
		return nil, nil
	}

	var upstreamPort int32

	for _, port := range b.Ports {
		if port.Name == "ui-http" {
			upstreamPort = port.ContainerPort
			break
		}
	}

	oidc := authz.NewOidc(
		string(b.GetClient().GetOwnerReference().GetUID()),
		b.GetImage(),
		upstreamPort,
		b.ClusterConfig.Authentication.Oidc,
		authClass.Spec.AuthenticationProvider.OIDC,
	)

	return oidc.GetContainer(), nil
}

func (b *StatefulSetBuilder) Build(ctx context.Context) (ctrlclient.Object, error) {
	b.AddContainers(b.buildContainer())
	b.AddVolumes(b.getVolumes())
	b.SetAffinity(b.GetDefaultAffinityBuilder().Build())

	if b.krb5Config != nil {
		b.AddVolumes(b.krb5Config.GetVolumes())
	}

	oidcContainer, err := b.buildOidcContainer(ctx)
	if err != nil {
		return nil, err
	}

	if oidcContainer != nil {
		b.AddContainer(oidcContainer)
	}

	return b.GetObject()
}

var _ reconciler.Reconciler = &StatefulSetReconciler{}

type StatefulSetReconciler struct {
	reconciler.StatefulSet
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func NewStatefulSetReconciler(
	client *client.Client,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleGroupInfo reconciler.RoleGroupInfo,
	ports []corev1.ContainerPort,
	image *util.Image,
	stopped bool,
	replicas *int32,
	options builder.WorkloadOptions,
) (*reconciler.StatefulSet, error) {

	krb5SecretClass, tlsSecretClass := "", ""

	if clusterConfig.Authentication != nil {
		krb5SecretClass = clusterConfig.Authentication.KerberosSecretClass
		tlsSecretClass = clusterConfig.Authentication.TlsSecretClass
	}

	stsBuilder := NewStatefulSetBuilder(
		client,
		roleGroupInfo.GetFullName(),
		clusterConfig,
		replicas,
		ports,
		image,
		krb5SecretClass,
		tlsSecretClass,
		options,
	)

	return reconciler.NewStatefulSet(
		client,
		roleGroupInfo.GetFullName(),
		stsBuilder,
		stopped,
	), nil
}
