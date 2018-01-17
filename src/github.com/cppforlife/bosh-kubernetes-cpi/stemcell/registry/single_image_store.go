package registry

import (
	"fmt"
	"time"

	"github.com/docker/docker/image"
	"github.com/docker/docker/layer"
)

type SingleImageStore struct {
	asset NamedAsset
}

func (s SingleImageStore) Get(id image.ID) (*image.Image, error) {
	// todo check on asset?

	// note that empty config is necessary, otherwise kube raises
	// `unable to convert a nil pointer to a runtime API image` error
	imgTpl := fmt.Sprintf(`{
    "architecture":"amd64",
    "os":"linux",
    "config": {},
    "rootfs":{"diff_ids":["%s"],"type":"layers"}
  }`, s.asset.Digest())

	img, err := image.NewFromJSON([]byte(imgTpl))
	if err != nil {
		return nil, fmt.Errorf("SingleImageStore#Get: %s", err)
	}

	return img, nil
}

// Not implemented

func (s SingleImageStore) Create(config []byte) (image.ID, error)        { panic("called") }
func (s SingleImageStore) Delete(id image.ID) ([]layer.Metadata, error)  { panic("called") }
func (s SingleImageStore) Search(partialID string) (image.ID, error)     { panic("called") }
func (s SingleImageStore) SetParent(id image.ID, parent image.ID) error  { panic("called") }
func (s SingleImageStore) GetParent(id image.ID) (image.ID, error)       { panic("called") }
func (s SingleImageStore) SetLastUpdated(id image.ID) error              { panic("called") }
func (s SingleImageStore) GetLastUpdated(id image.ID) (time.Time, error) { panic("called") }
func (s SingleImageStore) Children(id image.ID) []image.ID               { panic("called") }
func (s SingleImageStore) Map() map[image.ID]*image.Image                { panic("called") }
func (s SingleImageStore) Heads() map[image.ID]*image.Image              { panic("called") }
