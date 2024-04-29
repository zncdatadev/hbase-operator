package master

import (
	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/common/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ reconciler.Reconciler = &MasterStatefulSetReconciler{}

type MasterStatefulSetReconciler struct {
	reconciler.StatefulSetReconciler[*hbasev1alph1.MasterRoleGroupSpec]
}

func NewMasterStatefulSetReconciler(
	client client.Client,
	schema *runtime.Scheme,

	roleGroupName string,
	labels map[string]string,
	annotations map[string]string,

	ownerReference runtime.Object,

	spec *hbasev1alph1.MasterRoleGroupSpec,

) *MasterStatefulSetReconciler {
	return &MasterStatefulSetReconciler{
		StatefulSetReconciler: *reconciler.NewStatefulSetReconciler(
			client,
			schema,
			roleGroupName,
			labels,
			annotations,
			ownerReference,
			spec,
		),
	}
}
