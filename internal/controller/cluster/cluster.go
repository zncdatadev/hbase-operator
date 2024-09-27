package cluster

import (
	"context"

	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/master"
	"github.com/zncdatadev/hbase-operator/internal/controller/regionserver"
	"github.com/zncdatadev/hbase-operator/internal/controller/restserver"
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
	image := &util.Image{
		Repo:            hbasev1alpha1.DefaultRepository,
		ProductName:     hbasev1alpha1.DefaultProductName,
		KubedoopVersion: hbasev1alpha1.DefaultKubedoopVersion,
		ProductVersion:  hbasev1alpha1.DefaultProductVersion,
		PullPolicy:      corev1.PullIfNotPresent,
	}
	if r.Spec.Image != nil {
		image.Custom = r.Spec.Image.Custom
		image.Repo = r.Spec.Image.Repository
		image.ProductVersion = r.Spec.Image.ProductVersion
		image.KubedoopVersion = r.Spec.Image.KubedoopVersion
		image.PullPolicy = r.Spec.Image.PullPolicy
		image.PullSecretName = r.Spec.Image.PullSecretName
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
