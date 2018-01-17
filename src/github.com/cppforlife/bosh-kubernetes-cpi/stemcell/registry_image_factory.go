package stemcell

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"

	"github.com/cppforlife/bosh-kubernetes-cpi/stemcell/registry"
)

type RegistryImageFactory struct {
	imageName string
	registry  registry.Registry

	logTag string
	logger boshlog.Logger
}

func NewRegistryImageFactory(
	imageName string,
	registry registry.Registry,
	logger boshlog.Logger,
) RegistryImageFactory {
	return RegistryImageFactory{
		imageName: imageName,
		registry:  registry,

		logTag: "stemcell.RegistryImageFactory",
		logger: logger,
	}
}

func (i RegistryImageFactory) ImportFromPath(imagePath string, _ Props) (Stemcell, error) {
	i.logger.Debug(i.logTag, "Importing stemcell from path '%s' into registry", imagePath)

	ref, err := i.registry.Push(registry.NewFSTgzAsset(imagePath), i.imageName)
	if err != nil {
		return nil, bosherr.WrapError(err, "Pushing stemcell into registry")
	}

	i.logger.Debug(i.logTag, "Imported stemcell from path '%s' as '%s'", imagePath, ref)

	return NewRegistryImage(apiv1.NewStemcellCID(ref.FQ()), i.logger), nil
}

func (f RegistryImageFactory) Find(cid apiv1.StemcellCID) (Stemcell, error) {
	return NewRegistryImage(cid, f.logger), nil
}
