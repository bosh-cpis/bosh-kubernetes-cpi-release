package stemcell

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

const refImageStemcellPrefix = "scl-"

type RefImageFactory struct {
	uuidGen boshuuid.Generator

	logTag string
	logger boshlog.Logger
}

func NewRefImageFactory(
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) RefImageFactory {
	return RefImageFactory{
		uuidGen: uuidGen,

		logTag: "stemcell.RefImageFactory",
		logger: logger,
	}
}

func (f RefImageFactory) ImportFromPath(imagePath string, props Props) (Stemcell, error) {
	id, err := f.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating stemcell id")
	}

	id = refImageStemcellPrefix + id + "/" + props.Image

	return NewRefImage(apiv1.NewStemcellCID(id), f.logger), nil
}

func (f RefImageFactory) Find(cid apiv1.StemcellCID) (Stemcell, error) {
	if !strings.HasPrefix(cid.AsString(), refImageStemcellPrefix) {
		return nil, CIDDoesNotBelongError{}
	}
	return NewRefImage(cid, f.logger), nil
}
