package main

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cppforlife/bosh-kubernetes-cpi/cpi"
	"github.com/cppforlife/bosh-kubernetes-cpi/pdbctrl/director"
)

type Config struct {
	SyncIntervalStr string `json:"SyncInterval"`

	Kube     cpi.KubeClientOpts
	Director director.Config
}

func (c Config) Validate() error {
	_, err := time.ParseDuration(c.SyncIntervalStr)
	if err != nil {
		return bosherr.WrapError(err, "Validating SyncInterval")
	}

	err = c.Kube.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Kube configuration")
	}

	err = c.Director.Validate()
	if err != nil {
		return bosherr.WrapError(err, "Validating Director configuration")
	}

	return nil
}

func (c Config) SyncInterval() time.Duration {
	d, err := time.ParseDuration(c.SyncIntervalStr)
	if err != nil {
		panic(fmt.Sprintf("Unexpected SyncInterval: %s", err))
	}
	return d
}
