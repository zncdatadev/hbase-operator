package util

import "github.com/zncdatadev/operator-go/pkg/reconciler"

func CreateServiceMetricsName(roleGroupInfo *reconciler.RoleGroupInfo) string {
	return roleGroupInfo.GetFullName() + "-metrics"
}
