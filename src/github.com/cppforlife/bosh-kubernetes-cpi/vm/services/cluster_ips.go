package services

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type ClusterIPsService struct {
	cid            apiv1.VMCID
	ips            []ClusterIP
	servicesClient kcorev1.ServiceInterface
}

func NewClusterIPsService(cid apiv1.VMCID, ips []ClusterIP, servicesClient kcorev1.ServiceInterface) ClusterIPsService {
	return ClusterIPsService{cid, ips, servicesClient}
}

func (n ClusterIPsService) Create() error {
	vmLabel := bkube.NewVMLabel(n.cid)

	for _, ip := range n.ips {
		service := &kapiv1.Service{
			ObjectMeta: kmetav1.ObjectMeta{
				Name: n.cid.AsString() + "-" + ip.Name,
				Labels: map[string]string{
					vmLabel.Name(): vmLabel.Value(),
				},
			},
			Spec: kapiv1.ServiceSpec{
				Type:      kapiv1.ServiceTypeClusterIP,
				ClusterIP: ip.ClusterIP,
				Selector: map[string]string{
					vmLabel.Name(): vmLabel.Value(),
				},
			},
		}

		for _, port := range ip.Ports {
			service.Spec.Ports = append(service.Spec.Ports, kapiv1.ServicePort{
				Name:     port.Name,
				Protocol: kapiv1.Protocol(port.Protocol),
				Port:     port.Port,
			})
		}

		_, err := n.servicesClient.Create(service)
		if err != nil {
			return bosherr.WrapError(err, "Creating cluster IP service")
		}
	}

	return nil
}
