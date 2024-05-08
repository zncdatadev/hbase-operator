package builder

import (
	"context"

	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	"github.com/zncdata-labs/hbase-operator/pkg/image"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IStatefulSetBuilder interface {
	Builder

	GetObjectMeta() metav1.ObjectMeta
	GetContainers() []corev1.Container
	GetInitContainers() []corev1.Container
	GetVolumes() []corev1.Volume
	GetVolumesMounts() []corev1.VolumeMount
	GetVolumeClaimTemplates() []corev1.PersistentVolumeClaim
	GetTerminationGracePeriodSeconds() *int64
	GetAffinity() *corev1.Affinity
	GetTolerations() []corev1.Toleration
	GetEnvOverrides() map[string]string
	GetCommandOverrides() []string
	GetResourceRequirement() corev1.ResourceRequirements
}

var _ IStatefulSetBuilder = &BaseStatefulSetBuilder{}

type BaseStatefulSetBuilder struct {
	BaseResourceBuilder
	ClusterOperation *apiv1alpha1.ClusterOperationSpec
	EnvOverrides     map[string]string
	CommandOverrides []string
	Replicas         *int32
	Image            image.Image
	ServiceName      string
}

func NewBaseStatefulSetBuilder(
	client client.Client,
	name string,
	namespace string,
	labels map[string]string,
	annotations map[string]string,
	envOverrides map[string]string,
	CommandOverrides []string,
	replicas *int32,
	Image image.Image,
	serviceName string,
) *BaseStatefulSetBuilder {
	return &BaseStatefulSetBuilder{
		BaseResourceBuilder: BaseResourceBuilder{
			Client:      client,
			Name:        name,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		EnvOverrides:     envOverrides,
		CommandOverrides: CommandOverrides,
		Replicas:         replicas,
		Image:            Image,
		ServiceName:      serviceName,
	}
}

func (b *BaseStatefulSetBuilder) Build(ctx context.Context) (client.Object, error) {
	obj := &appv1.StatefulSet{
		ObjectMeta: b.GetObjectMeta(),
		Spec: appv1.StatefulSetSpec{
			Replicas: b.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.GetLabels(),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: b.GetLabels(),
				},
				Spec: corev1.PodSpec{
					Containers:     b.GetContainers(),
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
	return b.Replicas
}

func (b *BaseStatefulSetBuilder) GetServiceName() string {
	return b.ServiceName
}
func (b *BaseStatefulSetBuilder) GetObjectMeta() metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:        b.Name,
		Namespace:   b.GetNamespace(),
		Labels:      b.GetLabels(),
		Annotations: b.GetAnnotations(),
	}
}

func (b *BaseStatefulSetBuilder) GetEnvOverrides() map[string]string {
	return b.EnvOverrides
}

func (b *BaseStatefulSetBuilder) GetCommandOverrides() []string {
	return b.CommandOverrides
}

func (b *BaseStatefulSetBuilder) GetContainers() []corev1.Container {
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

func (b *BaseStatefulSetBuilder) GetResourceRequirement() corev1.ResourceRequirements {
	panic("Not implemented")
}

type GenericStatefulSetBuilder struct {
	BaseStatefulSetBuilder
}

func NewGenericStatefulSetBuilder(
	client client.Client,
	name string,
	namespace string,
	labels map[string]string,
	annotations map[string]string,
	envOverrides map[string]string,
	CommandOverrides []string,
	Image image.Image,
	replicas *int32, serviceName string,
) *GenericStatefulSetBuilder {
	return &GenericStatefulSetBuilder{
		BaseStatefulSetBuilder: *NewBaseStatefulSetBuilder(
			client,
			name,
			namespace,
			labels,
			annotations,
			envOverrides,
			CommandOverrides,
			replicas,
			Image,
			serviceName,
		),
	}
}

func (b *GenericStatefulSetBuilder) GetContainers() []corev1.Container {
	return []corev1.Container{
		{
			Name:      b.GetName(),
			Image:     b.Image.GetImageTag(),
			Resources: b.GetResourceRequirement(),
		},
	}
}

func (b *GenericStatefulSetBuilder) GetInitContainers() []corev1.Container {
	return []corev1.Container{}
}

func (b *GenericStatefulSetBuilder) GetVolumes() []corev1.Volume {
	return []corev1.Volume{}
}

func (b *GenericStatefulSetBuilder) GetVolumesMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{}
}

func (b *GenericStatefulSetBuilder) GetVolumeClaimTemplates() []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{}
}

func (b *GenericStatefulSetBuilder) GetTerminationGracePeriodSeconds() *int64 {
	return nil
}

func (b *GenericStatefulSetBuilder) GetAffinity() *corev1.Affinity {
	return &corev1.Affinity{}
}

func (b *GenericStatefulSetBuilder) GetTolerations() []corev1.Toleration {
	return []corev1.Toleration{}
}

func (b *GenericStatefulSetBuilder) GetResourceRequirement() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		// Limits: corev1.ResourceList{
		// 	corev1.ResourceCPU:    resource.MustParse("500m"),
		// 	corev1.ResourceMemory: resource.MustParse("512Mi"),
		// },
	}
}
