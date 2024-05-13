/*
Copyright 2024 zncdata-labs.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"github.com/zncdata-labs/hbase-operator/internal/controller/cluster"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
)

var (
	logger = log.Log.WithName("controller")
)

// HbaseClusterReconciler reconciles a HbaseCluster object
type HbaseClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hbase.zncdata.dev,resources=hbaseclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hbase.zncdata.dev,resources=hbaseclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hbase.zncdata.dev,resources=hbaseclusters/finalizers,verbs=update

func (r *HbaseClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger.V(1).Info("Reconciling HbaseCluster")

	// Fetch the HbaseCluster instance
	instance := &hbasev1alpha1.HbaseCluster{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(1).Info("Reconcile finished")
	r.Client.Scheme()
	client := reconciler.ResourceClient{
		Client: r.Client,
	}

	clusterReconciler := cluster.NewClusterReconciler(
		client,
		instance,
	)

	if err := clusterReconciler.RegisterResources(ctx); err != nil {
		return ctrl.Result{}, err
	}

	if result := clusterReconciler.Reconcile(); result.RequeueOrNot() {
		return result.Result()
	}

	if result := clusterReconciler.Ready(); result.RequeueOrNot() {
		return result.Result()
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HbaseClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hbasev1alpha1.HbaseCluster{}).
		Complete(r)
}
