package reconciler

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ Reconciler = &BaseReconciler[AnySpec]{}
var _ WorkloadReconciler[*appv1.StatefulSet] = &StatefulSetReconciler[AnySpec]{}

type StatefulSetReconciler[T AnySpec] struct {
	BaseReconciler[T]
}

func (s *StatefulSetReconciler[T]) GetSpec() T {
	return s.Spec
}

func (s *StatefulSetReconciler[T]) Build(ctx context.Context) (*appv1.StatefulSet, error) {
	panic("unimplemented")
}

func (s *StatefulSetReconciler[T]) CommandOverwide(obj *appv1.StatefulSet) {
	panic("unimplemented")
}

func (s *StatefulSetReconciler[T]) EnvOverwide(obj *appv1.StatefulSet) {
	panic("unimplemented")
}

// Ready implements reconciler.ResourceReconciler.
func (s *StatefulSetReconciler[T]) Ready() (bool, error) {
	panic("unimplemented")
}

// Reconcile implements reconciler.ResourceReconciler.
func (s *StatefulSetReconciler[T]) Reconcile() (ReconcileResult, error) {
	panic("unimplemented")
}

func (s *StatefulSetReconciler[T]) AddFinalizer(obj *appv1.StatefulSet) {
	panic("unimplemented")
}

func NewStatefulSetReconciler[O runtime.Object, T AnySpec](
	client client.Client,
	schema *runtime.Scheme,
	name string,
	labels map[string]string,
	annotations map[string]string,
	ownerReference O,
	spec T,

) *StatefulSetReconciler[T] {
	return &StatefulSetReconciler[T]{
		BaseReconciler[T]{
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
