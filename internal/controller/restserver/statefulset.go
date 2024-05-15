package restserver

import (
	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/controller/common"
	"github.com/zncdata-labs/hbase-operator/pkg/client"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
)

func NewStatefulSetReconciler(
	client client.ResourceClient,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleGroupInfo reconciler.RoleGroupInfo,
	ports []corev1.ContainerPort,
	spec *hbasev1alph1.RestServerRoleGroupSpec,
) *common.StatefulSetReconciler[*hbasev1alph1.RestServerRoleGroupSpec] {

	stsBuilder := common.NewStatefulSetBuilder(
		client,
		roleGroupInfo.GetFullName(),
		clusterConfig,
		ports,
		roleGroupInfo.Image,
		spec.EnvOverrides,
		spec.CommandOverrides,
	)

	return common.NewStatefulSetReconciler(
		client,
		roleGroupInfo,
		clusterConfig,
		ports,
		spec,
		stsBuilder,
	)
}
