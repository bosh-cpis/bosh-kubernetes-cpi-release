package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bdisk "github.com/cppforlife/bosh-kubernetes-cpi/disk"
	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

func (vm VMImpl) DiskIDs() ([]apiv1.DiskCID, error) {
	pod, err := vm.podsClient.Get(vm.cid.AsString(), kmetav1.GetOptions{})
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Getting pod")
	}

	cids := []apiv1.DiskCID{}

	for _, vol := range pod.Spec.Volumes {
		if vol.VolumeSource.PersistentVolumeClaim != nil {
			// todo use disk factory?
			pvc, err := vm.pvcsClient.Get(vol.VolumeSource.PersistentVolumeClaim.ClaimName, kmetav1.GetOptions{})
			if err != nil {
				return nil, bosherr.WrapErrorf(err, "Getting associated PVC")
			}

			if cidStr, found := pvc.Labels[bkube.NewDiskLabelName().AsString()]; found {
				cids = append(cids, apiv1.NewDiskCID(cidStr))
			}
		}
	}

	return cids, nil
}

func (vm VMImpl) AttachDisk(disk bdisk.Disk) error {
	diskHint := "/mnt/" + disk.ID().AsString() // todo warden-dev?

	return vm.recreatePod(
		func(pod *kapiv1.Pod) error { return vm.addVolumeToPodSpec(pod, disk, diskHint) },
		func(agentEnv apiv1.AgentEnv) { agentEnv.AttachPersistentDisk(disk.ID(), diskHint) },
	)
}

func (vm VMImpl) DetachDisk(disk bdisk.Disk) error {
	return vm.recreatePod(
		func(pod *kapiv1.Pod) error { return vm.removeVolumeFromPodSpec(pod, disk) },
		func(agentEnv apiv1.AgentEnv) { agentEnv.DetachPersistentDisk(disk.ID()) },
	)
}

func (vm VMImpl) recreatePod(
	updatePodFunc func(*kapiv1.Pod) error, updateAgentEnvFunc func(apiv1.AgentEnv)) error {

	pod, err := vm.podsClient.Get(vm.cid.AsString(), kmetav1.GetOptions{})
	if err != nil {
		return bosherr.WrapError(err, "Getting pod")
	}

	err = vm.deletePod()
	if err != nil {
		return bosherr.WrapError(err, "Deleting old pod")
	}

	// Reset in mem object
	pod.Status = kapiv1.PodStatus{}
	pod.ObjectMeta = kmetav1.ObjectMeta{
		Name:        pod.Name,
		Namespace:   pod.Namespace,
		Annotations: pod.Annotations,
		Labels:      pod.Labels,
	}

	err = updatePodFunc(pod)
	if err != nil {
		return bosherr.WrapError(err, "Updating pod spec")
	}

	err = vm.reconfigureAgent(updateAgentEnvFunc)
	if err != nil {
		return bosherr.WrapError(err, "Reconfiguring agent")
	}

	_, err = vm.podsClient.Create(pod)
	if err != nil {
		return bosherr.WrapError(err, "Creating updated pod")
	}

	err = vm.IsReady()
	if err != nil {
		return bosherr.WrapError(err, "Waiting for VM")
	}

	return nil
}

func (vm VMImpl) addVolumeToPodSpec(pod *kapiv1.Pod, disk bdisk.Disk, diskHint string) error {
	pod.Spec.Volumes = append(pod.Spec.Volumes, kapiv1.Volume{
		Name: disk.ID().AsString(),
		VolumeSource: kapiv1.VolumeSource{
			PersistentVolumeClaim: &kapiv1.PersistentVolumeClaimVolumeSource{
				ClaimName: disk.ID().AsString(),
			},
		},
	})

	var volMountAdded bool

	for i, cont := range pod.Spec.Containers {
		if cont.Name == vmContainerName {
			pod.Spec.Containers[i].VolumeMounts = append(cont.VolumeMounts, kapiv1.VolumeMount{
				Name:      disk.ID().AsString(),
				MountPath: diskHint,
			})
			volMountAdded = true
			break
		}
	}

	if !volMountAdded {
		return bosherr.Errorf("Expected to add volume mount '%s' to pod '%s'",
			disk.ID().AsString(), vm.cid.AsString())
	}

	return nil
}

func (vm VMImpl) removeVolumeFromPodSpec(pod *kapiv1.Pod, disk bdisk.Disk) error {
	var volRemoved, volMountRemoved bool

	for i, vol := range pod.Spec.Volumes {
		if vol.Name == disk.ID().AsString() {
			pod.Spec.Volumes = append(pod.Spec.Volumes[:i], pod.Spec.Volumes[i+1:]...)
			volRemoved = true
			break
		}
	}

	if !volRemoved {
		return bosherr.Errorf("Expected to remove volume '%s' from pod '%s'",
			disk.ID().AsString(), vm.cid.AsString())
	}

	for i, cont := range pod.Spec.Containers {
		// Assume there could be other containers
		if cont.Name == vmContainerName {
			for j, vol := range cont.VolumeMounts {
				if vol.Name == disk.ID().AsString() {
					pod.Spec.Containers[i].VolumeMounts = append(cont.VolumeMounts[:j], cont.VolumeMounts[j+1:]...)
					volMountRemoved = true
					break
				}
			}
		}
	}

	if !volMountRemoved {
		return bosherr.Errorf("Expected to remove volume mount '%s' from pod '%s'",
			disk.ID().AsString(), vm.cid.AsString())
	}

	return nil
}
