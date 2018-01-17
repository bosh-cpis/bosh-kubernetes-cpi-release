package cpi

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bstem "github.com/cppforlife/bosh-kubernetes-cpi/stemcell"
	bvm "github.com/cppforlife/bosh-kubernetes-cpi/vm"
)

type VMs struct {
	stemcellFinder bstem.Finder
	creator        bvm.Creator
	finder         bvm.Finder
}

func NewVMs(stemcellFinder bstem.Finder, creator bvm.Creator, finder bvm.Finder) VMs {
	return VMs{stemcellFinder, creator, finder}
}

func (a VMs) CreateVM(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	_ []apiv1.DiskCID, env apiv1.VMEnv) (apiv1.VMCID, error) {

	stemcell, err := a.stemcellFinder.Find(stemcellCID)
	if err != nil {
		return apiv1.VMCID{}, bosherr.WrapErrorf(err, "Finding stemcell '%s'", stemcellCID)
	}

	vm, err := a.creator.Create(agentID, stemcell, cloudProps, networks, env)
	if err != nil {
		return apiv1.VMCID{}, bosherr.WrapErrorf(err, "Creating VM with agent ID '%s'", agentID)
	}

	return vm.ID(), nil
}

func (a VMs) DeleteVM(cid apiv1.VMCID) error {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding vm '%s'", cid)
	}

	err = vm.Delete()
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting vm '%s'", cid)
	}

	return nil
}

func (a VMs) CalculateVMCloudProperties(res apiv1.VMResources) (apiv1.VMCloudProps, error) {
	cloudProps := apiv1.NewVMCloudPropsFromMap(map[string]interface{}{
		"limits": map[string]interface{}{
			"cpu":    res.CPU, // todo correct?
			"memory": fmt.Sprintf("%dMi", res.RAM),
		},
		"resources": map[string]interface{}{
			"cpu":    res.CPU,
			"memory": fmt.Sprintf("%dMi", res.RAM),
		},
		// todo ephemeral disk
	})

	return cloudProps, nil
}

func (a VMs) SetVMMetadata(cid apiv1.VMCID, metadata apiv1.VMMeta) error {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.SetMetadata(metadata)
}

func (a VMs) HasVM(cid apiv1.VMCID) (bool, error) {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return false, bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.Exists()
}

func (a VMs) RebootVM(cid apiv1.VMCID) error {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.Reboot()
}

func (a VMs) GetDisks(cid apiv1.VMCID) ([]apiv1.DiskCID, error) {
	vm, err := a.finder.Find(cid)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Finding VM '%s'", cid)
	}

	return vm.DiskIDs()
}
