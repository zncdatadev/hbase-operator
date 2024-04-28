package handler

import "github.com/zncdata-labs/hbase-operator/pkg/builder"

var _ ResourceHandler = &RoleReconciler{}

type RoleReconciler struct {
	BaseReconciler
}

func (h *RoleReconciler) Register() error {

	subs, err := h.Builder.GetSubBuilder()

	if err != nil {
		return err
	}

	for _, sub := range subs {
		reconciler, err := NewRoleGroupReconciler(h.Client, sub)
		if err != nil {
			return err
		}

		if err := reconciler.Register(); err != nil {
			return err
		}

		if err := h.AddReconciler(reconciler); err != nil {
			return err
		}
	}

	return nil
}

func NewRoleReconciler(
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
