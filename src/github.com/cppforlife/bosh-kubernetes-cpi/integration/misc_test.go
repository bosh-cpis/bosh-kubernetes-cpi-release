package integration_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("Misc", func() {
	Describe("info", func() {
		It("returns supported stemcells", func() {
			cpi := NewCPI()

			req := testlib.CPIRequest{
				Method:    "info",
				Arguments: []interface{}{},
			}
			resp := cpi.Exec(req)
			resp.ExpectResultToEqual(map[string]interface{}{
				"stemcell_formats": []interface{}{"warden-tar", "general-tar"},
			})
		})
	})
})
