package restserver

import (
	"context"

	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"
	ctrl "sigs.k8s.io/controller-runtime"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/common"
)

var (
	logger = ctrl.Log.WithName("controller").WithName("restserver")
)

var _ reconciler.RoleReconciler = &Reconciler{}

type Reconciler struct {
	reconciler.BaseRoleReconciler[*hbasev1alph1.RestServerSpec]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
	Image         *util.Image
}

func NewReconciler(
	client *client.Client,
	clusterStopped bool,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleInfo reconciler.RoleInfo,
	image *util.Image,
	spec *hbasev1alph1.RestServerSpec,
) *Reconciler {

	return &Reconciler{
		BaseRoleReconciler: *reconciler.NewBaseRoleReconciler(
			client,
			clusterStopped,
			roleInfo,
			spec,
		),
		Image:         image,
		ClusterConfig: clusterConfig,
	}
}

func (r *Reconciler) RegisterResources(ctx context.Context) error {
	for name, roleGroup := range r.Spec.RoleGroups {
		mergedConfig, err := util.MergeObject(r.Spec.Config, roleGroup.Config)
		if err != nil {
			return err
		}
		overrides, err := util.MergeObject(r.Spec.OverridesSpec, roleGroup.OverridesSpec)
		if err != nil {
			return err
		}
		info := reconciler.RoleGroupInfo{
			RoleInfo:      r.RoleInfo,
			RoleGroupName: name,
		}

		var roleGroupConfig *commonsv1alpha1.RoleGroupConfigSpec
		if mergedConfig != nil {
			roleGroupConfig = mergedConfig.RoleGroupConfigSpec
		}
		reconcilers, err := r.RegisterResourceWithRoleGroup(ctx, roleGroup.Replicas, info, overrides, roleGroupConfig)
		if err != nil {
			return err
		}

		for _, reconciler := range reconcilers {
			r.AddResource(reconciler)
			logger.Info("registered resource", "role", r.GetName(), "roleGroup", name, "reconciler", reconciler.GetName())
		}

	}
	return nil
}

func (r *Reconciler) RegisterResourceWithRoleGroup(
	_ context.Context,
	replicas *int32,
	info reconciler.RoleGroupInfo,
	overrides *commonsv1alpha1.OverridesSpec,
	roleGroupConfig *commonsv1alpha1.RoleGroupConfigSpec,
) ([]reconciler.Reconciler, error) {
	var reconcilers []reconciler.Reconciler

	// StatefulSet
	statefulSetReconciler, err := common.NewStatefulSetReconciler(
		r.Client,
		r.ClusterConfig,
		info,
		Ports,
		r.Image,
		r.ClusterStopped(),
		replicas,
		overrides,
		roleGroupConfig,
	)
	if err != nil {
		return nil, err
	}
	reconcilers = append(reconcilers, statefulSetReconciler)

	// Service
	serviceReconciler := reconciler.NewServiceReconciler(
		r.Client,
		info.GetFullName(),
		Ports,
		func(sbo *builder.ServiceBuilderOptions) {
			sbo.Labels = info.GetLabels()
			sbo.Annotations = info.GetAnnotations()
			sbo.ClusterName = r.GetClusterName()
			sbo.RoleName = r.RoleInfo.RoleName
			sbo.RoleGroupName = info.RoleGroupName
		},
	)
	reconcilers = append(reconcilers, serviceReconciler)

	// ConfigMap
	roleLoggingConfig := common.GetRoleLoggingConfig(roleGroupConfig, info.RoleName)
	configMapReconciler := common.NewConfigMapReconciler[*hbasev1alph1.RestServerRoleGroupSpec](
		r.GetClient(),
		r.ClusterConfig,
		roleLoggingConfig,
		info,
	)
	reconcilers = append(reconcilers, configMapReconciler)

	metricsServiceReconciler := common.NewRoleGroupMetricsService(
		r.Client,
		&info,
	)
	reconcilers = append(reconcilers, metricsServiceReconciler)

	return reconcilers, nil
}
