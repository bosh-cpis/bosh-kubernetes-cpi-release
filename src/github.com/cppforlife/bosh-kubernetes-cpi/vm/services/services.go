package services

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Services struct {
	cid   apiv1.VMCID
	group *apiv1.VMEnvGroup

	nodePorts []NodePort
	ips       []ClusterIP

	servicesClient kcorev1.ServiceInterface
}

func NewServices(cid apiv1.VMCID, group *apiv1.VMEnvGroup, nodePorts []NodePort, ips []ClusterIP, servicesClient kcorev1.ServiceInterface) Services {
	return Services{cid, group, nodePorts, ips, servicesClient}
}

func (n Services) Create() error {
	err := NewNodePortsService(n.cid, n.nodePorts, n.servicesClient).Create()
	if err != nil {
		return bosherr.WrapError(err, "Creating node ports")
	}

	ungroupedIPs, groupedIPs := ClusterIPs(n.ips).ByGrouped()

	err = NewClusterIPsService(n.cid, ungroupedIPs, n.servicesClient).Create()
	if err != nil {
		return bosherr.WrapError(err, "Creating cluster IPs")
	}

	if len(groupedIPs) > 0 {
		if n.group != nil {
			err = NewGroupedClusterIPsService(*n.group, groupedIPs, n.servicesClient).Create()
			if err != nil {
				return bosherr.WrapError(err, "Creating grouped cluster IPs")
			}
		} else {
			return bosherr.Errorf("Expected to find VM env group to be used with grouped cluster IPs")
		}
	}

	return nil
}
