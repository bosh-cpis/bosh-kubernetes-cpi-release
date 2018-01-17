package vm

import (
	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
	bstem "github.com/cppforlife/bosh-kubernetes-cpi/stemcell"
	bvmsrv "github.com/cppforlife/bosh-kubernetes-cpi/vm/services"
)

type FactoryOpts struct {
	ImagePullSecretName string
}

type Factory struct {
	opts         FactoryOpts
	agentOptions apiv1.AgentOptions
	client       bkube.Client

	uuidGen     boshuuid.Generator
	timeService clock.Clock

	logTag string
	logger boshlog.Logger
}

func NewFactory(
	opts FactoryOpts,
	agentOptions apiv1.AgentOptions,
	client bkube.Client,
	uuidGen boshuuid.Generator,
	timeService clock.Clock,
	logger boshlog.Logger,
) Factory {
	return Factory{
		opts:         opts,
		agentOptions: agentOptions,
		client:       client,

		uuidGen:     uuidGen,
		timeService: timeService,

		logTag: "vm.Factory",
		logger: logger,
	}
}

func (f Factory) Create(
	agentID apiv1.AgentID,
	stemcell bstem.Stemcell,
	cloudProps apiv1.VMCloudProps,
	networks apiv1.Networks,
	env apiv1.VMEnv,
) (VM, error) {

	props, err := NewProps(cloudProps)
	if err != nil {
		return nil, err
	}

	id, err := f.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating VM id")
	}

	vm := f.newVM(apiv1.NewVMCID("vm-" + id))

	err = bvmsrv.NewServices(vm.ID(), env.Group(), props.NodePorts, props.ClusterIPs, f.client.Services()).Create()
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating services")
	}

	// todo create the target namespace if it doesn't already exist

	if len(networks) == 0 {
		return nil, bosherr.Error("Expected exactly one network; received zero")
	}

	for _, net := range networks {
		net.SetPreconfigured()
	}

	initialAgentEnv := apiv1.NewAgentEnvFactory().ForVM(
		agentID, vm.ID(), networks, env, f.agentOptions)

	// todo initialAgentEnv.AttachSystemDisk("0")

	err = vm.ConfigureAgent(initialAgentEnv)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Initial agent configuration")
	}

	startOpts := StartOpts{
		Props:               props,
		Env:                 env,
		ImagePullSecretName: f.opts.ImagePullSecretName,
	}

	err = vm.Start(stemcell, startOpts)
	if err != nil {
		f.cleanUpPartialCreate(vm)
		return nil, bosherr.WrapError(err, "Starting VM")
	}

	return vm, nil
}

func (f Factory) Find(cid apiv1.VMCID) (VM, error) {
	return f.newVM(cid), nil
}

func (f Factory) newVM(cid apiv1.VMCID) VMImpl {
	return NewVMImpl(cid, f.client.Pods(), f.client.ConfigMaps(),
		f.client.PVCs(), f.client.Services(), f.timeService, f.logger)
}

func (f Factory) cleanUpPartialCreate(vm VM) {
	err := vm.Delete()
	if err != nil {
		f.logger.Error(f.logTag, "Failed to clean up partially created VM: %s", err)
	}
}
