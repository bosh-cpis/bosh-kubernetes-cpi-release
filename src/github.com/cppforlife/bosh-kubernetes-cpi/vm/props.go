package vm

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"

	bvmnet "github.com/cppforlife/bosh-kubernetes-cpi/vm/network"
)

type Network bvmnet.Props // aliased for json inlining

type Props struct {
	Region string `json:"region"` // value for failure-domain.beta.kubernetes.io/region
	Zone   string `json:"zone"`   // value for failure-domain.beta.kubernetes.io/zone

	NodeLabels map[string]string `json:"node_labels"`

	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`

	kapiv1.ResourceRequirements `json:"resources"`

	Network
}

func NewProps(props apiv1.VMCloudProps) (Props, error) {
	p1 := Props{}

	err := props.As(&p1)
	if err != nil {
		return Props{}, err
	}

	return p1, p1.Validate()
}

func (p Props) Validate() error {
	return bvmnet.Props(p.Network).Validate()
}
