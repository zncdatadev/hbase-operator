package builder

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AffinityStrength string

const (
	StrengthPrefer  AffinityStrength = "prefer"
	StrengthRequire AffinityStrength = "require"
)

type AffinityBuilder struct {
	PodAffinity []*PodAffinity
	//todo nodeAffinity
}

type PodAffinity struct {
	strength      AffinityStrength
	anti          bool
	weight        int32
	labelSelector metav1.LabelSelector
}

func NewPodAffinity(strength AffinityStrength, anti bool, labelSelector metav1.LabelSelector) *PodAffinity {
	return &PodAffinity{
		strength:      strength,
		anti:          anti,
		labelSelector: labelSelector,
	}
}

func (p *PodAffinity) Weight(weight int32) *PodAffinity {
	p.weight = weight
	return p
}

func (p *AffinityBuilder) Build() *corev1.Affinity {
	if len(p.PodAffinity) == 0 {
		return nil
	}

	var preferTerms []corev1.WeightedPodAffinityTerm
	var requireTerms []corev1.PodAffinityTerm
	var antiPreferTerms []corev1.WeightedPodAffinityTerm
	var antiRequireTerms []corev1.PodAffinityTerm

	for _, v := range p.PodAffinity {
		if v.strength == StrengthPrefer {
			weightTerm := corev1.WeightedPodAffinityTerm{
				Weight: v.weight,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &v.labelSelector,
					TopologyKey:   corev1.LabelHostname,
				},
			}
			if v.anti {
				antiPreferTerms = append(antiPreferTerms, weightTerm)
			} else {
				preferTerms = append(preferTerms, weightTerm)
			}
		} else if v.strength == StrengthRequire {
			term := corev1.PodAffinityTerm{
				LabelSelector: &v.labelSelector,
				TopologyKey:   corev1.LabelHostname,
			}
			if v.anti {
				antiRequireTerms = append(antiRequireTerms, term)
			} else {
				requireTerms = append(requireTerms, term)
			}
		}
	}

	return &corev1.Affinity{
		PodAffinity: &corev1.PodAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: preferTerms,
			RequiredDuringSchedulingIgnoredDuringExecution:  requireTerms,
		},
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: antiPreferTerms,
			RequiredDuringSchedulingIgnoredDuringExecution:  antiRequireTerms,
		},
	}
}
