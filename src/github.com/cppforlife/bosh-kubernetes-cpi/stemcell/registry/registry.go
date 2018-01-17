package registry

import (
	"fmt"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/distribution"
	"github.com/docker/docker/distribution/xfer"
	"github.com/docker/docker/pkg/progress"
	"github.com/docker/docker/registry"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	// todo "github.com/docker/libtrust"
)

type Registry struct {
	opts RegistryOpts
}

type RegistryOpts struct {
	Host     string
	PullHost string

	Auth RegistryAuthOpts

	LogFunc func(string)
}

type RegistryAuthOpts struct {
	URL      string // e.g. "https://gcr.io"
	Username string
	Password string
}

type RegistryReference struct {
	Host   string
	Image  string
	Digest string
	// tag is not returned as it's not immutable
}

func (r RegistryReference) FQ() string {
	return r.Host + "/" + r.Image + "@" + r.Digest
}

func NewRegistry(opts RegistryOpts) Registry {
	if opts.LogFunc == nil {
		opts.LogFunc = func(string) {}
	}
	return Registry{opts}
}

func (r Registry) Push(asset NamelessAsset, image string) (RegistryReference, error) {
	log.SetLevel(log.PanicLevel) // todo better place?

	// e.g. "localhost:5000/bosh.io/stemcells"
	ref, err := reference.ParseNormalizedNamed(r.opts.Host + "/" + image)
	if err != nil {
		return RegistryReference{}, fmt.Errorf("Parsing image '%s': %s", image, err)
	}

	// Push by digest is not supported, so only tags are supported.
	ref, err = reference.WithTag(ref, "latest") // todo generate uuid
	if err != nil {
		return RegistryReference{}, fmt.Errorf("Adding latest tag: %s", err)
	}

	asset2, err := NewFSNamedAsset(ref, asset)
	if err != nil {
		return RegistryReference{}, fmt.Errorf("Building named asset: %s", err)
	}

	refDigest, err := r.pushRef(asset2)
	if err != nil {
		return RegistryReference{}, fmt.Errorf("Pushing named asset: %s", err)
	}

	regRef := RegistryReference{Host: r.opts.Host, Image: image, Digest: refDigest}

	if len(r.opts.PullHost) > 0 {
		regRef.Host = r.opts.PullHost
	}

	return regRef, nil
}

func (r Registry) pushRef(asset NamedAsset) (string, error) {
	registryService, err := registry.NewService(registry.ServiceOptions{
		InsecureRegistries: []string{"127.0.0.0/8", "192.168.99.0/24"},
	})
	if err != nil {
		return "", fmt.Errorf("Building registry client: %s", err)
	}

	progressChan := make(chan progress.Progress, 100)
	writesDone := make(chan string)

	go func() {
		var reportedDigest string
		for prog := range progressChan {
			if prog.Aux != nil {
				if typedProg, ok := prog.Aux.(types.PushResult); ok {
					reportedDigest = typedProg.Digest
				}
			}
		}
		writesDone <- reportedDigest
	}()

	pushConfig := &distribution.ImagePushConfig{
		Config: distribution.Config{
			MetaHeaders: nil,
			AuthConfig: &types.AuthConfig{
				ServerAddress: r.opts.Auth.URL,
				Username:      r.opts.Auth.Username,
				Password:      r.opts.Auth.Password,
			},
			ProgressOutput:  progress.ChanOutput(progressChan),
			RegistryService: registryService,
			ImageEventLogger: func(imageID, refName, action string) {
				r.opts.LogFunc(fmt.Sprintf("%s %s %s", imageID, refName, action))
			},
			ImageStore:     distribution.NewImageConfigStoreFromStore(SingleImageStore{asset}), // todo simplify
			ReferenceStore: SingleReferenceStore{asset},
		},
		ConfigMediaType: schema2.MediaTypeImageConfig,
		LayerStore:      SingleLayerProvider{asset},
		TrustKey:        nil,                            // todo daemon.trustKey,
		UploadManager:   xfer.NewLayerUploadManager(10), // todo why 10?
	}

	pushErr := distribution.Push(context.TODO(), asset.Named(), pushConfig)

	close(progressChan)
	reportedDigest := <-writesDone

	if pushErr != nil {
		return "", fmt.Errorf("Pushing image: %s", pushErr)
	}
	if len(reportedDigest) == 0 {
		return "", fmt.Errorf("Unexpected empty digest for pushed image")
	}

	return reportedDigest, nil
}
