package common

import (
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	ZookeeperHostsKey      = "ZOOKEEPER_HOSTS"
	ZookeeperClientPortKey = "ZOOKEEPER_CLIENT_PORT"
	ZookeeperChrootKey     = "ZOOKEEPER_CHROOT"
	ZookeeperKey           = "zookeeper"
)

type ZnodeConfiguration struct {
	ConfigMap *corev1.ConfigMap
}

// zookeeper example:
// simple-zk-server-primary-0.simple-zk-server-primary.default.svc.cluster.local:2181/znode-0f8ba74a-7fb1-4d0a-81a5-7259a56defda
func (c *ZnodeConfiguration) AddData(zookeeper string) error {
	parts := strings.Split(zookeeper, "/")
	if len(parts) != 3 {
		return errors.New("invalid zookeeper string")
	}

	hostAndPort := strings.Split(parts[0], ":")
	if len(hostAndPort) != 2 {
		return errors.New("invalid host and port")
	}

	c.ConfigMap.Data[ZookeeperHostsKey] = hostAndPort[0]
	c.ConfigMap.Data[ZookeeperClientPortKey] = hostAndPort[1]
	c.ConfigMap.Data[ZookeeperChrootKey] = parts[2]
	c.ConfigMap.Data[ZookeeperKey] = zookeeper
	return nil
}

func (c *ZnodeConfiguration) GetQuorum() (string, error) {
	value, ok := c.ConfigMap.Data[ZookeeperHostsKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap", ZookeeperHostsKey)
	}
	return value, nil
}

func (c *ZnodeConfiguration) GetClientPort() (string, error) {
	value, ok := c.ConfigMap.Data[ZookeeperClientPortKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap", ZookeeperClientPortKey)
	}
	return value, nil
}

func (c *ZnodeConfiguration) GetChroot() (string, error) {
	value, ok := c.ConfigMap.Data[ZookeeperChrootKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap", ZookeeperChrootKey)
	}
	return value, nil
}

func (c *ZnodeConfiguration) GetZookeeper() (string, error) {
	value, ok := c.ConfigMap.Data[ZookeeperKey]
	if !ok {
		return "", fmt.Errorf("key %s not found in configmap", ZookeeperKey)
	}
	return value, nil
}
