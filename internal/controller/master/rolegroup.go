package master

import (
	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/common"
	"github.com/zncdata-labs/hbase-operator/pkg/handler"
)

var _ handler.Attribute = &MasterRolgGroupAttribute{}

type MasterRolgGroupAttribute struct {
	common.BaseAttribute

	Spec *hbasev1alpha1.MasterRoleGroupSpec
}

func (m *MasterRolgGroupAttribute) GetSubAttributes() ([]handler.Attribute, error) {
	var subAttributes []handler.Attribute

	return subAttributes, nil
}
