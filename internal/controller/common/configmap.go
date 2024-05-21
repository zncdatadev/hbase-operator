package common

import (
	"context"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/pkg/builder"
	"github.com/zncdatadev/hbase-operator/pkg/client"
	"github.com/zncdatadev/hbase-operator/pkg/reconciler"
	"github.com/zncdatadev/hbase-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HbaseKey = "hbase-site.xml"
	LogKey   = "log4j.properties"
)

var _ builder.ConfigBuilder = &ConfigMapBuilder{}

type ConfigMapBuilder struct {
	builder.ConfigMapBuilder
	ZnodeConfig *util.ZnodeConfiguration
}

func NewConfigMapBuilder(
	client client.ResourceClient,
	name string,
	ZnodeConfig *util.ZnodeConfiguration,
) *ConfigMapBuilder {
	return &ConfigMapBuilder{
		ConfigMapBuilder: *builder.NewConfigMapBuilder(
			client,
			name,
		),
		ZnodeConfig: ZnodeConfig,
	}
}

func (b *ConfigMapBuilder) Build(ctx context.Context) (ctrlclient.Object, error) {
	hbaseConfig, err := b.getHbaseConfig(b.ZnodeConfig)
	if err != nil {
		return nil, err
	}
	b.AddData(HbaseKey, hbaseConfig)

	return b.ConfigMapBuilder.Build(ctx)
}

func (r *ConfigMapBuilder) getHbaseConfig(znode *util.ZnodeConfiguration) (string, error) {

	configuration := util.XMLConfiguration{}

	quorum, err := znode.GetQuorum()
	if err != nil {
		return "", err
	}
	configuration.AddPropertyWithKV("hbase.zookeeper.quorum", quorum, "")

	chroot, err := znode.GetChroot()
	if err != nil {
		return "", err
	}
	configuration.AddPropertyWithKV("zookeeper.znode.parent", chroot+"/hbase", "")

	clientPort, err := znode.GetClientPort()
	if err != nil {
		return "", err
	}
	configuration.AddPropertyWithKV("hbase.zookeeper.property.clientPort", clientPort, "")

	configuration.AddPropertyWithKV("hbase.cluster.distributed", "true", "")
	configuration.AddPropertyWithKV("hbase.rootdir", "/hbase", "")
	configuration.AddPropertyWithKV("hbase.unsafe.regionserver.hostname.disable.master.reversedns", "true", "")

	data, err := configuration.Marshal()
	if err != nil {
		return "", err
	}

	return data, nil
}

var _ reconciler.ResourceReconciler[builder.ConfigBuilder] = &ConfigMapReconciler[reconciler.AnySpec]{}

type ConfigMapReconciler[T reconciler.AnySpec] struct {
	reconciler.BaseResourceReconciler[T, builder.ConfigBuilder]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func (r *ConfigMapReconciler[T]) GetZKZnodeConfig(ctx context.Context) (*util.ZnodeConfiguration, error) {
	obj := &corev1.ConfigMap{}
	name := r.ClusterConfig.ZookeeperConfigMap
	namespace := r.Client.GetOwnerNamespace()
	if err := r.Client.Get(ctx, ctrlclient.ObjectKey{Name: name, Namespace: namespace}, obj); err != nil {
		return nil, err
	}

	return &util.ZnodeConfiguration{ConfigMap: obj}, nil
}

func NewConfigMapReconciler[T reconciler.AnySpec](
	client client.ResourceClient,
	roleGroupName string,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	spec T,
) *ConfigMapReconciler[T] {

	cmBuilder := builder.NewConfigMapBuilder(
		client,
		roleGroupName,
	)

	return &ConfigMapReconciler[T]{
		BaseResourceReconciler: *reconciler.NewBaseResourceReconciler[T, builder.ConfigBuilder](
			client,
			roleGroupName,
			spec,
			cmBuilder,
		),
		ClusterConfig: clusterConfig,
	}
}
