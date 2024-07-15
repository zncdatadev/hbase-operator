package restserver

import corev1 "k8s.io/api/core/v1"

var (
	Ports = []corev1.ContainerPort{
		{
			Name:          "rest-http",
			ContainerPort: 8080,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "ui-http",
			ContainerPort: 8085,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "metrics",
			ContainerPort: 9100,
			Protocol:      corev1.ProtocolTCP,
		},
	}
)
