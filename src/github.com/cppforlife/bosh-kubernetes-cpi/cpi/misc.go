package cpi

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type Misc struct{}

func NewMisc() Misc {
	return Misc{}
}

func (m Misc) Info() (apiv1.Info, error) {
	return apiv1.Info{
		StemcellFormats: []string{"warden-tar", "general-tar"},
	}, nil
}
