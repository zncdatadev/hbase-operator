package builder

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder interface {
	Build(ctx context.Context) (client.Object, error)
	GetObjectMeta() metav1.ObjectMeta
	GetClient() client.Client
	GetName() string
	GetNamespace() string
	GetLabels() map[string]string
	GetAnnotations() map[string]string
}

var _ ResourceBuilder = &BaseResourceBuilder{}

type BaseResourceBuilder struct {
	Client client.Client

	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

func (b *BaseResourceBuilder) GetClient() client.Client {
	return b.Client
}

func (b *BaseResourceBuilder) GetName() string {
	return b.Name
}

func (b *BaseResourceBuilder) GetNamespace() string {
	return b.Namespace
}

func (b *BaseResourceBuilder) GetObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        b.Name,
		Namespace:   b.Namespace,
		Labels:      b.Labels,
		Annotations: b.Annotations,
	}

}

func (b *BaseResourceBuilder) GetLabels() map[string]string {
	return b.Labels
}

func (b *BaseResourceBuilder) GetAnnotations() map[string]string {
	return b.Annotations
}

func (b *BaseResourceBuilder) Build(ctx context.Context) (client.Object, error) {
	panic("implement me")
}
