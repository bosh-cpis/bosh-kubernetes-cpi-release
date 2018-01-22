package network

import (
	"fmt"

	kapiv1 "k8s.io/api/core/v1"
)

func (is Ingresses) AppliesToNetwork(networkName string) []Ingress {
	var result []Ingress
	for _, i := range is {
		if i.AppliesToNetwork(networkName) {
			result = append(result, i)
		}
	}
	return result
}

func (is Ingresses) ToPorts() []kapiv1.ServicePort {
	var result []kapiv1.ServicePort
	for i, ingress := range is {
		for j, port := range ingress.Ports {
			for k, proto := range ingress.ToProtocols() {
				result = append(result, kapiv1.ServicePort{
					Name:     fmt.Sprintf("bosh-port-%d-%d-%d", i, j, k),
					Protocol: proto,
					Port:     port,
				})
			}
		}
	}
	return result
}

func (i Ingress) AppliesToNetwork(networkName string) bool {
	if len(i.Networks) == 0 {
		return true
	}
	for _, name := range i.Networks {
		if name == networkName {
			return true
		}
	}
	return false
}

func (i Ingress) ToProtocols() []kapiv1.Protocol {
	if len(i.Protocols) == 0 {
		return []kapiv1.Protocol{kapiv1.ProtocolTCP}
	}
	var result []kapiv1.Protocol
	for _, proto := range i.Protocols {
		switch proto {
		case "tcp", "TCP":
			result = append(result, kapiv1.ProtocolTCP)
		case "udp", "UDP":
			result = append(result, kapiv1.ProtocolUDP)
		default:
			result = append(result, kapiv1.Protocol(proto))
		}
	}
	return result
}
