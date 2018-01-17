package registry

import (
	"fmt"

	"github.com/docker/distribution/reference"
	refstore "github.com/docker/docker/reference"
	"github.com/opencontainers/go-digest"
)

type SingleReferenceStore struct {
	asset NamedAsset
}

var _ refstore.Store = SingleReferenceStore{}

func (s SingleReferenceStore) ReferencesByName(ref reference.Named) []refstore.Association {
	if s.asset.MatchesNamed(ref) {
		return []refstore.Association{{Ref: ref, ID: s.asset.Digest()}}
	}

	return nil
}

func (s SingleReferenceStore) Get(ref reference.Named) (digest.Digest, error) {
	if s.asset.MatchesNamed(ref) {
		return s.asset.Digest(), nil
	}

	return "", fmt.Errorf("Did not find reference '%s' in store", ref)
}

// Not implemented

func (s SingleReferenceStore) References(id digest.Digest) []reference.Named { panic("called") }
func (s SingleReferenceStore) AddTag(ref reference.Named, id digest.Digest, force bool) error {
	panic("called")
}
func (s SingleReferenceStore) AddDigest(ref reference.Canonical, id digest.Digest, force bool) error {
	panic("called")
}
func (s SingleReferenceStore) Delete(ref reference.Named) (bool, error) { panic("called") }
