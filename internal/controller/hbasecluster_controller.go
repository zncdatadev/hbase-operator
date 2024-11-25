/*
Copyright 2024 zncdatadev.

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

	"github.com/zncdatadev/hbase-operator/internal/controller/cluster"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

var (
	logger = log.Log.WithName("controller")
)

// HbaseClusterReconciler reconciles a HbaseCluster object
type HbaseClusterReconciler struct {
	ctrlclient.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hbase.kubedoop.dev,resources=hbaseclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hbase.kubedoop.dev,resources=hbaseclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hbase.kubedoop.dev,resources=hbaseclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=authentication.kubedoop.dev,resources=authenticationclasses,verbs=get;list;watch
// +kubebuilder:rbac:groups=policy,resources=poddisruptionbudgets,verbs=get;list;watch;create;update;patch;delete

func (r *HbaseClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger.V(1).Info("Reconciling HbaseCluster")

	instance := &hbasev1alpha1.HbaseCluster{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if ctrlclient.IgnoreNotFound(err) == nil {
			logger.V(1).Info("HbaseCluster not found, may have been deleted")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	logger.V(1).Info("HbaseCluster found", "namespace", instance.Namespace, "name", instance.Name)

	// Labels: map[string]string{
	// 	"app.kubernetes.io/name":       "hbase",
	// 	"app.kubernetes.io/managed-by": "hbase.kubedoop.dev",
	// 	"app.kubernetes.io/instance":   instance.Name,
	// },
	resourceClient := &client.Client{
		Client:         r.Client,
		OwnerReference: instance,
	}

	gvk := instance.GetObjectKind().GroupVersionKind()

	clusterReconciler := cluster.NewClusterReconciler(
		resourceClient,
		reconciler.ClusterInfo{
			GVK: &metav1.GroupVersionKind{
				Group:   gvk.Group,
				Version: gvk.Version,
				Kind:    gvk.Kind,
			},
			ClusterName: instance.Name,
		},
		&instance.Spec,
	)

	if err := clusterReconciler.RegisterResources(ctx); err != nil {
		return ctrl.Result{}, err
	}

	if result, err := clusterReconciler.Reconcile(ctx); err != nil {
		return ctrl.Result{}, err
	} else if !result.IsZero() {
		return result, nil
	}

	logger.Info("Cluster resource reconciled, checking if ready.", "cluster", instance.Name, "namespace", instance.Namespace)

	if result, err := clusterReconciler.Ready(ctx); err != nil {
		return ctrl.Result{}, err
	} else if !result.IsZero() {
		return result, nil
	}

	logger.V(1).Info("Reconcile finished.", "cluster", instance.Name, "namespace", instance.Namespace)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HbaseClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hbasev1alpha1.HbaseCluster{}).
		Complete(r)
}
