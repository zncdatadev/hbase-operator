package controller

import (
	"context"
	"fmt"

	authv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/authentication/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/sidecar"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

// OidcCookieSecretName returns the name of the operator-managed Secret carrying the
// oauth2-proxy session cookie secret (key sidecar.OIDCCookieSecretKey). The user-facing
// client-credentials Secret keeps only CLIENT_ID/CLIENT_SECRET, matching the platform docs;
// the cookie secret is generated once by the OidcCookieSecretExtension.
func OidcCookieSecretName(clusterName string) string {
	return clusterName + "-oidc-cookie"
}

// oidcSpecFor returns the cluster's OIDC settings when OIDC authentication is configured
// (an AuthenticationClass reference plus an oidc block), else ("", nil).
func oidcSpecFor(cr *hbasev1alpha1.HbaseCluster) (string, *hbasev1alpha1.OidcSpec) {
	clusterConfig := cr.Spec.ClusterConfigSpec
	if clusterConfig == nil || clusterConfig.Authentication == nil {
		return "", nil
	}
	auth := clusterConfig.Authentication
	if auth.AuthenticationClass == "" || auth.Oidc == nil {
		return "", nil
	}
	return auth.AuthenticationClass, auth.Oidc
}

// registerOidcSidecar wires the framework's oauth2-proxy data-path sidecar in front of the
// role's UI port when the cluster references an AuthenticationClass with an OIDC provider.
// Registration happens on the per-build-context SidecarManager, so nothing leaks across
// role groups or reconciles; the base handler's InjectAll then injects the container.
func registerOidcSidecar(
	ctx context.Context,
	k8sClient client.Client,
	cr *hbasev1alpha1.HbaseCluster,
	buildCtx *reconciler.RoleGroupBuildContext,
	upstreamPort int32,
) error {
	authClassName, oidcSpec := oidcSpecFor(cr)
	if authClassName == "" {
		return nil
	}

	// AuthenticationClass is cluster-scoped.
	authClass := &authv1alpha1.AuthenticationClass{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: authClassName}, authClass); err != nil {
		return fmt.Errorf("failed to get AuthenticationClass %q: %w", authClassName, err)
	}

	if authClass.Spec.AuthenticationProvider == nil || authClass.Spec.AuthenticationProvider.OIDC == nil {
		// The class exists but is not an OIDC class: nothing to inject (mirrors the pre-Gen3
		// behavior, which silently skipped non-OIDC classes).
		return nil
	}

	if upstreamPort == 0 {
		return fmt.Errorf("role %q has no ui-http port to front with oauth2-proxy", buildCtx.RoleName)
	}

	provider := sidecar.NewOAuth2ProxySidecarProvider(
		authClass.Spec.AuthenticationProvider.OIDC,
		oidcSpec.ClientCredentialsSecret,
		upstreamPort,
		// The pre-Gen3 proxy ran with OAUTH2_PROXY_EMAIL_DOMAINS="*": authorization is the
		// IdP realm's concern for these clusters. Kept as an explicit decision.
		sidecar.WithOAuth2ProxyAllowAllEmails(),
		sidecar.WithOAuth2ProxyExtraScopes(oidcSpec.ExtraScopes...),
		// The session cookie secret lives in an operator-managed Secret (created by
		// OidcCookieSecretExtension), not in the user's client-credentials Secret.
		sidecar.WithOAuth2ProxyCookieSecretRef(&corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: OidcCookieSecretName(cr.Name)},
			Key:                  sidecar.OIDCCookieSecretKey,
		}),
	)

	// Keep the framework's limits, but state smaller requests explicitly. Kubernetes otherwise
	// defaults an omitted request to the limit, so the three HBase roles reserve 1800m CPU just
	// for oauth2-proxy and the final role cannot be scheduled on a small single-node cluster.
	buildCtx.SidecarManager.Register(provider, &sidecar.SidecarConfig{
		Enabled: true,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("600m"),
				corev1.ResourceMemory: resource.MustParse("512Mi"),
			},
		},
	})
	return nil
}
