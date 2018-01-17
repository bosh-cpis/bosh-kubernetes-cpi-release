package stemcell

import (
	"github.com/cppforlife/bosh-cpi-go/apiv1"
)

type ErrorImageFactory struct {
	err error
}

func NewErrorImageFactory(err error) ErrorImageFactory {
	return ErrorImageFactory{err}
}

func (i ErrorImageFactory) ImportFromPath(_ string, _ Props) (Stemcell, error) {
	return nil, i.err
}

func (i ErrorImageFactory) Find(cid apiv1.StemcellCID) (Stemcell, error) {
	return nil, i.err
}
