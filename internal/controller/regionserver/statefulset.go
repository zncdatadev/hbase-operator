package regionserver

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
	spec *hbasev1alph1.RegionServerRoleGroupSpec,
) *common.StatefulSetReconciler[*hbasev1alph1.RegionServerRoleGroupSpec] {
	statefulSetBuilder := common.NewStatefulSetBuilder(
		client,
		roleGroupInfo.GetFullName(),
		clusterConfig,
		ports,
		roleGroupInfo.Image,
		spec.EnvOverrides,
		spec.CommandOverrides,
	)
	// add affinity
	var specAffinity *corev1.Affinity
	if conf := spec.Config; conf != nil {
		specAffinity = conf.Affinity
	}
	var hdfsInstanceName string
	if clusterConfig != nil {
		hdfsInstanceName = clusterConfig.HdfsConfigMap
	}
	addAffinityToStatefulSetBuilder(statefulSetBuilder, specAffinity, roleGroupInfo.ClusterInfo.Name, hdfsInstanceName)

	return common.NewStatefulSetReconciler(
		client,
		roleGroupInfo,
		clusterConfig,
		ports,
		spec,
		statefulSetBuilder,
	)
}

func addAffinityToStatefulSetBuilder(statefulSetBuilder *common.StatefulSetBuilder, specAffinity *corev1.Affinity,
	instanceName string, hdfsInstanceName string) {
	affinityLabels := metav1.LabelSelector{
		MatchLabels: map[string]string{
			reconciler.LabelInstance: instanceName,
			reconciler.LabelServer:   "hbase",
		},
	}
	// affinity with hdfs's datanode
	hdfsDatanodeLabels := metav1.LabelSelector{
		MatchLabels: map[string]string{
			reconciler.LabelInstance:  hdfsInstanceName,
			reconciler.LabelServer:    "hdfs",
			reconciler.LabelComponent: "datanode",
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
		builder.NewPodAffinity(builder.StrengthPrefer, false, hdfsDatanodeLabels).Weight(50),
		builder.NewPodAffinity(builder.StrengthPrefer, true, antiAffinityLabels).Weight(70),
	}}

	statefulSetBuilder.Affinity(specAffinity, defaultAffinityBuilder.Build())
}
