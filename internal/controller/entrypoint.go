package controller

import (
	"fmt"
	"path"
	"strings"

	hbasev1alpha1 "github.com/zncdatadev/hbase-operator/api/v1alpha1"
)

// roleNameToCommandArg maps a role to the `bin/hbase <arg> start` subcommand.
var roleNameToCommandArg = map[string]string{
	hbasev1alpha1.MasterRole:       hbasev1alpha1.MasterRole,
	hbasev1alpha1.RegionServerRole: hbasev1alpha1.RegionServerRole,
	hbasev1alpha1.RestServerRole:   "rest",
}

// mainContainerCommand is the shell the entrypoint script runs under.
var mainContainerCommand = []string{"/bin/bash", "-x", "-euo", "pipefail", "-c"}

// mainContainerArgs renders the entrypoint script of the primary container: copy the mounted
// config into the effective config dir, resolve the Kerberos realm (when enabled), then run the
// role's HBase process under a TERM-forwarding wrapper so a pod deletion shuts HBase down
// gracefully. The script is byte-identical to the pre-Gen3 operator's.
func mainContainerArgs(roleName string, krb5Config *HbaseKerberosConfig) []string {
	hbaseSubArg := roleNameToCommandArg[roleName]

	setupKrb5 := ""
	if krb5Config != nil {
		setupKrb5 = krb5Config.GetContainerCommands()
	}

	arg := `mkdir -p ` + HbaseConfigDir + `
cp ` + path.Join(HbaseMountConfigDir, "*") + ` ` + HbaseConfigDir + `
cp ` + path.Join(HdfsMountConfigDir, "*") + ` ` + HbaseConfigDir + `

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
	return []string{indentTab4Spaces(arg)}
}

// hbaseEnvSh renders hbase-env.sh: HBASE_MANAGES_ZK off (ZooKeeper is external) plus the
// role-scoped JVM options (Kerberos krb5.conf when enabled). HBase's hbase-env.sh contract names
// these variables HBASE_<ROLE>_OPTS with an uppercase role; a lowercase role is silently ignored.
func hbaseEnvSh(roleName string, krb5Config *HbaseKerberosConfig) string {
	var opts []string
	if krb5Config != nil {
		for k, v := range krb5Config.GetJVMOPTS() {
			opts = append(opts, "-D"+k+"="+v)
		}
	}
	jvmOpts := fmt.Sprintf(
		`export HBASE_%s_OPTS="$HBASE_OPTS %s"`,
		strings.ToUpper(roleName),
		strings.Join(opts, " "),
	)

	hbaseEnv := fmt.Sprintf(`
export HBASE_MANAGES_ZK=false
%s

`,
		jvmOpts,
	)
	return indentTab4Spaces(hbaseEnv)
}
