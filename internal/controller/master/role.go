package master

import (
	"context"
	"time"

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
	logger = ctrl.Log.WithName("controller").WithName("master")
)

var _ reconciler.RoleReconciler = &Reconciler{}

type Reconciler struct {
	reconciler.BaseRoleReconciler[*hbasev1alph1.MasterSpec]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
	Image         *util.Image
}

func NewReconciler(
	client *client.Client,
	clusterStopped bool,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleInfo reconciler.RoleInfo,
	image *util.Image,
	spec *hbasev1alph1.MasterSpec,
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
		mergedRoleGroup := r.MergeRoleGroupSpec(&roleGroup)

		info := reconciler.RoleGroupInfo{
			RoleInfo:      r.RoleInfo,
			RoleGroupName: name,
		}

		reconcilers, err := r.RegisterResourceWithRoleGroup(ctx, info, mergedRoleGroup)
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

func (r *Reconciler) RegisterResourceWithRoleGroup(_ context.Context, info reconciler.RoleGroupInfo, roleGroupSpec any) ([]reconciler.Reconciler, error) {
	spec := roleGroupSpec.(*hbasev1alph1.MasterRoleGroupSpec)
	var reconcilers []reconciler.Reconciler
	var loggingConfig *commonsv1alpha1.LoggingConfigSpec

	if spec.Config != nil && spec.Config.Logging != nil {
		loggingConfig = spec.Config.Logging.Containers[info.RoleName]
	}

	options := builder.WorkloadOptions{
		Option: builder.Option{
			ClusterName:   info.GetClusterName(),
			RoleName:      info.GetRoleName(),
			RoleGroupName: info.GetGroupName(),
			Labels:        info.GetLabels(),
			Annotations:   info.GetAnnotations(),
		},
		PodOverrides: spec.PodOverrides,
		CliOverrides: spec.CliOverrides,
		EnvOverrides: spec.EnvOverrides,
	}

	if spec.Config != nil {
		var gracefulShutdownTimeout time.Duration
		var err error

		if spec.Config.GracefulShutdownTimeout != nil {
			gracefulShutdownTimeout, err = time.ParseDuration(*spec.Config.GracefulShutdownTimeout)
			if err != nil {
				return nil, err
			}
		}

		options.TerminationGracePeriod = &gracefulShutdownTimeout
		options.Resource = spec.Config.Resources
		options.Affinity = spec.Config.Affinity
	}

	statefulSetReconciler, err := common.NewStatefulSetReconciler(
		r.Client,
		r.ClusterConfig,
		info,
		Ports,
		r.Image,
		r.ClusterStopped,
		spec.Replicas,
		options,
	)

	if err != nil {
		return nil, err
	}

	reconcilers = append(reconcilers, statefulSetReconciler)

	serviceReconciler := reconciler.NewServiceReconciler(
		r.Client,
		info.GetFullName(),
		Ports,
		func(sbo *builder.ServiceBuilderOption) {
			sbo.Labels = info.GetLabels()
			sbo.Annotations = info.GetAnnotations()
			sbo.ClusterName = r.ClusterInfo.ClusterName
			sbo.RoleName = r.RoleInfo.RoleName
			sbo.RoleGroupName = info.RoleGroupName
		},
	)

	reconcilers = append(reconcilers, serviceReconciler)

	configMapReconciler := common.NewConfigMapReconciler(
		r.GetClient(),
		r.ClusterConfig,
		loggingConfig,
		info,
		roleGroupSpec.(*hbasev1alph1.MasterRoleGroupSpec),
	)
	reconcilers = append(reconcilers, configMapReconciler)

	return reconcilers, nil
}
