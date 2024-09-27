package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	DefaultRepository      = "quay.io/zncdatadev"
	DefaultProductVersion  = "2.4.18"
	DefaultKubedoopVersion = "0.0.0-dev"
	DefaultProductName     = "hbase"
)

type ImageSpec struct {
	// +kubebuilder:validation:Optional
	Custom string `json:"custom,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=quay.io/zncdatadev
	Repository string `json:"repository,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="0.0.0-dev"
	KubedoopVersion string `json:"kubedoopVersion,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="2.4.18"
	ProductVersion string `json:"productVersion,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=IfNotPresent
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default="hbase"
	PullSecretName string `json:"pullSecretName,omitempty"`
}
