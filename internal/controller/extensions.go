package controller

import (
	"context"

	"github.com/zncdatadev/operator-go/pkg/common"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/sidecar"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

// OidcCookieSecretExtension ensures the oauth2-proxy session cookie Secret exists before the
// role groups are built. The framework's oauth2-proxy provider requires a COOKIE_SECRET (an
// inline, CR-derived value would be forgeable by anyone who can read the CR), but the platform's
// user contract for the client-credentials Secret is CLIENT_ID/CLIENT_SECRET only, so the
// operator owns a separate Secret for it.
//
// The generate-once-never-rewrite semantics belong to reconciler.EnsureGeneratedSecret: a fresh
// cookie secret on every pass would re-sign every session and log every user out, while a MISSING
// key must still be filled, or a Secret that lost one key wedges the cluster forever (the sidecar
// provider's Validate fails on it every reconcile).
type OidcCookieSecretExtension struct {
	common.BaseExtension
	scheme *runtime.Scheme
}

var _ common.ClusterExtension[*hbasev1alpha1.HbaseCluster] = &OidcCookieSecretExtension{}

func NewOidcCookieSecretExtension(scheme *runtime.Scheme) *OidcCookieSecretExtension {
	return &OidcCookieSecretExtension{
		BaseExtension: common.NewBaseExtension("oidc-cookie-secret"),
		scheme:        scheme,
	}
}

// PreReconcile ensures the cookie Secret when OIDC is configured.
func (e *OidcCookieSecretExtension) PreReconcile(ctx context.Context, c client.Client, cr *hbasev1alpha1.HbaseCluster) error {
	if authClass, _ := oidcSpecFor(cr); authClass == "" {
		return nil
	}

	_, err := reconciler.EnsureGeneratedSecret(
		ctx, c, e.scheme, cr, OidcCookieSecretName(cr.Name),
		map[string]func() (string, error){
			sidecar.OIDCCookieSecretKey: sidecar.GenerateCookieSecret,
		},
		reconciler.WithGeneratedSecretProductName(hbasev1alpha1.DefaultProductName),
	)
	return err
}

func (e *OidcCookieSecretExtension) PostReconcile(_ context.Context, _ client.Client, _ *hbasev1alpha1.HbaseCluster) error {
	return nil
}

func (e *OidcCookieSecretExtension) OnReconcileError(_ context.Context, _ client.Client, _ *hbasev1alpha1.HbaseCluster, _ error) error {
	return nil
}
