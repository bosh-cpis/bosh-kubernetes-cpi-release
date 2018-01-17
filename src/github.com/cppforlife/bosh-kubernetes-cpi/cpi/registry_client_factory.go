package cpi

import (
	"github.com/cppforlife/bosh-kubernetes-cpi/stemcell/registry"
)

type RegistryClientFactory struct {
	opts RegistryClientOpts
}

func (f RegistryClientFactory) Build() (*registry.Registry, error) {
	if !f.opts.IsPresent() {
		return nil, nil // no options, no client
	}

	reg := registry.NewRegistry(registry.RegistryOpts{
		Host:     f.opts.Host,
		PullHost: f.opts.PullHost,

		Auth: registry.RegistryAuthOpts{
			URL:      f.opts.Auth.URL,
			Username: f.opts.Auth.Username,
			Password: f.opts.Auth.Password,
		},
	})

	return &reg, nil
}
