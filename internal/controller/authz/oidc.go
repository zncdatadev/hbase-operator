package authz

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	authv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/authentication/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

var (
	OidcContainerPort = corev1.ContainerPort{
		Name:          "oidc",
		ContainerPort: 4180,
		Protocol:      corev1.ProtocolTCP,
	}
)

type Oidc struct {
	Image        *util.Image
	ClusterUid   string
	UpstreamPort int32
	Oidc         *hbasev1alph1.OidcSpec
	OidcProvider *authv1alpha1.OIDCProvider
}

func NewOidc(
	clusterUid string,
	image *util.Image,
	upstreamPort int32,
	oidc *hbasev1alph1.OidcSpec,
	OidcProvider *authv1alpha1.OIDCProvider,
) *Oidc {
	return &Oidc{
		Image:        image,
		ClusterUid:   clusterUid,
		UpstreamPort: upstreamPort,
		Oidc:         oidc,
		OidcProvider: OidcProvider,
	}
}

func (o *Oidc) GetContainer() *corev1.Container {
	container := builder.NewContainer(
		"oidc",
		o.Image,
	)
	container.AddEnvVars(o.getEnvVars()).
		SetCommand(o.getCommands()).
		AddPorts([]corev1.ContainerPort{OidcContainerPort})
	return container.Build()
}

func (o *Oidc) getCommands() []string {
	return []string{
		"sh",
		"-c",
		"/kubedoop/oauth2-proxy/oauth2-proxy --upstream=${UPSTREAM}",
	}
}

func (o *Oidc) getEnvVars() []corev1.EnvVar {

	scopes := []string{"openid", "email", "profile"}

	if o.Oidc.ExtraScopes != nil {
		scopes = append(scopes, o.Oidc.ExtraScopes...)
	}

	issuer := url.URL{
		Scheme: "http",
		Host:   o.OidcProvider.Hostname,
		Path:   o.OidcProvider.RootPath,
	}

	if o.OidcProvider.Port != 0 && o.OidcProvider.Port != 80 {
		issuer.Host += ":" + strconv.Itoa(o.OidcProvider.Port)
	}

	providerHint := o.OidcProvider.ProviderHint
	if providerHint == "keycloak" {
		providerHint = "keycloak-oidc"
	}

	clientCredentialsSecretName := o.Oidc.ClientCredentialsSecret

	hash := sha256.Sum256([]byte(o.ClusterUid))
	hashStr := hex.EncodeToString(hash[:])
	tokenBytes := []byte(hashStr[:16])

	cookieSecret := base64.StdEncoding.EncodeToString([]byte(base64.StdEncoding.EncodeToString(tokenBytes)))

	return []corev1.EnvVar{
		{
			Name:  "OAUTH2_PROXY_COOKIE_SECRET",
			Value: cookieSecret,
		},
		{
			Name: "OAUTH2_PROXY_CLIENT_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: clientCredentialsSecretName,
					},
					Key: "CLIENT_ID",
				},
			},
		},
		{
			Name: "OAUTH2_PROXY_CLIENT_SECRET",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: clientCredentialsSecretName,
					},
					Key: "CLIENT_SECRET",
				},
			},
		},
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name:  "OAUTH2_PROXY_OIDC_ISSUER_URL",
			Value: issuer.String(),
		},
		{
			Name:  "OAUTH2_PROXY_SCOPE",
			Value: strings.Join(scopes, " "),
		},
		{
			Name:  "OAUTH2_PROXY_PROVIDER",
			Value: providerHint,
		},
		{
			Name:  "UPSTREAM",
			Value: "http://$(POD_IP):" + strconv.Itoa(int(o.UpstreamPort)),
		},
		{
			Name:  "OAUTH2_PROXY_HTTP_ADDRESS",
			Value: "0.0.0.0:" + strconv.Itoa(int(OidcContainerPort.ContainerPort)),
		},
		{
			Name:  "OAUTH2_PROXY_REDIRECT_URL",
			Value: "http://localhost:4180/oauth2/callback",
		},
		{
			Name:  "OAUTH2_PROXY_CODE_CHALLENGE_METHOD",
			Value: "S256",
		},
		{
			Name:  "OAUTH2_PROXY_EMAIL_DOMAINS",
			Value: "*",
		},
	}
}
