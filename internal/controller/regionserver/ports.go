package regionserver

import corev1 "k8s.io/api/core/v1"

var (
	Ports = []corev1.ContainerPort{
		{
			Name:          "master",
			ContainerPort: 16000,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "http",
			ContainerPort: 16010,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "9003",
			ContainerPort: 16030,
			Protocol:      corev1.ProtocolTCP,
		},
	}
)
