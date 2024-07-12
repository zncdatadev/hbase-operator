package common

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

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

	hbaseConfig *util.XMLConfiguration
}

func NewConfigMapBuilder(
	client *client.Client,
	name string,
	options builder.ResourceOptions,
) *ConfigMapBuilder {
	return &ConfigMapBuilder{
		ConfigMapBuilder: *builder.NewConfigMapBuilder(
			client,
			name,
			options.Labels,
			options.Annotations,
		),
	}
}

func (b *ConfigMapBuilder) Build(ctx context.Context) (ctrlclient.Object, error) {
	hbaseConfig, err := b.hbaseConfig.Marshal()
	if err != nil {
		return nil, err
	}
	b.AddData(map[string]string{hbaseConfig: hbaseConfig})

	return b.ConfigMapBuilder.Build(ctx)
}

func (r *ConfigMapBuilder) AddZnode(znode *ZnodeConfiguration) error {

	quorum, err := znode.GetQuorum()
	if err != nil {
		return err
	}
	r.hbaseConfig.AddPropertyWithString("hbase.zookeeper.quorum", quorum, "")

	chroot, err := znode.GetChroot()
	if err != nil {
		return err
	}
	r.hbaseConfig.AddPropertyWithString("zookeeper.znode.parent", chroot+"/hbase", "")

	clientPort, err := znode.GetClientPort()
	if err != nil {
		return err
	}
	r.hbaseConfig.AddPropertyWithString("hbase.zookeeper.property.clientPort", clientPort, "")

	r.hbaseConfig.AddPropertyWithString("hbase.cluster.distributed", "true", "")
	r.hbaseConfig.AddPropertyWithString("hbase.rootdir", "/hbase", "")
	r.hbaseConfig.AddPropertyWithString("hbase.unsafe.regionserver.hostname.disable.master.reversedns", "true", "")

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
	if err := builder.AddZnode(znode); err != nil {
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
		builder.ResourceOptions{
			Labels:      options.GetLabels(),
			Annotations: options.GetAnnotations(),
			RoleGroupInfo: &builder.RoleGroupInfo{
				ClusterName:   options.ClusterName,
				RoleName:      options.RoleName,
				RoleGroupName: options.RoleGroupName,
			},
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
