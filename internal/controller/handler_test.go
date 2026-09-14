package controller

import (
	"context"
	"strings"
	"testing"

	authv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/authentication/v1alpha1"
	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/config"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/sidecar"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

const (
	testCluster         = "hbase"
	testNamespace       = "default"
	testGroup           = "default"
	testZnodeCM         = "hbase-znode"
	testHdfsCM          = "hdfs"
	testOIDCAuthClass   = "oidc"
	testOIDCCredentials = "oidc-credentials"
	testMasterResource  = "hbase-master-default"
)

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := hbasev1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := authv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return scheme
}

// znodeConfigMap is a stand-in for the zookeeper-operator's ZookeeperZnode discovery ConfigMap.
func znodeConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: testZnodeCM, Namespace: testNamespace},
		Data: map[string]string{
			ZookeeperHostsKey:      "zk-server-default-0.zk-server-default.default.svc.cluster.local",
			ZookeeperClientPortKey: "2181",
			ZookeeperChrootKey:     "/znode-abc",
		},
	}
}

func testCR(mutators ...func(*hbasev1alpha1.HbaseCluster)) *hbasev1alpha1.HbaseCluster {
	cr := &hbasev1alpha1.HbaseCluster{
		ObjectMeta: metav1.ObjectMeta{Name: testCluster, Namespace: testNamespace},
		Spec: hbasev1alpha1.HbaseClusterSpec{
			Image: &hbasev1alpha1.ImageSpec{ProductVersion: "2.6.1"},
			ClusterConfigSpec: &hbasev1alpha1.ClusterConfigSpec{
				ZookeeperConfigMapName: testZnodeCM,
				HdfsConfigMapName:      testHdfsCM,
			},
			MasterSpec: &hbasev1alpha1.MasterSpec{
				RoleGroups: map[string]hbasev1alpha1.MasterRoleGroupSpec{testGroup: {}},
			},
			RegionServerSpec: &hbasev1alpha1.RegionServerSpec{
				RoleGroups: map[string]hbasev1alpha1.RegionServerRoleGroupSpec{testGroup: {}},
			},
			RestServerSpec: &hbasev1alpha1.RestServerSpec{
				RoleGroups: map[string]hbasev1alpha1.RestServerRoleGroupSpec{testGroup: {}},
			},
		},
	}
	for _, m := range mutators {
		m(cr)
	}
	return cr
}

// buildFor renders one role group exactly as GenericReconciler would: it folds the product
// config under the role/role-group overrides and hands the handler a build context of the same
// shape the framework builds.
func buildFor(t *testing.T, cr *hbasev1alpha1.HbaseCluster, roleName string, objs ...client.Object) *reconciler.RoleGroupResources {
	t.Helper()

	scheme := testScheme(t)
	seed := append([]client.Object{znodeConfigMap(), cr}, objs...)
	k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(seed...).Build()

	genericSpec := cr.GetSpec()
	roleSpec := genericSpec.Roles[roleName]
	groupSpec := roleSpec.RoleGroups[testGroup]
	ctx := context.Background()

	handler := NewHbaseRoleGroupHandler(scheme)

	// Mirror the reconciler's own staging: DECLARE the role, then DERIVE from the effective
	// config, then MERGE with the product's contribution beneath the CR's overrides.
	catalog, err := handler.DeclareRoles(ctx, k8sClient, cr)
	if err != nil {
		t.Fatalf("DeclareRoles failed: %v", err)
	}
	decl, ok := catalog[roleName]
	if !ok {
		t.Fatalf("role %q missing from the catalog", roleName)
	}

	// FOLD the framework-owned half of the config: the product's declared defaults beneath the
	// CR's role and role group levels. This is what carries the default affinity through.
	foldedConfig, _, err := reconciler.FoldCommonConfig(
		decl.ConfigDefaults, roleSpec.GetConfig(), groupSpec.GetConfig())
	if err != nil {
		t.Fatalf("FoldCommonConfig failed: %v", err)
	}
	foldedGroupSpec := *groupSpec.DeepCopy()
	foldedGroupSpec.Config = foldedConfig

	buildCtx := &reconciler.RoleGroupBuildContext{
		ClusterName:      cr.Name,
		ClusterNamespace: cr.Namespace,
		ClusterSpec:      genericSpec,
		RoleName:         roleName,
		RoleSpec:         &roleSpec,
		RoleGroupName:    testGroup,
		RoleGroupSpec:    foldedGroupSpec,
		ResourceName:     cr.Name + "-" + roleName + "-" + testGroup,
		SidecarManager:   sidecar.NewSidecarManager(),
		Declaration:      decl,
		ProductName:      hbasev1alpha1.DefaultProductName,
		ResolvedImage: reconciler.ResolvedImage{
			Reference:      "quay.io/zncdatadev/hbase:2.6.1-kubedoop0.0.0-dev",
			PullPolicy:     corev1.PullIfNotPresent,
			ProductVersion: "2.6.1",
		},
	}

	contribution, err := handler.ResolveRoleGroup(ctx, k8sClient, cr, buildCtx)
	if err != nil {
		t.Fatalf("ResolveRoleGroup failed for role %q: %v", roleName, err)
	}
	buildCtx.MergedConfig = config.NewConfigMerger().Merge(
		contributionOverrides(t, contribution),
		roleSpec.GetOverrides(),
		groupSpec.GetOverrides(),
	)

	resources, err := handler.BuildResources(ctx, k8sClient, cr, buildCtx)
	if err != nil {
		t.Fatalf("BuildResources failed for role %q: %v", roleName, err)
	}
	return resources
}

// contributionOverrides renders a Contribution as the lowest override layer, the way the
// reconciler folds it. The framework's own renderer is unexported, so the two fields this
// operator contributes are mapped here.
func contributionOverrides(t *testing.T, c *reconciler.Contribution) *commonsv1alpha1.OverridesSpec {
	t.Helper()
	if c == nil {
		return nil
	}
	return &commonsv1alpha1.OverridesSpec{
		ConfigOverrides: c.ConfigOverrides,
		EnvOverrides:    c.EnvVars,
	}
}

func mainContainer(t *testing.T, sts *appsv1.StatefulSet, roleName string) corev1.Container {
	t.Helper()
	for _, c := range sts.Spec.Template.Spec.Containers {
		if c.Name == roleName {
			return c
		}
	}
	t.Fatalf("primary container %q not found; containers: %v", roleName, sts.Spec.Template.Spec.Containers)
	return corev1.Container{}
}

func TestConfigMapCarriesTheHbaseFileSet(t *testing.T) {
	cr := testCR()
	resources := buildFor(t, cr, hbasev1alpha1.MasterRole)

	data := resources.ConfigMap.Data
	for _, key := range []string{HbaseSiteFileName, HbaseEnvFileName, "log4j.properties"} {
		if _, ok := data[key]; !ok {
			t.Errorf("ConfigMap missing key %q; got keys %v", key, keysOf(data))
		}
	}

	site := data[HbaseSiteFileName]
	// Zookeeper connection info derived from the znode discovery ConfigMap.
	for _, want := range []string{
		HbaseZookeeperQuorumKey,
		"zk-server-default-0.zk-server-default.default.svc.cluster.local",
		"zookeeper.znode.parent",
		"/znode-abc/hbase",
		"hbase.zookeeper.property.clientPort",
		"2181",
		// Product-intrinsic keys from ComputeProductConfig.
		"hbase.cluster.distributed",
		HbaseRootDirKey,
	} {
		if !strings.Contains(site, want) {
			t.Errorf("hbase-site.xml missing %q; got:\n%s", want, site)
		}
	}

	if env := data[HbaseEnvFileName]; !strings.Contains(env, "export HBASE_MANAGES_ZK=false") {
		t.Errorf("hbase-env.sh missing HBASE_MANAGES_ZK; got:\n%s", env)
	}
}

func TestHbaseEnvUsesTheRoleVariableHBaseReads(t *testing.T) {
	for _, role := range []string{
		hbasev1alpha1.MasterRole,
		hbasev1alpha1.RegionServerRole,
		hbasev1alpha1.RestServerRole,
	} {
		t.Run(role, func(t *testing.T) {
			env := buildFor(t, testCR(), role).ConfigMap.Data[HbaseEnvFileName]
			want := "HBASE_" + strings.ToUpper(role) + "_OPTS"
			if !strings.Contains(env, want) {
				t.Errorf("hbase-env.sh does not set %s; got:\n%s", want, env)
			}
			if wrong := "HBASE_" + role + "_OPTS"; wrong != want && strings.Contains(env, wrong) {
				t.Errorf("hbase-env.sh still contains ignored lowercase variable %s; got:\n%s", wrong, env)
			}
		})
	}
}

// A user's configOverrides must beat both the product config and the zookeeper-derived keys.
func TestUserConfigOverridesWin(t *testing.T) {
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.MasterSpec.OverridesSpec = &commonsv1alpha1.OverridesSpec{
			ConfigOverrides: map[string]map[string]string{
				HbaseSiteFileName: {
					"hbase.rootdir":          "/custom-root",
					"hbase.zookeeper.quorum": "my-own-zk",
				},
			},
		}
	})

	site := buildFor(t, cr, hbasev1alpha1.MasterRole).ConfigMap.Data[HbaseSiteFileName]
	for _, want := range []string{"/custom-root", "my-own-zk"} {
		if !strings.Contains(site, want) {
			t.Errorf("user override %q lost; got:\n%s", want, site)
		}
	}
	if strings.Contains(site, "zk-server-default-0") {
		t.Errorf("zookeeper-derived quorum overwrote the user's override; got:\n%s", site)
	}
}

func TestStatefulSetShapePerRole(t *testing.T) {
	cases := []struct {
		role          string
		commandArg    string
		probePortName string
	}{
		{hbasev1alpha1.MasterRole, "bin/hbase master start", hbasev1alpha1.MasterRole},
		{hbasev1alpha1.RegionServerRole, "bin/hbase regionserver start", hbasev1alpha1.RegionServerRole},
		{hbasev1alpha1.RestServerRole, "bin/hbase rest start", RestHTTPPortName},
	}

	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			cr := testCR()
			resources := buildFor(t, cr, tc.role)
			sts := resources.StatefulSet
			container := mainContainer(t, sts, tc.role)

			// The whole entrypoint lives in Command: RoleDeclaration has no Args, because args
			// are the user's channel (cliOverrides) and a product-declared list would replace
			// what the user wrote.
			if len(container.Args) != 0 {
				t.Errorf("args must stay empty (they are the user's cliOverrides channel); got %v", container.Args)
			}
			if len(container.Command) < 2 {
				t.Fatalf("entrypoint command too short: %v", container.Command)
			}
			script := container.Command[len(container.Command)-1]
			if !strings.Contains(script, tc.commandArg) {
				t.Errorf("entrypoint does not start the role process (%q); command:\n%v", tc.commandArg, container.Command)
			}
			// The entrypoint copies both mounted config dirs into the effective config dir.
			for _, want := range []string{HbaseMountConfigDir, HdfsMountConfigDir, "wait_for_termination"} {
				if !strings.Contains(script, want) {
					t.Errorf("entrypoint missing %q", want)
				}
			}

			for name, probe := range map[string]*corev1.Probe{
				"startup":   container.StartupProbe,
				"liveness":  container.LivenessProbe,
				"readiness": container.ReadinessProbe,
			} {
				if probe == nil || probe.TCPSocket == nil {
					t.Fatalf("%s probe missing or not TCP", name)
				}
				if got := probe.TCPSocket.Port.String(); got != tc.probePortName {
					t.Errorf("%s probe targets %q, want %q", name, got, tc.probePortName)
				}
			}

			assertVolumeMounted(t, sts, container, hdfsConfigVolume, HdfsMountConfigDir)
			assertVolumeMounted(t, sts, container, "config", HbaseMountConfigDir)

			if sts.Spec.Template.Spec.Affinity == nil {
				t.Error("default affinity not applied")
			}
		})
	}
}

func TestMetricsServiceMatchesTheObservabilityContract(t *testing.T) {
	// The observability e2e suite asserts this Service field by field.
	wantPorts := map[string]int32{
		hbasev1alpha1.MasterRole:       16010,
		hbasev1alpha1.RegionServerRole: 16030,
		hbasev1alpha1.RestServerRole:   8085,
	}

	for role, wantPort := range wantPorts {
		t.Run(role, func(t *testing.T) {
			resources := buildFor(t, testCR(), role)
			svc := resources.MetricsService
			if svc == nil {
				t.Fatal("no metrics service built")
			}

			wantName := testCluster + "-" + role + "-" + testGroup + "-metrics"
			if svc.Name != wantName {
				t.Errorf("metrics service name %q, want %q", svc.Name, wantName)
			}
			if svc.Spec.ClusterIP != corev1.ClusterIPNone {
				t.Errorf("metrics service is not headless: clusterIP=%q", svc.Spec.ClusterIP)
			}
			if svc.Labels["prometheus.io/scrape"] != "true" {
				t.Error("metrics service missing the prometheus.io/scrape label")
			}
			for key, want := range map[string]string{
				"prometheus.io/scrape": "true",
				"prometheus.io/path":   "/prometheus",
				"prometheus.io/scheme": "http",
			} {
				if svc.Annotations[key] != want {
					t.Errorf("annotation %s = %q, want %q", key, svc.Annotations[key], want)
				}
			}
			if len(svc.Spec.Ports) != 1 || svc.Spec.Ports[0].Port != wantPort {
				t.Errorf("metrics port = %v, want %d", svc.Spec.Ports, wantPort)
			}
			if got := svc.Spec.Ports[0].TargetPort.String(); got != uiPortName {
				t.Errorf("metrics targetPort = %q, want the HBase UI port %q", got, uiPortName)
			}
			container := mainContainer(t, resources.StatefulSet, role)
			foundTarget := false
			for _, port := range container.Ports {
				if port.Name == uiPortName && port.ContainerPort == wantPort {
					foundTarget = true
				}
			}
			if !foundTarget {
				t.Errorf("container has no %s port at %d; ports: %v", uiPortName, wantPort, container.Ports)
			}
			// The selector must match the pods of exactly this role group.
			for _, key := range []string{"app.kubernetes.io/instance", "app.kubernetes.io/component"} {
				if _, ok := svc.Spec.Selector[key]; !ok {
					t.Errorf("selector missing %q; got %v", key, svc.Spec.Selector)
				}
			}
			if svc.Spec.Selector["app.kubernetes.io/component"] != role {
				t.Errorf("selector component = %q, want %q", svc.Spec.Selector["app.kubernetes.io/component"], role)
			}
		})
	}
}

func TestKerberosWiring(t *testing.T) {
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.ClusterConfigSpec.Authentication = &hbasev1alpha1.AuthenticationSpec{
			KerberosSecretClass: kerberosAuthenticationType,
			TlsSecretClass:      TLSVolumeName,
		}
	})

	resources := buildFor(t, cr, hbasev1alpha1.MasterRole)
	data := resources.ConfigMap.Data

	for _, key := range []string{SSLClientFileName, SSLServerFileName} {
		if _, ok := data[key]; !ok {
			t.Errorf("kerberos enabled but %q missing; keys: %v", key, keysOf(data))
		}
	}
	site := data[HbaseSiteFileName]
	for _, want := range []string{"hbase.security.authentication", "kerberos", "hbase.master.kerberos.principal"} {
		if !strings.Contains(site, want) {
			t.Errorf("hbase-site.xml missing kerberos key %q", want)
		}
	}

	sts := resources.StatefulSet
	container := mainContainer(t, sts, hbasev1alpha1.MasterRole)
	// The entrypoint must resolve the realm before starting HBase.
	if len(container.Command) == 0 || !strings.Contains(container.Command[len(container.Command)-1], "KERBEROS_REALM") {
		t.Error("entrypoint does not resolve the kerberos realm")
	}
	assertVolumeMounted(t, sts, container, KerberosVolumeName, KerberosDir)
	assertVolumeMounted(t, sts, container, TLSVolumeName, TlsStoreDir)

	// The keytab volume must carry the secret-operator annotations verbatim.
	for _, v := range sts.Spec.Template.Spec.Volumes {
		if v.Name != KerberosVolumeName {
			continue
		}
		annotations := v.Ephemeral.VolumeClaimTemplate.Annotations
		if annotations["secrets.kubedoop.dev/class"] != kerberosAuthenticationType {
			t.Errorf("kerberos volume secret class = %q", annotations["secrets.kubedoop.dev/class"])
		}
		if annotations["secrets.kubedoop.dev/scope"] != "service="+testCluster {
			t.Errorf("kerberos volume scope = %q", annotations["secrets.kubedoop.dev/scope"])
		}
		if annotations["secrets.kubedoop.dev/kerberosServiceNames"] != "HTTP,hbase" {
			t.Errorf("kerberos service names = %q", annotations["secrets.kubedoop.dev/kerberosServiceNames"])
		}
	}
}

func TestOidcInjectsTheOauth2ProxySidecar(t *testing.T) {
	cr := testCR(func(cr *hbasev1alpha1.HbaseCluster) {
		cr.Spec.ClusterConfigSpec.Authentication = &hbasev1alpha1.AuthenticationSpec{
			AuthenticationClass: testOIDCAuthClass,
			Oidc: &hbasev1alpha1.OidcSpec{
				ClientCredentialsSecret: testOIDCCredentials,
			},
		}
	})

	authClass := &authv1alpha1.AuthenticationClass{
		ObjectMeta: metav1.ObjectMeta{Name: testOIDCAuthClass},
		Spec: authv1alpha1.AuthenticationClassSpec{
			AuthenticationProvider: &authv1alpha1.AuthenticationProvider{
				OIDC: &authv1alpha1.OIDCProvider{
					Hostname:       "keycloak.default.svc.cluster.local",
					Port:           80,
					RootPath:       "/realms/kubedoop",
					ProviderHint:   "keycloak",
					PrincipalClaim: "preferred_username",
				},
			},
		},
	}

	sts := buildFor(t, cr, hbasev1alpha1.RestServerRole, authClass).StatefulSet

	// The framework injects data-path sidecars as native sidecars (init containers).
	var proxy *corev1.Container
	for i := range sts.Spec.Template.Spec.InitContainers {
		if sts.Spec.Template.Spec.InitContainers[i].Name == sidecar.OAuth2ProxySidecarName {
			proxy = &sts.Spec.Template.Spec.InitContainers[i]
		}
	}
	if proxy == nil {
		t.Fatalf("oauth2-proxy sidecar not injected; init containers: %v", sts.Spec.Template.Spec.InitContainers)
	}

	env := map[string]corev1.EnvVar{}
	for _, e := range proxy.Env {
		env[e.Name] = e
	}
	// The restserver's ui-http port (8085) is what the proxy fronts.
	if got := env["OAUTH2_PROXY_UPSTREAMS"].Value; !strings.Contains(got, "8085") {
		t.Errorf("proxy upstream = %q, want the ui-http port 8085", got)
	}
	// The cookie secret must be referenced, never inlined.
	cookie := env["OAUTH2_PROXY_COOKIE_SECRET"]
	if cookie.Value != "" || cookie.ValueFrom == nil || cookie.ValueFrom.SecretKeyRef == nil {
		t.Fatalf("cookie secret is not a secret reference: %+v", cookie)
	}
	if got := cookie.ValueFrom.SecretKeyRef.Name; got != OidcCookieSecretName(testCluster) {
		t.Errorf("cookie secret ref = %q, want %q", got, OidcCookieSecretName(testCluster))
	}
	// The user-facing credentials Secret keeps its documented two-key contract.
	if got := env["OAUTH2_PROXY_CLIENT_ID"].ValueFrom.SecretKeyRef.Name; got != testOIDCCredentials {
		t.Errorf("client id secret = %q", got)
	}
	if got := proxy.Resources.Requests.Cpu().String(); got != "100m" {
		t.Errorf("proxy CPU request = %q, want 100m", got)
	}
	if got := proxy.Resources.Requests.Memory().String(); got != "128Mi" {
		t.Errorf("proxy memory request = %q, want 128Mi", got)
	}
	if got := proxy.Resources.Limits.Cpu().String(); got != "600m" {
		t.Errorf("proxy CPU limit = %q, want 600m", got)
	}
	if got := proxy.Resources.Limits.Memory().String(); got != "512Mi" {
		t.Errorf("proxy memory limit = %q, want 512Mi", got)
	}
}

// A cluster without OIDC must not gain a proxy.
func TestNoOidcMeansNoSidecar(t *testing.T) {
	sts := buildFor(t, testCR(), hbasev1alpha1.RestServerRole).StatefulSet
	for _, c := range sts.Spec.Template.Spec.InitContainers {
		if c.Name == sidecar.OAuth2ProxySidecarName {
			t.Error("oauth2-proxy injected although OIDC is not configured")
		}
	}
}

func assertVolumeMounted(t *testing.T, sts *appsv1.StatefulSet, container corev1.Container, volumeName, mountPath string) {
	t.Helper()

	found := false
	for _, v := range sts.Spec.Template.Spec.Volumes {
		if v.Name == volumeName {
			found = true
		}
	}
	if !found {
		t.Errorf("pod has no volume %q", volumeName)
	}

	for _, m := range container.VolumeMounts {
		if m.Name == volumeName {
			if m.MountPath != mountPath {
				t.Errorf("volume %q mounted at %q, want %q", volumeName, m.MountPath, mountPath)
			}
			return
		}
	}
	t.Errorf("container does not mount %q", volumeName)
}

func keysOf(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
