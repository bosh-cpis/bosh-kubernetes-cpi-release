package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("VMs and disks", func() {
	It("can attach disk to vm", func() {
		props := NewCPIProps()
		cpi := NewCPI()

		vmResp := cpi.Exec(testlib.CPIRequest{
			Method: "create_vm",
			Arguments: []interface{}{
				"agent-id",
				"scl-123/" + props.StemcellImage(),
				map[string]interface{}{"instance_type": "m1.small"},
				props.RandomNetwork(),
				nil,
				nil,
			},
		})

		diskResp := cpi.Exec(testlib.CPIRequest{
			Method:    "create_disk",
			Arguments: []interface{}{2000, map[string]interface{}{}, nil},
		})

		diskResp2 := cpi.Exec(testlib.CPIRequest{
			Method:    "create_disk",
			Arguments: []interface{}{3000, map[string]interface{}{}, nil},
		})

		By("checking no disks attached", func() {
			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "get_disks",
				Arguments: []interface{}{vmResp.StringResult()},
			})
			resp.ExpectResultToEqual([]string{})
		})

		By("attaching first disk", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "attach_disk",
				Arguments: []interface{}{vmResp.StringResult(), diskResp.StringResult()},
			})

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "get_disks",
				Arguments: []interface{}{vmResp.StringResult()},
			})
			resp.ExpectResultToEqual([]string{diskResp.StringResult()})
		})

		By("attaching second disk", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "attach_disk",
				Arguments: []interface{}{vmResp.StringResult(), diskResp2.StringResult()},
			})

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "get_disks",
				Arguments: []interface{}{vmResp.StringResult()},
			})
			resp.ExpectResultToEqual([]string{diskResp.StringResult(), diskResp2.StringResult()})
		})

		By("detaching first disk", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "detach_disk",
				Arguments: []interface{}{vmResp.StringResult(), diskResp.StringResult()},
			})

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "get_disks",
				Arguments: []interface{}{vmResp.StringResult()},
			})
			resp.ExpectResultToEqual([]string{diskResp2.StringResult()})
		})

		By("detaching second disk", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "detach_disk",
				Arguments: []interface{}{vmResp.StringResult(), diskResp2.StringResult()},
			})

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "get_disks",
				Arguments: []interface{}{vmResp.StringResult()},
			})
			resp.ExpectResultToEqual([]string{})
		})

		By("clean up", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_vm",
				Arguments: []interface{}{vmResp.StringResult()},
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{diskResp.StringResult()},
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{diskResp2.StringResult()},
			})
		})
	})

	It("does not delete persistent disks when deleting vm", func() {
		props := NewCPIProps()
		cpi := NewCPI()

		vmResp := cpi.Exec(testlib.CPIRequest{
			Method: "create_vm",
			Arguments: []interface{}{
				"agent-id",
				"scl-123/" + props.StemcellImage(),
				map[string]interface{}{},
				props.RandomNetwork(),
				nil,
				nil,
			},
		})

		diskResp := cpi.Exec(testlib.CPIRequest{
			Method:    "create_disk",
			Arguments: []interface{}{2000, map[string]interface{}{}, nil},
		})

		diskResp2 := cpi.Exec(testlib.CPIRequest{
			Method:    "create_disk",
			Arguments: []interface{}{3000, map[string]interface{}{}, nil},
		})

		By("attaching two disks", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "attach_disk",
				Arguments: []interface{}{vmResp.StringResult(), diskResp.StringResult()},
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "attach_disk",
				Arguments: []interface{}{vmResp.StringResult(), diskResp2.StringResult()},
			})

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "get_disks",
				Arguments: []interface{}{vmResp.StringResult()},
			})
			resp.ExpectResultToEqual([]string{diskResp.StringResult(), diskResp2.StringResult()})
		})

		By("deleting vm keeps disks", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_vm",
				Arguments: []interface{}{vmResp.StringResult()},
			})

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "has_disk",
				Arguments: []interface{}{diskResp.StringResult()},
			})
			Expect(resp.BoolResult()).To(BeTrue())

			resp = cpi.Exec(testlib.CPIRequest{
				Method:    "has_disk",
				Arguments: []interface{}{diskResp2.StringResult()},
			})
			Expect(resp.BoolResult()).To(BeTrue())
		})

		By("clean up", func() {
			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{diskResp.StringResult()},
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_disk",
				Arguments: []interface{}{diskResp2.StringResult()},
			})
		})
	})
})
