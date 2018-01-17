package registry

import (
	"fmt"
	"io"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/docker/distribution"
	"github.com/docker/docker/layer"
	"github.com/opencontainers/go-digest"
)

type SingleLayerProvider struct {
	asset NamedAsset
}

type SingleLayer struct {
	chainID layer.ChainID
	asset   NamedAsset
}

var _ distribution.PushLayerProvider = SingleLayerProvider{}
var _ distribution.PushLayer = SingleLayer{}

func (p SingleLayerProvider) Get(chainID layer.ChainID) (distribution.PushLayer, error) {
	if p.asset.MatchesDigest(digest.Digest(chainID)) {
		return SingleLayer{chainID, p.asset}, nil
	}

	return nil, fmt.Errorf("Did not find single layer")
}

func (l SingleLayer) ChainID() layer.ChainID         { return l.chainID }
func (l SingleLayer) DiffID() layer.DiffID           { return layer.DiffID(l.chainID) }
func (l SingleLayer) Parent() distribution.PushLayer { return nil }
func (l SingleLayer) Open() (io.ReadCloser, error)   { return l.asset.Stream() }
func (l SingleLayer) Size() (int64, error)           { return 0, nil }
func (l SingleLayer) MediaType() string              { return schema2.MediaTypeLayer } // already gzipped
func (l SingleLayer) Release()                       {}
