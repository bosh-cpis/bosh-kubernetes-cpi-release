package network

import (
	"encoding/json"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	kapiv1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ServiceKubeObj struct {
	service        *kapiv1.Service
	servicesClient kcorev1.ServiceInterface
}

func (n ServiceKubeObj) CreateOrUpdate() error {
	_, createErr := n.servicesClient.Create(n.service)
	if createErr != nil {
		if kerrors.IsAlreadyExists(createErr) || kerrors.IsInvalid(createErr) {
			return n.update()
		} else {
			return bosherr.WrapError(createErr, "Creating cluster IP service")
		}
	}
	return nil
}

func (n ServiceKubeObj) update() error {
	var existingService *kapiv1.Service

	{
		var err error

		for i := 0; i < 60; i++ {
			existingService, err = n.servicesClient.Get(n.service.ObjectMeta.Name, kmetav1.GetOptions{})
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
		return bosherr.WrapError(err, "Marshalling existing service")
	}

	existingService.Labels = n.service.Labels
	existingService.Spec = n.service.Spec

	updatedBytes, err := json.Marshal(existingService)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling updated service")
	}

	patch, err := strategicpatch.CreateTwoWayMergePatch(existingBytes, updatedBytes, n.service)
	if err != nil {
		return bosherr.WrapError(err, "CreateTwoWayMergePatch of service")
	}

	_, err = n.servicesClient.Patch(n.service.ObjectMeta.Name, ktypes.StrategicMergePatchType, patch)
	if err != nil {
		return bosherr.WrapError(err, "Patching service")
	}

	return nil
}
