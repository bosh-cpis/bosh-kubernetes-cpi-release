package stemcell

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	dkrclient "github.com/docker/engine-api/client"
	dkrtypes "github.com/docker/engine-api/types"
)

type DockerImageFactory struct {
	dkrClient *dkrclient.Client

	fs      boshsys.FileSystem
	uuidGen boshuuid.Generator

	logTag string
	logger boshlog.Logger
}

func NewDockerImageFactory(
	dkrClient *dkrclient.Client,
	fs boshsys.FileSystem,
	uuidGen boshuuid.Generator,
	logger boshlog.Logger,
) DockerImageFactory {
	return DockerImageFactory{
		dkrClient: dkrClient,

		fs:      fs,
		uuidGen: uuidGen,

		logTag: "stemcell.DockerImageFactory",
		logger: logger,
	}
}

func (i DockerImageFactory) ImportFromPath(imagePath string, _ Props) (Stemcell, error) {
	i.logger.Debug(i.logTag, "Importing stemcell from path '%s'", imagePath)

	id, err := i.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating stemcell id")
	}

	id = "img-" + id

	file, err := i.fs.OpenFile(imagePath, os.O_RDONLY, os.ModeDir)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Opening image archive '%s'", imagePath)
	}

	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Reading image archive '%s'", imagePath)
	}

	src := dkrtypes.ImageImportSource{
		Source:     gzipReader,
		SourceName: "-",
	}

	opts := dkrtypes.ImageImportOptions{
		Message: "bosh",
		Tag:     id,
	}

	repo := "bosh.io.invalid/stemcells"

	responseBody, err := i.dkrClient.ImageImport(context.TODO(), src, repo, opts)
	if err != nil {
		return nil, bosherr.WrapErrorf(err, "Starting image import")
	}

	defer responseBody.Close()

	i.logger.Debug(i.logTag, "Waiting for import to finish")

	dec := json.NewDecoder(responseBody)

	for {
		var jm dockerJSONMessage

		err := dec.Decode(&jm)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, bosherr.WrapErrorf(err, "Decoding event from importing")
		}

		if jm.Error != nil {
			return nil, bosherr.WrapErrorf(jm.Error.Error(), "Importing error event")
		}
	}

	i.logger.Debug(i.logTag, "Imported stemcell from path '%s'", imagePath)

	cid := repo + ":" + id

	return NewDockerImage(apiv1.NewStemcellCID(cid), i.logger), nil
}

func (f DockerImageFactory) Find(cid apiv1.StemcellCID) (Stemcell, error) {
	return NewDockerImage(cid, f.logger), nil
}

// todo should be in docker client?
type dockerJSONMessage struct {
	Error *dockerJSONError `json:"errorDetail,omitempty"`
}

type dockerJSONError struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func (e dockerJSONError) Error() error {
	return fmt.Errorf("%s (%d)", e.Message, e.Code)
}
