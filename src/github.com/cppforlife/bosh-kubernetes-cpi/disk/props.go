package disk

import (
	"fmt"

	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kapiv1 "k8s.io/api/core/v1"
)

type Props struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`

	kapiv1.ResourceRequirements `json:"resources"`

	StorageClass       string `json:"storage_class"`
	StorageProvisioner string `json:"storage_provisioner"`
}

func NewProps(props apiv1.DiskCloudProps) (Props, error) {
	// todo configurable options?
	p1 := Props{
		StorageClass:       "standard",
		StorageProvisioner: "default",

		ResourceRequirements: kapiv1.ResourceRequirements{
			Requests: kapiv1.ResourceList{}, // empty map
		},
	}

	err := props.As(&p1)
	if err != nil {
		return Props{}, err
	}

	if len(p1.StorageClass) == 0 {
		return p1, fmt.Errorf("Expected to find non-empty 'storage_class'")
	}

	if len(p1.StorageProvisioner) == 0 {
		return p1, fmt.Errorf("Expected to find non-empty 'storage_provisioner'")
	}

	return p1, nil
}
