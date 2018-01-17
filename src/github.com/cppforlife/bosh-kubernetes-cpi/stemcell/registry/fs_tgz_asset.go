package registry

import (
	"compress/gzip"
	_ "crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/opencontainers/go-digest"
)

type FSTgzAsset struct {
	path string
}

var _ NamelessAsset = FSTgzAsset{}

func NewFSTgzAsset(path string) FSTgzAsset {
	return FSTgzAsset{path}
}

// "/Users/pivotal/Downloads/bosh-stemcell-3468.15-warden-boshlite-ubuntu-trusty-go_agent/image.gz"
func (a FSTgzAsset) TgzStream() (io.ReadCloser, error) {
	file, err := os.OpenFile(a.path, os.O_RDONLY, os.ModeDir)
	if err != nil {
		return nil, fmt.Errorf("Opening tgz asset: %s", err)
	}

	return file, nil
}

func (a FSTgzAsset) TarDigest() (digest.Digest, error) {
	stream, err := a.TgzStream()
	if err != nil {
		return "", fmt.Errorf("Preparing tgz stream: %s", err)
	}

	defer stream.Close()

	gzipReader, err := gzip.NewReader(stream)
	if err != nil {
		return "", fmt.Errorf("Unzipping stream: %s", err)
	}

	readerDgst, err := digest.SHA256.FromReader(gzipReader)
	if err != nil {
		return "", fmt.Errorf("Calculating digest: %s", err)
	}

	return readerDgst, nil
}
