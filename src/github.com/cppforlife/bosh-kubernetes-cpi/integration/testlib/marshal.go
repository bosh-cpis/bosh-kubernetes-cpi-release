package testlib

import (
	"encoding/json"

	. "github.com/onsi/gomega"
)

func RemarshalWithJSON(in interface{}) interface{} {
	expectedBs, err := json.Marshal(in)
	Expect(err).ToNot(HaveOccurred())

	var out interface{}

	err = json.Unmarshal(expectedBs, &out)
	Expect(err).ToNot(HaveOccurred())

	return out
}
