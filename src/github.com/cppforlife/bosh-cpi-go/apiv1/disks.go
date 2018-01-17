package apiv1

type Disks interface {
	CreateDisk(int, DiskCloudProps, *VMCID) (DiskCID, error)
	DeleteDisk(DiskCID) error

	AttachDisk(VMCID, DiskCID) error
	DetachDisk(VMCID, DiskCID) error

	HasDisk(DiskCID) (bool, error)
}

type DiskCloudProps interface {
	As(interface{}) error
	_final() // interface unimplementable from outside
}

type DiskCID struct {
	cloudID
}

func NewDiskCID(cid string) DiskCID {
	if cid == "" {
		panic("Internal incosistency: Disk CID must not be empty")
	}
	return DiskCID{cloudID{cid}}
}
