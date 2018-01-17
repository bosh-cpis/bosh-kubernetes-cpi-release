package testlib

import (
	"bytes"
	"encoding/json"
	"fmt"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cppforlife/bosh-cpi-go/apiv1"
	"github.com/cppforlife/bosh-cpi-go/rpc"
	. "github.com/onsi/gomega"
)

type CPI struct {
	cpiFactory apiv1.CPIFactory
	logger     boshlog.Logger
}

func NewCPI(cpiFactory apiv1.CPIFactory, logger boshlog.Logger) CPI {
	return CPI{cpiFactory, logger}
}

func (c CPI) Exec(req CPIRequest) CPIResponse {
	resp := c.ExecWithoutErrorCheck(req)
	Expect(resp.HasError()).To(
		BeFalse(), fmt.Sprintf("Expected response '%#v' to be successful", resp))

	return resp
}

func (c CPI) ExecWithoutErrorCheck(req CPIRequest) CPIResponse {
	req = req.Backfill()

	inBs, err := json.Marshal(req)
	Expect(err).ToNot(HaveOccurred())

	resp := CPIResponse{}

	err = json.Unmarshal(c.execBytes(inBs), &resp)
	Expect(err).ToNot(HaveOccurred())

	return resp
}

func (c CPI) execBytes(inBs []byte) []byte {
	in := bytes.NewBuffer(inBs)
	out := bytes.NewBuffer([]byte{})

	rpcFactory := rpc.NewFactory(c.logger)
	cli := rpcFactory.NewCLIWithInOut(in, out, c.cpiFactory)

	Expect(cli.ServeOnce()).ToNot(HaveOccurred())
	return out.Bytes()
}
