package services

import (
	"encoding/json"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type GroupedClusterIPsService struct {
	group          apiv1.VMEnvGroup
	ips            []ClusterIP
	servicesClient kcorev1.ServiceInterface
}

func NewGroupedClusterIPsService(group apiv1.VMEnvGroup, ips []ClusterIP, servicesClient kcorev1.ServiceInterface) GroupedClusterIPsService {
	return GroupedClusterIPsService{group, ips, servicesClient}
}

func (n GroupedClusterIPsService) Create() error {
	groupLabel := bkube.NewVMEnvGroupLabel(n.group)

	for _, ip := range n.ips {
		service := &kapiv1.Service{
			ObjectMeta: kmetav1.ObjectMeta{
				Name: n.group.AsString() + "-" + ip.Name,
				Labels: map[string]string{
					groupLabel.Name(): groupLabel.Value(),
				},
			},
			Spec: kapiv1.ServiceSpec{
				Type:      kapiv1.ServiceTypeClusterIP,
				ClusterIP: ip.ClusterIP,
				Selector: map[string]string{
					groupLabel.Name(): groupLabel.Value(),
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

		_, createErr := n.servicesClient.Create(service)
		if createErr != nil {
			if kerrors.IsAlreadyExists(createErr) || kerrors.IsInvalid(createErr) {
				return n.updateService(service)
			} else {
				return bosherr.WrapError(createErr, "Creating grouped cluster IP service")
			}
		}
	}

	return nil
}

func (n GroupedClusterIPsService) updateService(service *kapiv1.Service) error {
	var existingService *kapiv1.Service

	{
		var err error

		for i := 0; i < 60; i++ {
			existingService, err = n.servicesClient.Get(service.ObjectMeta.Name, kmetav1.GetOptions{})
			if err == nil {
				break
			}
			time.Sleep(500 * time.Millisecond) // todo clock
		}
		if err != nil {
			return bosherr.WrapErrorf(err, "Getting service")
		}
	}

	existingBytes, err := json.Marshal(existingService)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling existing cluster IP")
	}

	existingService.Labels = service.Labels
	existingService.Spec = service.Spec

	updatedBytes, err := json.Marshal(existingService)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling updated cluster IP")
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(existingBytes, updatedBytes, service)
	if err != nil {
		return bosherr.WrapError(err, "CreateTwoWayMergePatch of cluster IP")
	}

	_, err = n.servicesClient.Patch(service.ObjectMeta.Name, ktypes.StrategicMergePatchType, patch)
	if err != nil {
		return bosherr.WrapError(err, "Patching grouped cluster IP service")
	}

	return nil
}
