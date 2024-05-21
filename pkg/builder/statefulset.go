package builder

import (
	"context"

	resourceClient "github.com/zncdatadev/hbase-operator/pkg/client"
	"github.com/zncdatadev/hbase-operator/pkg/util"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSetBuilder interface {
	Builder
	GetObject() *appv1.StatefulSet
	SetReplicas(replicas *int32) StatefulSetBuilder
	AddContainers([]corev1.Container) StatefulSetBuilder
	AddInitContainers([]corev1.Container) StatefulSetBuilder
	AddVolumes([]corev1.Volume) StatefulSetBuilder
	AddVolumeClaimTemplates([]corev1.PersistentVolumeClaim) StatefulSetBuilder
	AddTerminationGracePeriodSeconds(int64) StatefulSetBuilder
	AddAffinity(corev1.Affinity) StatefulSetBuilder
}

var _ StatefulSetBuilder = &GenericStatefulSetBuilder{}

type GenericStatefulSetBuilder struct {
	BaseResourceBuilder
	EnvOverrides     map[string]string
	CommandOverrides []string
	Image            util.Image

	obj *appv1.StatefulSet
}

func NewGenericStatefulSetBuilder(
	client resourceClient.ResourceClient,
	name string,
	envOverrides map[string]string,
	commandOverrides []string,
	image util.Image,
) *GenericStatefulSetBuilder {
	return &GenericStatefulSetBuilder{
		BaseResourceBuilder: BaseResourceBuilder{
			Client: client,
			Name:   name,
		},
		EnvOverrides:     envOverrides,
		CommandOverrides: commandOverrides,
		Image:            image,
	}
}

func (b *GenericStatefulSetBuilder) GetObject() *appv1.StatefulSet {
	if b.obj == nil {
		b.obj = &appv1.StatefulSet{
			ObjectMeta: b.GetObjectMeta(),
			Spec: appv1.StatefulSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: b.Client.GetMatchingLabels(),
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels:      b.Client.GetLabels(),
						Annotations: b.Client.GetAnnotations(),
					},
				},
			},
		}
	}
	return b.obj
}

func (b *GenericStatefulSetBuilder) Build(ctx context.Context) (ctrlclient.Object, error) {
	obj := b.GetObject()

	if len(obj.Spec.Template.Spec.Containers) == 0 {
		obj.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:    b.Name,
				Image:   b.Image.String(),
				Env:     util.EnvsToEnvVars(b.EnvOverrides),
				Command: b.CommandOverrides,
			},
		}
	}

	if obj.Spec.Replicas == nil {
		replicas := int32(1)
		obj.Spec.Replicas = &replicas
	}

	return obj, nil
}

func (b *GenericStatefulSetBuilder) SetReplicas(replicas *int32) StatefulSetBuilder {
	b.GetObject().Spec.Replicas = replicas
	return b
}

func (b *GenericStatefulSetBuilder) AddContainers(containers []corev1.Container) StatefulSetBuilder {
	b.GetObject().Spec.Template.Spec.Containers = containers
	return b
}

func (b *GenericStatefulSetBuilder) AddInitContainers(containers []corev1.Container) StatefulSetBuilder {
	b.GetObject().Spec.Template.Spec.InitContainers = containers
	return b
}

func (b *GenericStatefulSetBuilder) AddVolumes(volumes []corev1.Volume) StatefulSetBuilder {
	b.GetObject().Spec.Template.Spec.Volumes = volumes
	return b
}

func (b *GenericStatefulSetBuilder) AddVolumeClaimTemplates(claims []corev1.PersistentVolumeClaim) StatefulSetBuilder {
	b.GetObject().Spec.VolumeClaimTemplates = claims
	return b
}

func (b *GenericStatefulSetBuilder) AddTerminationGracePeriodSeconds(i int64) StatefulSetBuilder {
	b.GetObject().Spec.Template.Spec.TerminationGracePeriodSeconds = &i
	return b
}

func (b *GenericStatefulSetBuilder) AddAffinity(affinity corev1.Affinity) StatefulSetBuilder {
	b.GetObject().Spec.Template.Spec.Affinity = &affinity
	return b
}
