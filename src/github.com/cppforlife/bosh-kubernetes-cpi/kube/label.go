package kube

import (
	"crypto/md5"
	"fmt"
	"io"
	"strings"

	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabelName struct {
	org          string
	resourceType string
}

func NewDiskLabelName() LabelName {
	return LabelName{"bosh.io", "disk-cid"}
}

func NewVMLabelName() LabelName {
	return LabelName{"bosh.io", "vm-cid"}
}

func NewVMEnvGroupName() LabelName {
	return LabelName{"bosh.io", "group"}
}

func (l LabelName) AsString() string { return l.org + "/" + l.resourceType }

type Label struct {
	name  LabelName
	value string
}

func NewDiskLabel(cid apiv1.DiskCID) Label {
	return Label{NewDiskLabelName(), cid.AsString()}
}

func NewVMLabel(cid apiv1.VMCID) Label {
	return Label{NewVMLabelName(), cid.AsString()}
}

func NewVMEnvGroupLabel(group apiv1.VMEnvGroup) Label {
	val := group.AsString()

	if len(val) > 63 {
		h := md5.New()
		_, err := io.WriteString(h, val)
		if err != nil {
			panic("Unexpected io.WriteString error during label truncation")
		}
		val = fmt.Sprintf("bosh-md5-label-%x", h.Sum(nil))
	}

	return Label{NewVMEnvGroupName(), val}
}

func NewCustomLabel(name, value string) Label {
	// todo any validation of labels and values?
	return Label{LabelName{"bosh.io", strings.ToLower(name)}, value}
}

func (l Label) Name() string  { return l.name.AsString() }
func (l Label) Value() string { return l.value }

func (l Label) AsListOpts() kmetav1.ListOptions {
	return kmetav1.ListOptions{LabelSelector: l.Name() + "=" + l.Value()}
}
