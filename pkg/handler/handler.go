package handler

import "k8s.io/apimachinery/pkg/runtime"

type ResourceHandler interface {
	GetOwnerResource() runtime.Object
	DoReconcile() (ReconcileResult, error)
	Ready() bool

	Register() error
	AddReconciler(ResourceHandler) error
}
