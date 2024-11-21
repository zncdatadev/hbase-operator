package v1alpha1

import (
	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
)

type MasterSpec struct {
	// +kubebuilder:validation:Optional
	Config *MasterConfigSpec `json:"config,omitempty"`

	// +kubebuilder:validation:Required
	RoleGroups map[string]MasterRoleGroupSpec `json:"roleGroups,omitempty"`

	// +kubebuilder:validation:Optional
	RoleConfig *commonsv1alpha1.RoleConfigSpec `json:"roleConfig,omitempty"`

	*commonsv1alpha1.OverridesSpec `json:",inline"`
}

type MasterConfigSpec struct {
	*commonsv1alpha1.RoleGroupConfigSpec `json:",inline"`
}

type MasterRoleGroupSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	Config *MasterConfigSpec `json:"config,omitempty"`

	*commonsv1alpha1.OverridesSpec `json:",inline"`
}
