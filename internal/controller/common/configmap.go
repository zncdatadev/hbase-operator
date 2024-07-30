package common

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	util "github.com/zncdatadev/operator-go/pkg/util"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

const (
	HbaseKey = "hbase-site.xml"
	LogKey   = "log4j.properties"
)

var _ builder.ConfigBuilder = &ConfigMapBuilder{}

type ConfigMapBuilder struct {
	builder.ConfigMapBuilder

	ClusterName   string
	RoleName      string
	RolegroupName string
}

func NewConfigMapBuilder(
	client *client.Client,
	name string,
	options builder.Options,
) *ConfigMapBuilder {
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
	}
}

func (b *ConfigMapBuilder) AddHbaseENVSh() error {
	hbaseENVSh := `
export HBASE_MANAGES_ZK="false"
`
	b.AddData(map[string]string{"hbase-env.sh": util.IndentTab4Spaces(hbaseENVSh)})

	return nil
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
log4j.appender.FILE.File=/stackable/log/hbase/hbase.log4j.xml
log4j.appender.FILE.MaxFileSize=5MB
log4j.appender.FILE.MaxBackupIndex=1
log4j.appender.FILE.layout=org.apache.log4j.xml.XMLLayout

`
	b.AddData(map[string]string{LogKey: util.IndentTab4Spaces(log4j)})
	return nil
}

func (b *ConfigMapBuilder) AddSSLClientXML() error {
	return nil
}

func (b *ConfigMapBuilder) AddServerXML() error {
	return nil
}

func (b *ConfigMapBuilder) AddHbaseSite(znode *ZnodeConfiguration) error {
	hbaseSite := util.NewXMLConfiguration()
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

	hbaseSiteXML, err := hbaseSite.Marshal()
	if err != nil {
		return err
	}

	b.AddData(map[string]string{HbaseKey: hbaseSiteXML})
	return nil
}

var _ reconciler.ResourceReconciler[*ConfigMapBuilder] = &ConfigMapReconciler[reconciler.AnySpec]{}

type ConfigMapReconciler[T reconciler.AnySpec] struct {
	reconciler.ResourceReconciler[*ConfigMapBuilder]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func (r *ConfigMapReconciler[T]) GetZKZnodeConfig(ctx context.Context) (*ZnodeConfiguration, error) {
	name := r.ClusterConfig.ZookeeperConfigMapName
	namespace := r.GetClient().GetOwnerNamespace()
	obj := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := r.GetClient().Get(ctx, obj); err != nil {
		return nil, err
	}

	return &ZnodeConfiguration{ConfigMap: obj}, nil
}

func (r *ConfigMapReconciler[T]) Reconcile(ctx context.Context) *reconciler.Result {
	znode, err := r.GetZKZnodeConfig(ctx)
	if err != nil {
		return reconciler.NewResult(true, 0, err)
	}
	builder := r.GetBuilder()

	if err := builder.AddHbaseSite(znode); err != nil {
		return reconciler.NewResult(true, 0, err)
	}

	if err := builder.AddHbaseENVSh(); err != nil {
		return reconciler.NewResult(true, 0, err)
	}

	if err := builder.AddLog4jProperties(); err != nil {
		return reconciler.NewResult(true, 0, err)
	}

	if err := builder.AddSSLClientXML(); err != nil {
		return reconciler.NewResult(true, 0, err)
	}

	if err := builder.AddServerXML(); err != nil {
		return reconciler.NewResult(true, 0, err)
	}

	resource, err := builder.Build(ctx)

	if err != nil {
		return reconciler.NewResult(true, 0, err)
	}

	return r.ResourceReconcile(ctx, resource)
}

func NewConfigMapReconciler[T reconciler.AnySpec](
	client *client.Client,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	options reconciler.RoleGroupInfo,
	spec T,
) *ConfigMapReconciler[T] {

	cmBuilder := NewConfigMapBuilder(
		client,
		options.GetFullName(),
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
