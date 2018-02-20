package main

import (
	"time"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/cppforlife/bosh-kubernetes-cpi/pdbctrl/director"
)

type Controller struct {
	syncInterval    time.Duration
	directorFactory director.Factory
	igFactory       InstanceGroupFactory

	logTag string
	logger boshlog.Logger
}

func NewController(syncInterval time.Duration, directorFactory director.Factory,
	igFactory InstanceGroupFactory, logger boshlog.Logger) *Controller {

	return &Controller{
		syncInterval:    syncInterval,
		directorFactory: directorFactory,
		igFactory:       igFactory,

		logTag: "pdbctrl.Controller",
		logger: logger,
	}
}

func (c Controller) Run() error {
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

func (c Controller) sync() error {
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

		instancesByGroup := map[string][]boshdir.Instance{}

		for _, inst := range instances {
			if _, found := instancesByGroup[inst.Group]; !found {
				instancesByGroup[inst.Group] = []boshdir.Instance{}
			}
			instancesByGroup[inst.Group] = append(instancesByGroup[inst.Group], inst)
		}

		for name, insts := range instancesByGroup {
			ig := c.igFactory.New(name, insts)

			err := ig.SetUpPDB()
			if err != nil {
				err = bosherr.WrapErrorf(err, "Setting up PDB dep '%s' ig '%s'", dep.Name(), name)
				errs = append(errs, err)
				continue
			}
		}
	}

	// todo delete unnecessary PDBs

	if len(errs) > 0 {
		return bosherr.NewMultiError(errs...)
	}

	return nil
}
