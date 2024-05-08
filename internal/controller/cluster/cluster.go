package cluster

import (
	"context"

	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/controller/master"
	"github.com/zncdata-labs/hbase-operator/internal/controller/regionserver"
	"github.com/zncdata-labs/hbase-operator/internal/controller/restserver"
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ reconciler.Reconciler = &Reconciler{}

type Reconciler struct {
	reconciler.BaseClusterReconciler[*hbasev1alpha1.HbaseClusterSpec]
	ClusterOperation *apiv1alpha1.ClusterOperationSpec
	ClusterConfig    *hbasev1alpha1.ClusterConfigSpec
	Image            *hbasev1alpha1.ImageSpec
}

func (r *Reconciler) RegisterResources(_ context.Context) error {

	masterReconciler := master.NewReconciler(
		r.GetClient(),
		r.GetSchema(),
		r.GetOwnerReference(),
		r.ClusterOperation,
		r.ClusterConfig,
		r.Image,
		"master",
		r.Spec.MasterSpec,
	)
	r.AddResource(masterReconciler)

	regionServerReconciler := regionserver.NewReconciler(
		r.GetClient(),
		r.GetSchema(),
		r.GetOwnerReference(),
		r.ClusterOperation,
		r.ClusterConfig,
		r.Image,
		"regionserver",
		r.Spec.RegionServerSpec,
	)
	r.AddResource(regionServerReconciler)

	restServerReconciler := restserver.NewReconciler(
		r.GetClient(),
		r.GetSchema(),
		r.GetOwnerReference(),
		r.ClusterOperation,
		r.ClusterConfig,
		r.Image,
		"restserver",
		r.Spec.RestServerSpec,
	)
	r.AddResource(restServerReconciler)
	return nil
}

func (r *Reconciler) Reconcile() reconciler.Result {
	return r.BaseClusterReconciler.Reconcile()
}

func NewClusterReconciler(
	client client.Client,
	schema *runtime.Scheme,
	owner *hbasev1alpha1.HbaseCluster,
) *Reconciler {
	owner = owner.DeepCopy()
	obj := &Reconciler{
		BaseClusterReconciler: *reconciler.NewBaseClusterReconciler[*hbasev1alpha1.HbaseClusterSpec](
			client,
			schema,
			owner,
			owner.GetName(),
			owner.GetLabels(),
			owner.GetAnnotations(),
			&owner.Spec,
		),
	}
	return obj
}
