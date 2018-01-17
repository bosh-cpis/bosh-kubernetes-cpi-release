package main

import (
	"encoding/json"
	"os"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	"github.com/bosh-tools/bosh-cron/director"
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

	director, err := directorFactory.New()
	ensureNoErr(logger, "Failed building director", err)

	err = NewScheduler(director, logger).Run()
	ensureNoErr(logger, "Running scheduler", err)
}

func basicDeps() (boshlog.Logger, boshsys.FileSystem) {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	return logger, fs
}

func ensureNoErr(logger boshlog.Logger, errPrefix string, err error) {
	if err != nil {
		logger.Error("[bosh-cron]", "%s: %s", errPrefix, err)
		os.Exit(1)
	}
}
