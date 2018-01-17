package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwatch "k8s.io/apimachinery/pkg/watch"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

func (vm VMImpl) IsReady() error {
	pod, err := vm.podsClient.Get(vm.cid.AsString(), kmetav1.GetOptions{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting PVC")
	}

	if vm.isReady(pod) {
		return nil
	}

	listOpts := bkube.NewVMLabel(vm.cid).AsListOpts()
	listOpts.ResourceVersion = pod.ResourceVersion // todo should we take resource version?
	listOpts.Watch = true

	events, err := vm.podsClient.Watch(listOpts)
	if err != nil {
		return bosherr.WrapError(err, "Watching Pod")
	}

	defer events.Stop()

	timer := vm.timeService.NewTimer(vm.readyTimeout)
	defer timer.Stop()

	vm.logger.Debug(vm.logTag, "Starting following events")

	for {
		select {
		case event := <-events.ResultChan():
			if event.Type == kwatch.Modified {
				pod, ok := event.Object.(*kapiv1.Pod)
				if !ok {
					return bosherr.Errorf("Expected object to be a Pod but found '%T'", event.Object)
				}

				if vm.isReady(pod) {
					return nil
				} else {
					vm.logger.Debug(vm.logTag, "Pod not ready yet")
				}
			}

		case <-timer.C():
			return bosherr.Error("Expected Pod to become ready")
		}
	}
}

func (vm VMImpl) isReady(pod *kapiv1.Pod) bool {
	if pod.Status.Phase != kapiv1.PodRunning {
		return false
	}

	for _, status := range pod.Status.ContainerStatuses {
		if status.Name == vmContainerName {
			return status.Ready && status.State.Running != nil
		}
	}

	return false
}
