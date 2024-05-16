package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type ImageSpec struct {
	// +kubebuilder:validation:Optional
	
	Custom string `json:"custom,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=docker.stackable.tech/stackable/hbase
	Repo string `json:"repo,omitempty"`

	// +kubebuilder:validation:Optional
	KDSVersion string `json:"kdsVersion,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="2.4.17"
	ProductVersion string `json:"productVersion,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`
}
