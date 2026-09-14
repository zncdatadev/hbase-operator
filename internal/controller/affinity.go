package controller

import (
	"encoding/json"

	commonsv1alpha1 "github.com/zncdatadev/operator-go/pkg/apis/commons/v1alpha1"
	"github.com/zncdatadev/operator-go/pkg/constant"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	affinityLogger = ctrl.Log.WithName("controller").WithName("affinity")
)

// PodAffinity declares one (anti-)affinity term of the default scheduling policy.
type PodAffinity struct {
	affinityRequired bool
	anti             bool
	weight           int32
	labels           map[string]string
}

func NewPodAffinity(labels map[string]string, affinityRequired, anti bool) *PodAffinity {
	return &PodAffinity{
		affinityRequired: affinityRequired,
		anti:             anti,
		labels:           labels,
	}
}

func (p *PodAffinity) Weight(weight int32) *PodAffinity {
	p.weight = weight
	return p
}

// AffinityBuilder renders the default pod (anti-)affinity the pre-Gen3 operator applied. It is
// used only when neither the CRD config nor podOverrides supply an affinity — the framework
// leaves .spec.template.spec.affinity nil in that case, so the nil-guard default stays correct.
type AffinityBuilder struct {
	PodAffinity []PodAffinity
}

func NewAffinityBuilder(podAffinity ...PodAffinity) *AffinityBuilder {
	return &AffinityBuilder{PodAffinity: podAffinity}
}

func (a *AffinityBuilder) buildPodAffinity() (*corev1.PodAffinity, *corev1.PodAntiAffinity) {
	var preferTerms []corev1.WeightedPodAffinityTerm
	var requireTerms []corev1.PodAffinityTerm
	var antiPreferTerms []corev1.WeightedPodAffinityTerm
	var antiRequireTerms []corev1.PodAffinityTerm

	for _, pa := range a.PodAffinity {
		if pa.affinityRequired {
			term := corev1.PodAffinityTerm{
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: pa.labels,
				},
			}
			if pa.anti {
				antiRequireTerms = append(antiRequireTerms, term)
			} else {
				requireTerms = append(requireTerms, term)
			}
		} else {
			if pa.weight == 0 {
				pa.weight = corev1.DefaultHardPodAffinitySymmetricWeight
				affinityLogger.V(1).Info("Weight not set for preferred pod affinity, using default", "weight", pa.weight)
			}
			weightTerm := corev1.WeightedPodAffinityTerm{
				Weight: pa.weight,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: pa.labels,
					},
					TopologyKey: corev1.LabelHostname,
				},
			}
			if pa.anti {
				antiPreferTerms = append(antiPreferTerms, weightTerm)
			} else {
				preferTerms = append(preferTerms, weightTerm)
			}
		}
	}

	podAffinity := &corev1.PodAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution:  requireTerms,
		PreferredDuringSchedulingIgnoredDuringExecution: preferTerms,
	}

	podAntiAffinity := &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution:  antiRequireTerms,
		PreferredDuringSchedulingIgnoredDuringExecution: antiPreferTerms,
	}

	return podAffinity, podAntiAffinity
}

func (a *AffinityBuilder) Build() *corev1.Affinity {
	podAffinity, podAntiAffinity := a.buildPodAffinity()

	return &corev1.Affinity{
		PodAffinity:     podAffinity,
		PodAntiAffinity: podAntiAffinity,
	}
}

// defaultAffinityConfig wraps DefaultAffinity as the product's default for the framework-owned
// half of a role's config block. The framework folds it BENEATH the CR's role and role group
// levels, so a user's `config.affinity` still wins — without the product patching the built pod,
// which would land after podOverrides and beat the user.
//
// The CRD carries affinity as a schema-free RawExtension, so the typed value is marshalled here.
// A marshalling failure is impossible for a value this package constructs, and dropping the
// default is strictly better than failing every reconcile, so it degrades to "no default".
func defaultAffinityConfig(clusterName, roleName string) *commonsv1alpha1.RoleGroupConfigSpec {
	raw, err := json.Marshal(DefaultAffinity(clusterName, roleName))
	if err != nil {
		affinityLogger.Error(err, "encoding the default affinity", "cluster", clusterName, "role", roleName)
		return nil
	}
	return &commonsv1alpha1.RoleGroupConfigSpec{
		Affinity: &k8sruntime.RawExtension{Raw: raw},
	}
}

// DefaultAffinity reproduces the pre-Gen3 default scheduling policy: prefer co-locating with
// the cluster's other pods (weight 20) and prefer spreading same-role pods apart (weight 70).
func DefaultAffinity(clusterName, roleName string) *corev1.Affinity {
	affinityLabels := map[string]string{
		constant.LabelKubernetesInstance: clusterName,
		constant.LabelKubernetesName:     "hbase",
	}
	antiAffinityLabels := map[string]string{
		constant.LabelKubernetesInstance:  clusterName,
		constant.LabelKubernetesName:      "hbase",
		constant.LabelKubernetesComponent: roleName,
	}

	return NewAffinityBuilder(
		*NewPodAffinity(affinityLabels, false, false).Weight(20),
		*NewPodAffinity(antiAffinityLabels, false, true).Weight(70),
	).Build()
}
