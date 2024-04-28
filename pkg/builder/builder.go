package builder

import (
	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceBuilder interface {
	GetName() string
	GetNamespace() string
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetOwnerResource() client.Object

	Build() (client.Object, error)

	GetSpec() any
	GetSubBuilder() ([]ResourceBuilder, error)
}

var _ ResourceBuilder = &BaseResourceBuilder{}

type BaseResourceBuilder struct {
	Client        Client
	Name          string
	OwnerResource client.Object
	Labels        map[string]string
	Annotations   map[string]string

	Image            *hbasev1alpha1.ImageSpec
	ClusterConfig    *hbasev1alpha1.ClusterConfigSpec
	ClusterOperation *hbasev1alpha1.ClusterOperationSpec
}

func (r *BaseResourceBuilder) GetName() string {
	return r.Name
}

func (r *BaseResourceBuilder) GetOwnerResource() client.Object {
	return r.OwnerResource
}

func (r *BaseResourceBuilder) GetNamespace() string {
	return r.OwnerResource.GetNamespace()
}

func (r *BaseResourceBuilder) GetSubBuilder() ([]ResourceBuilder, error) {
	panic("unimplemented")
}

func (r *BaseResourceBuilder) GetAnnotations() map[string]string {
	panic("unimplemented")
}

func (r *BaseResourceBuilder) GetLabels() map[string]string {
	panic("unimplemented")
}

func (r *BaseResourceBuilder) Build() (client.Object, error) {
	panic("implement me")
}

func (r *BaseResourceBuilder) GetSubResourceBuilder() ([]ResourceBuilder, error) {
	panic("implement me")
}

func (r *BaseResourceBuilder) GetSpec() any {
	panic("implement me")
}
