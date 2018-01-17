package disk

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
)

type DiskImpl struct {
	cid        apiv1.DiskCID
	pvcsClient kcorev1.PersistentVolumeClaimInterface

	timeService  clock.Clock
	readyTimeout time.Duration

	logTag string
	logger boshlog.Logger
}

func NewDiskImpl(
	cid apiv1.DiskCID,
	pvcsClient kcorev1.PersistentVolumeClaimInterface,
	timeService clock.Clock,
	logger boshlog.Logger,
) DiskImpl {
	return DiskImpl{cid, pvcsClient, timeService, 20 * time.Minute, "disk.DiskImpl", logger}
}

func (d DiskImpl) ID() apiv1.DiskCID { return d.cid }

func (d DiskImpl) SetMetadata(meta apiv1.VMMeta) error {
	retryErr := kretry.RetryOnConflict(kretry.DefaultRetry, func() error {
		pvc, getErr := d.pvcsClient.Get(d.cid.AsString(), kmetav1.GetOptions{})
		if getErr != nil {
			return bosherr.WrapErrorf(getErr, "Getting PVC")
		}

		if pvc.ObjectMeta.Annotations == nil {
			pvc.ObjectMeta.Annotations = map[string]string{}
		}

		for k, v := range meta.StringedMap() {
			lbl := bkube.NewCustomLabel(k, v)
			pvc.ObjectMeta.Annotations[lbl.Name()] = lbl.Value()
		}

		_, updateErr := d.pvcsClient.Update(pvc)
		return updateErr
	})
	if retryErr != nil {
		return bosherr.WrapErrorf(retryErr, "Updating PVC")
	}

	return nil
}

func (d DiskImpl) Exists() (bool, error) {
	pvcs, err := d.pvcsClient.List(bkube.NewDiskLabel(d.cid).AsListOpts())
	if err != nil {
		return false, bosherr.WrapError(err, "Listing persistent volume claims")
	}

	if len(pvcs.Items) > 1 {
		return false, bosherr.Errorf("Expected to find exactly one PVC but found '%d' PVCs", len(pvcs.Items))
	}

	return len(pvcs.Items) == 1, nil
}

func (d DiskImpl) Delete() error {
	err := d.pvcsClient.Delete(d.cid.AsString(), kmetav1.NewDeleteOptions(0))
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
	}
	return nil
}
