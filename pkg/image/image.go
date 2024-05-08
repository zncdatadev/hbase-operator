package image

import (
	corev1 "k8s.io/api/core/v1"
)

type Image struct {
	Repository        string            `json:"repository,omitempty"`
	Tag               string            `json:"tag,omitempty"`
	ProductionVersion string            `json:"productionVersion,omitempty"`
	PullPolicy        corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

func (i *Image) GetImageTag() string {
	return i.Repository + ":" + i.Tag
}

func (i *Image) GetPullPolicy() corev1.PullPolicy {
	return i.PullPolicy
}
