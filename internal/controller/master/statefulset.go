package master

import (
	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/controller/common"
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/image"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
)

type StatefulSetReconciler struct {
	common.StatefulSetReconciler[*hbasev1alph1.MasterRoleGroupSpec]
}

func (r *StatefulSetReconciler) GetResourceRequirements() *apiv1alpha1.ResourcesSpec {
	return r.Spec.Config.Resources
}

func (r *StatefulSetReconciler) GetRoleCommandName() string {
	return "master"
}

func NewStatefulSetReconciler(
	client reconciler.ResourceClient,
	roleGroupName string,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	ports []corev1.ContainerPort,
	image image.Image,
	spec *hbasev1alph1.MasterRoleGroupSpec,
) *StatefulSetReconciler {
	return &StatefulSetReconciler{
		StatefulSetReconciler: *common.NewStatefulSetReconciler(
			client,
			roleGroupName,
			clusterConfig,
			ports,
			image,
			spec,
		),
	}
}
