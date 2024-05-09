package reconciler

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceClient struct {
	client.Client

	OwnerReference client.Object

	labels      map[string]string
	annotations map[string]string
}

func (c *ResourceClient) AddLabels(labels map[string]string, override bool) {

	if labels == nil {
		labels = c.OwnerReference.GetLabels()
	}

	for k, v := range c.labels {
		if _, ok := labels[k]; !ok || override {
			labels[k] = v
		}
	}

	c.labels = labels

}

func (c *ResourceClient) AddAnnotations(annotations map[string]string, override bool) {

	if annotations == nil {
		annotations = c.OwnerReference.GetAnnotations()
	}

	for k, v := range c.annotations {
		if _, ok := annotations[k]; !ok || override {
			annotations[k] = v
		}
	}

	c.annotations = annotations

}

func (c *ResourceClient) GetLabels() map[string]string {
	return c.labels
}

func (c *ResourceClient) GetAnnotations() map[string]string {
	return c.annotations
}

func (c *ResourceClient) GetOwnerReference() client.Object {
	return c.OwnerReference
}
func (c *ResourceClient) GetOwnerNamespace() string {
	return c.OwnerReference.GetNamespace()
}

func (c *ResourceClient) GetOwnerName() string {
	return c.OwnerReference.GetName()
}
