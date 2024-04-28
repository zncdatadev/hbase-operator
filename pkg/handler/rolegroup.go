package handler

import "github.com/zncdata-labs/hbase-operator/pkg/builder"

var _ ResourceHandler = &RoleGroupReconciler{}

type RoleGroupReconciler struct {
	BaseReconciler
}

func (h *RoleGroupReconciler) Register() error {

	return nil
}

func NewRoleGroupReconciler(
	client builder.Client,
	builder builder.ResourceBuilder,
) (*RoleReconciler, error) {
	obj := &RoleReconciler{
		BaseReconciler: BaseReconciler{
			Client:  client,
			Builder: builder,
		},
	}

	if err := obj.Register(); err != nil {
		return nil, err
	}

	return obj, nil
}
