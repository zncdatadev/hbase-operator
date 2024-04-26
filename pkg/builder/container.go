package builder

import (
	corev1 "k8s.io/api/core/v1"
)

type ContainerBuilder interface {
	Args() []string
	Ports() []corev1.ContainerPort
	Env() []corev1.EnvVar
	VolumeMounts() []corev1.VolumeMount
	LivenessProbe() *corev1.Probe
	ReadinessProbe() *corev1.Probe
	ImagePullPolicy() corev1.PullPolicy

	Build() *corev1.Container
}

var _ ContainerBuilder = &Container{}

type Container struct {
	Name string
}

func NewContainer() (*Container, error) {
	return &Container{}, nil
}

func (c *Container) Build() *corev1.Container {
	return &corev1.Container{
		Name:            c.Name,
		Image:           "docker.stackable.tech/stackable/hbase:2.4.17-stackable24.3.0",
		Args:            c.Args(),
		Ports:           c.Ports(),
		Env:             c.Env(),
		VolumeMounts:    c.VolumeMounts(),
		LivenessProbe:   c.LivenessProbe(),
		ReadinessProbe:  c.ReadinessProbe(),
		ImagePullPolicy: c.ImagePullPolicy(),
	}

}

func (c *Container) Args() []string {
	return []string{}
}

func (c *Container) Ports() []corev1.ContainerPort {
	return []corev1.ContainerPort{}
}

func (c *Container) Env() []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func (c *Container) VolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{}
}

func (c *Container) LivenessProbe() *corev1.Probe {
	return nil
}

func (c *Container) ReadinessProbe() *corev1.Probe {
	return nil
}

func (c *Container) ImagePullPolicy() corev1.PullPolicy {
	return corev1.PullIfNotPresent
}
