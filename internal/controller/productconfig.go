package controller

import (
	"context"
	"fmt"

	"github.com/zncdatadev/operator-go/pkg/reconciler"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

var _ reconciler.RoleGroupResolver[*hbasev1alpha1.HbaseCluster] = &HbaseRoleGroupHandler{}

// ResolveRoleGroup implements reconciler.RoleGroupResolver: it derives HBase's own configuration
// from the CR and from the live cluster, and returns it as the LOWEST merge layer — beneath the
// CR's role and role group overrides, so anything a user states always wins.
//
// The zookeeper connection keys are the reason this seam has a client: they come from the
// ZookeeperZnode discovery ConfigMap the zookeeper-operator publishes, which is live state no
// pure function of the CR could reach.
func (h *HbaseRoleGroupHandler) ResolveRoleGroup(
	ctx context.Context,
	k8sClient client.Client,
	cr *hbasev1alpha1.HbaseCluster,
	rg *reconciler.RoleGroupBuildContext,
) (*reconciler.Contribution, error) {
	znode, err := resolveZnode(ctx, k8sClient, cr)
	if err != nil {
		return nil, err
	}

	hbaseSite := map[string]string{
		"hbase.cluster.distributed": valueTrue,
		HbaseRootDirKey:             "/hbase",
		"hbase.unsafe.regionserver.hostname.disable.master.reversedns": valueTrue,

		HbaseZookeeperQuorumKey:               znode.quorum,
		"zookeeper.znode.parent":              znode.chroot + "/hbase",
		"hbase.zookeeper.property.clientPort": znode.clientPort,
	}

	envs := map[string]string{
		// The entrypoint copies the mounted ConfigMaps into HbaseConfigDir; both HBase and the
		// Hadoop client libraries read their config from there.
		"HBASE_CONF_DIR":  HbaseConfigDir,
		"HADOOP_CONF_DIR": HbaseConfigDir,
	}

	contribution := &reconciler.Contribution{
		ConfigOverrides: map[string]map[string]string{
			HbaseSiteFileName: hbaseSite,
		},
		EnvVars: envs,
	}

	if krb := KerberosConfigFor(cr); krb != nil {
		for k, v := range krb.GetHbaseSite() {
			hbaseSite[k] = v
		}
		contribution.ConfigOverrides[SSLClientFileName] = krb.GetSSLClientSettings()
		contribution.ConfigOverrides[SSLServerFileName] = krb.GetSSLServerSettings()
		for k, v := range krb.GetEnvOverrides() {
			envs[k] = v
		}
	}

	return contribution, nil
}

// znodeConnection is the connection info a ZookeeperZnode discovery ConfigMap carries.
type znodeConnection struct {
	quorum     string
	chroot     string
	clientPort string
}

// resolveZnode reads the znode discovery ConfigMap the CR references.
func resolveZnode(ctx context.Context, k8sClient client.Client, cr *hbasev1alpha1.HbaseCluster) (*znodeConnection, error) {
	clusterConfig := cr.Spec.ClusterConfigSpec
	if clusterConfig == nil || clusterConfig.ZookeeperConfigMapName == "" {
		return nil, fmt.Errorf("spec.clusterConfig.zookeeperConfigMapName is required")
	}

	cm := &corev1.ConfigMap{}
	key := types.NamespacedName{Namespace: cr.Namespace, Name: clusterConfig.ZookeeperConfigMapName}
	if err := k8sClient.Get(ctx, key, cm); err != nil {
		return nil, fmt.Errorf("failed to get zookeeper discovery ConfigMap %q: %w", clusterConfig.ZookeeperConfigMapName, err)
	}

	znode := &ZnodeConfiguration{ConfigMap: cm}
	quorum, err := znode.GetQuorum()
	if err != nil {
		return nil, err
	}
	chroot, err := znode.GetChroot()
	if err != nil {
		return nil, err
	}
	clientPort, err := znode.GetClientPort()
	if err != nil {
		return nil, err
	}

	return &znodeConnection{quorum: quorum, chroot: chroot, clientPort: clientPort}, nil
}
