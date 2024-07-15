package regionserver

import corev1 "k8s.io/api/core/v1"

var (
	Ports = []corev1.ContainerPort{
		{
			Name:          "regionserver",
			ContainerPort: 16020,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "ui-http",
			ContainerPort: 16030,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "metrics",
			ContainerPort: 9100,
			Protocol:      corev1.ProtocolTCP,
		},
	}
)
