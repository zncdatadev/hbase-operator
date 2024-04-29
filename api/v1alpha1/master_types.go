package v1alpha1

import (
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type MasterSpec struct {
	// +kubebuilder:validation:Optional
	Config *MasterConfigSpec `json:"config,omitempty"`

	// +kubebuilder:validation:Required
	RoleGroups map[string]MasterRoleGroupSpec `json:"roleGroups,omitempty"`

	// +kubebuilder:validation:Required
	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// +kubebuilder:validation:Optional
	CommandArgsOverrides []string `json:"commandArgsOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	ConfigOverrides *MasterConfigOverrideSpec `json:"configOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	EnvOverrides map[string]string `json:"envOverrides,omitempty"`

	// // +kubebuilder:validation:Optional
	// PodOverride *corev1.PodTemplateSpec `json:"podOverride,omitempty"`
}

type MasterConfigSpec struct {
	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`

	// +kubebuilder:validation:Optional
	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// Use time.ParseDuration to parse the string
	// +kubebuilder:validation:Optional
	GracefulShutdownTimeout *string `json:"gracefulShutdownTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	Logging *ContainerLoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *apiv1alpha1.ResourcesSpec `json:"resources,omitempty"`
}

type MasterConfigOverrideSpec struct {
}

type MasterRoleGroupSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	Config *MasterConfigSpec `json:"config,omitempty"`

	// +kubebuilder:validation:Optional
	PodDisruptionBudget *PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// +kubebuilder:validation:Optional
	CommandArgsOverrides []string `json:"commandArgsOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	ConfigOverrides *MasterConfigOverrideSpec `json:"configOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	EnvOverrides map[string]string `json:"envOverrides,omitempty"`

	// // +kubebuilder:validation:Optional
	// PodOverride *corev1.PodTemplateSpec `json:"podOverride,omitempty"`
}
