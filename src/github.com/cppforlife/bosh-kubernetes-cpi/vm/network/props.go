package network

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Props struct {
	NodePorts NodePorts `json:"node_ports"`
	Ingresses Ingresses

	// todo load balancers, external names
}

type NodePorts []NodePort

type NodePort struct {
	Name     string
	Protocol string
	Port     int32
	NodePort int32 `json:"node_port"`
}

type Ingresses []Ingress

type Ingress struct {
	Ports     []int32
	Protocols []string // if none specified, assume tcp
	Networks  []string // if none specified, assume applies to all
}

func (p Props) Validate() error {
	for i, srv := range p.NodePorts {
		err := srv.Validate()
		if err != nil {
			return bosherr.WrapErrorf(err, "Validating node_ports[%d]", i)
		}
	}

	err := p.NodePorts.ValidateNameUniqueness()
	if err != nil {
		return err
	}

	return nil
}

func (s NodePort) Validate() error {
	if s.Name == "" {
		return bosherr.Error("Must provide non-empty 'name'")
	}

	if s.Protocol == "" {
		return bosherr.Error("Must provide non-empty 'protocol'")
	}

	if s.Port == 0 {
		return bosherr.Error("Must provide non-empty 'port'")
	}

	if s.NodePort == 0 {
		return bosherr.Error("Must provide non-empty 'node_port'")
	}

	return nil
}

func (nps NodePorts) ValidateNameUniqueness() error {
	names := map[string]struct{}{}
	for _, np := range nps {
		if _, found := names[np.Name]; found {
			return bosherr.Errorf("Must have unique node port name '%s'", np.Name)
		}
		names[np.Name] = struct{}{}
	}
	return nil
}
