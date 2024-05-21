package builder

import (
	"context"

	resourceClient "github.com/zncdatadev/hbase-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigBuilder interface {
	Builder
	AddData(key, value string) ConfigBuilder
	AddDecodeData(key string, value []byte) ConfigBuilder
	GetData() map[string]string
	GetEncodeData() map[string][]byte
}

var _ ConfigBuilder = &BaseConfigBuilder{}

type BaseConfigBuilder struct {
	BaseResourceBuilder

	data map[string]string
}

func NewBaseConfigBuilder(
	client resourceClient.ResourceClient,
	name string,
) *BaseConfigBuilder {
	return &BaseConfigBuilder{
		BaseResourceBuilder: BaseResourceBuilder{
			Client: client,
			Name:   name,
		},
		data: make(map[string]string),
	}
}

func (b *BaseConfigBuilder) AddData(key, value string) ConfigBuilder {
	b.data[key] = value
	return b
}

func (b *BaseConfigBuilder) AddDecodeData(key string, value []byte) ConfigBuilder {
	b.data[key] = string(value)
	return b
}

func (b *BaseConfigBuilder) GetEncodeData() map[string][]byte {
	data := make(map[string][]byte)
	for k, v := range b.data {
		data[k] = []byte(v)
	}
	return data
}

func (b *BaseConfigBuilder) GetData() map[string]string {
	return b.data
}

var _ ConfigBuilder = &ConfigMapBuilder{}

type ConfigMapBuilder struct {
	BaseConfigBuilder
}

func NewConfigMapBuilder(
	client resourceClient.ResourceClient,
	name string,
) *ConfigMapBuilder {
	return &ConfigMapBuilder{
		BaseConfigBuilder: *NewBaseConfigBuilder(client, name),
	}
}

func (b *ConfigMapBuilder) GetObject() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: b.GetObjectMeta(),
		Data:       b.GetData(),
	}
}

func (b *ConfigMapBuilder) Build(_ context.Context) (client.Object, error) {
	return b.GetObject(), nil
}

type SecretBuilder struct {
	BaseConfigBuilder
}

var _ ConfigBuilder = &SecretBuilder{}

func NewSecretBuilder(
	client resourceClient.ResourceClient,
	name string,
) *SecretBuilder {
	return &SecretBuilder{
		BaseConfigBuilder: *NewBaseConfigBuilder(client, name),
	}
}

func (b *SecretBuilder) GetObject() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: b.GetObjectMeta(),
		Data:       b.GetEncodeData(),
	}
}

func (b *SecretBuilder) Build(_ context.Context) (client.Object, error) {
	return b.GetObject(), nil
}
