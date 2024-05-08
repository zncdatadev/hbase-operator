package reconciler

import (
	"context"
	"reflect"

	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/image"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("common").WithName("reconciler")
)

type ClusterReconciler interface {
	Reconciler
	GetResources() []Reconciler
	AddResource(resource Reconciler)
	RegisterResources(ctx context.Context) error
}

type BaseClusterReconciler[T AnySpec] struct {
	BaseReconciler[T]
	resources []Reconciler
}

func NewBaseClusterReconciler[T AnySpec](
	client client.Client,
	schema *runtime.Scheme,
	ownerReference client.Object,
	name string,
	labels map[string]string,
	annotations map[string]string,
	spec T,
) *BaseClusterReconciler[T] {
	return &BaseClusterReconciler[T]{
		BaseReconciler: BaseReconciler[T]{
			Client:         client,
			Schema:         schema,
			OwnerReference: ownerReference,
			Name:           name,
			Labels:         labels,
			Annotations:    annotations,
			Spec:           spec,
		},
	}
}

func (r *BaseClusterReconciler[T]) GetResources() []Reconciler {
	return r.resources
}

func (r *BaseClusterReconciler[T]) AddResource(resource Reconciler) {
	r.resources = append(r.resources, resource)
}

func (r *BaseClusterReconciler[T]) RegisterResources(ctx context.Context) error {
	panic("unimplemented")
}

func (b *BaseClusterReconciler[T]) Reconcile() Result {
	for _, resource := range b.resources {
		result := resource.Reconcile()
		if result.RequeueOrNot() {
			return result
		}
	}
	return NewResult(false, 0, nil)

}

type RoleReconciler interface {
	Reconciler

	GetClusterOperation() *apiv1alpha1.ClusterOperationSpec
	GetImage() image.Image
}

var _ RoleReconciler = &BaseRoleReconciler[AnySpec]{}

type BaseRoleReconciler[T AnySpec] struct {
	BaseClusterReconciler[T]

	ClusterOperation *apiv1alpha1.ClusterOperationSpec
	Image            image.Image
}

// MergeRoleGroup
// merge right to left, if field of right not exist in left, add it to left.
// else skip it.
// merge will modify left, so left must be a pointer.
func (b *BaseRoleReconciler[T]) MergeRoleGroup(roleGroup any) {
	leftValue := reflect.ValueOf(roleGroup)
	rightValue := reflect.ValueOf(b.Spec)

	if leftValue.Kind() == reflect.Ptr {
		leftValue = leftValue.Elem()
	} else {
		panic("roleGroup is not a pointer")
	}

	if rightValue.Kind() == reflect.Ptr {
		rightValue = rightValue.Elem()
	}

	for i := 0; i < rightValue.NumField(); i++ {
		rightField := rightValue.Field(i)
		rightFieldName := rightValue.Type().Field(i).Name

		if rightField.IsZero() {
			continue
		}

		leftField := leftValue.FieldByName(rightFieldName)

		if !leftField.IsValid() {
			leftValue.Field(i).Set(rightField)
			logger.V(5).Info("Merge role group", "field", rightFieldName, "value", rightField)
		}
	}
}

func (b *BaseRoleReconciler[T]) GetClusterOperation() *apiv1alpha1.ClusterOperationSpec {
	return b.ClusterOperation
}

func (b *BaseRoleReconciler[T]) GetImage() image.Image {
	return b.Image
}

func (b *BaseRoleReconciler[T]) Ready() Result {
	for _, resource := range b.resources {
		if result := resource.Ready(); result.RequeueOrNot() {
			return result
		}
	}
	return NewResult(false, 0, nil)
}

func NewBaseRoleReconciler[T AnySpec](
	client client.Client,
	schema *runtime.Scheme,
	ownerReference client.Object,
	name string,
	labels map[string]string,
	annotations map[string]string,
	clusterOperation *apiv1alpha1.ClusterOperationSpec,
	image image.Image,
	spec T,
) *BaseRoleReconciler[T] {
	return &BaseRoleReconciler[T]{
		BaseClusterReconciler: *NewBaseClusterReconciler[T](
			client,
			schema,
			ownerReference,
			name,
			labels,
			annotations,
			spec,
		),
		ClusterOperation: clusterOperation,
		Image:            image,
	}
}
