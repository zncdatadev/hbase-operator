package controller

import (
	"reflect"

	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/internal/common"
	"github.com/zncdata-labs/hbase-operator/internal/controller/master"
	"github.com/zncdata-labs/hbase-operator/pkg/handler"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	ExcludeFields = []string{"ConfigOverrides", "ClusterConfig", "PodOverride", "Image"}
)

const (
	MasterRoleName = "Master"
)

var _ handler.Attribute = &HbaseClusterAttribute{}

type HbaseClusterAttribute struct {
	handler.BaseAttribute

	Spec *hbasev1alpha1.HbaseClusterSpec
}

func (h *HbaseClusterAttribute) GetSubAttributes() ([]handler.Attribute, error) {
	var subAttributes []handler.Attribute

	specValue := reflect.ValueOf(h.Spec).Elem()
	specType := specValue.Type()

	for i := 0; i < specValue.NumField(); i++ {
		field := specValue.Field(i)
		fieldName := specType.Field(i).Name

		if slices.Contains(ExcludeFields, fieldName) {
			continue
		}

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				continue
			}
			field = field.Elem()
		}

		subAttribute := RoleAttributeFactory(
			fieldName,
			h.OwnerResource,
			field.Interface(),

			h.Spec.ClusterOperationSpec,
			h.Spec.ClusterConfigSpec,
			h.Spec.Image,
			h.GetLabels(),
			h.GetAnnotations(),
		)
		logger.Info("subAttribute", "name", subAttribute.GetName())
		subAttributes = append(subAttributes, subAttribute)
	}

	return subAttributes, nil
}

func RoleAttributeFactory(
	Name string,
	ownerResource client.Object,
	spec interface{},
	ClusterOperation *hbasev1alpha1.ClusterOperationSpec,
	ClusterConfig *hbasev1alpha1.ClusterConfigSpec,
	Image *hbasev1alpha1.ImageSpec,
	Labels map[string]string,
	Annotations map[string]string,
) handler.Attribute {

	base := common.BaseAttribute{
		BaseAttribute: handler.BaseAttribute{
			Name:          Name,
			OwnerResource: ownerResource,
			Labels:        Labels,
			Annotations:   Annotations,
		},
		ClusterOperation: ClusterOperation,
		ClusterConfig:    ClusterConfig,
		Image:            Image,
	}

	switch Name {
	case MasterRoleName:
		return &master.MasterRoleAttribute{
			BaseAttribute: base,
			Spec:          spec.(*hbasev1alpha1.MasterSpec),
		}

	default:
		panic("unimplemented")

	}
}

var _ handler.ResourceHandler = &ClusterReconciler{}

type ClusterReconciler struct {
	handler.GenericReconciler
}
