package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"

	bvmsrv "github.com/cppforlife/bosh-kubernetes-cpi/vm/services"
)

type Props struct {
	Region string `json:"region"` // value for failure-domain.beta.kubernetes.io/region
	Zone   string `json:"zone"`   // value for failure-domain.beta.kubernetes.io/zone

	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`

	kapiv1.ResourceRequirements `json:"resources"`

	NodePorts  []bvmsrv.NodePort  `json:"node_ports"`
	ClusterIPs []bvmsrv.ClusterIP `json:"cluster_ips"`

	// todo load balancers, external names
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
	for i, srv := range p.NodePorts {
		err := srv.Validate()
		if err != nil {
			return bosherr.WrapErrorf(err, "Validating node_ports[%d]", i)
		}
	}

	err := bvmsrv.NodePorts(p.NodePorts).ValidateNameUniqueness()
	if err != nil {
		return err
	}

	for i, srv := range p.ClusterIPs {
		err := srv.Validate()
		if err != nil {
			return bosherr.WrapErrorf(err, "Validating cluster_ips[%d]", i)
		}
	}

	err = bvmsrv.ClusterIPs(p.ClusterIPs).ValidateNameUniqueness()
	if err != nil {
		return err
	}

	return nil
}
