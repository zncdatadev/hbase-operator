package builder

import (
	apiv1alpha1 "github.com/zncdata-labs/hbase-operator/pkg/apis/v1alpha1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	GetClusterOperation() *apiv1alpha1.ClusterOperationSpec
	GetResourceRequirment() corev1.ResourceRequirements
}

var _ IStatefulSetBuilder = &BaseStatefulSetBuilder{}

type BaseStatefulSetBuilder struct {
	BaseResourceBuilder

	Replccas    *int32
	ServiceName string
}

func NewBaseStatefulSetBuilder(
	client client.Client,
	schema *runtime.Scheme,
	owner runtime.Object,
	name string,
	labels map[string]string,
	annotations map[string]string,
	clusterOperation *apiv1alpha1.ClusterOperationSpec,
	replicas *int32,
	serviceName string,
) *BaseStatefulSetBuilder {
	return &BaseStatefulSetBuilder{
		BaseResourceBuilder: BaseResourceBuilder{
			Client:           client,
			Schema:           schema,
			OwnerReference:   owner,
			Name:             name,
			Labels:           labels,
			Annotations:      annotations,
			ClusterOperation: clusterOperation,
		},
		Replccas:    replicas,
		ServiceName: serviceName,
	}
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
	return metav1.ObjectMeta{
		Name:        b.Name,
		Labels:      b.GetLabels(),
		Annotations: b.GetAnnotations(),
	}
}

func (b *BaseResourceBuilder) GetClusterOperation() *apiv1alpha1.ClusterOperationSpec {
	return b.ClusterOperation
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

func (b *BaseStatefulSetBuilder) GetResourceRequirment() corev1.ResourceRequirements {
	panic("Not implemented")
}

type GenericStatefulSetBuilder struct {
	BaseStatefulSetBuilder

	Image string
}

func NewGenericStatefulSetBuilder(
	client client.Client,
	schema *runtime.Scheme,
	owner runtime.Object,
	name string,
	labels map[string]string,
	annotations map[string]string,
	clusterOperation *apiv1alpha1.ClusterOperationSpec,
	replicas *int32, serviceName string,
) *GenericStatefulSetBuilder {
	return &GenericStatefulSetBuilder{
		BaseStatefulSetBuilder: *NewBaseStatefulSetBuilder(
			client,
			schema,
			owner,
			name,
			labels,
			annotations,
			clusterOperation,
			replicas,
			serviceName,
		),
	}
}

func (b *GenericStatefulSetBuilder) GetContaiers() []corev1.Container {
	return []corev1.Container{
		{
			Name:      b.GetName(),
			Image:     b.Image,
			Resources: b.GetResourceRequirment(),
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

func (b *GenericStatefulSetBuilder) GetResourceRequirment() corev1.ResourceRequirements {
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
