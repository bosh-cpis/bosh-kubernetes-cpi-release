package vm

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

const agentCfgMapKey = "instance_settings"

func (vm VMImpl) ConfigureAgent(agentEnv apiv1.AgentEnv) error {
	bytes, err := agentEnv.AsBytes()
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent env")
	}

	cfgMap := &kapiv1.ConfigMap{
		ObjectMeta: kmetav1.ObjectMeta{
			Name: vm.cid.AsString(),
			Labels: map[string]string{
				bkube.NewVMLabel(vm.cid).Name(): bkube.NewVMLabel(vm.cid).Value(),
			},
		},
		Data: map[string]string{
			agentCfgMapKey: string(bytes),
		},
	}

	_, err = vm.configMapsClient.Create(cfgMap)
	if err != nil {
		return bosherr.WrapError(err, "Creating agent env config map")
	}

	return nil
}

func (vm VMImpl) reconfigureAgent(agentEnvFunc func(apiv1.AgentEnv)) error {
	cfgMap, err := vm.configMapsClient.Get(vm.cid.AsString(), kmetav1.GetOptions{})
	if err != nil {
		return bosherr.WrapError(err, "Getting agent settings config map")
	}

	agentEnv, err := apiv1.NewAgentEnvFactory().FromBytes([]byte(cfgMap.Data[agentCfgMapKey]))
	if err != nil {
		return bosherr.WrapError(err, "Unmarshalling agent env")
	}

	agentEnvFunc(agentEnv)

	bytes, err := agentEnv.AsBytes()
	if err != nil {
		return bosherr.WrapError(err, "Marshalling agent env")
	}

	cfgMap.Data[agentCfgMapKey] = string(bytes)

	_, err = vm.configMapsClient.Update(cfgMap)
	if err != nil {
		return bosherr.WrapError(err, "Updating agent env config map")
	}

	return nil
}
