package reconciler

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RoleReconciler interface {
	Reconciler

	GetResources() []Reconciler
	AddResource(resource Reconciler)
	RegisterResources(ctx context.Context) error
}

var _ RoleReconciler = &BaseRoleReconciler[AnySpec]{}

type BaseRoleReconciler[T AnySpec] struct {
	BaseReconciler[T]
	resources []Reconciler
}

func (b *BaseRoleReconciler[T]) GetResources() []Reconciler {
	return b.resources
}

func (b *BaseRoleReconciler[T]) AddResource(resource Reconciler) {
	b.resources = append(b.resources, resource)
}

func (b *BaseRoleReconciler[T]) RegisterResources(ctx context.Context) error {
	panic("implement me")
}

func (b *BaseRoleReconciler[T]) Ready() (bool, error) {
	for _, resource := range b.resources {
		ready, err := resource.Ready()
		if err != nil {
			return false, err
		}
		if !ready {
			return false, nil
		}
	}
	return true, nil
}

func (b *BaseRoleReconciler[T]) Reconcile() (ReconcileResult, error) {
	for _, resource := range b.resources {
		result, err := resource.Reconcile()
		if err != nil {
			return result, nil
		}
	}
	return ReconcileResult{}, nil
}

func NewBaseRoleReconciler[T AnySpec](
	client client.Client,
	schema *runtime.Scheme,
	ownerReference runtime.Object,
	name string,
	labels map[string]string,
	annotations map[string]string,
	spec T,
) *BaseRoleReconciler[T] {
	return &BaseRoleReconciler[T]{
		BaseReconciler: BaseReconciler[T]{
			Client:         client,
			Schema:         schema,
			OwnerReference: ownerReference,
			Name:           name,
			Labels:         labels,
			Annotations:    annotations,

			Spec: spec,
		},
	}
}
