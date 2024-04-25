package builder

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatefulSetBuilder interface {
	ResourceBuilder[appv1.StatefulSet]
	Replicas() int32
	Selector() *metav1.LabelSelector
	BuildObjectMeta() metav1.ObjectMeta
	Labels() map[string]string
	Volumes() []corev1.Volume
}

type StatefulSet struct {
	Resource

	Containers []corev1.Container
}

func (s *StatefulSet) Build() *appv1.StatefulSet {
	var replicas = s.Replicas()
	obj := &appv1.StatefulSet{
		ObjectMeta: s.BuildObjectMeta(),
		Spec: appv1.StatefulSetSpec{
			Replicas: &replicas,
			Selector: s.Selector(),
			Template: s.BuildPodTemplateSpec(),
		},
	}

	return obj
}

func (s *StatefulSet) Labels() map[string]string {
	return map[string]string{
		"app": s.Name(),
	}
}

func (s *StatefulSet) Replicas() int32 {
	return 1
}

func (s *StatefulSet) Selector() *metav1.LabelSelector {
	return &metav1.LabelSelector{
		MatchLabels: s.Labels(),
	}
}

func (s *StatefulSet) BuildObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      s.Name(),
		Namespace: s.Namespace(),
		Labels:    s.Labels(),
	}
}

func (s *StatefulSet) BuildPodTemplateSpec() corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: s.Labels(),
		},
		Spec: corev1.PodSpec{
			Containers: s.Containers,
			Volumes:    s.Volumes(),
		},
	}
}

func (s *StatefulSet) Volumes() []corev1.Volume {
	return []corev1.Volume{}
}
