package master

import (
	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceReconciler struct {
	reconciler.BaseResourceReconciler[*hbasev1alph1.MasterRoleGroupSpec]
}

func NewServiceReconciler(
	client client.Client,
	schema *runtime.Scheme,
	roleGroupName string,
	labels map[string]string,
	annotations map[string]string,
	ownerReference client.Object,
	spec *hbasev1alph1.MasterRoleGroupSpec,
) *ServiceReconciler {

	ports := []corev1.ServicePort{
		{
			Name:       "thrift",
			Port:       9090,
			Protocol:   corev1.ProtocolTCP,
			TargetPort: intstr.FromInt(9090),
		},
	}
	serviceBuilder := builder.NewServiceBuilder(
		client,
		roleGroupName,
		labels,
		annotations,
		ports,
		corev1.ServiceTypeClusterIP,
	)

	return &ServiceReconciler{
		BaseResourceReconciler: *reconciler.NewBaseResourceReconciler(
			client,
			schema,
			ownerReference,
			roleGroupName,
			spec,
			serviceBuilder,
		),
	}
}
