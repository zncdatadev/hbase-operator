package restserver

import (
	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"github.com/zncdata-labs/hbase-operator/pkg/image"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ reconciler.Reconciler = &StatefulSetReconciler{}

type StatefulSetReconciler struct {
	reconciler.StatefulSetReconciler[*hbasev1alph1.RestServerRoleGroupSpec]
}

type StatefulSetBuilder struct {
	builder.BaseStatefulSetBuilder
}

func NewStatefulSetBuilder(
	client client.Client,
	name string,
	namespace string,
	labels map[string]string,
	annotations map[string]string,
	envOverrides map[string]string,
	commandOverrides []string,
	replicas *int32,
	image image.Image,
	serviceName string,
) *StatefulSetBuilder {
	return &StatefulSetBuilder{
		BaseStatefulSetBuilder: *builder.NewBaseStatefulSetBuilder(
			client,
			name,
			namespace,
			labels,
			annotations,
			envOverrides,
			commandOverrides,
			replicas,
			image,
			serviceName,
		),
	}
}

func NewStatefulSetReconciler(
	client client.Client,
	schema *runtime.Scheme,

	roleGroupName string,
	labels map[string]string,
	annotations map[string]string,

	ownerReference client.Object,

	image image.Image,

	spec *hbasev1alph1.RestServerRoleGroupSpec,

) *StatefulSetReconciler {
	b := NewStatefulSetBuilder(
		client,
		roleGroupName,
		ownerReference.GetNamespace(),
		labels,
		annotations,
		spec.EnvOverrides,
		spec.CommandOverrides,
		spec.Replicas,
		image,
		roleGroupName,
	)

	return &StatefulSetReconciler{
		StatefulSetReconciler: *reconciler.NewStatefulSetReconciler(
			client,
			schema,
			roleGroupName,
			ownerReference,
			b,
			spec,
		),
	}
}
