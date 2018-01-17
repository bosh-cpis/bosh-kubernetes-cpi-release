package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type DockerImage struct {
	cid    apiv1.StemcellCID
	logger boshlog.Logger
}

func NewDockerImage(
	cid apiv1.StemcellCID,
	logger boshlog.Logger,
) DockerImage {
	return DockerImage{cid, logger}
}

func (s DockerImage) ID() apiv1.StemcellCID { return s.cid }
func (s DockerImage) Image() string         { return s.cid.AsString() }

func (s DockerImage) Exists() (bool, error) { return true, nil }
func (s DockerImage) Delete() error         { return nil } // todo detag?
