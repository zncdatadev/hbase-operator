package common

import (
	"fmt"
	"strings"

	"github.com/zncdata-labs/hbase-operator/pkg/builder"
	"github.com/zncdata-labs/hbase-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

type ContainerBuilder struct {
	builder.GenericContainerBuilder
}

func NewContainerBuilder(
	name string,
	image string,
	pullPolicy corev1.PullPolicy,
) *ContainerBuilder {
	return &ContainerBuilder{
		GenericContainerBuilder: *builder.NewGenericContainerBuilder(
			name,
			image,
			pullPolicy,
		),
	}
}

func (b *ContainerBuilder) Build() *corev1.Container {
	b.SetCommand(b.getCommand())
	b.AddEnvVars(b.getEnvVars())

	return b.GenericContainerBuilder.Build()
}

func (b *ContainerBuilder) getRoleCommandArg() string {
	s := "["
	for k, v := range roleCommandArgMapping {
		s += fmt.Sprintf(`"%s": "%s", `, k, v)
		if strings.Contains(b.Name, k) {
			return v
		}
	}
	s += "]"
	panic("Unknown name: " + b.Name + ", cannot determine role command. Supported: " + s)
}

func (b *ContainerBuilder) getCommand() []string {
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

prepare_signal_handlers
bin/hbase ` + b.getRoleCommandArg() + ` start &
wait_for_termination $!
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

func (b *ContainerBuilder) getEnvVars() []corev1.EnvVar {
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
