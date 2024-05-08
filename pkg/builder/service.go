package builder

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IServiceBuilder interface {
	Builder
	AddPort(name string, port int32, protocol corev1.Protocol, targetPort int32)
	GetPorts() []corev1.ServicePort
	GetServiceType() corev1.ServiceType
}

var _ IServiceBuilder = &BaseServiceBuilder{}

type BaseServiceBuilder struct {
	BaseResourceBuilder
	ServiceType corev1.ServiceType

	ports []corev1.ServicePort
}

func (b *BaseServiceBuilder) AddPort(name string, port int32, protocol corev1.Protocol, targetPort int32) {
	b.ports = append(b.ports, corev1.ServicePort{
		Name:       name,
		Port:       port,
		Protocol:   protocol,
		TargetPort: intstr.FromInt(int(targetPort)),
	})
}

func (b *BaseServiceBuilder) GetPorts() []corev1.ServicePort {
	return b.ports
}

func (b *BaseServiceBuilder) GetServiceType() corev1.ServiceType {
	return b.ServiceType
}

func (b *BaseServiceBuilder) Build(_ context.Context) (client.Object, error) {

	obj := &corev1.Service{
		ObjectMeta: b.GetObjectMeta(),
		Spec: corev1.ServiceSpec{
			Ports:    b.GetPorts(),
			Selector: b.GetLabels(),
			Type:     b.GetServiceType(),
		},
	}

	return obj, nil

}

func NewServiceBuilder(
	client client.Client,
	name string,
	labels map[string]string,
	annotations map[string]string,
	ports []corev1.ServicePort,
	serviceType corev1.ServiceType,
) *BaseServiceBuilder {
	return &BaseServiceBuilder{
		BaseResourceBuilder: BaseResourceBuilder{
			Client:      client,
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
		ServiceType: serviceType,
		ports:       ports,
	}
}
