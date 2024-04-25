package handler

import "sigs.k8s.io/controller-runtime/pkg/client"

type ResourceHandler interface {
	DoReconcile() (ReconcileResult, error)
	Ready() bool
	Register() error
	AddReconciler(ResourceHandler) error
}

type Attribute interface {
	GetSubAttributes() ([]Attribute, error)
	GetName() string
	GetLabels() map[string]string
	GetAnnotations() map[string]string
	GetOwnerResource() client.Object
}
