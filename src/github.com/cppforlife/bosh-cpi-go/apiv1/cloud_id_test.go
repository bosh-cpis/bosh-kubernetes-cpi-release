package apiv1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("CloudID", func() {
	It("marshals into json", func() {
		cid := NewCloudID("cid")
		bytes, err := json.Marshal(cid)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(Equal([]byte(`"cid"`)))
	})

	It("unmarshals from json", func() {
		cid := NewCloudID("cid")
		err := json.Unmarshal([]byte(`"cid2"`), &cid)
		Expect(err).ToNot(HaveOccurred())
		Expect(cid).To(Equal(NewCloudID("cid2")))
	})

	It("returns error if cid is empty", func() {
		cid := NewCloudID("cid")
		err := json.Unmarshal([]byte(`""`), &cid)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("Expected CID to be non-empty"))
	})

	It("returns error if is not a string", func() {
		cid := NewCloudID("cid")
		err := json.Unmarshal([]byte(`123`), &cid)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("json: cannot unmarshal number into Go value of type string"))
	})

	It("returns error if unmarshaling fails due to empty input", func() {
		cid := NewCloudID("cid")
		err := json.Unmarshal([]byte{}, &cid)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("unexpected end of JSON input"))
	})
})
