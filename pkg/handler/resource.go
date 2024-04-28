package handler

import (
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ ResourceHandler = &BaseReconciler{}

type BaseReconciler struct {
	Client builder.Client

	Builder builder.ResourceBuilder

	handlers []ResourceHandler
}

func (r *BaseReconciler) GetOwnerResource() runtime.Object {
	return r.Builder.GetOwnerResource()
}

func (r *BaseReconciler) Ready() bool {
	for _, resourceHandler := range r.handlers {
		if !resourceHandler.Ready() {
			return false
		}
	}
	return true
}

func (r *BaseReconciler) GetClient() client.Client {
	return r.Client
}

func (r *BaseReconciler) GetSchema() *runtime.Scheme {
	return r.Client.Schema
}

func (r *BaseReconciler) DoReconcile() (ReconcileResult, error) {
	panic("implement me")
}

func (r *BaseReconciler) Register() error {
	panic("implement me")
}

func (r *BaseReconciler) AddReconciler(resourceHandler ResourceHandler) error {
	r.handlers = append(r.handlers, resourceHandler)
	return nil
}
