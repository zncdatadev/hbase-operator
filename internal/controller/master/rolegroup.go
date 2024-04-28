package master

import (
	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ builder.ResourceBuilder = &MasterRoleGroupBuilder{}

type MasterRoleGroupBuilder struct {
	builder.BaseResourceBuilder

	Image            *hbasev1alpha1.ImageSpec
	ClusterConfig    *hbasev1alpha1.ClusterConfigSpec
	ClusterOperation *hbasev1alpha1.ClusterOperationSpec
	Spec             *hbasev1alpha1.MasterRoleGroupSpec
}

func (m *MasterRoleGroupBuilder) Build() (client.Object, error) {
	return nil, nil
}

func (m *MasterRoleGroupBuilder) GetSubResourceBuilder() ([]builder.ResourceBuilder, error) {
	objs := make([]builder.ResourceBuilder, 0)

	if m.Spec == nil {
		return objs, nil
	}

	base := builder.BaseResourceBuilder{
		Name:          m.GetName(),
		Client:        m.Client,
		OwnerResource: m.GetOwnerResource(),
		Labels:        m.GetLabels(),
		Annotations:   m.GetAnnotations(),
	}

	statefulSetBuilder := &StatefulSetBuilder{
		StatefulSet: builder.StatefulSet{
			BaseResourceBuilder: base,
		},
	}

	objs = append(objs, statefulSetBuilder)

	return objs, nil
}
