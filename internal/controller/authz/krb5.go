package authz

import (
	"fmt"
	"path"

	"github.com/zncdatadev/operator-go/pkg/constants"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	TlsStorePassword = "changeit"

	TlsStoreDir  = path.Join(constants.KubedoopRoot, "tls")
	TrustoreFile = path.Join(TlsStoreDir, "truststore.p12")
	KeystoreFile = path.Join(TlsStoreDir, "keystore.p12")
	TrustoreType = "pkcs12"
	KeystoreType = "pkcs12"

	ConfigDir = path.Join(constants.KubedoopRoot, "conf")

	AuthenticationType = "kerberos"
	KerberosDir        = path.Join(constants.KubedoopRoot, "kerberos")
	Krb5ConfigFile     = path.Join(KerberosDir, "krb5.conf")
	KetytabFile        = path.Join(KerberosDir, "keytab")
)

type HbaseKerberosConfig struct {
	Namespace   string
	ClusterName string

	KerberosSecretClass string
	TlsSecretClass      string
}

func NewHbaseKerberosConfig(
	namespace string,
	clustername string,
	rolename string,
	rolegroupname string,
	kerberosSecretClass string,
	tlsSecretClass string,
) *HbaseKerberosConfig {
	return &HbaseKerberosConfig{
		Namespace:           namespace,
		ClusterName:         clustername,
		KerberosSecretClass: kerberosSecretClass,
		TlsSecretClass:      tlsSecretClass,
	}
}

func (c *HbaseKerberosConfig) getPrincipal(service string) string {
	host := fmt.Sprintf("%s.%s.svc.cluster.local", c.ClusterName, c.Namespace)
	return fmt.Sprintf("%s/%s@${env.KERBEROS_REALM}", service, host)
}

func (c *HbaseKerberosConfig) GetJVMOPTS() map[string]string {
	return map[string]string{
		"java.security.krb5.conf": Krb5ConfigFile,
	}
}

func (c *HbaseKerberosConfig) GetContainerEnvvars() []corev1.EnvVar {

	return []corev1.EnvVar{
		{
			Name:  "KRB5_CONFIG",
			Value: Krb5ConfigFile,
		},
		{
			Name: "HBASE_OPTS",
			Value: fmt.Sprintf(
				"-Djava.security.krb5.conf=%s",
				Krb5ConfigFile,
			),
		},
	}
}

func (c *HbaseKerberosConfig) GetHbaseSite() map[string]string {
	return map[string]string{
		"hbase.security.authentication": AuthenticationType,
		"hbase.security.authorization":  "true",
		"hbase.rpc.protection":          "privacy",
		"dfs.data.transfer.protection":  "privacy",
		"hbase.rpc.engine":              "org.apache.hadoop.hbase.ipc.SecureRpcEngine",

		"hbase.master.kerberos.principal":       c.getPrincipal("hbase"),
		"hbase.regionserver.kerberos.principal": c.getPrincipal("hbase"),
		"hbase.rest.kerberos.principal":         c.getPrincipal("HTTP"),

		"hbase.master.keytab.file":       KetytabFile,
		"hbase.regionserver.keytab.file": KetytabFile,
		"hbase.rest.keytab.file":         KetytabFile,

		"hbase.coprocessor.master.classes": "org.apache.hadoop.hbase.security.access.AccessController",
		"hbase.coprocessor.region.classes": "org.apache.hadoop.hbase.security.token.TokenProvider,org.apache.hadoop.hbase.security.access.AccessController",

		"hbase.rest.authentication.type":               AuthenticationType,
		"hbase.rest.authentication.kerberos.principal": c.getPrincipal("HTTP"),
		"hbase.rest.authentication.kerberos.keytab":    KetytabFile,

		"hbase.ssl.enabled": "true",
		"hbase.http.policy": "HTTPS_ONLY",
		// Recommended by the docs https://hbase.apache.org/book.html#hbase.ui.cache
		"hbase.http.filter.no-store.enable": "true",

		"hbase.rest.ssl.enabled":           "true",
		"hbase.rest.ssl.keystore.store":    path.Join(TlsStoreDir, "keystore.p12"),
		"hbase.rest.ssl.keystore.password": TlsStorePassword,
		"hbase.rest.ssl.keystore.type":     "pkcs12",
	}
}

func (c *HbaseKerberosConfig) GetDiscoveryConfig() map[string]string {
	return map[string]string{
		"hbase.security.authentication":         AuthenticationType,
		"hbase.rpc.protection":                  "privacy",
		"hbase.ssl.enabled":                     "true",
		"hbase.maser.kerberos.principal":        c.getPrincipal("hbase"),
		"hbase.regionserver.kerberos.principal": c.getPrincipal("hbase"),
		"hbase.rest.kerberos.principal":         c.getPrincipal("hbase"),
	}
}

func (c *HbaseKerberosConfig) GetSSLServerSttings() map[string]string {
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

func (c *HbaseKerberosConfig) GetVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "kerberos",
			MountPath: KerberosDir,
		},
		{
			Name:      "tls",
			MountPath: TlsStoreDir,
		},
	}
}

func (c *HbaseKerberosConfig) GetVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "kerberos",
			VolumeSource: corev1.VolumeSource{
				Ephemeral: &corev1.EphemeralVolumeSource{
					VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"secrets.zncdata.dev/class":                c.KerberosSecretClass,
								"secrets.zncdata.dev/scope":                fmt.Sprintf("service=%s", c.ClusterName),
								"secrets.zncdata.dev/kerberosServiceNames": "HTTP,hbase",
							},
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							StorageClassName: &[]string{"secrets.zncdata.dev"}[0],
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
			Name: "tls",
			VolumeSource: corev1.VolumeSource{
				Ephemeral: &corev1.EphemeralVolumeSource{
					VolumeClaimTemplate: &corev1.PersistentVolumeClaimTemplate{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"secrets.zncdata.dev/class":             c.TlsSecretClass,
								"secrets.zncdata.dev/scope":             "node,pod",
								"secrets.zncdata.dev/format":            "tls-p12",
								"secrets.zncdata.dev/tlsPKCS12Password": TlsStorePassword,
							},
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							StorageClassName: &[]string{"secrets.zncdata.dev"}[0],
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

func (c *HbaseKerberosConfig) GetContainerCommands() string {
	cmds := `
export KERBEROS_REALM=$(grep -oP 'default_realm = \K.*' ` + Krb5ConfigFile + `)
sed -i -e 's/${env.KERBEROS_REALM}/'"$KERBEROS_REALM/g" ` + path.Join(constants.KubedoopConfigDir, "core-site.xml") + `
sed -i -e 's/${env.KERBEROS_REALM}/'"$KERBEROS_REALM/g"  ` + path.Join(constants.KubedoopConfigDir, "hdfs-site.xml") + `
sed -i -e 's/${env.KERBEROS_REALM}/'"$KERBEROS_REALM/g"  ` + path.Join(constants.KubedoopConfigDir, "hbase-site.xml") + `
`

	return util.IndentTab4Spaces(cmds)
}
