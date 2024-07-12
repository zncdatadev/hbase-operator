package regionserver

import (
	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/internal/controller/common"
	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

func NewStatefulSetReconciler(
	client *client.Client,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	roleGroupInfo reconciler.RoleGroupInfo,
	ports []corev1.ContainerPort,
	image *util.Image,
	spec *hbasev1alph1.RegionServerRoleGroupSpec,
) reconciler.ResourceReconciler[builder.StatefulSetBuilder] {
	stsBuilder := common.NewStatefulSetBuilder(
		client,
		roleGroupInfo.GetFullName(),
		clusterConfig,
		&builder.RoleGroupInfo{},
		spec.Replicas,
		ports,
		image,
	)

	affinityBuilder := stsBuilder.GetDefaultAffinityBuilder()

	hdfsDatanodeLabels := map[string]string{
		util.AppKubernetesInstanceName:  clusterConfig.HdfsConfigMapName,
		util.AppKubernetesNameName:      "hdfs",
		util.AppKubernetesComponentName: "datanode",
	}

	affinityBuilder.AddPodAffinity(
		*common.NewPodAffinity(hdfsDatanodeLabels, false, false).Weight(50),
	)

	stsBuilder.AddAffinity(affinityBuilder.Build())

	return reconciler.NewStatefulSet(
		client,
		roleGroupInfo.GetFullName(),
		stsBuilder,
	)
}

// func addAffinityToStatefulSetBuilder(statefulSetBuilder *common.StatefulSetBuilder, specAffinity *corev1.Affinity,
// 	instanceName string, hdfsInstanceName string) {
// 	affinityLabels := metav1.LabelSelector{
// 		MatchLabels: map[string]string{
// 			reconciler.LabelInstance: instanceName,
// 			reconciler.LabelServer:   "hbase",
// 		},
// 	}
// 	// affinity with hdfs's datanode
// 	hdfsDatanodeLabels := metav1.LabelSelector{
// 		MatchLabels: map[string]string{
// 			reconciler.LabelInstance:  hdfsInstanceName,
// 			reconciler.LabelServer:    "hdfs",
// 			reconciler.LabelComponent: "datanode",
// 		},
// 	}
// 	antiAffinityLabels := metav1.LabelSelector{
// 		MatchLabels: map[string]string{
// 			reconciler.LabelInstance:  instanceName,
// 			reconciler.LabelServer:    "hbase",
// 			reconciler.LabelComponent: roleName(),
// 		},
// 	}
// 	defaultAffinityBuilder := builder.AffinityBuilder{PodAffinity: []*builder.PodAffinity{
// 		builder.NewPodAffinity(builder.StrengthPrefer, false, affinityLabels).Weight(20),
// 		builder.NewPodAffinity(builder.StrengthPrefer, false, hdfsDatanodeLabels).Weight(50),
// 		builder.NewPodAffinity(builder.StrengthPrefer, true, antiAffinityLabels).Weight(70),
// 	}}

// 	statefulSetBuilder.Affinity(specAffinity, defaultAffinityBuilder.Build())
// }
