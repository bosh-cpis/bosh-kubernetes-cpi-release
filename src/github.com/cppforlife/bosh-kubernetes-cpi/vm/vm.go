package vm

import (
	"time"

	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	kretry "k8s.io/client-go/util/retry"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
	bvmnet "github.com/cppforlife/bosh-kubernetes-cpi/vm/network"
)

type VMImpl struct {
	cid        apiv1.VMCID
	networking bvmnet.Networking

	podsClient       kcorev1.PodInterface
	configMapsClient kcorev1.ConfigMapInterface
	pvcsClient       kcorev1.PersistentVolumeClaimInterface

	timeService  clock.Clock
	readyTimeout time.Duration

	logTag string
	logger boshlog.Logger
}

func NewVMImpl(
	cid apiv1.VMCID,
	networking bvmnet.Networking,
	podsClient kcorev1.PodInterface,
	configMapsClient kcorev1.ConfigMapInterface,
	pvcsClient kcorev1.PersistentVolumeClaimInterface,
	timeService clock.Clock,
	logger boshlog.Logger,
) VMImpl {
	return VMImpl{cid, networking,
		podsClient, configMapsClient, pvcsClient,
		timeService, 5 * time.Minute, "vm.VMImpl", logger}
}

func (vm VMImpl) ID() apiv1.VMCID { return vm.cid }

func (vm VMImpl) SetMetadata(meta apiv1.VMMeta) error {
	retryErr := kretry.RetryOnConflict(kretry.DefaultRetry, func() error {
		pod, getErr := vm.podsClient.Get(vm.cid.AsString(), kmetav1.GetOptions{})
		if getErr != nil {
			return bosherr.WrapErrorf(getErr, "Getting pod")
		}

		if pod.ObjectMeta.Annotations == nil {
			pod.ObjectMeta.Annotations = map[string]string{}
		}

		for k, v := range meta.StringedMap() {
			lbl := bkube.NewCustomLabel(k, v)
			pod.ObjectMeta.Annotations[lbl.Name()] = lbl.Value()
		}

		_, updateErr := vm.podsClient.Update(pod)
		return updateErr
	})
	if retryErr != nil {
		return bosherr.WrapErrorf(retryErr, "Updating pod")
	}

	return nil
}

func (vm VMImpl) Reboot() error {
	return bosherr.Errorf("Rebooting is not supported")
}

func (vm VMImpl) Exists() (bool, error) {
	pods, err := vm.podsClient.List(bkube.NewVMLabel(vm.cid).AsListOpts())
	if err != nil {
		return false, bosherr.WrapError(err, "Listing pods")
	}

	if len(pods.Items) > 1 {
		return false, bosherr.Errorf("Expected to find exactly one pod but found '%d' pods", len(pods.Items))
	}

	return len(pods.Items) == 1, nil
}

func (vm VMImpl) Delete() error {
	err := vm.deletePod()
	if err != nil {
		return err
	}

	err = vm.configMapsClient.Delete(vm.cid.AsString(), kmetav1.NewDeleteOptions(0))
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return bosherr.WrapError(err, "Deleting config map")
		}
	}

	err = vm.networking.Delete()
	if err != nil {
		return bosherr.WrapError(err, "Deleting networking")
	}

	return nil
}

func (vm VMImpl) deletePod() error {
	err := vm.podsClient.Delete(vm.cid.AsString(), kmetav1.NewDeleteOptions(0))
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return bosherr.WrapError(err, "Deleting pod")
		}
	}

	return nil
}
