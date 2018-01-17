package integration_test

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/clock"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	. "github.com/onsi/gomega"

	"github.com/cppforlife/bosh-kubernetes-cpi/cpi"
	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type CPI struct {
	testlib.CPI
	kubeClientFactory cpi.KubeClientFactory
}

func NewCPI() CPI {
	logger, fs, _, uuidGen := basicDeps()

	configPath := os.Getenv("BOSH_KUBE_CPI_KUBE_CONFIG_PATH")
	if len(configPath) == 0 {
		panic("Expected to find BOSH_KUBE_CPI_KUBE_CONFIG_PATH")
	}

	config, err := ioutil.ReadFile(configPath)
	Expect(err).ToNot(HaveOccurred())

	opts := cpi.FactoryOpts{
		FactoryInnerOpts: cpi.FactoryInnerOpts{
			Kube: cpi.KubeClientOpts{
				Config:    string(yaml2json(config)),
				Namespace: "bosh-kubernetes-cpi-integration-tests",
			},
		},
	}

	cpiFactory := cpi.NewFactory(fs, uuidGen, clock.NewClock(), opts, logger)
	kubeClientFactory := cpi.NewKubeClientFactory(fs, opts.Kube)

	return CPI{testlib.NewCPI(cpiFactory, logger), kubeClientFactory}
}

func (c CPI) Kube() bkube.Client {
	client, err := c.kubeClientFactory.Build()
	Expect(err).ToNot(HaveOccurred())

	return client
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem, boshsys.CmdRunner, boshuuid.Generator) {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	cmdRunner := boshsys.NewExecCmdRunner(logger)
	uuidGen := boshuuid.NewGenerator()
	return logger, fs, cmdRunner, uuidGen
}
