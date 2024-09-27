package common

import (
	"context"
	"fmt"
	"path"
	"strings"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/config/xml"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	authz "github.com/zncdatadev/hbase-operator/internal/controller/authz"
)

const (
	HbaseKey          = "hbase-site.xml"
	LogKey            = "log4j.properties"
	SSLClientFileName = "ssl-client.xml"
	SSLServerFileName = "ssl-server.xml"
)

var (
	logger = ctrl.Log.WithName("common")
)

var _ builder.ConfigBuilder = &ConfigMapBuilder{}

type ConfigMapBuilder struct {
	builder.ConfigMapBuilder

	ClusterName   string
	RoleName      string
	RolegroupName string

	krb5Config *authz.HbaseKerberosConfig
}

func NewConfigMapBuilder(
	client *client.Client,
	name string,
	krb5SecretClass string,
	tlsSecretClass string,
	options builder.Options,
) *ConfigMapBuilder {
	var krb5Config *authz.HbaseKerberosConfig
	if krb5SecretClass != "" && tlsSecretClass != "" {
		krb5Config = authz.NewHbaseKerberosConfig(
			client.GetOwnerNamespace(),
			options.ClusterName,
			options.RoleName,
			options.RoleGroupName,
			krb5SecretClass,
			tlsSecretClass,
		)

	}

	return &ConfigMapBuilder{
		ConfigMapBuilder: *builder.NewConfigMapBuilder(
			client,
			name,
			options.Labels,
			options.Annotations,
		),
		ClusterName:   options.ClusterName,
		RoleName:      options.RoleName,
		RolegroupName: options.RoleGroupName,
		krb5Config:    krb5Config,
	}
}

func (b *ConfigMapBuilder) AddHbaseENVSh() error {

	jvmOpts := b.getJVMOPTS()

	hbaseENVSh := fmt.Sprintf(`
export HBASE_MANAGES_ZK=false
%s

`,
		jvmOpts,
	)
	b.AddData(map[string]string{"hbase-env.sh": util.IndentTab4Spaces(hbaseENVSh)})
	logger.V(15).Info("Hbase ENV", "hbaseENVSh", hbaseENVSh, "clusterName", b.ClusterName, "roleName", b.RoleName, "roleGroupName", b.RolegroupName)

	return nil
}

func (b *ConfigMapBuilder) getJVMOPTS() string {
	var opts []string
	if b.krb5Config != nil {
		for k, v := range b.krb5Config.GetJVMOPTS() {
			opts = append(opts, "-D"+k+"="+v)
		}
	}
	return fmt.Sprintf(`export HBASE_%s_OPTS="$HBASE_OPTS %s"`, b.RoleName, strings.Join(opts, " "))
}

func (b *ConfigMapBuilder) AddLog4jProperties() error {
	log4j := `
log4j.rootLogger=INFO, CONSOLE, FILE

log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender
log4j.appender.CONSOLE.Threshold=INFO
log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout
log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} %-5p [%t] %c{2}: %.1000m%n

log4j.appender.FILE=org.apache.log4j.RollingFileAppender
log4j.appender.FILE.Threshold=INFO
log4j.appender.FILE.File=` + path.Join(HBaseConfigDir, "hbase.log4j.xml") + `
log4j.appender.FILE.MaxFileSize=5MB
log4j.appender.FILE.MaxBackupIndex=1
log4j.appender.FILE.layout=org.apache.log4j.xml.XMLLayout

`
	b.AddData(map[string]string{LogKey: util.IndentTab4Spaces(log4j)})
	logger.V(15).Info("Log4j", "log4j", log4j, "clusterName", b.ClusterName, "roleName", b.RoleName, "roleGroupName", b.RolegroupName)
	return nil
}

func (b *ConfigMapBuilder) AddSSLClientXML() error {
	if b.krb5Config != nil {
		data := b.krb5Config.GetSSLClientSettings()
		sslClientXML := xml.NewXMLConfigurationFromMap(data)
		sslClientXMLXML, err := sslClientXML.Marshal()
		if err != nil {
			return err
		}
		b.AddData(map[string]string{SSLClientFileName: sslClientXMLXML})
		logger.V(15).Info("SSL Client XML", "sslClientXML", sslClientXML, "clusterName", b.ClusterName, "roleName", b.RoleName, "roleGroupName", b.RolegroupName)
	}
	return nil
}

func (b *ConfigMapBuilder) AddSSLServerXML() error {
	if b.krb5Config != nil {
		data := b.krb5Config.GetSSLServerSttings()
		serverXML := xml.NewXMLConfigurationFromMap(data)
		serverXMLXML, err := serverXML.Marshal()
		if err != nil {
			return err
		}
		b.AddData(map[string]string{SSLServerFileName: serverXMLXML})
		logger.V(15).Info("Server XML", "serverXML", serverXML, "clusterName", b.ClusterName, "roleName", b.RoleName, "roleGroupName", b.RolegroupName)
	}
	return nil
}

func (b *ConfigMapBuilder) AddHbaseSite(znode *ZnodeConfiguration) error {
	hbaseSite := xml.NewXMLConfiguration()
	quorum, err := znode.GetQuorum()
	if err != nil {
		return err
	}
	hbaseSite.AddPropertyWithString("hbase.zookeeper.quorum", quorum, "")

	chroot, err := znode.GetChroot()
	if err != nil {
		return err
	}
	hbaseSite.AddPropertyWithString("zookeeper.znode.parent", chroot+"/hbase", "")

	clientPort, err := znode.GetClientPort()
	if err != nil {
		return err
	}
	hbaseSite.AddPropertyWithString("hbase.zookeeper.property.clientPort", clientPort, "")

	hbaseSite.AddPropertyWithString("hbase.cluster.distributed", "true", "")
	hbaseSite.AddPropertyWithString("hbase.rootdir", "/hbase", "")
	hbaseSite.AddPropertyWithString("hbase.unsafe.regionserver.hostname.disable.master.reversedns", "true", "")

	if b.krb5Config != nil {
		hbaseSite.AddPropertiesWithMap(b.krb5Config.GetHbaseSite())
	}

	hbaseSiteXML, err := hbaseSite.Marshal()
	if err != nil {
		return err
	}

	b.AddData(map[string]string{HbaseKey: hbaseSiteXML})
	logger.V(15).Info("Hbase Site", "hbaseSite", hbaseSite, "clusterName", b.ClusterName, "roleName", b.RoleName, "roleGroupName", b.RolegroupName)
	return nil
}

var _ reconciler.ResourceReconciler[*ConfigMapBuilder] = &ConfigMapReconciler[reconciler.AnySpec]{}

type ConfigMapReconciler[T reconciler.AnySpec] struct {
	reconciler.ResourceReconciler[*ConfigMapBuilder]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func (r *ConfigMapReconciler[T]) GetZKZnodeConfig(ctx context.Context) (*ZnodeConfiguration, error) {
	name := r.ClusterConfig.ZookeeperConfigMapName
	obj := &corev1.ConfigMap{}

	if err := r.GetClient().GetWithOwnerNamespace(ctx, name, obj); err != nil {
		return nil, err
	}

	return &ZnodeConfiguration{ConfigMap: obj}, nil
}

func (r *ConfigMapReconciler[T]) Reconcile(ctx context.Context) (ctrl.Result, error) {
	znode, err := r.GetZKZnodeConfig(ctx)
	if err != nil {
		return ctrl.Result{}, err
	}
	builder := r.GetBuilder()

	if err := builder.AddHbaseSite(znode); err != nil {
		return ctrl.Result{}, err
	}

	if err := builder.AddHbaseENVSh(); err != nil {
		return ctrl.Result{}, err
	}

	if err := builder.AddLog4jProperties(); err != nil {
		return ctrl.Result{}, err
	}

	if err := builder.AddSSLClientXML(); err != nil {
		return ctrl.Result{}, err
	}

	if err := builder.AddSSLServerXML(); err != nil {
		return ctrl.Result{}, err
	}

	resource, err := builder.Build(ctx)

	if err != nil {
		return ctrl.Result{}, err
	}

	return r.ResourceReconcile(ctx, resource)
}

func NewConfigMapReconciler[T reconciler.AnySpec](
	client *client.Client,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	options reconciler.RoleGroupInfo,
	spec T,
) *ConfigMapReconciler[T] {
	krb5SecretClass, tlsSecretClass := "", ""
	if clusterConfig.Authentication != nil {
		krb5SecretClass = clusterConfig.Authentication.KerberosSecretClass
		tlsSecretClass = clusterConfig.Authentication.TlsSecretClass
		logger.Info("Using authentication configuration", "kerberosSecretClass", krb5SecretClass, "tlsSecretClass", tlsSecretClass, "clusterName", options.GetClusterName(), "roleName", options.GetRoleName(), "roleGroupName", options.GetGroupName())
	}

	cmBuilder := NewConfigMapBuilder(
		client,
		options.GetFullName(),
		krb5SecretClass,
		tlsSecretClass,
		builder.Options{
			ClusterName:   options.GetClusterName(),
			RoleName:      options.GetRoleName(),
			RoleGroupName: options.GetGroupName(),
			Labels:        options.GetLabels(),
			Annotations:   options.GetAnnotations(),
		},
	)

	return &ConfigMapReconciler[T]{
		ResourceReconciler: reconciler.NewGenericResourceReconciler[*ConfigMapBuilder](
			client,
			options.GetFullName(),
			cmBuilder,
		),
		ClusterConfig: clusterConfig,
	}
}
