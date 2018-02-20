package main

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type InstanceGroupFactory struct {
	client bkube.Client
}

func NewInstanceGroupFactory(client bkube.Client) InstanceGroupFactory {
	return InstanceGroupFactory{client}
}

func (f InstanceGroupFactory) New(name string, instances []boshdir.Instance) InstanceGroup {
	return InstanceGroup{name, instances, f.client.Pods(), f.client.PDBs()}
}
