package builder

import (
	"context"

	resourceClient "github.com/zncdatadev/hbase-operator/pkg/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("builder")
)

type Builder interface {
	Build(ctx context.Context) (ctrlclient.Object, error)
	GetObjectMeta() metav1.ObjectMeta
	GetClient() resourceClient.ResourceClient
	GetName() string
}

var _ Builder = &BaseResourceBuilder{}

type BaseResourceBuilder struct {
	Client resourceClient.ResourceClient

	Name string
}

func (b *BaseResourceBuilder) GetClient() resourceClient.ResourceClient {
	return b.Client
}

func (b *BaseResourceBuilder) GetName() string {
	return b.Name
}

func (b *BaseResourceBuilder) GetObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        b.Name,
		Namespace:   b.Client.GetOwnerNamespace(),
		Labels:      b.Client.GetLabels(),
		Annotations: b.Client.GetAnnotations(),
	}
}

// GetObjectMetaWithClusterScope returns the object meta with cluster scope
func (b *BaseResourceBuilder) GetObjectMetaWithClusterScope() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        b.Name,
		Labels:      b.Client.GetLabels(),
		Annotations: b.Client.GetAnnotations(),
	}
}

func (b *BaseResourceBuilder) Build(ctx context.Context) (ctrlclient.Object, error) {
	panic("implement me")
}
