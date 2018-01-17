package cpi

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type FactoryOpts struct {
	FactoryInnerOpts `json:",inline"`
	Agent            apiv1.AgentOptions
}

type FactoryInnerOpts struct {
	Kube     KubeClientOpts `json:",inline"`
	Docker   DockerClientOpts
	Registry RegistryClientOpts
}

func (o FactoryOpts) Validate() error {
	err := o.FactoryInnerOpts.Validate()
	if err != nil {
		return err
	}

	err = o.Agent.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Agent configuration")
	}

	return nil
}

func (o FactoryInnerOpts) Validate() error {
	if o.Docker.IsPresent() {
		err := o.Docker.Validate()
		if err != nil {
			return bosherr.WrapError(err, "Validating Docker configuration")
		}
	}

	if o.Registry.IsPresent() {
		err := o.Registry.Validate()
		if err != nil {
			return bosherr.WrapError(err, "Validating Registry configuration")
		}
	}

	err := o.Kube.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Kube configuration")
	}

	return nil
}
