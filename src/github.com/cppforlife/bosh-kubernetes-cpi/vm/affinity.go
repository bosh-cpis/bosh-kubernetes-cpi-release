package vm

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type Affinity struct{}

func (a Affinity) PlacePodOnNodes(pod *kapiv1.Pod, props Props) {
	reqs := []kapiv1.NodeSelectorRequirement{}
	reqs = append(reqs, a.placePodIntoRegionAndZone(pod, props)...)
	reqs = append(reqs, a.placePodOnSpecificNodes(pod, props)...)

	if len(reqs) > 0 {
		if pod.Spec.Affinity == nil {
			pod.Spec.Affinity = &kapiv1.Affinity{}
		}

		pod.Spec.Affinity.NodeAffinity = &kapiv1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &kapiv1.NodeSelector{
				NodeSelectorTerms: []kapiv1.NodeSelectorTerm{{
					MatchExpressions: reqs,
				}},
			},
		}
	}
}

func (Affinity) placePodIntoRegionAndZone(pod *kapiv1.Pod, props Props) []kapiv1.NodeSelectorRequirement {
	reqs := []kapiv1.NodeSelectorRequirement{}

	if len(props.Region) > 0 {
		reqs = append(reqs, kapiv1.NodeSelectorRequirement{
			Key:      "failure-domain.beta.kubernetes.io/region",
			Operator: kapiv1.NodeSelectorOpIn,
			Values:   []string{props.Region},
		})
	}

	if len(props.Zone) > 0 {
		reqs = append(reqs, kapiv1.NodeSelectorRequirement{
			Key:      "failure-domain.beta.kubernetes.io/zone",
			Operator: kapiv1.NodeSelectorOpIn,
			Values:   []string{props.Zone},
		})
	}

	return reqs
}

func (Affinity) placePodOnSpecificNodes(pod *kapiv1.Pod, props Props) []kapiv1.NodeSelectorRequirement {
	reqs := []kapiv1.NodeSelectorRequirement{}

	for k, v := range props.NodeLabels {
		reqs = append(reqs, kapiv1.NodeSelectorRequirement{
			Key:      k,
			Operator: kapiv1.NodeSelectorOpIn,
			Values:   []string{v},
		})
	}

	return reqs
}

func (Affinity) SpreadOutPodsAcrossNodes(pod *kapiv1.Pod, env apiv1.VMEnv) {
	if group := env.Group(); group != nil {
		groupLabel := bkube.NewVMEnvGroupLabel(*group)

		if pod.Spec.Affinity == nil {
			pod.Spec.Affinity = &kapiv1.Affinity{}
		}

		pod.Spec.Affinity.PodAntiAffinity = &kapiv1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []kapiv1.WeightedPodAffinityTerm{{
				Weight: 100,
				PodAffinityTerm: kapiv1.PodAffinityTerm{
					LabelSelector: &kmetav1.LabelSelector{
						MatchExpressions: []kmetav1.LabelSelectorRequirement{{
							Key:      groupLabel.Name(),
							Operator: kmetav1.LabelSelectorOpIn,
							Values:   []string{groupLabel.Value()},
						}},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			}},
		}
	}
}
