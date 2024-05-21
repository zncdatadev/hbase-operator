package builder

import (
	"context"

	resourceClient "github.com/zncdatadev/hbase-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceBuilder interface {
	Builder
	AddPort(name string, port int32, protocol corev1.Protocol, targetPort int32)
	GetPorts() []corev1.ServicePort
	GetServiceType() corev1.ServiceType
}

var _ ServiceBuilder = &BaseServiceBuilder{}

type BaseServiceBuilder struct {
	BaseResourceBuilder

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
	return corev1.ServiceTypeClusterIP
}

func (b *BaseServiceBuilder) Build(_ context.Context) (client.Object, error) {

	obj := &corev1.Service{
		ObjectMeta: b.GetObjectMeta(),
		Spec: corev1.ServiceSpec{
			Ports:    b.GetPorts(),
			Selector: b.Client.GetMatchingLabels(),
			Type:     b.GetServiceType(),
		},
	}

	return obj, nil

}

// NewGenericServiceBuilder creates a new service builder with generic ports
// FIXME: There may be performance issues or fatal exceptions in the following code.
func NewGenericServiceBuilder[T corev1.ContainerPort | corev1.ServicePort](
	client resourceClient.ResourceClient,
	name string,
	ports []T,
) *BaseServiceBuilder {

	var svcPorts []corev1.ServicePort

	switch any(new(T)).(type) {
	case corev1.ContainerPort:
		for _, tPort := range ports {
			port, ok := any(tPort).(corev1.ContainerPort)
			if !ok {
				panic("invalid type")
			}

			svcPorts = append(svcPorts, corev1.ServicePort{
				Name:     port.Name,
				Port:     port.ContainerPort,
				Protocol: port.Protocol,
			})
		}
	case corev1.ServicePort:
		var ok bool
		svcPorts, ok = any(ports).([]corev1.ServicePort)
		if !ok {
			panic("invalid type")
		}
	}

	return &BaseServiceBuilder{
		BaseResourceBuilder: BaseResourceBuilder{
			Client: client,
			Name:   name,
		},
		ports: svcPorts,
	}
}

func NewServiceBuilder(
	client resourceClient.ResourceClient,
	name string,
	ports []corev1.ContainerPort,
) *BaseServiceBuilder {
	var svcPorts []corev1.ServicePort

	for _, port := range ports {
		svcPorts = append(svcPorts, corev1.ServicePort{
			Name:     port.Name,
			Port:     port.ContainerPort,
			Protocol: port.Protocol,
		})
	}

	return &BaseServiceBuilder{
		BaseResourceBuilder: BaseResourceBuilder{
			Client: client,
			Name:   name,
		},
		ports: svcPorts,
	}
}
