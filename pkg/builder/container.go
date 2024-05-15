package builder

import (
	"slices"

	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type ContainerBuilder interface {
	Build() *corev1.Container
	AddVolumeMounts(mounts []corev1.VolumeMount) ContainerBuilder
	AddEnvVars(envVars []corev1.EnvVar) ContainerBuilder
	AddEnvs(envs map[string]string) ContainerBuilder
	AddPorts(ports []corev1.ContainerPort) ContainerBuilder
	SetResources(resources apiv1alpha1.ResourcesSpec) ContainerBuilder
	SetLiveProbe(probe *corev1.Probe) ContainerBuilder
	SetReadinessProbe(probe *corev1.Probe) ContainerBuilder
	SetStartupProbe(probe *corev1.Probe) ContainerBuilder
	SetSecurityContext(user int64, group int64, nonRoot bool) ContainerBuilder
	SetCommand(command []string) ContainerBuilder
	SetArgs(args []string) ContainerBuilder
	OverrideEnv(envs map[string]string) ContainerBuilder
	AutomaticSetProbe() ContainerBuilder
}

var _ ContainerBuilder = &GenericContainerBuilder{}

type GenericContainerBuilder struct {
	Name       string
	Image      string
	PullPolicy corev1.PullPolicy

	obj *corev1.Container
}

func NewGenericContainerBuilder(
	name, image string,
	pullPolicy corev1.PullPolicy,
) *GenericContainerBuilder {
	return &GenericContainerBuilder{
		Name:       name,
		Image:      image,
		PullPolicy: pullPolicy,
	}
}

func (b *GenericContainerBuilder) getObject() *corev1.Container {
	if b.obj == nil {
		b.obj = &corev1.Container{
			Name:            b.Name,
			Image:           b.Image,
			ImagePullPolicy: b.PullPolicy,
		}
	}
	return b.obj
}

func (b *GenericContainerBuilder) Build() *corev1.Container {
	obj := b.getObject()
	return obj
}

func (b *GenericContainerBuilder) AddVolumeMounts(mounts []corev1.VolumeMount) ContainerBuilder {
	b.getObject().VolumeMounts = mounts
	return b
}

func (b *GenericContainerBuilder) AddEnvVars(envVars []corev1.EnvVar) ContainerBuilder {
	envs := b.getObject().Env
	envs = append(envs, envVars...)
	var envNames []string
	for _, env := range envs {
		if slices.Contains(envNames, env.Name) {
			logger.V(2).Info("EnvVar already exists, it may be overwritten", "env", env.Name)
		}
		envNames = append(envNames, env.Name)
	}
	b.getObject().Env = envs
	return b
}

func (b *GenericContainerBuilder) AddEnvs(envs map[string]string) ContainerBuilder {
	var envVars []corev1.EnvVar
	for name, value := range envs {
		envVars = append(envVars, corev1.EnvVar{
			Name:  name,
			Value: value,
		})
	}
	return b.AddEnvVars(envVars)
}

func (b *GenericContainerBuilder) AddPorts(ports []corev1.ContainerPort) ContainerBuilder {
	p := b.getObject().Ports

	p = append(p, ports...)
	b.getObject().Ports = p

	return b
}

// SetCommand sets the command for the container
// and clears the args.
func (b *GenericContainerBuilder) SetCommand(command []string) ContainerBuilder {
	b.getObject().Command = command
	b.getObject().Args = []string{}
	return b
}

func (b *GenericContainerBuilder) SetArgs(args []string) ContainerBuilder {
	b.getObject().Args = args
	return b
}

func (b *GenericContainerBuilder) OverrideEnv(envs map[string]string) ContainerBuilder {
	b.getObject().Env = []corev1.EnvVar{}
	return b.AddEnvs(envs)
}

func (b *GenericContainerBuilder) SetResources(resources apiv1alpha1.ResourcesSpec) ContainerBuilder {
	obj := b.getObject()
	if resources.CPU != nil {
		obj.Resources.Requests[corev1.ResourceCPU] = resources.CPU.Min
		obj.Resources.Limits[corev1.ResourceCPU] = resources.CPU.Max
	}
	if resources.Memory != nil {
		obj.Resources.Requests[corev1.ResourceMemory] = resources.Memory.Limit
	}
	return b

}

func (b *GenericContainerBuilder) SetLiveProbe(probe *corev1.Probe) ContainerBuilder {
	b.getObject().LivenessProbe = probe
	return b
}

func (b *GenericContainerBuilder) SetReadinessProbe(probe *corev1.Probe) ContainerBuilder {
	b.getObject().ReadinessProbe = probe
	return b
}

func (b *GenericContainerBuilder) SetStartupProbe(probe *corev1.Probe) ContainerBuilder {
	b.getObject().StartupProbe = probe
	return b
}

func (b *GenericContainerBuilder) SetSecurityContext(user int64, group int64, nonRoot bool) ContainerBuilder {
	b.getObject().SecurityContext = &corev1.SecurityContext{
		RunAsUser:                &user,
		RunAsGroup:               &group,
		AllowPrivilegeEscalation: &nonRoot,
	}
	return b
}

// AutomaticSetProbe sets the liveness, readiness and startup probes
// policy:
// - handle policy:
//   - if name of ports contains "http", "ui", "metrics" or "health", use httpGet
//   - if name of ports contains "master", use tcpSocket
//   - todo: add more rules
//
// - startupProbe:
//   - failureThreshold: 30
//   - initialDelaySeconds: 4
//   - periodSeconds: 6
//   - successThreshold: 1
//   - timeoutSeconds: 3
//
// - livenessProbe:
//   - failureThreshold: 3
//   - periodSeconds: 10
//   - successThreshold: 1
//   - timeoutSeconds: 3
//
// - readinessProbe:
//   - failureThreshold: 3
//   - periodSeconds: 10
//   - successThreshold: 1
//   - timeoutSeconds: 3
func (b *GenericContainerBuilder) AutomaticSetProbe() ContainerBuilder {

	probeHandler := b.getProbeHandler()

	if probeHandler == nil {
		logger.V(2).Info("No probe handler found, skip setting probes")
		return b
	}

	// Set startup probe
	startupProbe := &corev1.Probe{
		FailureThreshold:    30,
		InitialDelaySeconds: 4,
		PeriodSeconds:       6,
		SuccessThreshold:    1,
		TimeoutSeconds:      3,
		ProbeHandler:        *probeHandler,
	}
	b.SetStartupProbe(startupProbe)

	// Set liveness probe
	livenessProbe := &corev1.Probe{
		FailureThreshold: 3,
		PeriodSeconds:    10,
		SuccessThreshold: 1,
		TimeoutSeconds:   3,
		ProbeHandler:     *probeHandler,
	}
	b.SetLiveProbe(livenessProbe)

	// Set readiness probe
	readinessProbe := &corev1.Probe{
		FailureThreshold: 3,
		PeriodSeconds:    10,
		SuccessThreshold: 1,
		TimeoutSeconds:   3,
		ProbeHandler:     *probeHandler,
	}
	b.SetReadinessProbe(readinessProbe)

	return b
}

// getProbeHandler returns the handler for the probe
// policy:
// - handle policy:
//   - if name of ports contains "http", "ui", "metrics" or "health", use httpGet
//   - if name of ports contains "master", use tcpSocket
//   - todo: add more rules
func (b *GenericContainerBuilder) getProbeHandler() *corev1.ProbeHandler {
	for _, port := range b.getObject().Ports {
		if slices.Contains(HTTPGetProbHandler2PortNames, port.Name) {
			return &corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: "/",
					Port: intstr.FromString(port.Name),
				},
			}
		}
		if slices.Contains(TCPProbHandler2PortNames, port.Name) {
			return &corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromString(port.Name),
				},
			}
		}
	}
	return nil
}

var (
	HTTPGetProbHandler2PortNames = []string{"http", "ui", "metrics", "health"}
	TCPProbHandler2PortNames     = []string{"master"}
)
