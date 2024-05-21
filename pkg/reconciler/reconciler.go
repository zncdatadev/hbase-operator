package reconciler

import (
	"context"
	"time"

	"github.com/zncdatadev/hbase-operator/pkg/builder"
	resourceClient "github.com/zncdatadev/hbase-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AnySpec any

type Reconciler interface {
	GetClient() resourceClient.ResourceClient
	Reconcile(ctx context.Context) Result
	Ready(ctx context.Context) Result
	GetNameWithSuffix(suffix string) string
}

var _ Reconciler = &BaseReconciler[AnySpec]{}

type BaseReconciler[T AnySpec] struct {
	Client resourceClient.ResourceClient
	Name   string

	Spec T
}

func (b *BaseReconciler[T]) GetClient() resourceClient.ResourceClient {
	return b.Client
}

func (b *BaseReconciler[T]) GetName() string {
	return b.Name
}

func (b *BaseReconciler[T]) GetNameWithSuffix(suffix string) string {
	return b.Name + "-" + suffix
}

func (b *BaseReconciler[T]) GetScheme() *runtime.Scheme {
	return b.Client.Scheme()
}

func (b *BaseReconciler[T]) Ready(ctx context.Context) Result {
	panic("unimplemented")
}

func (b *BaseReconciler[T]) Reconcile(ctx context.Context) Result {
	panic("unimplemented")
}

func (b *BaseReconciler[T]) GetSpec() T {
	return b.Spec
}

type ResourceReconciler[B builder.Builder] interface {
	Reconciler
	GetBuilder() B
	ResourceReconcile(ctx context.Context, resource client.Object) Result
}

var _ ResourceReconciler[builder.Builder] = &BaseResourceReconciler[AnySpec, builder.Builder]{}

type BaseResourceReconciler[T AnySpec, B builder.Builder] struct {
	BaseReconciler[T]
	Builder B
}

func (r *BaseResourceReconciler[T, B]) GetBuilder() B {
	return r.Builder
}

func (r *BaseResourceReconciler[T, B]) ResourceReconcile(ctx context.Context, resource client.Object) Result {

	if err := ctrl.SetControllerReference(r.Client.OwnerReference, resource, r.Client.Scheme()); err != nil {
		return NewResult(true, 0, err)
	}

	if mutation, err := r.Client.CreateOrUpdate(ctx, resource); err != nil {
		return NewResult(true, 0, err)
	} else if mutation {
		return NewResult(true, time.Second, nil)
	}
	return NewResult(false, 0, nil)
}

func (r *BaseResourceReconciler[T, B]) Reconcile(ctx context.Context) Result {
	resource, err := r.GetBuilder().Build(ctx)

	if err != nil {
		return NewResult(true, 0, err)
	}
	return r.ResourceReconcile(ctx, resource)
}

func (r *BaseResourceReconciler[T, B]) Ready(ctx context.Context) Result {
	return NewResult(false, 0, nil)
}

func NewBaseResourceReconciler[T AnySpec, B builder.Builder](
	client resourceClient.ResourceClient,
	name string,
	spec T,
	builder B,
) *BaseResourceReconciler[T, B] {
	return &BaseResourceReconciler[T, B]{
		BaseReconciler: BaseReconciler[T]{
			Client: client,
			Name:   name,
			Spec:   spec,
		},
		Builder: builder,
	}
}

// NewSimpleResourceReconciler creates a new resource reconciler with a simple builder
// that does not require a spec, and can not use the spec.
func NewSimpleResourceReconciler[B builder.Builder](
	client resourceClient.ResourceClient,
	name string,
	builder B,
) *BaseResourceReconciler[AnySpec, B] {
	return NewBaseResourceReconciler[AnySpec, B](client, name, nil, builder)
}
