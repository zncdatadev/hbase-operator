package cluster

import (
	"context"

	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/controller/master"
	"github.com/zncdata-labs/hbase-operator/internal/controller/regionserver"
	"github.com/zncdata-labs/hbase-operator/internal/controller/restserver"
	"github.com/zncdata-labs/hbase-operator/pkg/client"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	"github.com/zncdata-labs/hbase-operator/pkg/util"
)

var _ reconciler.Reconciler = &Reconciler{}

type Reconciler struct {
	reconciler.BaseClusterReconciler[*hbasev1alpha1.HbaseClusterSpec]
	ClusterConfig *hbasev1alpha1.ClusterConfigSpec
}

func (r *Reconciler) RegisterResources(ctx context.Context) error {
	masterReconciler := master.NewReconciler(
		r.GetClient(),
		r.ClusterConfig,
		reconciler.RoleInfo{
			ClusterInfo: r.ClusterInfo,
			Name:        "master",
		},
		r.Spec.MasterSpec,
	)

	if err := masterReconciler.RegisterResources(ctx); err != nil {
		return err
	}

	r.AddResource(masterReconciler)

	regionServerReconciler := regionserver.NewReconciler(
		r.GetClient(),
		r.ClusterConfig,
		reconciler.RoleInfo{
			ClusterInfo: r.ClusterInfo,
			Name:        "regionserver",
		},
		r.Spec.RegionServerSpec,
	)
	if err := regionServerReconciler.RegisterResources(ctx); err != nil {
		return err
	}

	r.AddResource(regionServerReconciler)

	restServerReconciler := restserver.NewReconciler(
		r.GetClient(),
		r.ClusterConfig,
		reconciler.RoleInfo{
			ClusterInfo: r.ClusterInfo,
			Name:        "restserver",
		},
		r.Spec.RestServerSpec,
	)
	if err := restServerReconciler.RegisterResources(ctx); err != nil {
		return err
	}
	r.AddResource(restServerReconciler)

	return nil
}

func (r *Reconciler) Reconcile(ctx context.Context) reconciler.Result {
	return r.BaseClusterReconciler.Reconcile(ctx)
}

func NewClusterReconciler(
	client client.ResourceClient,
	instance *hbasev1alpha1.HbaseCluster,
) *Reconciler {
	image := util.Image{
		Custom:         instance.Spec.Image.Custom,
		Repo:           instance.Spec.Image.Repo,
		KDSVersion:     instance.Spec.Image.KDSVersion,
		ProductVersion: instance.Spec.Image.ProductVersion,
	}

	clusterInfo := reconciler.ClusterInfo{
		Name:             instance.Name,
		Namespace:        instance.Namespace,
		ClusterOperation: instance.Spec.ClusterOperationSpec,
		Image:            image,
	}

	obj := &Reconciler{
		BaseClusterReconciler: *reconciler.NewBaseClusterReconciler[*hbasev1alpha1.HbaseClusterSpec](
			client,
			clusterInfo,
			&instance.Spec,
		),
		ClusterConfig: instance.Spec.ClusterConfigSpec,
	}
	return obj
}
