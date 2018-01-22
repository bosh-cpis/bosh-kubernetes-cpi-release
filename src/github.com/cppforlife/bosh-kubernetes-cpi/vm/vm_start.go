package vm

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
	bstem "github.com/cppforlife/bosh-kubernetes-cpi/stemcell"
	bvmnet "github.com/cppforlife/bosh-kubernetes-cpi/vm/network"
)

const (
	vmContainerName  = "bosh-container"
	agentSettingFile = "warden-cpi-agent-env.json"
)

type StartOpts struct {
	Stemcell bstem.Stemcell
	Networks apiv1.Networks

	Props Props
	Env   apiv1.VMEnv

	ImagePullSecretName string
}

func (vm VMImpl) Start(opts StartOpts) error {
	trueBool := true
	falseBool := true
	zeroInt64 := int64(0)

	networkInitBashCmd, err := vm.networking.Create(opts.Networks, bvmnet.Props(opts.Props.Network))
	if err != nil {
		return bosherr.WrapErrorf(err, "Creating networking")
	}

	pod := &kapiv1.Pod{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:        vm.cid.AsString(),
			Labels:      vm.buildLabels(opts.Props, opts.Env),
			Annotations: opts.Props.Annotations,
		},
		Spec: kapiv1.PodSpec{
			Hostname: "bosh-stemcell", // changed by the agent

			RestartPolicy:                kapiv1.RestartPolicyNever,
			AutomountServiceAccountToken: &falseBool,

			// todo set dns policy to none
			// todo set terminationGracePeriodSeconds

			Containers: []kapiv1.Container{{
				Name: vmContainerName,

				Image:           opts.Stemcell.Image(),
				ImagePullPolicy: kapiv1.PullIfNotPresent, // supports imported local docker image

				Resources: opts.Props.ResourceRequirements,
				SecurityContext: &kapiv1.SecurityContext{
					Privileged: &trueBool,
					RunAsUser:  &zeroInt64, // 0 is root
				},

				Command: []string{"/bin/bash"},
				Args: []string{"-c", strings.Join([]string{
					// Remove default Docker/kube integration
					`umount /etc/resolv.conf`,
					`umount /etc/hosts`,
					`umount /etc/hostname`,
					networkInitBashCmd,
					// todo move initialization to agent
					`rm -rf /var/vcap/data/sys`,
					`mkdir -p /var/vcap/data/sys`,
					`mkdir -p /var/vcap/store`,
					// todo expose as agent api for disable time mgmt
					`echo '#!/bin/bash' > /var/vcap/bosh/bin/ntpdate`,
					// start the agent
					`exec env -i /usr/sbin/runsvdir-start`,
				}, " && ")},

				VolumeMounts: []kapiv1.VolumeMount{
					{
						Name:      "bosh-agent-settings",
						MountPath: "/var/vcap/bosh/" + agentSettingFile,
						SubPath:   agentSettingFile,
					},
					{
						Name:      "bosh-ephemeral-disk",
						MountPath: "/var/vcap/data",
					},
				},
			}},

			Volumes: []kapiv1.Volume{
				{
					Name: "bosh-agent-settings",
					VolumeSource: kapiv1.VolumeSource{
						ConfigMap: &kapiv1.ConfigMapVolumeSource{
							LocalObjectReference: kapiv1.LocalObjectReference{
								Name: vm.cid.AsString(),
							},
							Items: []kapiv1.KeyToPath{{
								Key:  agentCfgMapKey,
								Path: agentSettingFile,
							}},
						},
					},
				},
				{
					Name: "bosh-ephemeral-disk",
					VolumeSource: kapiv1.VolumeSource{
						EmptyDir: &kapiv1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	if len(opts.ImagePullSecretName) > 0 {
		pod.Spec.ImagePullSecrets = []kapiv1.LocalObjectReference{
			{Name: opts.ImagePullSecretName},
		}
	}

	Affinity{}.PlacePodIntoRegionAndZone(pod, opts.Props)
	Affinity{}.SpreadOutPodsAcrossNodes(pod, opts.Env)

	_, err = vm.podsClient.Create(pod)
	if err != nil {
		return bosherr.WrapError(err, "Creating pod")
	}

	return nil
}

func (vm VMImpl) buildLabels(props Props, env apiv1.VMEnv) map[string]string {
	labels := map[string]string{
		bkube.NewVMLabel(vm.cid).Name(): bkube.NewVMLabel(vm.cid).Value(),
	}

	if group := env.Group(); group != nil {
		groupLabel := bkube.NewVMEnvGroupLabel(*group)
		labels[groupLabel.Name()] = groupLabel.Value()
	}

	for k, v := range props.Labels {
		labels[k] = v
	}

	return labels
}
