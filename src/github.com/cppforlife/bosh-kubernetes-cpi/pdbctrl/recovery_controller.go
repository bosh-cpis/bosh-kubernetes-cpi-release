package main

import (
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/bosh-kubernetes-cpi/pdbctrl/director"
)

type RecoveryController struct {
	syncInterval    time.Duration
	directorFactory director.Factory
	igFactory       InstanceGroupFactory

	logTag string
	logger boshlog.Logger
}

func NewRecoveryController(syncInterval time.Duration, directorFactory director.Factory,
	igFactory InstanceGroupFactory, logger boshlog.Logger) *RecoveryController {

	return &RecoveryController{
		syncInterval:    syncInterval,
		directorFactory: directorFactory,
		igFactory:       igFactory,

		logTag: "pdbctrl.RecoveryController",
		logger: logger,
	}
}

func (c RecoveryController) Run() error {
	c.logger.Debug(c.logTag, "Starting scheduler with interval '%s'", c.syncInterval)

	ticker := time.NewTicker(c.syncInterval)

	for {
		select {
		case <-ticker.C:
			c.logger.Debug(c.logTag, "Reloading instances")

			err := c.sync()
			if err != nil {
				c.logger.Error(c.logTag, "Failed syncing: %s", err)
			}

			c.logger.Debug(c.logTag, "Finished syncing")
		}
	}

	return nil
}

func (c RecoveryController) sync() error {
	director, err := c.directorFactory.New()
	if err != nil {
		return err
	}

	deps, err := director.Deployments()
	if err != nil {
		return err
	}

	var errs []error

	for _, dep := range deps {
		instances, err := dep.Instances()
		if err != nil {
			errs = append(errs, err)
			continue
		}

		for _, inst := range instances {
			instance := c.igFactory.NewInstance(inst, dep, director)

			err := instance.ResurrectIfNecessary()
			if err != nil {
				err = bosherr.WrapErrorf(err, "Resurrecting dep '%s' ig '%s/%s'", dep.Name(), inst.Group, inst.ID)
				errs = append(errs, err)
				continue
			}
		}
	}

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}
