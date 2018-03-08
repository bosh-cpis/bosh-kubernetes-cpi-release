package main

import (
	"encoding/json"
	"os"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/cppforlife/bosh-kubernetes-cpi/cpi"
	"github.com/cppforlife/bosh-kubernetes-cpi/pdbctrl/director"
)

func main() {
	logger, fs := basicDeps()
	defer logger.HandlePanic("Main")

	cfgBytes, err := fs.ReadFile(os.Args[1])
	ensureNoErr(logger, "Failed to read config", err)

	config := Config{}

	err = json.Unmarshal(cfgBytes, &config)
	ensureNoErr(logger, "Failed to parse config", err)

	directorFactory := director.NewFactory(config.Director, logger)

	kubeClient, err := cpi.NewKubeClientFactory(fs, config.Kube).Build()
	ensureNoErr(logger, "Failed building Kubernetes client", err)

	igFactory := NewInstanceGroupFactory(kubeClient)

	errCh := make(chan error, 1)

	go func() {
		errCh <- NewController(config.SyncInterval(), directorFactory, igFactory, logger).Run()
	}()

	go func() {
		errCh <- NewRecoveryController(5*time.Second, directorFactory, igFactory, logger).Run()
	}()

	ensureNoErr(logger, "Running", <-errCh)
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem) {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	return logger, fs
}

func ensureNoErr(logger boshlog.Logger, errPrefix string, err error) {
	if err != nil {
		logger.Error("[bosh-kubernetes-cpi/pdbctrl]", "%s: %s", errPrefix, err)
		os.Exit(1)
	}
}
