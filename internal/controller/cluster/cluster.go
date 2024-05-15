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

func (r *Reconciler) RegisterResources(_ context.Context) error {
	roleInfo := reconciler.RoleInfo{}
	masterReconciler := master.NewReconciler(
		r.GetClient(),
		r.ClusterConfig,
		roleInfo,
		r.Spec.MasterSpec,
	)
	r.AddResource(masterReconciler)

	regionServerReconciler := regionserver.NewReconciler(
		r.GetClient(),
		r.ClusterConfig,
		roleInfo,
		r.Spec.RegionServerSpec,
	)
	r.AddResource(regionServerReconciler)

	restServerReconciler := restserver.NewReconciler(
		r.GetClient(),
		r.ClusterConfig,
		roleInfo,
		r.Spec.RestServerSpec,
	)
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
