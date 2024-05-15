package client

import (
	"context"
	"fmt"

	"github.com/cisco-open/k8s-objectmatcher/patch"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	clientLogger = ctrl.Log.WithName("resourceClient")
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

func (c *ResourceClient) CreateOrUpdate(ctx context.Context, obj client.Object) (mutation bool, err error) {

	key := client.ObjectKeyFromObject(obj)
	namespace := obj.GetNamespace()
	kinds, _, _ := c.Scheme().ObjectKinds(obj)
	name := obj.GetName()

	clientLogger.V(5).Info("Creating or updating object", "Kind", kinds, "Namespace", namespace, "Name", name)

	current := obj.DeepCopyObject().(client.Object)
	// Check if the object exists, if not create a new one
	err = c.Get(ctx, key, current)
	var calculateOpt = []patch.CalculateOption{
		patch.IgnoreStatusFields(),
	}
	if errors.IsNotFound(err) {
		if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(obj); err != nil {
			return false, err
		}
		clientLogger.Info("Creating a new object", "Kind", kinds, "Namespace", namespace, "Name", name)

		if err := c.Create(ctx, obj); err != nil {
			return false, err
		}
		return true, nil
	} else if err == nil {
		switch obj.(type) {
		case *corev1.Service:
			currentSvc := current.(*corev1.Service)
			svc := obj.(*corev1.Service)
			// Preserve the ClusterIP when updating the service
			svc.Spec.ClusterIP = currentSvc.Spec.ClusterIP
			// Preserve the annotation when updating the service, ensure any updated annotation is preserved
			//for key, value := range currentSvc.Annotations {
			//	if _, present := svc.Annotations[key]; !present {
			//		svc.Annotations[key] = value
			//	}
			//}

			if svc.Spec.Type == corev1.ServiceTypeNodePort || svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				for i := range svc.Spec.Ports {
					svc.Spec.Ports[i].NodePort = currentSvc.Spec.Ports[i].NodePort
				}
			}
		case *appsv1.StatefulSet:
			calculateOpt = append(calculateOpt, patch.IgnoreVolumeClaimTemplateTypeMetaAndStatus())
		}
		result, err := patch.DefaultPatchMaker.Calculate(current, obj, calculateOpt...)
		if err != nil {
			clientLogger.Error(err, "failed to calculate patch to match objects, moving on to update")
			// if there is an error with matching, we still want to update
			resourceVersion := current.(metav1.ObjectMetaAccessor).GetObjectMeta().GetResourceVersion()
			obj.(metav1.ObjectMetaAccessor).GetObjectMeta().SetResourceVersion(resourceVersion)

			if err := c.Update(ctx, obj); err != nil {
				return false, err
			}
			return true, nil
		}

		if !result.IsEmpty() {
			clientLogger.Info(
				fmt.Sprintf("Resource update for object %s:%s", kinds, obj.(metav1.ObjectMetaAccessor).GetObjectMeta().GetName()),
				"patch", string(result.Patch),
			)

			if err := patch.DefaultAnnotator.SetLastAppliedAnnotation(obj); err != nil {
				clientLogger.Error(err, "failed to annotate modified object", "object", obj)
			}

			resourceVersion := current.(metav1.ObjectMetaAccessor).GetObjectMeta().GetResourceVersion()
			obj.(metav1.ObjectMetaAccessor).GetObjectMeta().SetResourceVersion(resourceVersion)

			if err = c.Update(ctx, obj); err != nil {
				return false, err
			}
			return true, nil
		}
		clientLogger.V(1).Info(fmt.Sprintf("Skipping update for object %s:%s", kinds, obj.(metav1.ObjectMetaAccessor).GetObjectMeta().GetName()))

	}
	return false, err
}
