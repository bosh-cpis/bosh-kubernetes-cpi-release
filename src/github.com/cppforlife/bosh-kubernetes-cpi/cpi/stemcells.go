package cpi

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"

	bstem "github.com/cppforlife/bosh-kubernetes-cpi/stemcell"
)

type Stemcells struct {
	importer bstem.Importer
	finder   bstem.Finder
}

func NewStemcells(importer bstem.Importer, finder bstem.Finder) Stemcells {
	return Stemcells{importer, finder}
}

func (a Stemcells) CreateStemcell(
	imagePath string, cloudProps apiv1.StemcellCloudProps) (apiv1.StemcellCID, error) {

	props, err := bstem.NewProps(cloudProps)
	if err != nil {
		return apiv1.StemcellCID{}, bosherr.WrapErrorf(err, "Unmarshaling cloud properties")
	}

	stemcell, err := a.importer.ImportFromPath(imagePath, props)
	if err != nil {
		return apiv1.StemcellCID{}, bosherr.WrapErrorf(err, "Importing stemcell from '%s'", imagePath)
	}

	return stemcell.ID(), nil
}

func (a Stemcells) DeleteStemcell(cid apiv1.StemcellCID) error {
	stemcell, err := a.finder.Find(cid)
	if err != nil {
		return bosherr.WrapErrorf(err, "Finding stemcell '%s'", cid)
	}

	err = stemcell.Delete()
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting stemcell '%s'", cid)
	}

	return nil
}
