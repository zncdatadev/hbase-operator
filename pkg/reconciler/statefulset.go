package reconciler

import (
	"context"
	"errors"

	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ ResourceReconciler[*appv1.StatefulSet] = &StatefulSetReconciler[AnySpec]{}

type StatefulSetReconciler[T AnySpec] struct {
	BaseResourceReconciler[T]
}

func (s *StatefulSetReconciler[T]) GetSpec() T {
	return s.Spec
}

func (s *StatefulSetReconciler[T]) Build(ctx context.Context) (*appv1.StatefulSet, error) {
	obj, err := s.Builder.Build(ctx)
	if err != nil {
		return nil, err
	}

	statefulSet, ok := obj.(*appv1.StatefulSet)
	if !ok {
		return nil, errors.New("invalid type")
	}

	return statefulSet, nil
}

func (s *StatefulSetReconciler[T]) Ready() Result {
	panic("unimplemented")
}

func (s *StatefulSetReconciler[T]) Reconcile() Result {
	panic("unimplemented")
}

func (s *StatefulSetReconciler[T]) AddFinalizer(obj *appv1.StatefulSet) {
	panic("unimplemented")
}

func NewStatefulSetReconciler[T AnySpec](
	client client.Client,
	schema *runtime.Scheme,
	name string,
	ownerReference client.Object,
	builder builder.Builder,
	spec T,
) *StatefulSetReconciler[T] {
	return &StatefulSetReconciler[T]{
		BaseResourceReconciler: BaseResourceReconciler[T]{
			BaseReconciler: BaseReconciler[T]{
				Client:         client,
				Schema:         schema,
				OwnerReference: ownerReference,
				Name:           name,
				Spec:           spec,
			},
			Builder: builder,
		},
	}
}
