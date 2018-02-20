package main

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
	kapipolicyv1 "k8s.io/api/policy/v1beta1" // todo rename
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kintstr "k8s.io/apimachinery/pkg/util/intstr"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	kv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
)

type InstanceGroup struct {
	name      string
	instances []boshdir.Instance

	podsClient kcorev1.PodInterface
	pdbsClient kv1beta1.PodDisruptionBudgetInterface
}

func NewInstanceGroup(
	name string,
	instances []boshdir.Instance,
	podsClient kcorev1.PodInterface,
	pdbsClient kv1beta1.PodDisruptionBudgetInterface,
) InstanceGroup {
	return InstanceGroup{name, instances, podsClient, pdbsClient}
}

func (ig InstanceGroup) ExpectedMinAvailable() int {
	// todo check for expects_vm?
	return len(ig.instances) - 1
}

func (ig InstanceGroup) SetUpPDB() error {
	groupLbl, err := ig.groupLabel()
	if err != nil {
		return bosherr.WrapErrorf(err, "Getting VM group label from instances")
	}

	// todo what should be name?
	pdbName := groupLbl.Value()

	pdb, err := ig.pdbsClient.Get(pdbName, kmetav1.GetOptions{})
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return bosherr.WrapErrorf(err, "Getting PDB")
		} else {
			pdb = nil
		}
	}

	if pdb != nil {
		if pdb.Spec.MinAvailable.IntValue() != ig.ExpectedMinAvailable() {
			return ig.updatePDB(pdbName, groupLbl)
		}
	} else {
		// may conflict
		return ig.createPDB(pdbName, groupLbl)
	}

	return nil
}

func (ig InstanceGroup) groupLabel() (bkube.Label, error) {
	if len(ig.instances) == 0 {
		return bkube.Label{}, bosherr.Errorf("Expected more than 0 instances in an instance group")
	}

	var errs []error

	for _, inst := range ig.instances {
		lbl, err := ig.groupLabelForInstance(inst)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		return lbl, nil
	}

	err := bosherr.NewMultiError(errs...)

	return bkube.Label{}, bosherr.WrapErrorf(err, "Expected to find at least one pod")
}

func (ig InstanceGroup) groupLabelForInstance(inst boshdir.Instance) (bkube.Label, error) {
	if len(inst.VMID) == 0 {
		return bkube.Label{}, bosherr.Errorf("Expected instance to have VM CID")
	}

	pod, err := ig.podsClient.Get(inst.VMID, kmetav1.GetOptions{})
	if err != nil {
		return bkube.Label{}, bosherr.WrapErrorf(err, "Getting pod")
	}

	if pod.ObjectMeta.Labels == nil {
		return bkube.Label{}, bosherr.Errorf("Expected pod labels to be non-nil")
	}

	groupLblName := bkube.NewVMEnvGroupName()
	groupLblValue := pod.ObjectMeta.Labels[groupLblName.AsString()]

	if len(groupLblValue) == 0 {
		return bkube.Label{}, bosherr.Errorf("Expected pod group label to be non-empty")
	}

	return bkube.NewLabel(groupLblName, groupLblValue), nil
}

func (ig InstanceGroup) updatePDB(pdbName string, groupLbl bkube.Label) error {
	err := ig.pdbsClient.Delete(pdbName, kmetav1.NewDeleteOptions(0))
	if err != nil {
		if !kerrors.IsNotFound(err) {
			return bosherr.WrapError(err, "Deleting PDB")
		}
	}

	return ig.createPDB(pdbName, groupLbl)
}

func (ig InstanceGroup) createPDB(pdbName string, groupLbl bkube.Label) error {
	minAvail := kintstr.FromInt(ig.ExpectedMinAvailable())

	pdb := &kapipolicyv1.PodDisruptionBudget{
		ObjectMeta: kmetav1.ObjectMeta{
			Name: pdbName,
			// todo labels
		},
		Spec: kapipolicyv1.PodDisruptionBudgetSpec{
			// MinAvailable is the only specification that does not couple PDB to any controller
			// (since currently controller list if hard coded to several well known controllers)
			// this is the only way to have a standalong PDB.
			// https://github.com/kubernetes/kubernetes/issues/59839
			MinAvailable: &minAvail,

			Selector: &kmetav1.LabelSelector{
				MatchLabels: map[string]string{
					groupLbl.Name(): groupLbl.Value(),
				},
			},
		},
	}

	_, err := ig.pdbsClient.Create(pdb)
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating PDB")
	}

	return nil
}
