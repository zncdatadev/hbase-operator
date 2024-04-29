package reconciler

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceReconciler[T client.Object] interface {
	Build(ctx context.Context) (T, error)
}

type WorkloadReconciler[T client.Object] interface {
	ResourceReconciler[T]
	CommandOverwide(obj T)
	EnvOverwide(obj T)
}

type ConfigurationReconciler[T client.Object] interface {
	ResourceReconciler[T]

	LogOverwide(obj T)
	ConfigurationOverwide(obj T)
}
