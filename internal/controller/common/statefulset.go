package common

import (
	hbasev1alph1 "github.com/zncdata-labs/hbase-operator/api/v1alpha1"
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/image"
	"github.com/zncdata-labs/hbase-operator/pkg/reconciler"
	"github.com/zncdata-labs/hbase-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

var _ reconciler.Reconciler = &StatefulSetReconciler[reconciler.AnySpec]{}

type StatefulSetReconciler[T reconciler.AnySpec] struct {
	reconciler.StatefulSetReconciler[T]
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}

func NewStatefulSetReconciler[T reconciler.AnySpec](
	client reconciler.ResourceClient,

	roleGroupName string,

	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	ports []corev1.ContainerPort,
	image image.Image,

	spec T,

) *StatefulSetReconciler[T] {

	return &StatefulSetReconciler[T]{
		StatefulSetReconciler: *reconciler.NewStatefulSetReconciler[T](
			client,
			roleGroupName,
			image,
			ports,
			spec,
		),
		ClusterConfig: clusterConfig,
	}
}

func (r *StatefulSetReconciler[T]) GetContainers() []corev1.Container {
	objs := []corev1.Container{
		{
			Name:            r.Name,
			Image:           r.Image.GetImageTag(),
			ImagePullPolicy: r.Image.GetPullPolicy(),
			Env:             r.GetEnvVars(),
			Command:         r.GetCommand(),
			Resources:       r.convertResource(),
			VolumeMounts:    r.GetVolumeMounts(),
			Ports:           r.Ports,
		},
	}

	return objs
}

func (r *StatefulSetReconciler[T]) GetRoleCommandName() string {
	panic("unimplemented")
}

func (r *StatefulSetReconciler[T]) GetCommand() []string {
	args := `
mkdir -p /stackable/conf
cp /stackable/tmp/hdfs/* /stackable/conf
cp /stackable/tmp/hbase/* /stackable/conf


prepare_signal_handlers()
{
	unset term_child_pid
	unset term_kill_needed
	trap 'handle_term_signal' TERM
}

handle_term_signal()
{
	if [ "${term_child_pid}" ]; then
		kill -TERM "${term_child_pid}" 2>/dev/null
	else
		term_kill_needed="yes"
	fi
}

wait_for_termination()
{
	set +e
	term_child_pid=$1
	if [[ -v term_kill_needed ]]; then
		kill -TERM "${term_child_pid}" 2>/dev/null
	fi
	wait ${term_child_pid} 2>/dev/null
	trap - TERM
	wait ${term_child_pid} 2>/dev/null
	set -e
}

rm -f /stackable/log/_vector/shutdown
prepare_signal_handlers
bin/hbase master start &
wait_for_termination $!
mkdir -p /stackable/log/_vector && touch /stackable/log/_vector/shutdown
	`

	return []string{
		"/bin/bash",
		"-x",
		"-euo",
		"pipefail",
		"-c",
		util.IndentTab4Spaces(args),
	}
}

func (r *StatefulSetReconciler[T]) GetEnvVars() []corev1.EnvVar {
	objs := []corev1.EnvVar{
		{
			Name:  "HBASE_CONF_DIR",
			Value: "/stackable/conf",
		},
		{
			Name:  "HDFS_CONF_DIR",
			Value: "/stackable/conf",
		},
	}

	return objs
}

func (r *StatefulSetReconciler[T]) GetVolumes() []corev1.Volume {

	defaultMode := int32(420)

	objs := []corev1.Volume{
		{
			Name: "hdfs-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					DefaultMode: &defaultMode,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.ClusterConfig.HdfsConfigMap,
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
						Name: r.GetName(),
					},
				},
			},
		},
	}

	return objs
}

func (r *StatefulSetReconciler[T]) GetVolumeMounts() []corev1.VolumeMount {
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

func (r *StatefulSetReconciler[T]) GetVolumeClaimTemplate() *corev1.PersistentVolumeClaim {
	return nil
}
func (r *StatefulSetReconciler[T]) GetResourceRequirements() *apiv1alpha1.ResourcesSpec {
	panic("unimplemented")
}

func (r *StatefulSetReconciler[T]) convertResource() corev1.ResourceRequirements {
	resources := r.GetResourceRequirements()
	return corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    *resources.CPU.Max,
			corev1.ResourceMemory: *resources.Memory.Limit,
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    *resources.CPU.Min,
			corev1.ResourceMemory: *resources.Memory.Limit,
		},
	}
}
