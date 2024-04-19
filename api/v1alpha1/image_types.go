package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type ImageSpec struct {
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=bitnami/kafka
	Repository string `json:"repository,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="3.7.0-debian-12-r2"
	Tag string `json:"tag,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}
