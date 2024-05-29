package restserver

import (
	"context"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/common"
	"github.com/zncdatadev/hbase-operator/pkg/client"
	"github.com/zncdatadev/hbase-operator/pkg/reconciler"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("controller").WithName("regionServer")
)

type Reconciler struct {
	reconciler.BaseRoleReconciler[*hbasev1alph1.RestServerSpec]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func (r *Reconciler) RegisterResources(ctx context.Context) error {
	for name, roleGroup := range r.Spec.RoleGroups {
		mergedRoleGroup := roleGroup.DeepCopy()
		r.MergeRoleGroupSpec(mergedRoleGroup)

		if err := r.RegisterResourceWithRoleGroup(ctx, name, mergedRoleGroup); err != nil {
			return err
		}
		logger.V(1).Info("Master role group registered", "roleGroup", name)
	}
	return nil
}

func (r *Reconciler) RegisterResourceWithRoleGroup(_ context.Context, name string, roleGroup *hbasev1alph1.RestServerRoleGroupSpec) error {

	roleGroupInfo := reconciler.RoleGroupInfo{
		RoleInfo:            r.RoleInfo,
		Name:                name,
		Replicas:            roleGroup.Replicas,
		PodDisruptionBudget: roleGroup.PodDisruptionBudget,
		CommandOverrides:    roleGroup.CommandOverrides,
		EnvOverrides:        roleGroup.EnvOverrides,
		//PodOverrides:        roleGroup.PodOverrides,	TODO: Uncomment this line
	}
	statefulSetReconciler := NewStatefulSetReconciler(
		r.GetClient(),
		r.ClusterConfig,
		roleGroupInfo,
		Ports,
		roleGroup,
	)
	r.AddResource(statefulSetReconciler)

	serviceReconciler := common.NewServiceReconciler(
		r.GetClient(),
		roleGroupInfo.GetFullName(),
		Ports,
		roleGroup,
	)
	r.AddResource(serviceReconciler)

	configMapReconciler := common.NewConfigMapReconciler[*hbasev1alph1.RestServerRoleGroupSpec](
		r.GetClient(),
		roleGroupInfo.GetFullName(),
		r.ClusterConfig,
		roleGroup,
	)
	r.AddResource(configMapReconciler)

	return nil
}

func NewReconciler(
	client client.ResourceClient,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleInfo reconciler.RoleInfo,
	spec *hbasev1alph1.RestServerSpec,
) *Reconciler {

	return &Reconciler{
		BaseRoleReconciler: *reconciler.NewBaseRoleReconciler(
			client,
			roleInfo,
			spec,
		),
		ClusterConfig: clusterConfig,
	}
}

func roleName() string {
	return "restserver"
}
