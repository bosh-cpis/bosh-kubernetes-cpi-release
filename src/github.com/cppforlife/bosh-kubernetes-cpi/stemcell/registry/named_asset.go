package registry

import (
	"io"

	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

type NamedAsset struct {
	named  reference.Named
	digest digest.Digest
	asset  NamelessAsset
}

type NamelessAsset interface {
	TgzStream() (io.ReadCloser, error)
	TarDigest() (digest.Digest, error)
}

func NewFSNamedAsset(named reference.Named, asset NamelessAsset) (NamedAsset, error) {
	digest, err := asset.TarDigest()
	if err != nil {
		return NamedAsset{}, err
	}

	return NamedAsset{named, digest, asset}, nil
}

func (a NamedAsset) MatchesNamed(n reference.Named) bool {
	return a.named.Name() == n.Name()
}

func (a NamedAsset) MatchesDigest(d digest.Digest) bool {
	return a.digest.String() == d.String()
}

func (a NamedAsset) Named() reference.Named { return a.named }
func (a NamedAsset) Digest() digest.Digest  { return a.digest }

func (a NamedAsset) Stream() (io.ReadCloser, error) { return a.asset.TgzStream() }
