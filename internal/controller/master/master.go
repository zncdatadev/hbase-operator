package master

import (
	"context"

	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/common/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("controller-master")
)

var _ reconciler.RoleReconciler = &MasterReconciler{}

type MasterReconciler struct {
	reconciler.BaseRoleReconciler[*hbasev1alph1.MasterSpec]
}

func (m *MasterReconciler) RegisterResources(ctx context.Context) error {
	for name, roleGroup := range m.Spec.RoleGroups {
		statefulSetReconciler := NewMasterStatefulSetReconciler(
			m.GetClient(),
			m.GetSchema(),
			name,
			m.GetLabels(),
			m.GetAnnotations(),
			m.GetOwnerReference(),
			&roleGroup,
		)
		m.AddResource(statefulSetReconciler)

		logger.V(1).Info("Registered Master StatefulSet reconciler", "roleGroup", name)
	}
	return nil
}

func NewMasterReconciler(
	client client.Client,
	schema *runtime.Scheme,
	ownerReference *hbasev1alph1.HbaseCluster,
	spec *hbasev1alph1.MasterSpec,
) *MasterReconciler {
	return &MasterReconciler{
		BaseRoleReconciler: *reconciler.NewBaseRoleReconciler(
			client,
			schema,
			ownerReference,
			"master",
			ownerReference.GetLabels(),
			ownerReference.GetAnnotations(),
			spec,
		),
	}
}
