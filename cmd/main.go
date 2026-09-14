/*
Copyright 2024 zncdatadev.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	authv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/authentication/v1alpha1"
	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/common"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller"
	"github.com/zncdatadev/hbase-operator/internal/util/version"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(authv1alpha1.AddToScheme(scheme))
	utilruntime.Must(hbasev1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

// newExtensionRegistry builds the extension registry for HbaseCluster reconciliation. The
// registry is instantiated for the product's own CR type and handed to exactly one reconciler
// (GenericReconcilerConfig.ExtensionRegistry).
func newExtensionRegistry(scheme *runtime.Scheme) *common.ExtensionRegistry[*hbasev1alpha1.HbaseCluster] {
	registry := common.NewExtensionRegistry[*hbasev1alpha1.HbaseCluster]()

	// Ensures the oauth2-proxy session cookie Secret exists before the role groups build the
	// sidecar that references it.
	registry.RegisterClusterExtension(controller.NewOidcCookieSecretExtension(scheme))

	return registry
}

// dependencies declares the external ConfigMaps/Secrets an HbaseCluster references but does not
// create, so a missing one fails the cycle with a Degraded condition instead of crash-looping
// pods on an absent mount.
func dependencies(cr *hbasev1alpha1.HbaseCluster) []reconciler.Dependency {
	clusterConfig := cr.Spec.ClusterConfigSpec
	if clusterConfig == nil {
		return nil
	}

	deps := []reconciler.Dependency{}
	if name := clusterConfig.ZookeeperConfigMapName; name != "" {
		deps = append(deps, reconciler.Dependency{Kind: reconciler.DependencyConfigMap, Name: name})
	}
	if name := clusterConfig.HdfsConfigMapName; name != "" {
		deps = append(deps, reconciler.Dependency{Kind: reconciler.DependencyConfigMap, Name: name})
	}
	if auth := clusterConfig.Authentication; auth != nil && auth.Oidc != nil && auth.Oidc.ClientCredentialsSecret != "" {
		deps = append(deps, reconciler.Dependency{Kind: reconciler.DependencySecret, Name: auth.Oidc.ClientCredentialsSecret})
	}
	return deps
}

func main() {
	var metricsAddr string
	var metricsCertPath, metricsCertName, metricsCertKey string
	var webhookCertPath, webhookCertName, webhookCertKey string
	var enableLeaderElection bool
	var probeAddr string
	var secureMetrics bool
	var enableHTTP2 bool
	var showVersion bool
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.StringVar(&webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	flag.StringVar(&webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	flag.StringVar(&webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	flag.BoolVar(&showVersion, "version", false, "Print version information and exit.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	if showVersion {
		fmt.Println(version.NewAppInfo("hbase-operator").String())
		os.Exit(0)
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Initial webhook TLS options
	webhookServerOptions := webhook.Options{
		TLSOpts: tlsOpts,
	}

	if len(webhookCertPath) > 0 {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", webhookCertPath, "webhook-cert-name", webhookCertName, "webhook-cert-key", webhookCertKey)

		webhookServerOptions.CertDir = webhookCertPath
		webhookServerOptions.CertName = webhookCertName
		webhookServerOptions.KeyName = webhookCertKey
	}

	webhookServer := webhook.NewServer(webhookServerOptions)

	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	if len(metricsCertPath) > 0 {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", metricsCertPath, "metrics-cert-name", metricsCertName, "metrics-cert-key", metricsCertKey)

		metricsServerOptions.CertDir = metricsCertPath
		metricsServerOptions.CertName = metricsCertName
		metricsServerOptions.KeyName = metricsCertKey
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		WebhookServer:          webhookServer,
		LeaderElectionID:       "8a493e83.kubedoop.dev",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	hbaseHandler := controller.NewHbaseRoleGroupHandler(mgr.GetScheme())

	// The Gen 3 wiring: a GenericReconciler drives the HbaseRoleGroupHandler; product config
	// flows through the merge pipeline and extensions own the cluster-scoped side effects.
	reconcilerCfg := &reconciler.GenericReconcilerConfig[*hbasev1alpha1.HbaseCluster]{
		Client: mgr.GetClient(),
		// Uncached: used to refresh the resourceVersion after a conflicting status write, which
		// the informer cache is by definition too stale to serve.
		APIReader: mgr.GetAPIReader(),
		Scheme:    mgr.GetScheme(),
		//nolint:staticcheck // TODO: migrate to GetEventRecorder when the SDK supports the new events API
		Recorder: mgr.GetEventRecorderFor("hbasecluster-controller"),

		// The handler is all three product seams at once: it builds the role group resources,
		// declares what each role is made of (DeclareRoles), and derives HBase's own config
		// from the CR plus the live zookeeper discovery ConfigMap (ResolveRoleGroup).
		RoleGroupHandler:  hbaseHandler,
		RoleProvider:      hbaseHandler,
		RoleGroupResolver: hbaseHandler,

		// spec.image resolves to "{repo}/hbase:{productVersion}-kubedoop{kubedoopVersion}";
		// the CR's own spec.image outranks these defaults.
		ImageResolution: reconciler.ImageResolution{
			ProductName: hbasev1alpha1.DefaultProductName,
			Defaults: commonsv1alpha1.ImageSpec{
				Repo:            hbasev1alpha1.DefaultRepository,
				ProductVersion:  hbasev1alpha1.DefaultProductVersion,
				KubedoopVersion: version.BuildVersion,
			},
		},

		Dependencies: dependencies,

		HealthCheckInterval: 120 * time.Second,
		HealthCheckTimeout:  300 * time.Second,

		ExtensionRegistry: newExtensionRegistry(mgr.GetScheme()),
		Prototype:         &hbasev1alpha1.HbaseCluster{},
	}

	hbaseReconciler, err := reconciler.NewGenericReconciler(reconcilerCfg)
	if err != nil {
		setupLog.Error(err, "unable to create reconciler")
		os.Exit(1)
	}

	if err := controller.SetupReferenceIndexes(context.Background(), mgr.GetFieldIndexer()); err != nil {
		setupLog.Error(err, "unable to index HbaseCluster external references")
		os.Exit(1)
	}

	if err := hbaseReconciler.SetupWithManagerOpts(mgr, reconciler.SetupWithManagerOptions{
		Watches: controller.ReferenceWatches(mgr.GetClient()),
	}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "HbaseCluster")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
