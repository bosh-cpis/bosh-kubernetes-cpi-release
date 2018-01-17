package integration_test

import (
	"fmt"
	"strings"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
	. "github.com/onsi/gomega"
)

type KubeObj struct {
	Obj interface{}
}

func (o KubeObj) Find(path string) KubeObj {
	objNormalized := testlib.RemarshalWithJSON(o.Obj)
	return KubeObj{o.findPiece(strings.Split(path, "."), objNormalized)}
}

func (o KubeObj) Remove(field string) KubeObj {
	objNormalized := testlib.RemarshalWithJSON(o.Obj)

	m, ok := objNormalized.(map[string]interface{})
	Expect(ok).To(BeTrue(), fmt.Sprintf("Expected interface '%#v' to be a map", objNormalized))

	delete(m, field)

	return KubeObj{m}
}

func (o KubeObj) ExpectToEqual(expected interface{}) {
	objNormalized := testlib.RemarshalWithJSON(o.Obj)
	expectedNormalized := testlib.RemarshalWithJSON(expected)
	Expect(objNormalized).To(BeEquivalentTo(expectedNormalized))
}

func (o KubeObj) findPiece(path []string, i interface{}) interface{} {
	if len(path) == 0 {
		return i
	}

	m, ok := i.(map[string]interface{})
	Expect(ok).To(BeTrue(), fmt.Sprintf("Expected interface '%#v' to be a map", i))

	return o.findPiece(path[1:], m[path[0]])
}
