package network

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Networking struct {
	cid            apiv1.VMCID
	servicesClient kcorev1.ServiceInterface
	nodePorts      NodePortsService

	logTag string
	logger boshlog.Logger
}

func NewNetworking(
	cid apiv1.VMCID,
	servicesClient kcorev1.ServiceInterface,
	logger boshlog.Logger,
) Networking {
	return Networking{cid, servicesClient,
		NewNodePortsService(cid, servicesClient),
		"vm.network.Networking", logger}
}

func (n Networking) Create(networks apiv1.Networks, props Props) (string, error) {
	n.logger.Debug(n.logTag, "Creating networking for VM '%s' with networks '%#v'", n.cid.AsString(), networks)

	err := n.nodePorts.Create(props.NodePorts)
	if err != nil {
		return "", bosherr.WrapError(err, "Creating node ports")
	}

	manualNetworkInitBashCmd, err := n.configureManual(networks, props)
	if err != nil {
		return "", bosherr.WrapError(err, "Configuring manual networks")
	}

	err = n.configureVIP(networks, props)
	if err != nil {
		return "", bosherr.WrapError(err, "Configuring VIP networks")
	}

	return manualNetworkInitBashCmd, nil
}

func (n Networking) Delete() error {
	err := n.nodePorts.Delete()
	if err != nil {
		return bosherr.WrapError(err, "Deleting node ports")
	}

	return nil
}

func (n Networking) configureManual(networks apiv1.Networks, props Props) (string, error) {
	var singleManualNet *apiv1.Network
	var singleManualNetName string

	for netName, net := range networks {
		netName, net := netName, net // copy

		if net.Type() == apiv1.NetworkTypeManual {
			if singleManualNet != nil {
				return "", bosherr.Errorf("Expected to find exactly one manual network")
			}
			singleManualNetName = netName
			singleManualNet = &net
		}
	}

	if singleManualNet != nil {
		n.logger.Debug(n.logTag, "Found manual network for VM '%s' with IP '%s'",
			n.cid.AsString(), (*singleManualNet).IP())

		ingresses := props.Ingresses.AppliesToNetwork(singleManualNetName)
		manualNetwork := NewManualNetwork(
			n.cid, *singleManualNet, ingresses, n.servicesClient, n.logger)

		err := manualNetwork.Create()
		if err != nil {
			return "", bosherr.WrapError(err, "Creating manual network")
		}

		return manualNetwork.ToBashCmd(), nil
	}

	return " : ", nil // bash noop
}

func (n Networking) configureVIP(networks apiv1.Networks, props Props) error {
	for netName, net := range networks {
		if net.Type() == apiv1.NetworkTypeVIP {
			ingresses := props.Ingresses.AppliesToNetwork(netName)
			vipNetwork := NewVIPNetwork(n.cid, net, ingresses, n.servicesClient, n.logger)

			err := vipNetwork.Create()
			if err != nil {
				return bosherr.WrapError(err, "Creating VIP network")
			}
		}
	}

	return nil
}
