package stemcell

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type MuxFactory struct {
	heavyFactory ImporterFinder
	lightFactory ImporterFinder
	logger       boshlog.Logger
}

func NewMuxFactory(
	heavyFactory ImporterFinder,
	lightFactory ImporterFinder,
	logger boshlog.Logger,
) MuxFactory {
	return MuxFactory{
		heavyFactory: heavyFactory,
		lightFactory: lightFactory,
		logger:       logger,
	}
}

func (f MuxFactory) ImportFromPath(imagePath string, props Props) (Stemcell, error) {
	if props.HasImage() {
		return f.lightFactory.ImportFromPath(imagePath, props)
	}
	return f.heavyFactory.ImportFromPath(imagePath, props)
}

func (f MuxFactory) Find(cid apiv1.StemcellCID) (Stemcell, error) {
	stem, err := f.lightFactory.Find(cid)
	if err != nil {
		if _, ok := err.(CIDDoesNotBelongError); !ok {
			return nil, err
		}
	} else {
		return stem, nil
	}

	return f.heavyFactory.Find(cid)
}
