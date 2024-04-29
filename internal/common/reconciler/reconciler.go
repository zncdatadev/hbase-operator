package reconciler

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AnySpec any

type Reconciler interface {
	GetClient() client.Client
	GetSchema() *runtime.Scheme

	GetOwnerReference() runtime.Object

	GetLabels() map[string]string
	GetAnnotations() map[string]string

	Reconcile() (ReconcileResult, error)
	Ready() (bool, error)
}

var _ Reconciler = &BaseReconciler[AnySpec]{}

type BaseReconciler[T AnySpec] struct {
	Client client.Client
	Schema *runtime.Scheme

	OwnerReference runtime.Object

	Name        string
	Labels      map[string]string
	Annotations map[string]string

	Spec T
}

func (b *BaseReconciler[T]) GetAnnotations() map[string]string {
	return b.Annotations

}

// GetClient implements reconciler.ResourceReconciler.
func (b *BaseReconciler[T]) GetClient() client.Client {
	return b.Client
}

func (b *BaseReconciler[T]) GetOwnerReference() runtime.Object {
	return b.OwnerReference
}

func (b *BaseReconciler[T]) GetName() string {
	return b.Name
}

// GetLabels implements reconciler.ResourceReconciler.
func (b *BaseReconciler[T]) GetLabels() map[string]string {
	return b.Labels
}

// GetSchema implements reconciler.ResourceReconciler.
func (b *BaseReconciler[T]) GetSchema() *runtime.Scheme {
	return b.Schema
}

// Ready implements reconciler.ResourceReconciler.
func (b *BaseReconciler[T]) Ready() (bool, error) {
	panic("unimplemented")
}

// Reconcile implements reconciler.ResourceReconciler.
func (b *BaseReconciler[T]) Reconcile() (ReconcileResult, error) {
	panic("unimplemented")
}

func (b *BaseReconciler[T]) GetSpec() T {
	return b.Spec
}
