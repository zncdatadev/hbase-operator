package handler

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ ResourceHandler = &GenericReconciler{}

type GenericReconciler struct {
	Client Client

	Attribute Attribute

	handlers []ResourceHandler
}

func (r *GenericReconciler) GetResourceHandlers() ([]ResourceHandler, error) {
	objs := make([]ResourceHandler, 0)

	subResource, err := r.Attribute.GetSubAttributes()

	if err != nil {
		return nil, err
	}

	for _, resource := range subResource {
		resourceReconciler, err := NewGenericReconciler(
			r.Client,
			resource,
		)

		if err != nil {
			return nil, err
		}

		objs = append(objs, resourceReconciler)
	}

	return objs, nil
}

func (r *GenericReconciler) Register() error {

	objs, err := r.GetResourceHandlers()
	if err != nil {
		return err
	}

	r.handlers = objs
	return nil
}

func (r *GenericReconciler) AddReconciler(resourceHandler ResourceHandler) error {
	r.handlers = append(r.handlers, resourceHandler)
	return nil
}

func (r *GenericReconciler) DoReconcile() (ReconcileResult, error) {
	for _, resourceHandler := range r.handlers {
		result, err := resourceHandler.DoReconcile()
		if err != nil {
			return ReconcileResult{}, err
		}

		if result.Requeue {
			return result, nil
		}
	}

	return ReconcileResult{}, nil
}

func (r *GenericReconciler) Ready() bool {
	for _, resourceHandler := range r.handlers {
		if !resourceHandler.Ready() {
			return false
		}
	}

	return true
}

func (r *GenericReconciler) GetOwnerResource() client.Object {
	return r.Attribute.GetOwnerResource()
}

func (r *GenericReconciler) GetClient() client.Client {
	return r.Client
}

func (r *GenericReconciler) GetSchema() *runtime.Scheme {
	return r.Client.Schema
}

func NewGenericReconciler(
	Client Client,
	Attribute Attribute,
) (*GenericReconciler, error) {
	obj := &GenericReconciler{
		Client: Client,

		Attribute: Attribute,
	}

	err := obj.Register()
	if err != nil {
		return nil, err
	}

	return obj, nil
}

var _ Attribute = &BaseAttribute{}

type BaseAttribute struct {
	Name          string
	OwnerResource client.Object
	Labels        map[string]string
	Annotations   map[string]string
}

func (b *BaseAttribute) GetAnnotations() map[string]string {
	return b.Annotations
}

func (b *BaseAttribute) GetLabels() map[string]string {
	return b.Labels
}

func (b *BaseAttribute) GetName() string {
	return b.Name
}

func (b *BaseAttribute) GetOwnerResource() client.Object {
	return b.OwnerResource
}

// GetSubAttributes implements Attribute.
func (b *BaseAttribute) GetSubAttributes() ([]Attribute, error) {
	panic("unimplemented")
}
