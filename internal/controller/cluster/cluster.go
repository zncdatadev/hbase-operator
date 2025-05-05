package cluster

import (
	"context"

	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/master"
	"github.com/zncdatadev/hbase-operator/internal/controller/regionserver"
	"github.com/zncdatadev/hbase-operator/internal/controller/restserver"
	"github.com/zncdatadev/hbase-operator/internal/util/version"
)

var _ reconciler.Reconciler = &Reconciler{}

type Reconciler struct {
	reconciler.BaseCluster[*hbasev1alpha1.HbaseClusterSpec]
	ClusterConfig *hbasev1alpha1.ClusterConfigSpec
}

func NewClusterReconciler(
	client *client.Client,
	clusterInfo reconciler.ClusterInfo,
	spec *hbasev1alpha1.HbaseClusterSpec,
) *Reconciler {
	return &Reconciler{
		BaseCluster: *reconciler.NewBaseCluster(
			client,
			clusterInfo,
			spec.ClusterOperationSpec,
			spec,
		),
		ClusterConfig: spec.ClusterConfigSpec,
	}

}

func (r *Reconciler) GetImage() *util.Image {
	image := util.NewImage(
		hbasev1alpha1.DefaultProductName,
		version.BuildVersion,
		hbasev1alpha1.DefaultProductVersion,
		func(options *util.ImageOptions) {
			options.Custom = r.Spec.Image.Custom
			options.Repo = r.Spec.Image.Repo
			options.PullPolicy = r.Spec.Image.PullPolicy
		},
	)

	if r.Spec.Image.KubedoopVersion != "" {
		image.KubedoopVersion = r.Spec.Image.KubedoopVersion
	}

	return image
}

func (r *Reconciler) RegisterResources(ctx context.Context) error {

	master := master.NewReconciler(
		r.GetClient(),
		r.IsStopped(),
		r.ClusterConfig,
		reconciler.RoleInfo{
			ClusterInfo: r.ClusterInfo,
			RoleName:    "master",
		},

		r.GetImage(),
		r.Spec.MasterSpec,
	)
	if err := master.RegisterResources(ctx); err != nil {
		return err
	}
	r.AddResource(master)

	region := regionserver.NewReconciler(
		r.GetClient(),
		r.IsStopped(),
		r.ClusterConfig,
		reconciler.RoleInfo{
			ClusterInfo: r.ClusterInfo,
			RoleName:    "regionserver",
		},

		r.GetImage(),
		r.Spec.RegionServerSpec,
	)
	if err := region.RegisterResources(ctx); err != nil {
		return err
	}
	r.AddResource(region)

	rest := restserver.NewReconciler(
		r.GetClient(),
		r.IsStopped(),
		r.ClusterConfig,
		reconciler.RoleInfo{
			ClusterInfo: r.ClusterInfo,
			RoleName:    "restserver",
		},

		r.GetImage(),
		r.Spec.RestServerSpec,
	)
	if err := rest.RegisterResources(ctx); err != nil {
		return err
	}
	r.AddResource(rest)

	return nil

}
