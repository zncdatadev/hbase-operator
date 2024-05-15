package reconciler

import (
	"context"

	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"github.com/zncdata-labs/hbase-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

var _ ResourceReconciler[builder.StatefulSetBuilder] = &StatefulSetReconciler[AnySpec]{}

type StatefulSetReconciler[T AnySpec] struct {
	BaseResourceReconciler[T, builder.StatefulSetBuilder]
	Ports         []corev1.ContainerPort
	RoleGroupInfo RoleGroupInfo
}

// getReplicas returns the number of replicas for the role group.
// handle cluster operation stopped state.
func (r *StatefulSetReconciler[T]) getReplicas() *int32 {
	if r.RoleGroupInfo.ClusterOperation.Stopped {
		logger.Info("Cluster operation stopped, set replicas to 0")
		zero := int32(0)
		return &zero
	}
	return r.RoleGroupInfo.Replicas
}

func (r *StatefulSetReconciler[T]) Reconcile(ctx context.Context) Result {
	resource, err := r.GetBuilder().
		SetReplicas(r.getReplicas()).
		Build(ctx)

	if err != nil {
		return NewResult(true, 0, err)
	}
	return r.ResourceReconcile(ctx, resource)
}

func NewStatefulSetReconciler[T AnySpec](
	client client.ResourceClient,
	roleGroupInfo RoleGroupInfo,
	ports []corev1.ContainerPort,
	spec T,
	stsBuilder builder.StatefulSetBuilder,
) *StatefulSetReconciler[T] {
	return &StatefulSetReconciler[T]{
		BaseResourceReconciler: *NewBaseResourceReconciler[T, builder.StatefulSetBuilder](
			client,
			roleGroupInfo.GetFullName(),
			spec,
			stsBuilder,
		),
		RoleGroupInfo: roleGroupInfo,
		Ports:         ports,
	}
}
