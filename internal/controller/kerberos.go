package controller

import (
	"fmt"
	"path"
	"strings"

	"github.com/zncdatadev/operator-go/pkg/constant"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

const (
	securityEnabled    = valueTrue
	rpcProtectionLevel = "privacy"

	kerberosAuthenticationType = "kerberos"
)

var (
	TlsStorePassword = "changeit"

	TlsStoreDir  = path.Join(constant.KubedoopRoot, "tls")
	TrustoreFile = path.Join(TlsStoreDir, "truststore.p12")
	KeystoreFile = path.Join(TlsStoreDir, "keystore.p12")
	TrustoreType = "pkcs12"
	KeystoreType = "pkcs12"

	KerberosDir    = path.Join(constant.KubedoopRoot, "kerberos")
	Krb5ConfigFile = path.Join(KerberosDir, "krb5.conf")
	KeytabFile     = path.Join(KerberosDir, "keytab")
)

// HbaseKerberosConfig computes the Kerberos+TLS wiring for a cluster: hbase-site.xml security
// keys, ssl-client/server.xml store settings, the JVM/system properties, the secret-operator
// CSI volumes and the entrypoint preamble that resolves the realm.
type HbaseKerberosConfig struct {
	Namespace   string
	ClusterName string

	KerberosSecretClass string
	TlsSecretClass      string
}

// KerberosConfigFor returns the Kerberos config when the cluster enables Kerberos
// authentication (both secret classes set, mirroring the pre-Gen3 gate), else nil.
func KerberosConfigFor(cr *hbasev1alpha1.HbaseCluster) *HbaseKerberosConfig {
	clusterConfig := cr.Spec.ClusterConfigSpec
	if clusterConfig == nil || clusterConfig.Authentication == nil {
		return nil
	}
	auth := clusterConfig.Authentication
	if auth.KerberosSecretClass == "" || auth.TlsSecretClass == "" {
		return nil
	}
	return &HbaseKerberosConfig{
		Namespace:           cr.Namespace,
		ClusterName:         cr.Name,
		KerberosSecretClass: auth.KerberosSecretClass,
		TlsSecretClass:      auth.TlsSecretClass,
	}
}

func (c *HbaseKerberosConfig) getPrincipal(service string) string {
	host := fmt.Sprintf("%s.%s.svc.cluster.local", c.ClusterName, c.Namespace)
	return fmt.Sprintf("%s/%s@${env.KERBEROS_REALM}", service, host)
}

// GetJVMOPTS returns the system properties exported through HBASE_<role>_OPTS in hbase-env.sh.
func (c *HbaseKerberosConfig) GetJVMOPTS() map[string]string {
	return map[string]string{
		"java.security.krb5.conf": Krb5ConfigFile,
	}
}

// GetEnvOverrides returns the Kerberos container environment, flowing through the merge
// pipeline as env overrides so a user's CRD envOverrides win.
func (c *HbaseKerberosConfig) GetEnvOverrides() map[string]string {
	return map[string]string{
		"KRB5_CONFIG": Krb5ConfigFile,
		"HBASE_OPTS":  fmt.Sprintf("-Djava.security.krb5.conf=%s", Krb5ConfigFile),
	}
}

func (c *HbaseKerberosConfig) GetHbaseSite() map[string]string {
	return map[string]string{
		"hbase.security.authentication": kerberosAuthenticationType,
		"hbase.security.authorization":  securityEnabled,
		"hbase.rpc.protection":          rpcProtectionLevel,
		"dfs.data.transfer.protection":  rpcProtectionLevel,
		"hbase.rpc.engine":              "org.apache.hadoop.hbase.ipc.SecureRpcEngine",

		"hbase.master.kerberos.principal":       c.getPrincipal("hbase"),
		"hbase.regionserver.kerberos.principal": c.getPrincipal("hbase"),
		"hbase.rest.kerberos.principal":         c.getPrincipal("HTTP"),

		"hbase.master.keytab.file":       KeytabFile,
		"hbase.regionserver.keytab.file": KeytabFile,
		"hbase.rest.keytab.file":         KeytabFile,

		"hbase.coprocessor.master.classes": "org.apache.hadoop.hbase.security.access.AccessController",
		"hbase.coprocessor.region.classes": "org.apache.hadoop.hbase.security.token.TokenProvider,org.apache.hadoop.hbase.security.access.AccessController",

		"hbase.rest.authentication.type":               kerberosAuthenticationType,
		"hbase.rest.authentication.kerberos.principal": c.getPrincipal("HTTP"),
		"hbase.rest.authentication.kerberos.keytab":    KeytabFile,

		"hbase.ssl.enabled": securityEnabled,
		"hbase.http.policy": "HTTPS_ONLY",
		// Recommended by the docs https://hbase.apache.org/book.html#hbase.ui.cache
		"hbase.http.filter.no-store.enable": securityEnabled,

		"hbase.rest.ssl.enabled":           securityEnabled,
		"hbase.rest.ssl.keystore.store":    path.Join(TlsStoreDir, "keystore.p12"),
		"hbase.rest.ssl.keystore.password": TlsStorePassword,
		"hbase.rest.ssl.keystore.type":     "pkcs12",
	}
}

func (c *HbaseKerberosConfig) GetSSLServerSettings() map[string]string {
	return map[string]string{
		"ssl.server.truststore.location": TrustoreFile,
		"ssl.server.truststore.type":     TrustoreType,
		"ssl.server.truststore.password": TlsStorePassword,
		"ssl.server.keystore.location":   KeystoreFile,
		"ssl.server.keystore.type":       KeystoreType,
		"ssl.server.keystore.password":   TlsStorePassword,
	}
}

func (c *HbaseKerberosConfig) GetSSLClientSettings() map[string]string {
	return map[string]string{
		"ssl.client.truststore.location": TrustoreFile,
		"ssl.client.truststore.type":     TrustoreType,
		"ssl.client.truststore.password": TlsStorePassword,
	}
}

func (c *HbaseKerberosConfig) VolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      KerberosVolumeName,
			MountPath: KerberosDir,
		},
		{
			Name:      TLSVolumeName,
			MountPath: TlsStoreDir,
		},
	}
}

// GetVolumes returns the secret-operator CSI ephemeral volumes materializing the keytab and the
// TLS stores. The annotations are the secret-operator's public contract and must stay
// byte-identical across the Gen 3 migration.
func (c *HbaseKerberosConfig) Volumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: KerberosVolumeName,
			VolumeSource: corev1.VolumeSource{
				Ephemeral: &corev1.EphemeralVolumeSource{
					VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"secrets.kubedoop.dev/class":                c.KerberosSecretClass,
								"secrets.kubedoop.dev/scope":                fmt.Sprintf("service=%s", c.ClusterName),
								"secrets.kubedoop.dev/kerberosServiceNames": "HTTP,hbase",
							},
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							StorageClassName: &[]string{"secrets.kubedoop.dev"}[0],
							AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("1Mi"),
								},
							},
						},
					},
				},
			},
		},

		{
			Name: TLSVolumeName,
			VolumeSource: corev1.VolumeSource{
				Ephemeral: &corev1.EphemeralVolumeSource{
					VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"secrets.kubedoop.dev/class":             c.TlsSecretClass,
								"secrets.kubedoop.dev/scope":             "node,pod",
								"secrets.kubedoop.dev/format":            "tls-p12",
								"secrets.kubedoop.dev/tlsPKCS12Password": TlsStorePassword,
							},
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							StorageClassName: &[]string{"secrets.kubedoop.dev"}[0],
							AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.VolumeResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceStorage: resource.MustParse("1Mi"),
								},
							},
						},
					},
				},
			},
		},
	}
}

// GetContainerCommands returns the entrypoint preamble that resolves the Kerberos realm from
// krb5.conf and substitutes it into the copied Hadoop/HBase config files.
func (c *HbaseKerberosConfig) GetContainerCommands() string {
	cmds := `
export KERBEROS_REALM=$(grep -oP 'default_realm = \K.*' ` + Krb5ConfigFile + `)
sed -i -e 's/${env.KERBEROS_REALM}/'"$KERBEROS_REALM/g" ` + path.Join(HbaseConfigDir, "core-site.xml") + `
sed -i -e 's/${env.KERBEROS_REALM}/'"$KERBEROS_REALM/g"  ` + path.Join(HbaseConfigDir, "hdfs-site.xml") + `
sed -i -e 's/${env.KERBEROS_REALM}/'"$KERBEROS_REALM/g"  ` + path.Join(HbaseConfigDir, "hbase-site.xml") + `
`

	return indentTab4Spaces(cmds)
}

// indentTab4Spaces converts tab indentation to 4 spaces, matching the pre-Gen3
// util.IndentTab4Spaces helper so the rendered scripts stay byte-identical.
func indentTab4Spaces(s string) string {
	return strings.ReplaceAll(s, "\t", "    ")
}
