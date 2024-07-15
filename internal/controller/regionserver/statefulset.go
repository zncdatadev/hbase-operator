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
		builder.RoleGroupInfo{
			ClusterName:   roleGroupInfo.GetClusterName(),
			RoleName:      roleGroupInfo.GetRoleName(),
			RoleGroupName: roleGroupInfo.GetGroupName(),
		},
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
