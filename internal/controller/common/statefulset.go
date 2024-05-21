package common

import (
	"context"
	"fmt"
	"strings"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	"github.com/zncdatadev/hbase-operator/pkg/builder"
	"github.com/zncdatadev/hbase-operator/pkg/client"
	"github.com/zncdatadev/hbase-operator/pkg/reconciler"
	"github.com/zncdatadev/hbase-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	roleCommandArgMapping = map[string]string{
		"master": "master",
		"region": "region",
		"rest":   "rest",
	}
	containerNameMapping = map[string]string{
		"master": "master",
		"region": "region",
		"rest":   "rest",
	}
)

var _ builder.StatefulSetBuilder = &StatefulSetBuilder{}

type StatefulSetBuilder struct {
	builder.GenericStatefulSetBuilder
	Ports         []corev1.ContainerPort
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func NewStatefulSetBuilder(
	client client.ResourceClient,
	roleGroupName string,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	ports []corev1.ContainerPort,
	image util.Image,
	envOverrides map[string]string,
	commandOverrides []string,
) *StatefulSetBuilder {
	stsBuilder := builder.NewGenericStatefulSetBuilder(
		client,
		roleGroupName,
		envOverrides,
		commandOverrides,
		image,
	)
	return &StatefulSetBuilder{
		GenericStatefulSetBuilder: *stsBuilder,
		Ports:                     ports,
		ClusterConfig:             clusterConfig,
	}
}

func (b *StatefulSetBuilder) getContainerName() string {
	var name string
	s := "["
	for k, v := range containerNameMapping {
		s += fmt.Sprintf(`"%s": "%s", `, k, v)
		if strings.Contains(b.GetName(), k) {
			name = v
		}
	}
	s += "]"

	if name == "" {
		panic("Unknown name: " + b.GetName() + ", cannot determine container name. Supported: " + s)
	}
	return name
}

func (b *StatefulSetBuilder) getVolumeMounts() []corev1.VolumeMount {
	objs := []corev1.VolumeMount{
		{
			Name:      "hdfs-config",
			MountPath: "/stackable/tmp/hdfs",
		},
		{
			Name:      "hbase-config",
			MountPath: "/stackable/tmp/hbase",
		},
	}
	return objs
}

func (b *StatefulSetBuilder) getVolumes() []corev1.Volume {

	defaultMode := int32(420)

	objs := []corev1.Volume{
		{
			Name: "hdfs-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.ClusterConfig.HdfsConfigMap,
					},
				},
			},
		},
		{
			Name: "hbase-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: b.GetName(),
					},
				},
			},
		},
	}

	return objs
}

func (b *StatefulSetBuilder) buildContainer() *corev1.Container {
	containerBuilder := NewContainerBuilder(
		b.getContainerName(),
		b.Image.String(),
		b.Image.PullPolicy,
	)

	containerBuilder.AddVolumeMounts(b.getVolumeMounts())
	containerBuilder.AddPorts(b.Ports)
	containerBuilder.AutomaticSetProbe()
	return containerBuilder.Build()
}

func (b *StatefulSetBuilder) Build(ctx context.Context) (k8sclient.Object, error) {
	b.AddContainers([]corev1.Container{*b.buildContainer()})
	b.AddVolumes(b.getVolumes())
	return b.GetObject(), nil
}

var _ reconciler.Reconciler = &StatefulSetReconciler[reconciler.AnySpec]{}

type StatefulSetReconciler[T reconciler.AnySpec] struct {
	reconciler.StatefulSetReconciler[T]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func NewStatefulSetReconciler[T reconciler.AnySpec](
	client client.ResourceClient,
	roleGroupInfo reconciler.RoleGroupInfo,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	ports []corev1.ContainerPort,
	spec T,
	stsBuilder builder.StatefulSetBuilder,
) *StatefulSetReconciler[T] {

	return &StatefulSetReconciler[T]{
		StatefulSetReconciler: *reconciler.NewStatefulSetReconciler[T](
			client,
			roleGroupInfo,
			ports,
			spec,
			stsBuilder,
		),
		ClusterConfig: clusterConfig,
	}
}
