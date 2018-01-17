package services

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// todo kintstr "k8s.io/apimachinery/pkg/util/intstr"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type NodePortsService struct {
	cid            apiv1.VMCID
	nodePorts      []NodePort
	servicesClient kcorev1.ServiceInterface
}

func NewNodePortsService(cid apiv1.VMCID, nodePorts []NodePort, servicesClient kcorev1.ServiceInterface) NodePortsService {
	return NodePortsService{cid, nodePorts, servicesClient}
}

func (n NodePortsService) Create() error {
	vmLabel := bkube.NewVMLabel(n.cid)

	for _, nodePort := range n.nodePorts {
		service := &kapiv1.Service{
			ObjectMeta: kmetav1.ObjectMeta{
				Name: n.cid.AsString() + "-" + nodePort.Name,
				Labels: map[string]string{
					vmLabel.Name(): vmLabel.Value(),
				},
			},
			Spec: kapiv1.ServiceSpec{
				Type: kapiv1.ServiceTypeNodePort,
				Selector: map[string]string{
					vmLabel.Name(): vmLabel.Value(),
				},
				Ports: []kapiv1.ServicePort{{
					Name:     nodePort.Name,
					Protocol: kapiv1.Protocol(nodePort.Protocol),
					Port:     nodePort.Port,
					NodePort: nodePort.NodePort,
				}},
				// clusterIP cannot be set "None"
			},
		}

		_, err := n.servicesClient.Create(service)
		if err != nil {
			return bosherr.WrapError(err, "Creating node port service")
		}
	}

	return nil
}
