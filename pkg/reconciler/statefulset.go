package reconciler

import (
	"context"

	"github.com/zncdata-labs/hbase-operator/pkg/image"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ ResourceReconciler[*appv1.StatefulSet] = &StatefulSetReconciler[AnySpec]{}

type StatefulSetReconciler[T AnySpec] struct {
	BaseResourceReconciler[T]
	Ports []corev1.ContainerPort
	Image image.Image
}

func (s *StatefulSetReconciler[T]) GetSpec() T {
	return s.Spec
}

func (s *StatefulSetReconciler[T]) Build(ctx context.Context) (*appv1.StatefulSet, error) {
	panic("unimplemented")
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
	client ResourceClient,
	name string,
	image image.Image,
	ports []corev1.ContainerPort,
	spec T,
) *StatefulSetReconciler[T] {
	return &StatefulSetReconciler[T]{
		BaseResourceReconciler: BaseResourceReconciler[T]{
			BaseReconciler: BaseReconciler[T]{
				Client: client,
				Name:   name,
				Spec:   spec,
			},
		},
		Ports: ports,
		Image: image,
	}
}
