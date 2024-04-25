package master

import (
	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/common"
	"github.com/zncdata-labs/hbase-operator/pkg/handler"
)

var _ handler.Attribute = &MasterRoleAttribute{}

type MasterRoleAttribute struct {
	common.BaseAttribute
	Spec *hbasev1alpha1.MasterSpec
}

func (m *MasterRoleAttribute) GetSubAttributes() ([]handler.Attribute, error) {
	var subAttributes []handler.Attribute

	base := common.BaseAttribute{
		BaseAttribute: handler.BaseAttribute{
			Name:          m.Name,
			OwnerResource: m.OwnerResource,
			Labels:        m.GetLabels(),
			Annotations:   m.GetAnnotations(),
		},

		ClusterOperation: m.ClusterOperation,
		ClusterConfig:    m.ClusterConfig,
		Image:            m.Image,
	}

	for _, roleGroup := range m.Spec.RoleGroups {
		subAttribute := &MasterRolgGroupAttribute{
			BaseAttribute: base,
			Spec:          &roleGroup,
		}

		subAttributes = append(subAttributes, subAttribute)
	}

	return subAttributes, nil
}
