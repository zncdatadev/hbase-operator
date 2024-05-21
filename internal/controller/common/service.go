package common

import (
	"github.com/zncdatadev/hbase-operator/pkg/builder"
	"github.com/zncdatadev/hbase-operator/pkg/client"
	"github.com/zncdatadev/hbase-operator/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
)

var _ reconciler.ResourceReconciler[builder.ServiceBuilder] = &ServiceReconciler[reconciler.AnySpec]{}

type ServiceReconciler[T reconciler.AnySpec] struct {
	reconciler.BaseResourceReconciler[T, builder.ServiceBuilder]
	Ports []corev1.ContainerPort
}

func NewServiceReconciler[T reconciler.AnySpec](
	client client.ResourceClient,
	roleGroupName string,
	ports []corev1.ContainerPort,
	spec T,
) *ServiceReconciler[T] {

	svcBuilder := builder.NewServiceBuilder(
		client,
		roleGroupName,
		ports,
	)
	return &ServiceReconciler[T]{
		BaseResourceReconciler: *reconciler.NewBaseResourceReconciler[T, builder.ServiceBuilder](
			client,
			roleGroupName,
			spec,
			svcBuilder,
		),
		Ports: ports,
	}
}
