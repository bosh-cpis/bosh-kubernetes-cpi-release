package stemcell

import (
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type Props struct {
	Image string // e.g. "bosh.io/stemcells:tag..."
}

func NewProps(props apiv1.StemcellCloudProps) (Props, error) {
	p1 := Props{}

	err := props.As(&p1)
	if err != nil {
		return Props{}, err
	}

	return p1, nil
}

func (p Props) HasImage() bool { return len(p.Image) > 0 }
