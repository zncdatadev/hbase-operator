package common

import (
	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"k8s.io/utils/ptr"
)

func GetRoleLoggingConfig(
	roleGroupConfig *commonsv1alpha1.RoleGroupConfigSpec,
	role string,
) *commonsv1alpha1.LoggingConfigSpec {
	if roleGroupConfig != nil && roleGroupConfig.Logging != nil {
		contaienrsLogging := roleGroupConfig.Logging.Containers
		if contaienrsLogging != nil {
			if logging, ok := contaienrsLogging[role]; ok {
				return ptr.To(logging)
			}
		}
	}
	return nil
}
