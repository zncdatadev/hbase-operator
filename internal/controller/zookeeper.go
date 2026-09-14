package controller

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
)

const (
	ZookeeperHostsKey      = "ZOOKEEPER_HOSTS"
	ZookeeperClientPortKey = "ZOOKEEPER_PORT"
	ZookeeperChrootKey     = "ZOOKEEPER_CHROOT"
)

// ZnodeConfiguration reads the connection info a ZookeeperZnode discovery ConfigMap carries.
// The ConfigMap name comes from spec.clusterConfig.zookeeperConfigMapName; its keys are the
// zookeeper-operator's public discovery contract.
type ZnodeConfiguration struct {
	ConfigMap *corev1.ConfigMap
}

func (c *ZnodeConfiguration) GetQuorum() (string, error) {
	value, ok := c.ConfigMap.Data[ZookeeperHostsKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap %s", ZookeeperHostsKey, c.ConfigMap.Name)
	}
	return value, nil
}

func (c *ZnodeConfiguration) GetClientPort() (string, error) {
	value, ok := c.ConfigMap.Data[ZookeeperClientPortKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap %s", ZookeeperClientPortKey, c.ConfigMap.Name)
	}
	return value, nil
}

func (c *ZnodeConfiguration) GetChroot() (string, error) {
	value, ok := c.ConfigMap.Data[ZookeeperChrootKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap %s", ZookeeperChrootKey, c.ConfigMap.Name)
	}
	return value, nil
}
