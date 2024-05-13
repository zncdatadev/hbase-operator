package reconciler

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AnySpec any

type Reconciler interface {
	GetClient() ResourceClient
	Reconcile() Result
	Ready() Result
}

var _ Reconciler = &BaseReconciler[AnySpec]{}

type BaseReconciler[T AnySpec] struct {
	Client ResourceClient
	Name   string

	Spec T
}

func (b *BaseReconciler[T]) GetClient() ResourceClient {
	return b.Client
}

func (b *BaseReconciler[T]) GetName() string {
	return b.Name
}

func (b *BaseReconciler[T]) GetScheme() *runtime.Scheme {
	return b.Client.Scheme()
}

func (b *BaseReconciler[T]) Ready() Result {
	panic("unimplemented")
}

func (b *BaseReconciler[T]) Reconcile() Result {
	panic("unimplemented")
}

func (b *BaseReconciler[T]) GetSpec() T {
	return b.Spec
}

type ResourceReconciler[T client.Object] interface {
	Reconciler
	Build(ctx context.Context) (T, error)
	GetObjectMeta() metav1.ObjectMeta
}

var _ ResourceReconciler[client.Object] = &BaseResourceReconciler[AnySpec]{}

type BaseResourceReconciler[T AnySpec] struct {
	BaseReconciler[T]
}

func (b *BaseResourceReconciler[T]) GetObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        b.Name,
		Namespace:   b.Client.GetOwnerNamespace(),
		Labels:      b.Client.GetLabels(),
		Annotations: b.Client.GetAnnotations(),
	}
}

func NewBaseResourceReconciler[T AnySpec](
	client ResourceClient,
	name string,
	spec T,
) *BaseResourceReconciler[T] {
	return &BaseResourceReconciler[T]{
		BaseReconciler: BaseReconciler[T]{
			Client: client,
			Name:   name,
			Spec:   spec,
		},
	}
}

func (b *BaseResourceReconciler[T]) Build(ctx context.Context) (client.Object, error) {
	panic("unimplemented")
}
