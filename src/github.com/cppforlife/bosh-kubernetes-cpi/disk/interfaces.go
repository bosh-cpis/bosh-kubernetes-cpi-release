package disk

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type Creator interface {
	Create(int, Props) (Disk, error)
}

var _ Creator = Factory{}

type Finder interface {
	Find(apiv1.DiskCID) (Disk, error)
}

var _ Finder = Factory{}

type Disk interface {
	ID() apiv1.DiskCID
	SetMetadata(apiv1.VMMeta) error // todo

	Exists() (bool, error)
	Delete() error
}

var _ Disk = DiskImpl{}
