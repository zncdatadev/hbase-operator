package master

import (
	"context"

	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/image"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("controller").WithName("master")
)

var _ reconciler.RoleReconciler = &Reconciler{}

type Reconciler struct {
	reconciler.BaseRoleReconciler[*hbasev1alph1.MasterSpec]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func (m *Reconciler) RegisterResources(ctx context.Context) error {
	for name, roleGroup := range m.Spec.RoleGroups {
		mergedRoleGroup := roleGroup.DeepCopy()
		m.MergeRoleGroup(&mergedRoleGroup)

		if err := m.RegisterResourceWithRoleGroup(ctx, name, mergedRoleGroup); err != nil {
			return err
		}
		logger.V(1).Info("Master role group registered", "roleGroup", name)
	}
	return nil
}

func (m *Reconciler) RegisterResourceWithRoleGroup(_ context.Context, name string, roleGroup *hbasev1alph1.MasterRoleGroupSpec) error {
	mergedRoleGroup := roleGroup.DeepCopy()
	m.MergeRoleGroup(&mergedRoleGroup)

	statefulSetReconciler := NewStatefulSetReconciler(
		m.GetClient(),
		m.GetSchema(),
		name,
		m.GetLabels(),
		m.GetAnnotations(),
		m.GetOwnerReference(),
		m.GetImage(),
		roleGroup,
	)
	m.AddResource(statefulSetReconciler)

	serviceReconciler := NewServiceReconciler(
		m.GetClient(),
		m.GetSchema(),
		name,
		m.GetLabels(),
		m.GetAnnotations(),
		m.GetOwnerReference(),
		roleGroup,
	)
	m.AddResource(serviceReconciler)

	configMapReconciler := NewConfigMapReconciler(
		m.GetClient(),
		m.GetSchema(),
		name,
		m.GetLabels(),
		m.GetAnnotations(),
		m.GetOwnerReference(),
		m.ClusterConfig,
		roleGroup,
	)

	m.AddResource(configMapReconciler)

	return nil
}

func NewReconciler(
	client client.Client,
	schema *runtime.Scheme,
	ownerReference client.Object,
	clusterOperation *apiv1alpha1.ClusterOperationSpec,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	imageSpec *hbasev1alph1.ImageSpec,
	name string,
	spec *hbasev1alph1.MasterSpec,
) *Reconciler {

	image := image.Image{
		Repository: imageSpec.Repository,
		Tag:        imageSpec.Tag,
		PullPolicy: imageSpec.PullPolicy,
	}

	return &Reconciler{
		BaseRoleReconciler: *reconciler.NewBaseRoleReconciler(
			client,
			schema,
			ownerReference,
			name,
			ownerReference.GetLabels(),
			ownerReference.GetAnnotations(),
			clusterOperation,
			image,
			spec,
		),
		ClusterConfig: clusterConfig,
	}
}
