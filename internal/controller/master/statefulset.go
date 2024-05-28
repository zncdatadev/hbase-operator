package master

import (
	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/common"
	"github.com/zncdatadev/hbase-operator/pkg/builder"
	"github.com/zncdatadev/hbase-operator/pkg/client"
	"github.com/zncdatadev/hbase-operator/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStatefulSetReconciler(
	client client.ResourceClient,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleGroupInfo reconciler.RoleGroupInfo,
	ports []corev1.ContainerPort,
	spec *hbasev1alph1.MasterRoleGroupSpec,
) *common.StatefulSetReconciler[*hbasev1alph1.MasterRoleGroupSpec] {
	stsBuilder := common.NewStatefulSetBuilder(
		client,
		roleGroupInfo.GetFullName(),
		clusterConfig,
		ports,
		roleGroupInfo.Image,
		spec.EnvOverrides,
		spec.CommandOverrides,
	)

	var specAffinity *corev1.Affinity
	if conf := spec.Config; conf != nil {
		specAffinity = conf.Affinity
	}
	addAffinityToStatefulSetBuilder(stsBuilder, specAffinity, roleGroupInfo.ClusterInfo.Name)

	return common.NewStatefulSetReconciler(
		client,
		roleGroupInfo,
		clusterConfig,
		ports,
		spec,
		stsBuilder,
	)
}

func addAffinityToStatefulSetBuilder(statefulSetBuilder *common.StatefulSetBuilder, specAffinity *corev1.Affinity,
	instanceName string) {
	affinityLabels := metav1.LabelSelector{
		MatchLabels: map[string]string{
			reconciler.LabelInstance: instanceName,
			reconciler.LabelServer:   "hbase",
		},
	}
	antiAffinityLabels := metav1.LabelSelector{
		MatchLabels: map[string]string{
			reconciler.LabelInstance:  instanceName,
			reconciler.LabelServer:    "hbase",
			reconciler.LabelComponent: roleName(),
		},
	}
	defaultAffinityBuilder := builder.AffinityBuilder{PodAffinity: []*builder.PodAffinity{
		builder.NewPodAffinity(builder.StrengthPrefer, false, affinityLabels).Weight(20),
		builder.NewPodAffinity(builder.StrengthPrefer, true, antiAffinityLabels).Weight(70),
	}}

	statefulSetBuilder.Affinity(specAffinity, defaultAffinityBuilder.Build())
}
