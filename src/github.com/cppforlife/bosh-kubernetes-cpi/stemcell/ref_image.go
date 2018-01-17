package stemcell

import (
	"strings"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type RefImage struct {
	cid    apiv1.StemcellCID
	logger boshlog.Logger
}

func NewRefImage(
	cid apiv1.StemcellCID,
	logger boshlog.Logger,
) RefImage {
	return RefImage{cid, logger}
}

func (s RefImage) ID() apiv1.StemcellCID { return s.cid }

func (s RefImage) Image() string {
	pieces := strings.SplitN(s.cid.AsString(), "/", 2)
	return pieces[len(pieces)-1]
}

func (s RefImage) Exists() (bool, error) { return true, nil }
func (s RefImage) Delete() error         { return nil }
