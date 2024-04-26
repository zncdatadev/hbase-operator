package master

import (
	hbasev1alpha1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	logger = ctrl.Log.WithName("master")
)

var _ builder.ResourceBuilder = &MasterRoleBuilder{}

type MasterRoleBuilder struct {
	builder.BaseResourceBuilder

	Spec *hbasev1alpha1.MasterSpec
}

func (m *MasterRoleBuilder) Build() (client.Object, error) {
	logger.Info("Building MasterRole")
	return nil, nil
}

// var _ handler.ResourceHandler = &MasterRoleReconciler{}

// type MasterRoleReconciler struct {
// 	handler.BaseReconciler
// }

// func NewMasterRoleReconciler(
// 	client builder.Client,
// 	builder builder.ResourceBuilder,

// ) (*MasterRoleReconciler, error) {
// 	obj := &MasterRoleReconciler{
// 		BaseReconciler: handler.BaseReconciler{
// 			Client:  client,
// 			Builder: builder,
// 		},
// 	}

// 	if err := obj.Register(); err != nil {
// 		return nil, err
// 	}

// 	return obj, nil
// }
