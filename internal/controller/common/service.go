package common

import (
	"context"

	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"github.com/zncdata-labs/hbase-operator/pkg/client"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
)

var _ reconciler.ResourceReconciler[builder.ServiceBuilder] = &ServiceReconciler[reconciler.AnySpec]{}

type ServiceReconciler[T reconciler.AnySpec] struct {
	reconciler.BaseResourceReconciler[T, builder.ServiceBuilder]
	Ports []corev1.ContainerPort
}

func (r *ServiceReconciler[T]) Build(_ context.Context) (*corev1.Service, error) {
	var ports []corev1.ServicePort

	for _, port := range r.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:     port.Name,
			Port:     port.ContainerPort,
			Protocol: port.Protocol,
		})
	}

	obj := &corev1.Service{
		ObjectMeta: r.GetObjectMeta(),
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": r.GetName(),
			},
			Ports: ports,
		},
	}
	return obj, nil
}

func NewServiceReconciler[T reconciler.AnySpec](
	client client.ResourceClient,
	roleGroupName string,
	ports []corev1.ContainerPort,
	spec T,
) *ServiceReconciler[T] {

	var svcPorts []corev1.ServicePort

	for _, port := range ports {
		svcPorts = append(svcPorts, corev1.ServicePort{
			Name:     port.Name,
			Port:     port.ContainerPort,
			Protocol: port.Protocol,
		})
	}

	svcBuilder := builder.NewServiceBuilder(
		client,
		roleGroupName,
		client.GetLabels(),
		client.GetAnnotations(),
		svcPorts,
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
