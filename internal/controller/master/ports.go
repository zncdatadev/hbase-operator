package master

import corev1 "k8s.io/api/core/v1"

var (
	Ports = []corev1.ContainerPort{
		{
			Name:          "master",
			ContainerPort: 16000,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "ui-http",
			ContainerPort: 16010,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "metrics",
			ContainerPort: 9100,
			Protocol:      corev1.ProtocolTCP,
		},
	}
)
