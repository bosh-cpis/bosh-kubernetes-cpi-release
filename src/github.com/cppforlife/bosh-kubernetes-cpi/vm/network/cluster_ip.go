package network

import (
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type ClusterIP struct {
	ip             string
	servicesClient kcorev1.ServiceInterface

	logTag string
	logger boshlog.Logger
}

func NewClusterIP(ip string, servicesClient kcorev1.ServiceInterface, logger boshlog.Logger) ClusterIP {
	return ClusterIP{ip, servicesClient, "vm.networking.ClusterIP", logger}
}

func (n ClusterIP) Attach(cid apiv1.VMCID, ports []kapiv1.ServicePort) error {
	vmLabel := bkube.NewVMLabel(cid)

	n.logger.Debug(n.logTag, "Attaching VM '%s' to IP '%s'", cid.AsString(), n.ip)

	service := &kapiv1.Service{
		ObjectMeta: kmetav1.ObjectMeta{
			// todo assume that this will be the IP
			Name: "bosh-ip-" + strings.Replace(n.ip, ".", "-", -1),
			Labels: map[string]string{
				vmLabel.Name(): vmLabel.Value(), // for user searching
			},
		},
		Spec: kapiv1.ServiceSpec{
			Type: kapiv1.ServiceTypeClusterIP,
			Selector: map[string]string{
				vmLabel.Name(): vmLabel.Value(),
			},
			Ports:     ports,
			ClusterIP: n.ip, // todo find by ClusterIP
		},
	}

	return ServiceKubeObj{service, n.servicesClient}.CreateOrUpdate()
}
