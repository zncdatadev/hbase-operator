package master

import (
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ builder.ResourceBuilder = &StatefulSetBuilder{}

type StatefulSetBuilder struct {
	builder.StatefulSet
}

// Build implements builder.ResourceBuilder.
// Subtle: this method shadows the method (StatefulSet).Build of StatefulSetBuilder.StatefulSet.
func (s *StatefulSetBuilder) Build() (client.Object, error) {
	panic("unimplemented")
}

// GetAnnotations implements builder.ResourceBuilder.
func (s *StatefulSetBuilder) GetAnnotations() map[string]string {
	panic("unimplemented")
}

// GetLabels implements builder.ResourceBuilder.
func (s *StatefulSetBuilder) GetLabels() map[string]string {
	panic("unimplemented")
}

// GetName implements builder.ResourceBuilder.
func (s *StatefulSetBuilder) GetName() string {
	panic("unimplemented")
}

// GetNamespace implements builder.ResourceBuilder.
func (s *StatefulSetBuilder) GetNamespace() string {
	panic("unimplemented")
}

// GetOwnerResource implements builder.ResourceBuilder.
func (s *StatefulSetBuilder) GetOwnerResource() client.Object {
	panic("unimplemented")
}

// GetSpec implements builder.ResourceBuilder.
func (s *StatefulSetBuilder) GetSpec() any {
	panic("unimplemented")
}

// GetSubBuilder implements builder.ResourceBuilder.
func (s *StatefulSetBuilder) GetSubBuilder() ([]builder.ResourceBuilder, error) {
	panic("unimplemented")
}
