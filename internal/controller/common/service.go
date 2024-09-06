package common

import (
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
)

var _ reconciler.ResourceReconciler[builder.ServiceBuilder] = &ServiceReconciler{}

type ServiceReconciler struct {
	reconciler.Service
}

func NewServiceReconciler(
	client *client.Client,
	ports []corev1.ContainerPort,
	options reconciler.RoleGroupInfo,
) *ServiceReconciler {

	return &ServiceReconciler{
		Service: *reconciler.NewServiceReconciler(
			client,
			options.GetFullName(),
			options.GetLabels(),
			options.GetAnnotations(),
			ports,
			&[]corev1.ServiceType{corev1.ServiceTypeClusterIP}[0],
		),
	}
}
