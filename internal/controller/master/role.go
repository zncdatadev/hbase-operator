package master

import (
	"context"

	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/controller/common"
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/image"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("controller").WithName("master")
)

var _ reconciler.RoleReconciler = &Reconciler{}

type Reconciler struct {
	reconciler.BaseRoleReconciler[*hbasev1alph1.MasterSpec]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func (r *Reconciler) RegisterResources(ctx context.Context) error {
	for name, roleGroup := range r.Spec.RoleGroups {
		mergedRoleGroup := roleGroup.DeepCopy()
		r.MergeRoleGroup(&mergedRoleGroup)

		if err := r.RegisterResourceWithRoleGroup(ctx, name, mergedRoleGroup); err != nil {
			return err
		}
		logger.V(1).Info("Master role group registered", "roleGroup", name)
	}
	return nil
}

func (r *Reconciler) RegisterResourceWithRoleGroup(_ context.Context, name string, roleGroup *hbasev1alph1.MasterRoleGroupSpec) error {
	mergedRoleGroup := roleGroup.DeepCopy()
	r.MergeRoleGroup(&mergedRoleGroup)

	statefulSetReconciler := NewStatefulSetReconciler(
		r.GetClient(),
		name,
		r.ClusterConfig,
		Ports,
		r.GetImage(),
		roleGroup,
	)
	r.AddResource(statefulSetReconciler)

	serviceReconciler := common.NewServiceReconciler(
		r.GetClient(),
		name,
		Ports,
		roleGroup,
	)
	r.AddResource(serviceReconciler)

	configMapReconciler := common.NewConfigMapReconciler(
		r.GetClient(),
		name,
		r.ClusterConfig,
		roleGroup,
	)

	r.AddResource(configMapReconciler)

	return nil
}

func NewReconciler(
	client reconciler.ResourceClient,
	clusterOperation *apiv1alpha1.ClusterOperationSpec,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	imageSpec *hbasev1alph1.ImageSpec,
	name string,
	spec *hbasev1alph1.MasterSpec,
) *Reconciler {

	i := image.Image{
		Repository: imageSpec.Repository,
		Tag:        imageSpec.Tag,
		PullPolicy: imageSpec.PullPolicy,
	}

	return &Reconciler{
		BaseRoleReconciler: *reconciler.NewBaseRoleReconciler(
			client,
			name,
			clusterOperation,
			i,
			spec,
		),
		ClusterConfig: clusterConfig,
	}
}
