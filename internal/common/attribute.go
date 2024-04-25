package common

import (
	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/handler"
)

var _ handler.Attribute = &BaseAttribute{}

type BaseAttribute struct {
	handler.BaseAttribute
	ClusterOperation *hbasev1alpha1.ClusterOperationSpec
	ClusterConfig    *hbasev1alpha1.ClusterConfigSpec
	Image            *hbasev1alpha1.ImageSpec
}
