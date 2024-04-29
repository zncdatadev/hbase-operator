package builder

import (
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder interface {
	Build() (client.Object, error)

	GetClient() client.Client
	GetSchema() *runtime.Scheme
	GetOwnerReference() runtime.Object

	GetLabels() map[string]string
	GetAnnotations() map[string]string
}

var _ ResourceBuilder = &BaseResourceBuilder{}

type BaseResourceBuilder struct {
	Client client.Client
	Schema *runtime.Scheme

	OwnerReference runtime.Object

	Name             string
	Labels           map[string]string
	Annotations      map[string]string
	ClusterOperation *apiv1alpha1.ClusterOperationSpec

	// Spec T
}

func (b *BaseResourceBuilder) GetClient() client.Client {
	return b.Client
}

func (b *BaseResourceBuilder) GetSchema() *runtime.Scheme {
	return b.Schema
}

func (b *BaseResourceBuilder) GetOwnerReference() runtime.Object {
	return b.OwnerReference
}

func (b *BaseResourceBuilder) GetName() string {
	return b.Name
}

func (b *BaseResourceBuilder) GetLabels() map[string]string {
	return b.Labels
}

func (b *BaseResourceBuilder) GetAnnotations() map[string]string {
	return b.Annotations
}

// func (b *BaseResourceBuilder[T]) GetSpec() T {
// 	return b.Spec
// }

func (b *BaseResourceBuilder) Build() (client.Object, error) {
	panic("implement me")
}
