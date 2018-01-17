package cpi

import (
	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bdisk "github.com/cppforlife/bosh-kubernetes-cpi/disk"
	bstem "github.com/cppforlife/bosh-kubernetes-cpi/stemcell"
	bvm "github.com/cppforlife/bosh-kubernetes-cpi/vm"
)

type Factory struct {
	fs          boshsys.FileSystem
	uuidGen     boshuuid.Generator
	timeService clock.Clock
	defaultOpts FactoryOpts
	logger      boshlog.Logger
}

var _ apiv1.CPIFactory = Factory{}

type CPI struct {
	Misc
	Stemcells
	VMs
	Disks
}

var _ apiv1.CPI = CPI{}

func NewFactory(
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	timeService clock.Clock,
	defaultOpts FactoryOpts,
	logger boshlog.Logger,
) Factory {
	return Factory{fs, uuidGen, timeService, defaultOpts, logger}
}

func (f Factory) New(ctx apiv1.CallContext) (apiv1.CPI, error) {
	opts, err := f.buildOpts(ctx)
	if err != nil {
		return nil, err
	}

	lightImageFactory := bstem.NewRefImageFactory(f.uuidGen, f.logger)

	var heavyImageFactory bstem.ImporterFinder

	{
		err := bosherr.Errorf("Heavy stemcell importing is not supported due to missing Docker daemon or Registry configurations.")
		heavyImageFactory = bstem.NewErrorImageFactory(err)
	}

	{
		dockerClient, err := DockerClientFactory{opts.Docker}.Build()
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Building Docker client")
		}
		if dockerClient != nil {
			heavyImageFactory = bstem.NewDockerImageFactory(dockerClient, f.fs, f.uuidGen, f.logger)
		}
	}

	{
		registryClient, err := RegistryClientFactory{opts.Registry}.Build()
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Building Registry client")
		}
		if registryClient != nil {
			heavyImageFactory = bstem.NewRegistryImageFactory(
				opts.Registry.StemcellImageName, *registryClient, f.logger)
		}
	}

	stemcells := bstem.NewMuxFactory(heavyImageFactory, lightImageFactory, f.logger)

	kubeClient, err := KubeClientFactory{f.fs, opts.Kube}.Build()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Building Kubernetes client")
	}

	disks := bdisk.NewFactory(kubeClient.PVCs(), f.uuidGen, f.timeService, f.logger)

	vmOpts := bvm.FactoryOpts{ImagePullSecretName: opts.Kube.ImagePullSecretName}
	vms := bvm.NewFactory(vmOpts, f.defaultOpts.Agent, kubeClient, f.uuidGen, f.timeService, f.logger)

	return CPI{
		NewMisc(),
		NewStemcells(stemcells, stemcells),
		NewVMs(stemcells, vms, vms),
		NewDisks(disks, disks, vms),
	}, nil
}

func (f Factory) buildOpts(ctx apiv1.CallContext) (FactoryInnerOpts, error) {
	var opts FactoryInnerOpts

	err := ctx.As(&opts)
	if err != nil {
		return FactoryInnerOpts{}, bosherr.WrapError(err, "Parsing CPI context")
	}

	if len(opts.Kube.Config) > 0 { // todo more generic?
		err := opts.Validate()
		if err != nil {
			return FactoryInnerOpts{}, bosherr.WrapError(err, "Validating CPI context")
		}
	} else {
		opts = f.defaultOpts.FactoryInnerOpts
	}

	return opts, nil
}
