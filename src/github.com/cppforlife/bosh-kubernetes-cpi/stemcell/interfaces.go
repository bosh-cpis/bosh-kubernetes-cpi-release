package stemcell

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type Importer interface {
	ImportFromPath(string, Props) (Stemcell, error)
}

type Finder interface {
	Find(apiv1.StemcellCID) (Stemcell, error)
}

type ImporterFinder interface {
	Importer
	Finder
}

var _ ImporterFinder = MuxFactory{}
var _ ImporterFinder = RefImageFactory{}
var _ ImporterFinder = RegistryImageFactory{}
var _ ImporterFinder = DockerImageFactory{}
var _ ImporterFinder = ErrorImageFactory{}

type Stemcell interface {
	ID() apiv1.StemcellCID
	Image() string

	Exists() (bool, error)
	Delete() error
}

// RefImage represents a non-owned image (ie light stemcell)
var _ Stemcell = RefImage{}

// RegistryImage represents an owned image that CPI imported via Registry API
var _ Stemcell = RegistryImage{}

// DockerImage represents an owned image that CPI imported via Docker API
var _ Stemcell = DockerImage{}

type CIDDoesNotBelongError struct{}

func (e CIDDoesNotBelongError) Error() string { return "CIDDoesNotBelongError" }
