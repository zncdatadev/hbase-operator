package builder

import "sigs.k8s.io/controller-runtime/pkg/client"

type ResourceBuilder interface {
	Builder
}

type WorkloadOverwide[T client.Object] interface {
	CommandOverwide(obj T)
	EnvOverwide(obj T)
	LogOverwide(obj T)
	ConfigurationOverwide(obj T)
}
