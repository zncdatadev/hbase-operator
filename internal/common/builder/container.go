package builder

import (
	corev1 "k8s.io/api/core/v1"
)

// TODO: Add more fields to ContainerBuilder

type IContainerBuilder interface {
	Build() corev1.Container
}

type ContainerBuilder struct {
	Name       string
	Image      string
	PullPolicy corev1.PullPolicy
	Command    []string
	Args       []string
	Env        []corev1.EnvVar
	Ports      []corev1.ContainerPort
}
