package network

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type VIPNetwork struct {
	cid            apiv1.VMCID
	network        apiv1.Network
	ingress        []Ingress
	servicesClient kcorev1.ServiceInterface

	logTag string
	logger boshlog.Logger
}

func NewVIPNetwork(cid apiv1.VMCID, network apiv1.Network,
	ingress []Ingress, servicesClient kcorev1.ServiceInterface,
	logger boshlog.Logger) VIPNetwork {

	return VIPNetwork{cid, network,
		ingress, servicesClient,
		"vm.networking.VIPNetwork", logger}
}

func (n VIPNetwork) Create() error {
	ports := Ingresses(n.ingress).ToPorts()

	if len(ports) == 0 {
		// todo better error?
		return bosherr.Errorf("Expected to find at least one ingress" +
			" port configuration when using VIP networks")
	}

	clusterIP := NewClusterIP(n.network.IP(), n.servicesClient, n.logger)

	return clusterIP.Attach(n.cid, ports)
}
