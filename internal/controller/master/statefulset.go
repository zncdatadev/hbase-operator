package master

import (
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	appv1 "k8s.io/api/apps/v1"
)

var _ builder.ResourceBuilder[appv1.StatefulSet] = &StatefulSetBuilder{}

type StatefulSetBuilder struct {
	builder.StatefulSet
}

func NewStatefulSetBuilder() (*StatefulSetBuilder, error) {
	return &StatefulSetBuilder{}, nil
}

// var _ handler.ResourceHandler = &StatefulSetReconciler{}

// type StatefulSetReconciler struct {
// 	handler.ResourceReconciler
// }
