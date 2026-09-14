package controller

import (
	"path"

	"github.com/zncdatadev/operator-go/pkg/constant"
)

const (
	// HbaseSiteFileName is the main HBase configuration file rendered into the role group
	// ConfigMap (Hadoop-style XML).
	HbaseSiteFileName = "hbase-site.xml"
	// HbaseEnvFileName is the HBase environment script. It is a whole shell file, not a
	// key-value config, so it bypasses the merge pipeline and is written by the handler.
	HbaseEnvFileName = "hbase-env.sh"
	// SSLClientFileName / SSLServerFileName carry the TLS store settings when Kerberos+TLS
	// authentication is enabled.
	SSLClientFileName = "ssl-client.xml"
	SSLServerFileName = "ssl-server.xml"
)

var (
	// HbaseConfigDir is the effective config dir the processes read (the entrypoint copies the
	// mounted ConfigMaps here): /kubedoop/config
	HbaseConfigDir = path.Join(constant.KubedoopConfigDir)
	// HbaseMountConfigDir is where the role group ConfigMap is mounted: /kubedoop/mount/config/hbase
	HbaseMountConfigDir = path.Join(constant.KubedoopConfigDirMount, "hbase")
	// HdfsMountConfigDir is where the HDFS discovery ConfigMap is mounted: /kubedoop/mount/config/hdfs
	HdfsMountConfigDir = path.Join(constant.KubedoopConfigDirMount, "hdfs")
)

// ConsoleConversionPattern is the log4j console layout pattern HBase ships with.
const ConsoleConversionPattern = "%d{ISO8601} %-5p [%t] %c{2}: %.1000m%n"

// valueTrue is the canonical string form of a boolean in a Kubernetes label/annotation or a
// Hadoop-style config value, where everything is a string.
const valueTrue = "true"

// Volume names of the secret-operator CSI volumes backing Kerberos authentication.
const (
	KerberosVolumeName = "kerberos"
	TLSVolumeName      = "tls"
)

// RestHTTPPortName is the restserver's client port name; it is also its probe target.
const RestHTTPPortName = "rest-http"

// HbaseRootDirKey is the hbase-site.xml key naming the HDFS directory HBase stores data under.
const HbaseRootDirKey = "hbase.rootdir"

// HbaseZookeeperQuorumKey is the hbase-site.xml key carrying the ZooKeeper quorum, derived from
// the ZookeeperZnode discovery ConfigMap.
const HbaseZookeeperQuorumKey = "hbase.zookeeper.quorum"
