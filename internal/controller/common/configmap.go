package common

import (
	"context"

	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	"github.com/zncdata-labs/hbase-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	HbaseKey = "hbase-site.xml"
	LogKey   = "log4j.properties"
)

var _ reconciler.ResourceReconciler[*corev1.ConfigMap] = &ConfigMapReconciler{}

type ConfigMapReconciler struct {
	reconciler.BaseResourceReconciler[any]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec

	data map[string]string
}

func (r *ConfigMapReconciler) AddData(key, value string) {
	r.data[key] = value
}

func (r *ConfigMapReconciler) GetZKZnodeConfig(ctx context.Context) (*util.ZnodeConfiguration, error) {
	obj := &corev1.ConfigMap{}
	name := r.ClusterConfig.ZookeeperConfigMap
	namespace := r.Client.GetOwnerNamespace()
	if err := r.Client.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, obj); err != nil {
		return nil, err
	}

	return &util.ZnodeConfiguration{ConfigMap: obj}, nil
}

func (r *ConfigMapReconciler) AddLogConfig() error {
	r.AddData(LogKey, "")
	return nil
}

func (r *ConfigMapReconciler) AddHbaseConfig(znode *util.ZnodeConfiguration) error {

	configuration := util.XMLConfiguration{}

	quorum, err := znode.GetQuorum()
	if err != nil {
		return err
	}
	configuration.AddPropertyWithKV("hbase.zookeeper.quorum", quorum, "")

	chroot, err := znode.GetChroot()
	if err != nil {
		return err
	}
	configuration.AddPropertyWithKV("zookeeper.znode.parent", chroot+"/hbase", "")

	clientPort, err := znode.GetClientPort()
	if err != nil {
		return err
	}
	configuration.AddPropertyWithKV("hbase.zookeeper.property.clientPort", clientPort, "")

	configuration.AddPropertyWithKV("hbase.cluster.distributed", "true", "")
	configuration.AddPropertyWithKV("hbase.rootdir", "/hbase", "")
	configuration.AddPropertyWithKV("hbase.unsafe.regionserver.hostname.disable.master.reversedns", "true", "")

	data, err := configuration.Marshal()
	if err != nil {
		return err
	}
	r.AddData(HbaseKey, data)

	return nil
}

func (r *ConfigMapReconciler) OverrideHbaseConfig(_ *util.XMLConfiguration) error {
	panic("unimplemented")
}

func (r *ConfigMapReconciler) Build(ctx context.Context) (*corev1.ConfigMap, error) {

	znodeConfig, err := r.GetZKZnodeConfig(ctx)
	if err != nil {
		return nil, err
	}

	if err := r.AddHbaseConfig(znodeConfig); err != nil {
		return nil, err
	}

	if err := r.AddLogConfig(); err != nil {
		return nil, err
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: r.GetObjectMeta(),
		Data:       r.data,
	}
	return configMap, nil
}

func NewConfigMapReconciler(
	client reconciler.ResourceClient,
	roleGroupName string,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	spec any,
) *ConfigMapReconciler {
	return &ConfigMapReconciler{
		BaseResourceReconciler: *reconciler.NewBaseResourceReconciler[any](
			client,
			roleGroupName,
			spec,
		),
		ClusterConfig: clusterConfig,
	}
}
