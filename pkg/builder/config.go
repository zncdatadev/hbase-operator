package builder

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigBuilder interface {
	Builder
	AddData(key, value string)
	AddDecodeData(key string, value []byte)
	GetData() map[string]string
	GetEncodeData() map[string][]byte
}

var _ ConfigBuilder = &BaseConfigBuilder[corev1.Secret]{}

type BaseConfigBuilder[T corev1.Secret | corev1.ConfigMap] struct {
	BaseResourceBuilder

	data map[string]string
}

func NewConfigBuilder[T corev1.Secret | corev1.ConfigMap](
	client client.Client,
	name, namespace string,
	labels, annotations map[string]string,
) *BaseConfigBuilder[T] {
	return &BaseConfigBuilder[T]{
		BaseResourceBuilder: BaseResourceBuilder{
			Client:      client,
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		data: make(map[string]string),
	}
}

func (b *BaseConfigBuilder[T]) AddData(key, value string) {
	b.data[key] = value
}

func (b *BaseConfigBuilder[T]) AddDecodeData(key string, value []byte) {
	b.data[key] = string(value)
}

func (b *BaseConfigBuilder[T]) GetEncodeData() map[string][]byte {
	data := make(map[string][]byte)
	for k, v := range b.data {
		data[k] = []byte(v)
	}
	return data
}

func (b *BaseConfigBuilder[T]) GetData() map[string]string {
	return b.data
}

func (b *BaseConfigBuilder[T]) Build(_ context.Context) (client.Object, error) {
	var obj client.Object
	switch o := any(new(T)).(type) {
	case *corev1.Secret:
		obj = &corev1.Secret{
			ObjectMeta: b.GetObjectMeta(),
			Data:       b.GetEncodeData(),
		}
	case *corev1.ConfigMap:
		obj = &corev1.ConfigMap{
			ObjectMeta: b.GetObjectMeta(),
			Data:       b.GetData(),
		}
	default:
		panic(fmt.Sprintf("invalid type %T, must be *corev1.Secret or *corev1.ConfigMap", o))
	}
	return obj, nil
}
