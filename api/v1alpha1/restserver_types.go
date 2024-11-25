package v1alpha1

import (
	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
)

type RestServerSpec struct {
	// +kubebuilder:validation:Optional
	Config *RestServerConfigSpec `json:"config,omitempty"`

	// +kubebuilder:validation:Required
	RoleGroups map[string]RestServerRoleGroupSpec `json:"roleGroups,omitempty"`

	// +kubebuilder:validation:Optional
	RoleConfig *commonsv1alpha1.RoleConfigSpec `json:"roleConfig,omitempty"`

	*commonsv1alpha1.OverridesSpec `json:",inline"`
}

type RestServerConfigSpec struct {
	*commonsv1alpha1.RoleGroupConfigSpec `json:",inline"`
}

type RestServerRoleGroupSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	Config *RestServerConfigSpec `json:"config,omitempty"`

	*commonsv1alpha1.OverridesSpec `json:",inline"`
}
