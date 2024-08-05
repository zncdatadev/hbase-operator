package common

import (
	"context"

	"github.com/zncdatadev/operator-go/pkg/builder"
	"github.com/zncdatadev/operator-go/pkg/client"
	"github.com/zncdatadev/operator-go/pkg/reconciler"
	"github.com/zncdatadev/operator-go/pkg/util"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	hbasev1alph1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
	auzhz "github.com/zncdatadev/hbase-operator/internal/controller/authz"
)

var (
	roleNameToCommandArg = map[string]string{
		"master":       "master",
		"regionserver": "regionserver",
		"restserver":   "rest",
	}
)

var _ builder.StatefulSetBuilder = &StatefulSetBuilder{}

type StatefulSetBuilder struct {
	builder.StatefulSet
	Ports         []corev1.ContainerPort
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
	RoleName      string
	ClusterName   string

	krb5Config *auzhz.HbaseKerberosConfig
}

func NewStatefulSetBuilder(
	client *client.Client,
	name string,
	clusterConfig *hbasev1alph1.ClusterConfigSpec,
	replicas *int32,
	ports []corev1.ContainerPort,
	image *util.Image,
	krb5SecretClass string,
	tlsSecretClass string,
	options builder.WorkloadOptions,
) *StatefulSetBuilder {

	var krb5Config *auzhz.HbaseKerberosConfig
	if krb5SecretClass != "" && tlsSecretClass != "" {
		krb5Config = auzhz.NewHbaseKerberosConfig(
			client.GetOwnerNamespace(),
			options.ClusterName,
			options.RoleName,
			options.RoleGroupName,
			krb5SecretClass,
			tlsSecretClass,
		)
	}

	return &StatefulSetBuilder{
		StatefulSet:   *builder.NewStatefulSetBuilder(client, name, replicas, image, options),
		Ports:         ports,
		ClusterConfig: clusterConfig,
		RoleName:      options.RoleName,
		ClusterName:   options.ClusterName,
		krb5Config:    krb5Config,
	}

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
						Name: b.ClusterConfig.HdfsConfigMapName,
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

func (b *StatefulSetBuilder) GetMainContainerCommanArgs() []string {
	hbaseSubArg := roleNameToCommandArg[b.RoleName]

	setupKrb5 := ""
	if b.krb5Config != nil {
		setupKrb5 = b.krb5Config.GetContainerCommands()
	}

	arg := `mkdir -p /stackable/conf
cp /stackable/tmp/hdfs/* /stackable/conf
cp /stackable/tmp/hbase/* /stackable/conf

` + setupKrb5 + `

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

prepare_signal_handlers
bin/hbase ` + hbaseSubArg + ` start &
wait_for_termination $!
`
	intendedString := util.IndentTab4Spaces(arg)
	return []string{intendedString}
}

func (b *StatefulSetBuilder) getEnvVars() []corev1.EnvVar {
	objs := []corev1.EnvVar{
		{
			Name:  "HBASE_CONF_DIR",
			Value: "/stackable/conf",
		},
		{
			Name:  "HADOOP_CONF_DIR",
			Value: "/stackable/conf",
		},
	}

	return objs
}

func (b *StatefulSetBuilder) buildContainer() []corev1.Container {
	containers := []corev1.Container{}
	image := b.GetImageWithTag()
	mainContainerBuilder := builder.NewContainer(b.RoleName, image).
		SetCommand([]string{"/bin/bash", "-x", "-euo", "pipefail", "-c"}).
		SetArgs(b.GetMainContainerCommanArgs()).
		AddVolumeMounts(b.getVolumeMounts()).
		AddEnvVars(b.getEnvVars()).
		AddPorts(b.Ports)

	if b.krb5Config != nil {
		mainContainerBuilder.AddEnvVars(b.krb5Config.GetContainerEnvvars())
		mainContainerBuilder.AddVolumeMounts(b.krb5Config.GetVolumeMounts())
	}

	// TODO: add vector container

	containers = append(containers, *mainContainerBuilder.Build())

	return containers
}

func (b *StatefulSetBuilder) GetDefaultAffinityBuilder() *AffinityBuilder {
	affinityLabels := map[string]string{
		util.AppKubernetesInstanceName: b.ClusterName,
		util.AppKubernetesNameName:     "hbase",
	}
	antiAffinityLabels := map[string]string{
		util.AppKubernetesInstanceName:  b.ClusterName,
		util.AppKubernetesNameName:      "hbase",
		util.AppKubernetesComponentName: b.RoleName,
	}

	affinity := NewAffinityBuilder(
		*NewPodAffinity(affinityLabels, false, false).Weight(20),
		*NewPodAffinity(antiAffinityLabels, false, true).Weight(70),
	)

	return affinity
}

func (b *StatefulSetBuilder) Build(ctx context.Context) (ctrlclient.Object, error) {
	b.AddContainers(b.buildContainer())
	b.AddVolumes(b.getVolumes())
	b.SetAffinity(b.GetDefaultAffinityBuilder().Build())

	if b.krb5Config != nil {
		b.AddVolumes(b.krb5Config.GetVolumes())
	}
	return b.GetObject()
}

var _ reconciler.Reconciler = &StatefulSetReconciler{}

type StatefulSetReconciler struct {
	reconciler.StatefulSet
	ClusterConfig *hbasev1alph1.ClusterConfigSpec
}
