package disk

import (
	"fmt"

	"code.cloudfoundry.org/clock"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
	kresource "k8s.io/apimachinery/pkg/api/resource"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	bkube "github.com/cppforlife/bosh-kubernetes-cpi/kube"
)

type Factory struct {
	pvcsClient  kcorev1.PersistentVolumeClaimInterface
	uuidGen     boshuuid.Generator
	timeService clock.Clock

	logTag string
	logger boshlog.Logger
}

func NewFactory(
	pvcsClient kcorev1.PersistentVolumeClaimInterface,
	uuidGen boshuuid.Generator,
	timeService clock.Clock,
	logger boshlog.Logger,
) Factory {
	return Factory{
		pvcsClient:  pvcsClient,
		uuidGen:     uuidGen,
		timeService: timeService,

		logTag: "disk.Factory",
		logger: logger,
	}
}

func (f Factory) Create(size int, props Props) (Disk, error) {
	id, err := f.uuidGen.Generate()
	if err != nil {
		return nil, bosherr.WrapError(err, "Generating disk id")
	}

	cid := apiv1.NewDiskCID("disk-" + id)

	sizeResource, err := kresource.ParseQuantity(fmt.Sprintf("%dMi", size))
	if err != nil {
		return nil, bosherr.WrapError(err, "Parsing disk size")
	}

	props.ResourceRequirements.Requests[kapiv1.ResourceStorage] = sizeResource

	pvc := &kapiv1.PersistentVolumeClaim{
		ObjectMeta: kmetav1.ObjectMeta{
			Name:        cid.AsString(),
			Labels:      f.buildLabels(cid, props),
			Annotations: f.buildAnnotations(props),
		},
		Spec: kapiv1.PersistentVolumeClaimSpec{
			AccessModes: []kapiv1.PersistentVolumeAccessMode{kapiv1.ReadWriteOnce},
			Resources:   props.ResourceRequirements,
		},
	}

	pvc, err = f.pvcsClient.Create(pvc)
	if err != nil {
		return nil, bosherr.WrapError(err, "Creating PVC")
	}

	disk := f.newDisk(cid)

	err = disk.IsReady()
	if err != nil {
		f.cleanUpPartialCreate(disk)
		return nil, bosherr.WrapError(err, "Waiting for disk")
	}

	return disk, nil
}

func (f Factory) Find(cid apiv1.DiskCID) (Disk, error) {
	return f.newDisk(cid), nil
}

func (f Factory) buildLabels(cid apiv1.DiskCID, props Props) map[string]string {
	labels := map[string]string{
		bkube.NewDiskLabel(cid).Name(): bkube.NewDiskLabel(cid).Value(),
	}
	for k, v := range props.Labels {
		labels[k] = v
	}
	return labels
}

func (f Factory) buildAnnotations(props Props) map[string]string {
	anns := map[string]string{
		"volume.beta.kubernetes.io/storage-class":       props.StorageClass,
		"volume.beta.kubernetes.io/storage-provisioner": props.StorageProvisioner,
	}
	for k, v := range props.Annotations {
		anns[k] = v
	}
	return anns
}

func (f Factory) newDisk(cid apiv1.DiskCID) DiskImpl {
	return NewDiskImpl(cid, f.pvcsClient, f.timeService, f.logger)
}

func (f Factory) cleanUpPartialCreate(disk Disk) {
	err := disk.Delete()
	if err != nil {
		f.logger.Error(f.logTag, "Failed to clean up partially created disk: %s", err)
	}
}
