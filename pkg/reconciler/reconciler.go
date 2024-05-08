package reconciler

import (
	"context"

	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AnySpec any

type Reconciler interface {
	GetClient() client.Client
	GetSchema() *runtime.Scheme
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetOwnerReference() client.Object

	Reconcile() Result
	Ready() Result
}

var _ Reconciler = &BaseReconciler[AnySpec]{}

type BaseReconciler[T AnySpec] struct {
	Client client.Client
	Schema *runtime.Scheme

	OwnerReference client.Object

	Name        string
	Labels      map[string]string
	Annotations map[string]string

	Spec T
}

func (b *BaseReconciler[T]) GetClient() client.Client {
	return b.Client
}

func (b *BaseReconciler[T]) GetLabels() map[string]string {
	return b.Labels
}

func (b *BaseReconciler[T]) GetAnnotations() map[string]string {
	return b.Annotations
}

func (b *BaseReconciler[T]) GetOwnerReference() client.Object {
	return b.OwnerReference
}

func (b *BaseReconciler[T]) GetName() string {
	return b.Name
}

func (b *BaseReconciler[T]) GetSchema() *runtime.Scheme {
	return b.Schema
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
}

var _ ResourceReconciler[client.Object] = &BaseResourceReconciler[AnySpec]{}

type BaseResourceReconciler[T AnySpec] struct {
	BaseReconciler[T]
	Builder builder.Builder
}

func NewBaseResourceReconciler[T AnySpec](
	client client.Client,
	schema *runtime.Scheme,
	ownerReference client.Object,
	name string,
	spec T,
	builder builder.Builder,
) *BaseResourceReconciler[T] {
	return &BaseResourceReconciler[T]{
		BaseReconciler: BaseReconciler[T]{
			Client:         client,
			Schema:         schema,
			OwnerReference: ownerReference,
			Name:           name,
			Spec:           spec,
		},
		Builder: builder,
	}
}

func (b *BaseResourceReconciler[T]) GetBuilder() builder.Builder {
	return b.Builder
}

func (b *BaseResourceReconciler[T]) Build(ctx context.Context) (client.Object, error) {
	return b.Builder.Build(ctx)
}
