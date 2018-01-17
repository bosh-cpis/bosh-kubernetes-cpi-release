package integration_test

import (
	"encoding/json"

	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

// https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct-golang
func yaml2json(b []byte) []byte {
	var val interface{}

	err := yaml.Unmarshal(b, &val)
	Expect(err).ToNot(HaveOccurred())

	val = stripInterfaceKeys(val)

	b, err = json.Marshal(val)
	Expect(err).ToNot(HaveOccurred())

	return b
}

func stripInterfaceKeys(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = stripInterfaceKeys(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = stripInterfaceKeys(v)
		}
	}
	return i
}
