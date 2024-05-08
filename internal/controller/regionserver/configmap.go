package regionserver

import (
	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMapReconciler struct {
	reconciler.BaseResourceReconciler[*hbasev1alph1.RegionServerRoleGroupSpec]
}

type ConfigMapBuilder struct {
	builder.BaseConfigBuilder[corev1.ConfigMap]

	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func NewConfigMapReconciler(
	client client.Client,
	schema *runtime.Scheme,
	roleGroupName string,
	labels map[string]string,
	annotations map[string]string,
	ownerReference client.Object,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	spec *hbasev1alph1.RegionServerRoleGroupSpec,

) *ConfigMapReconciler {

	cmBuilder := &ConfigMapBuilder{
		BaseConfigBuilder: *builder.NewConfigBuilder[corev1.ConfigMap](
			client,
			roleGroupName,
			ownerReference.GetNamespace(),
			annotations,
			labels,
		),
		ClusterConfig: clusterConfig,
	}

	return &ConfigMapReconciler{
		BaseResourceReconciler: *reconciler.NewBaseResourceReconciler(
			client,
			schema,
			ownerReference,
			roleGroupName,
			spec,
			cmBuilder,
		),
	}
}
