package integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("Disks", func() {
	Describe("create_disk", func() {
		It("can create disk without any cloud properties", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_disk",
				Arguments: []interface{}{
					2000,
					map[string]interface{}{},
					nil,
				},
			})
			Expect(strings.HasPrefix(resp.StringResult(), "disk-")).To(
				BeTrue(), "disk does not have nice cid prefix")

			By("checking created pvc", func() {
				pvc, err := cpi.Kube().PVCs().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				Expect(pvc.ObjectMeta.Name).To(Equal(resp.StringResult()))
				Expect(pvc.ObjectMeta.Labels).To(Equal(map[string]string{
					"bosh.io/disk-cid": resp.StringResult(),
				}))
				Expect(pvc.ObjectMeta.Annotations["volume.beta.kubernetes.io/storage-class"]).To(Equal("standard"))
				Expect(pvc.ObjectMeta.Annotations["volume.beta.kubernetes.io/storage-provisioner"]).ToNot(BeEmpty())

				KubeObj{pvc}.Find("spec").Remove("volumeName").ExpectToEqual(map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{"storage": "2000Mi"},
					},
					"accessModes": []string{"ReadWriteOnce"},
				})

				KubeObj{pvc}.Find("status").ExpectToEqual(map[string]interface{}{
					"phase":       "Bound",
					"capacity":    map[string]interface{}{"storage": "2000Mi"},
					"accessModes": []string{"ReadWriteOnce"},
				})
			})

			By("clean up", func() {
				cpi.Exec(testlib.CPIRequest{
					Method:    "delete_disk",
					Arguments: []interface{}{resp.StringResult()},
				})

				_, err := cpi.Kube().PVCs().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).To(HaveOccurred())
				Expect(kerrors.IsNotFound(err)).To(BeTrue())
			})
		})

		It("can create disk with cloud properties", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_disk",
				Arguments: []interface{}{
					2000,
					map[string]interface{}{
						"labels": map[string]string{
							"lbl1_name": "lbl1_val",
							"lbl2_name": "lbl2_val",
						},
						"annotations": map[string]string{
							"ann1_name": "ann1_val",
							"ann2_name": "ann2_val",
						},
						"resources": map[string]interface{}{
							"requests": map[string]interface{}{
								"storage": "4000Mi", // gets replaced by above size
								"custom":  "100",
							},
						},
						// todo test custom storage*
						// "storage_class": "custom-storage-class",
						// "storage_provisioner": "custom-storage-provisioner",
					},
					nil,
				},
			})

			By("checking created pvc", func() {
				pvc, err := cpi.Kube().PVCs().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				Expect(pvc.ObjectMeta.Name).To(Equal(resp.StringResult()))
				Expect(pvc.ObjectMeta.Labels).To(Equal(map[string]string{
					"bosh.io/disk-cid": resp.StringResult(),
					"lbl1_name":        "lbl1_val",
					"lbl2_name":        "lbl2_val",
				}))
				Expect(pvc.ObjectMeta.Annotations["volume.beta.kubernetes.io/storage-class"]).To(Equal("standard"))
				Expect(pvc.ObjectMeta.Annotations["volume.beta.kubernetes.io/storage-provisioner"]).ToNot(BeEmpty())
				Expect(pvc.ObjectMeta.Annotations["ann1_name"]).To(Equal("ann1_val"))
				Expect(pvc.ObjectMeta.Annotations["ann2_name"]).To(Equal("ann2_val"))

				KubeObj{pvc}.Find("spec").Remove("volumeName").ExpectToEqual(map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": "2000Mi",
							"custom":  "100",
						},
					},
					"accessModes": []string{"ReadWriteOnce"},
				})

				KubeObj{pvc}.Find("status").ExpectToEqual(map[string]interface{}{
					"phase":       "Bound",
					"capacity":    map[string]interface{}{"storage": "2000Mi"}, // todo should custom be here?
					"accessModes": []string{"ReadWriteOnce"},
				})
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{resp.StringResult()},
			})
		})
	})

	Describe("delete_disk", func() {
		It("can delete non-existent disk", func() {
			cpi := NewCPI()

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{"fake-does-not-exist"},
			})
		})
	})

	XDescribe("set_disk_metadata", func() {
		It("succeeds setting annotations", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "create_disk",
				Arguments: []interface{}{2000, map[string]interface{}{}, nil},
			})

			cpi.Exec(testlib.CPIRequest{
				Method: "set_disk_metadata",
				Arguments: []interface{}{
					resp.StringResult(),
					map[string]interface{}{
						"meta-ann1": "meta-ann1_value",
						"meta-ann2": "meta-ann2_value",
					},
				},
			})

			By("checking pvc", func() {
				pvc, err := cpi.Kube().PVCs().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				Expect(pvc.ObjectMeta.Annotations).To(Equal(map[string]string{
					"bosh.io/meta-ann1_name": "meta-ann1_val",
					"bosh.io/meta-ann2_name": "meta-ann2_val",
				}))
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{resp.StringResult()},
			})
		})
	})

	Describe("has_disk", func() {
		It("returns true if disk exists", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_disk",
				Arguments: []interface{}{
					2000,
					map[string]interface{}{},
					nil,
				},
			})

			resp2 := cpi.Exec(testlib.CPIRequest{
				Method:    "has_disk",
				Arguments: []interface{}{resp.StringResult()},
			})
			Expect(resp2.BoolResult()).To(BeTrue())

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{resp.StringResult()},
			})
		})

		It("returns false if disk does not exists", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "has_disk",
				Arguments: []interface{}{"fake-does-not-exist"},
			})
			Expect(resp.BoolResult()).To(BeFalse())
		})
	})
})
