package common

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"

	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/config/xml"
	"github.com/zncdatadev/operator-go/pkg/productlogging"
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

	LoggingConfig       *commonsv1alpha1.LoggingConfigSpec
	VectorConfigMapName string

	ClusterName   string
	RoleName      string
	RolegroupName string

	krb5Config *authz.HbaseKerberosConfig
}

func NewConfigMapBuilder(
	client *client.Client,
	name string,
	loggingConfig *commonsv1alpha1.LoggingConfigSpec,
	vectorConfigMapName string,
	krb5SecretClass string,
	tlsSecretClass string,
	options ...builder.Option,
) *ConfigMapBuilder {
	var krb5Config *authz.HbaseKerberosConfig

	buidlerOptions := &builder.Options{}
	for _, o := range options {
		o(buidlerOptions)
	}

	if krb5SecretClass != "" && tlsSecretClass != "" {
		krb5Config = authz.NewHbaseKerberosConfig(
			client.GetOwnerNamespace(),
			buidlerOptions.ClusterName,
			buidlerOptions.RoleName,
			buidlerOptions.RoleGroupName,
			krb5SecretClass,
			tlsSecretClass,
		)
	}
	return &ConfigMapBuilder{
		ConfigMapBuilder: *builder.NewConfigMapBuilder(
			client,
			name,
			func(o *builder.Options) {
				o.Labels = buidlerOptions.Labels
				o.Annotations = buidlerOptions.Annotations
			},
		),
		ClusterName:         buidlerOptions.ClusterName,
		RoleName:            buidlerOptions.RoleName,
		RolegroupName:       buidlerOptions.RoleGroupName,
		krb5Config:          krb5Config,
		LoggingConfig:       loggingConfig,
		VectorConfigMapName: vectorConfigMapName,
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
	logGenerator, err := productlogging.NewConfigGenerator(
		b.LoggingConfig,
		b.RoleName,
		"hbase.log4j.xml",
		productlogging.LogTypeLog4j,
		func(cgo *productlogging.ConfigGeneratorOption) {
			cgo.ConsoleHandlerFormatter = ptr.To("%d{ISO8601} %-5p [%t] %c{2}: %.1000m%n")
		},
	)
	if err != nil {
		return err
	}

	s, err := logGenerator.Content()
	if err != nil {
		return err
	}

	b.AddItem(LogKey, s)
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

func (b *ConfigMapBuilder) AddVectorConfig(ctx context.Context) error {
	if b.VectorConfigMapName == "" {
		return nil
	}
	s, err := productlogging.MakeVectorYaml(
		ctx,
		b.Client.Client,
		b.Client.GetOwnerNamespace(),
		b.ClusterName,
		b.RoleName,
		b.RolegroupName,
		b.VectorConfigMapName,
	)
	if err != nil {
		return err
	}

	b.AddItem(builder.VectorConfigFile, s)
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

	if err := builder.AddVectorConfig(ctx); err != nil {
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
	loggingConfig *commonsv1alpha1.LoggingConfigSpec,
	options reconciler.RoleGroupInfo,
) *ConfigMapReconciler[T] {
	krb5SecretClass, tlsSecretClass := "", ""
	if clusterConfig.Authentication != nil {
		krb5SecretClass = clusterConfig.Authentication.KerberosSecretClass
		tlsSecretClass = clusterConfig.Authentication.TlsSecretClass
		logger.Info("Using authentication configuration", "kerberosSecretClass", krb5SecretClass, "tlsSecretClass",
			tlsSecretClass, "clusterName", options.GetClusterName(), "roleName", options.GetRoleName(), "roleGroupName", options.GetGroupName())
	}

	cmBuilder := NewConfigMapBuilder(
		client,
		options.GetFullName(),
		loggingConfig,
		clusterConfig.VectorAggregatorConfigMapName,
		krb5SecretClass,
		tlsSecretClass,
		func(o *builder.Options) {
			o.ClusterName = options.GetClusterName()
			o.RoleName = options.GetRoleName()
			o.RoleGroupName = options.GetGroupName()
			o.Labels = options.GetLabels()
			o.Annotations = options.GetAnnotations()
		},
	)

	return &ConfigMapReconciler[T]{
		ResourceReconciler: reconciler.NewGenericResourceReconciler[*ConfigMapBuilder](
			client,
			cmBuilder,
		),
		ClusterConfig: clusterConfig,
	}
}
