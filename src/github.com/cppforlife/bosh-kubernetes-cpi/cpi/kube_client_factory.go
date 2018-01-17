package cpi

import (
	"fmt"
	"os"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	krest "k8s.io/client-go/rest"
	kclientcmd "k8s.io/client-go/tools/clientcmd"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type KubeClientFactory struct {
	fs   boshsys.FileSystem
	opts KubeClientOpts
}

func NewKubeClientFactory(fs boshsys.FileSystem, opts KubeClientOpts) KubeClientFactory {
	return KubeClientFactory{fs, opts}
}

func (f KubeClientFactory) Build() (bkube.Client, error) {
	clientCfg, err := f.clusterConfig()
	if err != nil {
		return bkube.Client{}, err
	}

	clientset, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		return bkube.Client{}, bosherr.WrapErrorf(err, "Building Kubernetes client")
	}

	return bkube.NewClient(clientset, f.opts.Namespace), nil
}

func (f KubeClientFactory) clusterConfig() (*krest.Config, error) {
	if len(f.opts.Config) > 0 {
		// todo deal with inlines certificates
		cfg, err := kclientcmd.Load([]byte(f.opts.Config))
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Loading kubernetes config")
		}

		clientCfg, err := kclientcmd.NewDefaultClientConfig(*cfg, &kclientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			return nil, bosherr.WrapErrorf(err, "Building kubernetes client config")
		}

		return clientCfg, nil
	}

	// Work around https://github.com/kubernetes/kubernetes/issues/40973
	// See https://github.com/coreos/etcd-operator/issues/731#issuecomment-283804819
	if len(os.Getenv("KUBERNETES_SERVICE_HOST")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_HOST", f.opts.OverrideAPIHost)
	}
	if len(os.Getenv("KUBERNETES_SERVICE_PORT")) == 0 {
		os.Setenv("KUBERNETES_SERVICE_PORT", fmt.Sprintf("%d", f.opts.OverrideAPIPort))
	}

	clientCfg, err := krest.InClusterConfig()
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Loading in cluster config")
	}

	return clientCfg, nil
}
