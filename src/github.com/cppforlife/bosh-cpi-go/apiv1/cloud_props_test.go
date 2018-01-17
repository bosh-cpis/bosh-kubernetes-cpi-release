package apiv1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

type FakeCPs struct {
	CP string `json:"cp1"`
}

var _ = Describe("CloudPropsImpl", func() {
	It("allows customized unmarshaling", func() {
		var cloudProps CloudPropsImpl

		err := json.Unmarshal([]byte(`{"cp1": "cp1-val"}`), &cloudProps)
		Expect(err).ToNot(HaveOccurred())

		var cp FakeCPs

		err = cloudProps.As(&cp)
		Expect(err).ToNot(HaveOccurred())
		Expect(cp).To(Equal(FakeCPs{CP: "cp1-val"}))

		var wrongCP string

		err = cloudProps.As(&wrongCP)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("json: cannot unmarshal object into Go value of type string"))
	})

	It("does not marshal into json", func() {
		_, err := json.Marshal(CloudPropsImpl{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Expected to not marshal CloudPropsImpl as JSON"))
	})
})
