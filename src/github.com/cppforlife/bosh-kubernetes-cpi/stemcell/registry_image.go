package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type RegistryImage struct {
	cid    apiv1.StemcellCID
	logger boshlog.Logger
}

func NewRegistryImage(
	cid apiv1.StemcellCID,
	logger boshlog.Logger,
) RegistryImage {
	return RegistryImage{cid, logger}
}

func (s RegistryImage) ID() apiv1.StemcellCID { return s.cid }
func (s RegistryImage) Image() string         { return s.cid.AsString() }

func (s RegistryImage) Exists() (bool, error) { return true, nil }
func (s RegistryImage) Delete() error         { return nil } // todo detag?
