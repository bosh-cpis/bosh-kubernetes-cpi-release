package cpi

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type DockerClientOpts struct {
	Host       string
	APIVersion string `json:"api_version"`
	TLS        DockerClientOptsTLS
}

type DockerClientOptsTLS struct {
	// Assume always enabled
	Cert DockerClientOptsTLSCert
}

type DockerClientOptsTLSCert struct {
	CA          string
	Certificate string
	PrivateKey  string `json:"private_key"`
}

func (o DockerClientOpts) IsPresent() bool {
	return len(o.Host) > 0
}

func (o DockerClientOpts) RequiresTLS() bool {
	return !strings.HasPrefix(o.Host, "unix://")
}

func (o DockerClientOpts) Validate() error {
	if o.Host == "" {
		return bosherr.Error("Must provide non-empty Host")
	}

	if o.APIVersion == "" {
		return bosherr.Error("Must provide non-empty APIVersion")
	}

	if o.RequiresTLS() {
		if len(o.TLS.Cert.CA) == 0 {
			return bosherr.Error("Must provide non-empty CA")
		}

		if len(o.TLS.Cert.Certificate) == 0 {
			return bosherr.Error("Must provide non-empty Certificate")
		}

		if len(o.TLS.Cert.PrivateKey) == 0 {
			return bosherr.Error("Must provide non-empty PrivateKey")
		}
	}

	return nil
}
