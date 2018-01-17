package apiv1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("VMCloudPropsImpl", func() {
	It("marshals given data", func() {
		cps := NewVMCloudPropsFromMap(map[string]interface{}{"cp1": "cp1-val"})

		bytes, err := json.Marshal(cps)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(bytes)).To(Equal(`{"cp1":"cp1-val"}`))

		wrongCPs := NewVMCloudPropsFromMap(map[string]interface{}{"func": func() {}})

		_, err = json.Marshal(wrongCPs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("json: error calling MarshalJSON for type apiv1.VMCloudPropsImpl: json: unsupported type: func()"))
	})

	It("does not allow unmarshaling", func() {
		var cloudProps VMCloudPropsImpl

		err := cloudProps.As(nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Expected to not convert VMCloudPropsImpl"))
	})
})
