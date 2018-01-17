package testlib

import (
	. "github.com/onsi/gomega"
)

type CPIRequest struct {
	Method    string                 `json:"method"`
	Arguments []interface{}          `json:"arguments"`
	Context   map[string]interface{} `json:"context"`
}

func (c CPIRequest) Backfill() CPIRequest {
	if c.Context == nil {
		c.Context = map[string]interface{}{}
	}
	return c
}

type CPIResponse struct {
	Result interface{}
	Error  map[string]interface{}
}

func (r CPIResponse) StringResult() string { return r.Result.(string) }
func (r CPIResponse) BoolResult() bool     { return r.Result.(bool) }

func (r CPIResponse) MapResult() map[string]interface{} {
	return r.Result.(map[string]interface{})
}

func (r CPIResponse) ExpectResultToEqual(ex interface{}) {
	// remarshal to conveniently match types of deserialized result
	Expect(r.Result).To(Equal(RemarshalWithJSON(ex)))
}

func (r CPIResponse) HasError() bool {
	return r.Error != nil
}

func (r CPIResponse) ErrorMessage() string {
	return r.Error["message"].(string)
}
