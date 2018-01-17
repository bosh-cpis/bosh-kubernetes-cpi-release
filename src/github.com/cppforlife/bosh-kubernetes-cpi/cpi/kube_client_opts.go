package cpi

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type KubeClientOpts struct {
	Config    string // if empty, expect to use service account
	Namespace string

	OverrideAPIHost string
	OverrideAPIPort int

	ImagePullSecretName string
}

func (o KubeClientOpts) Validate() error {
	// todo is this ignored when running in in-cluster mode?
	if o.Namespace == "" {
		return bosherr.Error("Must provide non-empty 'Namespace'")
	}

	return nil
}
