package apiv1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("CloudKVs", func() {
	It("marshals into json", func() {
		kvs := NewCloudKVs(map[string]interface{}{"kv1": "kv1-val"})
		bytes, err := json.Marshal(kvs)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(Equal([]byte(`{"kv1":"kv1-val"}`)))
	})

	It("unmarshals from json", func() {
		kvs := NewCloudKVs(map[string]interface{}{})
		err := json.Unmarshal([]byte(`{"kv2": "kv2-val"}`), &kvs)
		Expect(err).ToNot(HaveOccurred())
		Expect(kvs).To(Equal(NewCloudKVs(map[string]interface{}{"kv2": "kv2-val"})))
	})

	It("returns error if is not a map", func() {
		kvs := NewCloudKVs(map[string]interface{}{})
		err := json.Unmarshal([]byte(`123`), &kvs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("json: cannot unmarshal number into Go value of type map[string]interface {}"))
	})

	It("returns error if unmarshaling fails due to empty input", func() {
		kvs := NewCloudKVs(map[string]interface{}{})
		err := json.Unmarshal([]byte{}, &kvs)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal("unexpected end of JSON input"))
	})
})
