package cpi

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type RegistryClientOpts struct {
	Host     string
	PullHost string `json:"pull_host"`

	StemcellImageName string `json:"stemcell_image_name"`

	Auth RegistryClientOptsAuth
}

type RegistryClientOptsAuth struct {
	URL      string
	Username string
	Password string
}

func (o RegistryClientOpts) IsPresent() bool {
	return len(o.Host) > 0
}

func (o RegistryClientOpts) Validate() error {
	if o.Host == "" {
		return bosherr.Error("Must provide non-empty Host")
	}

	if o.StemcellImageName == "" {
		return bosherr.Error("Must provide non-empty StemcellImageName")
	}

	return nil
}
