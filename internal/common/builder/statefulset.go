package builder

import (
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IStatefulSetBuilder interface {
	Builder

	GetObjectMeta() metav1.ObjectMeta
	GetContaiers() []corev1.Container
	GetInitContainers() []corev1.Container
	GetVolumes() []corev1.Volume
	GetVolumesMounts() []corev1.VolumeMount
	GetVolumeClaimTemplates() []corev1.PersistentVolumeClaim
	GetTerminationGracePeriodSeconds() *int64
	GetAffinity() *corev1.Affinity
	GetTolerations() []corev1.Toleration
}

var _ IStatefulSetBuilder = &BaseStatefulSetBuilder{}

type BaseStatefulSetBuilder struct {
	BaseResourceBuilder

	Replccas    *int32
	ServiceName string
}

func (b *BaseStatefulSetBuilder) Build() (client.Object, error) {
	obj := &appv1.StatefulSet{
		ObjectMeta: b.GetObjectMeta(),
		Spec: appv1.StatefulSetSpec{
			Replicas: b.Replccas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.GetLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: b.GetLabels(),
				},
				Spec: corev1.PodSpec{
					Containers:     b.GetContaiers(),
					InitContainers: b.GetInitContainers(),
					Volumes:        b.GetVolumes(),
					Affinity:       b.GetAffinity(),
					Tolerations:    b.GetTolerations(),
				},
			},
			VolumeClaimTemplates: b.GetVolumeClaimTemplates(),
			ServiceName:          b.ServiceName,
		},
	}

	return obj, nil
}

func (b *BaseStatefulSetBuilder) GetReplicas() *int32 {
	return b.Replccas
}

func (b *BaseStatefulSetBuilder) GetServiceName() string {
	return b.ServiceName
}
func (b *BaseStatefulSetBuilder) GetObjectMeta() metav1.ObjectMeta {
	panic("Not implemented")
}

func (b *BaseStatefulSetBuilder) GetContaiers() []corev1.Container {
	panic("Not implemented")
}

func (b *BaseStatefulSetBuilder) GetInitContainers() []corev1.Container {
	panic("Not implemented")
}

func (b *BaseStatefulSetBuilder) GetVolumes() []corev1.Volume {
	panic("Not implemented")
}

func (b *BaseStatefulSetBuilder) GetVolumesMounts() []corev1.VolumeMount {
	panic("Not implemented")
}

// GetVolumeClaimTemplates implements IStatefulSetBuilder.
func (b *BaseStatefulSetBuilder) GetVolumeClaimTemplates() []corev1.PersistentVolumeClaim {
	panic("unimplemented")
}

func (b *BaseStatefulSetBuilder) GetTerminationGracePeriodSeconds() *int64 {
	panic("Not implemented")
}

func (b *BaseStatefulSetBuilder) GetAffinity() *corev1.Affinity {
	panic("Not implemented")
}

func (b *BaseStatefulSetBuilder) GetTolerations() []corev1.Toleration {
	panic("Not implemented")
}
