package main

import (
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	"github.com/cppforlife/bosh-cpi-go/rpc"
)

type CPIFactory struct{}

type CPI struct{}

var _ apiv1.CPIFactory = CPIFactory{}
var _ apiv1.CPI = CPI{}

func main() {
	logger := boshlog.NewLogger(boshlog.LevelNone)

	cli := rpc.NewFactory(logger).NewCLI(CPIFactory{})

	err := cli.ServeOnce()
	if err != nil {
		logger.Error("main", "Serving once: %s", err)
		os.Exit(1)
	}
}

// Empty CPI implementation

func (f CPIFactory) New(_ apiv1.CallContext) (apiv1.CPI, error) {
	return CPI{}, nil
}

func (c CPI) Info() (apiv1.Info, error) {
	return apiv1.Info{}, nil
}

func (c CPI) CreateStemcell(imagePath string, _ apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {
	return apiv1.NewStemcellCID("stemcell-cid"), nil
}

func (c CPI) DeleteStemcell(cid apiv1.StemcellCID) error {
	return nil
}

func (c CPI) CreateVM(
	agentID apiv1.AgentID, stemcellCID apiv1.StemcellCID,
	cloudProps apiv1.VMCloudProps, networks apiv1.Networks,
	associatedDiskCIDs []apiv1.DiskCID, env apiv1.VMEnv) (apiv1.VMCID, error) {

	return apiv1.NewVMCID("vm-cid"), nil
}

func (c CPI) DeleteVM(cid apiv1.VMCID) error {
	return nil
}

func (c CPI) CalculateVMCloudProperties(res VMResources) (apiv1.VMCloudProps, error) {
	return apiv1.NewVMCloudPropsFromMap(map[string]interface{}{}), nil
}

func (c CPI) SetVMMetadata(cid apiv1.VMCID, metadata apiv1.VMMeta) error {
	return nil
}

func (c CPI) HasVM(cid apiv1.VMCID) (bool, error) {
	return false, nil
}

func (c CPI) RebootVM(cid apiv1.VMCID) error {
	return nil
}

func (c CPI) GetDisks(cid apiv1.VMCID) ([]apiv1.DiskCID, error) {
	return []apiv1.DiskCID{}, nil
}

func (c CPI) CreateDisk(size int,
	cloudProps apiv1.DiskCloudProps, associatedVMCID *apiv1.VMCID) (apiv1.DiskCID, error) {

	return apiv1.NewDiskCID("disk-cid"), nil
}

func (c CPI) DeleteDisk(cid apiv1.DiskCID) error {
	return nil
}

func (c CPI) AttachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	return nil
}

func (c CPI) DetachDisk(vmCID apiv1.VMCID, diskCID apiv1.DiskCID) error {
	return nil
}

func (c CPI) HasDisk(cid apiv1.DiskCID) (bool, error) {
	return false, nil
}
