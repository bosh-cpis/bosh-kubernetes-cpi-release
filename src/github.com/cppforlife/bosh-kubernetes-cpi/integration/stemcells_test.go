package integration_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("Stemcells", func() {
	Describe("create_stemcell", func() {
		It("can create light stemcell", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_stemcell",
				Arguments: []interface{}{
					"/tmp/bosh-kube-cpi-doesnt-exist",
					map[string]interface{}{"image": "fake-image-id"},
				},
			})
			Expect(strings.HasPrefix(resp.StringResult(), "scl-")).To(
				BeTrue(), "stemcell does not have light stemcell prefix")
			Expect(strings.Split(resp.StringResult(), "/")[1]).To(
				Equal("fake-image-id"), "stemcell does not have image id")

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_stemcell",
				Arguments: []interface{}{resp.StringResult()},
			})
		})

		XIt("create heavy stemcell via registry upload", func() {})

		XIt("create heavy stemcell via docker import", func() {})
	})

	Describe("delete_stemcell", func() {
		It("can delete non-existent light stemcell", func() {
			cpi := NewCPI()

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_stemcell",
				Arguments: []interface{}{"scl-123/fake-image-id"},
			})
		})
	})
})
