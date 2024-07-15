package master

import (
	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/common"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

func NewStatefulSetReconciler(
	client *client.Client,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleGroupInfo reconciler.RoleGroupInfo,
	ports []corev1.ContainerPort,
	image *util.Image,
	spec *hbasev1alph1.MasterRoleGroupSpec,
) reconciler.ResourceReconciler[builder.StatefulSetBuilder] {
	stsBuilder := common.NewStatefulSetBuilder(
		client,
		roleGroupInfo.GetFullName(),
		clusterConfig,
		builder.RoleGroupInfo{
			ClusterName:   roleGroupInfo.GetClusterName(),
			RoleName:      roleGroupInfo.GetRoleName(),
			RoleGroupName: roleGroupInfo.GetGroupName(),
		},
		spec.Replicas,
		ports,
		image,
	)

	return reconciler.NewStatefulSet(
		client,
		roleGroupInfo.GetFullName(),
		stsBuilder,
	)
}
