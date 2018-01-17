package services

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type NodePort struct {
	Name     string
	Protocol string
	Port     int32
	NodePort int32 `json:"node_port"`
}

type NodePorts []NodePort

type ClusterIP struct {
	Name      string
	ClusterIP string `json:"cluster_ip"`
	Ports     []ClusterIPPort

	Grouped bool // Indicates whether cluster IP is to be applied for an entire group
}

type ClusterIPs []ClusterIP

type ClusterIPPort struct {
	Name     string
	Protocol string
	Port     int32
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

func (s ClusterIP) Validate() error {
	if s.Name == "" {
		return bosherr.Error("Must provide non-empty 'name'")
	}

	if s.ClusterIP == "" {
		return bosherr.Error("Must provide non-empty 'cluster_ip'")
	}

	for i, port := range s.Ports {
		err := port.Validate()
		if err != nil {
			return bosherr.WrapErrorf(err, "Validating 'ports[%d]'", i)
		}
	}

	return nil
}

func (s ClusterIPPort) Validate() error {
	if s.Name == "" {
		return bosherr.Error("Must provide non-empty 'name'")
	}

	if s.Protocol == "" {
		return bosherr.Error("Must provide non-empty 'protocol'")
	}

	if s.Port == 0 {
		return bosherr.Error("Must provide non-empty 'port'")
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

func (ips ClusterIPs) ByGrouped() ([]ClusterIP, []ClusterIP) {
	var ungrouped []ClusterIP
	var grouped []ClusterIP

	for _, ip := range ips {
		if ip.Grouped {
			grouped = append(grouped, ip)
		} else {
			ungrouped = append(ungrouped, ip)
		}
	}

	return ungrouped, grouped
}

func (ips ClusterIPs) ValidateNameUniqueness() error {
	names := map[string]struct{}{}
	for _, ip := range ips {
		if _, found := names[ip.Name]; found {
			return bosherr.Errorf("Must have unique cluster IP name '%s'", ip.Name)
		}
		names[ip.Name] = struct{}{}
	}
	return nil
}
