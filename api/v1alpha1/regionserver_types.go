package v1alpha1

import (
	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

type RegionServerSpec struct {
	// +kubebuilder:validation:Optional
	Config *RegionConfigSpec `json:"config,omitempty"`

	// +kubebuilder:validation:Optional
	RoleGroups map[string]RegionServerRoleGroupSpec `json:"roleGroups,omitempty"`

	PodDisruptionBudget *commonsv1alpha1.PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// +kubebuilder:validation:Optional
	CommandOverrides []string `json:"commandOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	ConfigOverrides *RegionConfigOverrideSpec `json:"configOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	EnvOverrides map[string]string `json:"envOverrides,omitempty"`

	// // +kubebuilder:validation:Optional
	// PodOverride *corev1.PodTemplateSpec `json:"podOverride,omitempty"`
}

type RegionConfigSpec struct {
	// +kubebuilder:validation:Optional
	Affinity *corev1.Affinity `json:"affinity"`

	// +kubebuilder:validation:Optional
	Tolerations []corev1.Toleration `json:"tolerations"`

	// +kubebuilder:validation:Optional
	PodDisruptionBudget *commonsv1alpha1.PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// Use time.ParseDuration to parse the string
	// +kubebuilder:validation:Optional
	GracefulShutdownTimeout *string `json:"gracefulShutdownTimeout,omitempty"`

	// +kubebuilder:validation:Optional
	Logging *LoggingSpec `json:"logging,omitempty"`

	// +kubebuilder:validation:Optional
	Resources *commonsv1alpha1.ResourcesSpec `json:"resources,omitempty"`
}

type RegionConfigOverrideSpec struct{}

type RegionServerRoleGroupSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=1
	Replicas *int32 `json:"replicas,omitempty"`

	// +kubebuilder:validation:Optional
	Config *RegionConfigSpec `json:"config,omitempty"`

	// +kubebuilder:validation:Optional
	PodDisruptionBudget *commonsv1alpha1.PodDisruptionBudgetSpec `json:"podDisruptionBudget,omitempty"`

	// +kubebuilder:validation:Optional
	CommandOverrides []string `json:"commandOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	ConfigOverrides *RegionConfigOverrideSpec `json:"configOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	EnvOverrides map[string]string `json:"envOverrides,omitempty"`

	// +kubebuilder:validation:Optional
	PodOverrides *corev1.PodTemplateSpec `json:"podOverrides,omitempty"`
}
