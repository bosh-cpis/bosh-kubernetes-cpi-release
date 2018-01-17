package disk

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwatch "k8s.io/apimachinery/pkg/watch"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

func (d DiskImpl) IsReady() error {
	pvc, err := d.pvcsClient.Get(d.cid.AsString(), kmetav1.GetOptions{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting PVC")
	}

	if d.isPVCBound(pvc) {
		return nil
	}

	listOpts := bkube.NewDiskLabel(d.cid).AsListOpts()
	listOpts.ResourceVersion = pvc.ResourceVersion
	listOpts.Watch = true

	events, err := d.pvcsClient.Watch(listOpts)
	if err != nil {
		return bosherr.WrapError(err, "Watching PVC")
	}

	defer events.Stop()

	timer := d.timeService.NewTimer(d.readyTimeout)
	defer timer.Stop()

	d.logger.Debug(d.logTag, "Starting following events")

	for {
		select {
		case event := <-events.ResultChan():
			if event.Type == kwatch.Modified {
				pvc, ok := event.Object.(*kapiv1.PersistentVolumeClaim)
				if !ok {
					return bosherr.Errorf("Expected object to be a PVC but found '%T'", event.Object)
				}

				if d.isPVCBound(pvc) {
					return nil
				} else {
					d.logger.Debug(d.logTag, "PVC not ready yet")
				}
			}

		case <-timer.C():
			return bosherr.Error("Expected PVC to become ready (ClaimBound phase)")
		}
	}
}

func (d DiskImpl) isPVCBound(pvc *kapiv1.PersistentVolumeClaim) bool {
	return pvc.Status.Phase == kapiv1.ClaimBound
}
