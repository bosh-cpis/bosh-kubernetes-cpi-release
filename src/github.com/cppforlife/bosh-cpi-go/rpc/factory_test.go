package rpc_test

import (
	"bytes"
	"encoding/json"
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cppforlife/bosh-cpi-go/apiv1"
	"github.com/cppforlife/bosh-cpi-go/apiv1/apiv1fakes"
	. "github.com/cppforlife/bosh-cpi-go/rpc"
)

type FakeCPs struct {
	CP string `json:"cp1"`
}

var _ = Describe("Factory", func() {
	var (
		cpi        *apiv1fakes.FakeCPI
		cpiFactory *apiv1fakes.FakeCPIFactory
		factory    Factory
	)

	BeforeEach(func() {
		cpi = &apiv1fakes.FakeCPI{}
		cpiFactory = &apiv1fakes.FakeCPIFactory{
			NewStub: func(_ apiv1.CallContext) (apiv1.CPI, error) { return cpi, nil },
		}

		logger := boshlog.NewLogger(boshlog.LevelNone)
		factory = NewFactory(logger)
	})

	act := func(inStr string) (Response, string) {
		in := bytes.NewBufferString(inStr + "\n")
		out := bytes.NewBufferString("")

		cli := factory.NewCLIWithInOut(in, out, cpiFactory)
		err := cli.ServeOnce()
		Expect(err).ToNot(HaveOccurred())

		outBytes := out.Bytes()
		var resp Response

		err = json.Unmarshal(outBytes, &resp)
		Expect(err).ToNot(HaveOccurred())

		return resp, string(outBytes)
	}

	Describe("cpi context", func() {
		It("works", func() {
			act(`{"method":"info", "arguments":[], "context": {"cp1": "cp1-val"}}`)

			ctx := cpiFactory.NewArgsForCall(0)

			var cp FakeCPs
			Expect(ctx.As(&cp)).ToNot(HaveOccurred())
			Expect(cp).To(Equal(FakeCPs{CP: "cp1-val"}))
		})
	})

	Describe("info", func() {
		It("works", func() {
			cpi.InfoReturns(apiv1.Info{
				StemcellFormats: []string{"stemcell-fmt"},
			}, nil)

			resp, _ := act(`{"method":"info", "arguments":[]}`)
			Expect(resp).To(Equal(Response{
				Result: map[string]interface{}{
					"stemcell_formats": []interface{}{"stemcell-fmt"},
				},
			}))
		})

		It("errs", func() {
			cpi.InfoReturns(apiv1.Info{}, errors.New("err"))

			resp, _ := act(`{"method":"info", "arguments":[]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("create_stemcell", func() {
		It("works", func() {
			cpi.CreateStemcellReturns(apiv1.NewStemcellCID("stemcell-cid"), nil)

			resp, _ := act(`{"method":"create_stemcell", "arguments":["/image-path", {"cp1": "cp1-val"}]}`)
			Expect(resp).To(Equal(Response{Result: "stemcell-cid"}))

			imagePath, cloudProps := cpi.CreateStemcellArgsForCall(0)
			Expect(imagePath).To(Equal("/image-path"))

			var cp FakeCPs
			Expect(cloudProps.As(&cp)).ToNot(HaveOccurred())
			Expect(cp).To(Equal(FakeCPs{CP: "cp1-val"}))
		})

		It("errs", func() {
			cpi.CreateStemcellReturns(apiv1.StemcellCID{}, errors.New("err"))

			resp, _ := act(`{"method":"create_stemcell", "arguments":["/image-path", {"cp1": "cp1-val"}]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("delete_stemcell", func() {
		It("works", func() {
			cpi.DeleteStemcellReturns(nil)

			resp, _ := act(`{"method":"delete_stemcell", "arguments":["stemcell-cid"]}`)
			Expect(resp).To(Equal(Response{}))

			cid := cpi.DeleteStemcellArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewStemcellCID("stemcell-cid")))
		})

		It("errs", func() {
			cpi.DeleteStemcellReturns(errors.New("err"))

			resp, _ := act(`{"method":"delete_stemcell", "arguments":["stemcell-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("create_vm", func() {
		It("works without associated disks", func() {
			cpi.CreateVMReturns(apiv1.NewVMCID("vm-cid"), nil)

			resp, _ := act(`{"method":"create_vm", "arguments":[
        "agent-id", "stemcell-cid", {"cp1": "cp1-val"}, 
        {}, [], {"env1": "env1-val"}]}`)
			Expect(resp).To(Equal(Response{Result: "vm-cid"}))

			agentID, stemcellCID, cloudProps, nets, diskCIDs, env := cpi.CreateVMArgsForCall(0)
			Expect(agentID).To(Equal(apiv1.NewAgentID("agent-id")))
			Expect(stemcellCID).To(Equal(apiv1.NewStemcellCID("stemcell-cid")))
			Expect(nets).To(Equal(apiv1.Networks{}))
			Expect(diskCIDs).To(Equal([]apiv1.DiskCID{}))
			Expect(env).To(Equal(apiv1.NewVMEnv(map[string]interface{}{"env1": "env1-val"})))

			var cp FakeCPs
			Expect(cloudProps.As(&cp)).ToNot(HaveOccurred())
			Expect(cp).To(Equal(FakeCPs{CP: "cp1-val"}))
		})

		It("works with networks", func() {
			cpi.CreateVMReturns(apiv1.NewVMCID("vm-cid"), nil)

			resp, _ := act(`{"method":"create_vm", "arguments":[
        "agent-id", "stemcell-cid", {}, {
          "net1": {
            "type": "net1-type",
            "ip": "net1-ip",
            "netmask": "net1-netmask",
            "gateway": "net1-gateway",
            "dns": [],
            "default": [],
            "cloud_properties": {"cp1": "net1-cp1-val"}
          },
          "net2": {
            "type": "net2-type",
            "dns": ["net2-dns"],
            "default": ["net2-default"],
            "cloud_properties": {"cp1": "net2-cp1-val"}
          }
        }, [], {}]}`)
			Expect(resp).To(Equal(Response{Result: "vm-cid"}))

			_, _, _, nets, _, _ := cpi.CreateVMArgsForCall(0)

			net1 := nets["net1"]

			Expect(net1.Type()).To(Equal("net1-type"))
			Expect(net1.IP()).To(Equal("net1-ip"))
			Expect(net1.Netmask()).To(Equal("net1-netmask"))
			Expect(net1.Gateway()).To(Equal("net1-gateway"))
			Expect(net1.DNS()).To(Equal([]string{}))
			Expect(net1.Default()).To(Equal([]string{}))

			var net1CP FakeCPs
			Expect(net1.CloudProps().As(&net1CP)).ToNot(HaveOccurred())
			Expect(net1CP).To(Equal(FakeCPs{CP: "net1-cp1-val"}))

			net2 := nets["net2"]

			Expect(net2.Type()).To(Equal("net2-type"))
			Expect(net2.IP()).To(Equal(""))
			Expect(net2.Netmask()).To(Equal(""))
			Expect(net2.Gateway()).To(Equal(""))
			Expect(net2.DNS()).To(Equal([]string{"net2-dns"}))
			Expect(net2.Default()).To(Equal([]string{"net2-default"}))

			var net2CP FakeCPs
			Expect(net2.CloudProps().As(&net2CP)).ToNot(HaveOccurred())
			Expect(net2CP).To(Equal(FakeCPs{CP: "net2-cp1-val"}))
		})

		It("works with associated disks", func() {
			cpi.CreateVMReturns(apiv1.NewVMCID("vm-cid"), nil)

			resp, _ := act(`{"method":"create_vm", "arguments":[
        "agent-id", "stemcell-cid", {"cp1": "cp1-val"}, 
        {"net1": {}}, ["disk-cid1", "disk-cid2"], {"env1": "env1-val"}]}`)
			Expect(resp).To(Equal(Response{Result: "vm-cid"}))

			_, _, _, _, diskCIDs, _ := cpi.CreateVMArgsForCall(0)
			Expect(diskCIDs).To(Equal([]apiv1.DiskCID{
				apiv1.NewDiskCID("disk-cid1"),
				apiv1.NewDiskCID("disk-cid2"),
			}))
		})

		It("works even if associated disks are null", func() {
			cpi.CreateVMReturns(apiv1.NewVMCID("vm-cid"), nil)

			resp, _ := act(`{"method":"create_vm", "arguments":[
        "agent-id", "stemcell-cid", {"cp1": "cp1-val"}, 
        {"net1": {}}, null, {"env1": "env1-val"}]}`)
			Expect(resp).To(Equal(Response{Result: "vm-cid"}))

			_, _, _, _, diskCIDs, _ := cpi.CreateVMArgsForCall(0)
			Expect(diskCIDs).To(HaveLen(0))
		})

		It("errs", func() {
			cpi.CreateVMReturns(apiv1.VMCID{}, errors.New("err"))

			resp, _ := act(`{"method":"create_vm", "arguments":[
        "agent-id", "stemcell-cid", {"cp1": "cp1-val"}, 
        {"net1": {}}, ["disk-cid1", "disk-cid2"], {"env1": "env1-val"}]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("delete_vm", func() {
		It("works", func() {
			cpi.DeleteVMReturns(nil)

			resp, _ := act(`{"method":"delete_vm", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Result: nil}))

			cid := cpi.DeleteVMArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewVMCID("vm-cid")))
		})

		It("errs", func() {
			cpi.DeleteVMReturns(errors.New("err"))

			resp, _ := act(`{"method":"delete_vm", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("calculate_vm_cloud_properties", func() {
		It("works", func() {
			cpi.CalculateVMCloudPropertiesReturns(nil, nil)

			resp, _ := act(`{"method":"calculate_vm_cloud_properties", "arguments":[{"ram": 123, "cpu": 1, "ephemeral_disk_size": 1000}]}`)
			Expect(resp).To(Equal(Response{Result: nil}))

			vmRes := cpi.CalculateVMCloudPropertiesArgsForCall(0)
			Expect(vmRes).To(Equal(apiv1.VMResources{
				RAM:               123,
				CPU:               1,
				EphemeralDiskSize: 1000,
			}))
		})

		It("errs", func() {
			cpi.CalculateVMCloudPropertiesReturns(nil, errors.New("err"))

			resp, _ := act(`{"method":"calculate_vm_cloud_properties", "arguments":[{}]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("set_vm_metadata", func() {
		It("works", func() {
			cpi.SetVMMetadataReturns(nil)

			resp, _ := act(`{"method":"set_vm_metadata", "arguments":["vm-cid", {"meta1": "meta1-val"}]}`)
			Expect(resp).To(Equal(Response{Result: nil}))

			cid, meta := cpi.SetVMMetadataArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewVMCID("vm-cid")))
			Expect(meta).To(Equal(apiv1.NewVMMeta(map[string]interface{}{"meta1": "meta1-val"})))
		})

		It("errs", func() {
			cpi.SetVMMetadataReturns(errors.New("err"))

			resp, _ := act(`{"method":"set_vm_metadata", "arguments":["vm-cid", {}]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("has_vm", func() {
		It("works", func() {
			cpi.HasVMReturns(true, nil)

			resp, _ := act(`{"method":"has_vm", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Result: true}))

			cid := cpi.HasVMArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewVMCID("vm-cid")))
		})

		It("errs", func() {
			cpi.HasVMReturns(false, errors.New("err"))

			resp, _ := act(`{"method":"has_vm", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("reboot_vm", func() {
		It("works", func() {
			cpi.RebootVMReturns(nil)

			resp, _ := act(`{"method":"reboot_vm", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Result: ""}))

			cid := cpi.RebootVMArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewVMCID("vm-cid")))
		})

		It("errs", func() {
			cpi.RebootVMReturns(errors.New("err"))

			resp, _ := act(`{"method":"reboot_vm", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("get_disks", func() {
		It("works with empty result", func() {
			cpi.GetDisksReturns(nil, nil)

			resp, _ := act(`{"method":"get_disks", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Result: []interface{}{}})) // empty array, not nil

			cid := cpi.GetDisksArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewVMCID("vm-cid")))
		})

		It("works with non-empty result", func() {
			cpi.GetDisksReturns([]apiv1.DiskCID{
				apiv1.NewDiskCID("disk-cid1"),
				apiv1.NewDiskCID("disk-cid2"),
			}, nil)

			resp, _ := act(`{"method":"get_disks", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Result: []interface{}{"disk-cid1", "disk-cid2"}}))

			cid := cpi.GetDisksArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewVMCID("vm-cid")))
		})

		It("errs", func() {
			cpi.GetDisksReturns(nil, errors.New("err"))

			resp, _ := act(`{"method":"get_disks", "arguments":["vm-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("create_disk", func() {
		It("works without associated VM", func() {
			cpi.CreateDiskReturns(apiv1.NewDiskCID("disk-cid"), nil)

			resp, _ := act(`{"method":"create_disk", "arguments":[123, {"cp1": "cp1-val"}, null]}`)
			Expect(resp).To(Equal(Response{Result: "disk-cid"}))

			size, cloudProps, vmCID := cpi.CreateDiskArgsForCall(0)
			Expect(size).To(Equal(123))
			Expect(vmCID).To(BeNil())

			var cp FakeCPs
			Expect(cloudProps.As(&cp)).ToNot(HaveOccurred())
			Expect(cp).To(Equal(FakeCPs{CP: "cp1-val"}))
		})

		It("works with associated VM", func() {
			cpi.CreateDiskReturns(apiv1.NewDiskCID("disk-cid"), nil)

			resp, _ := act(`{"method":"create_disk", "arguments":[123, {"cp1": "cp1-val"}, "vm-cid"]}`)
			Expect(resp).To(Equal(Response{Result: "disk-cid"}))

			size, cloudProps, vmCID := cpi.CreateDiskArgsForCall(0)
			Expect(size).To(Equal(123))
			Expect(*vmCID).To(Equal(apiv1.NewVMCID("vm-cid")))

			var cp FakeCPs
			Expect(cloudProps.As(&cp)).ToNot(HaveOccurred())
			Expect(cp).To(Equal(FakeCPs{CP: "cp1-val"}))
		})

		It("errs", func() {
			cpi.CreateDiskReturns(apiv1.DiskCID{}, errors.New("err"))

			resp, _ := act(`{"method":"create_disk", "arguments":[123, {"cp1": "cp1-val"}, null]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("delete_disk", func() {
		It("works", func() {
			cpi.DeleteDiskReturns(nil)

			resp, _ := act(`{"method":"delete_disk", "arguments":["disk-cid"]}`)
			Expect(resp).To(Equal(Response{Result: nil}))

			cid := cpi.DeleteDiskArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewDiskCID("disk-cid")))
		})

		It("errs", func() {
			cpi.DeleteDiskReturns(errors.New("err"))

			resp, _ := act(`{"method":"delete_disk", "arguments":["disk-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("attach_disk", func() {
		It("works", func() {
			cpi.AttachDiskReturns(nil)

			resp, _ := act(`{"method":"attach_disk", "arguments":["vm-cid", "disk-cid"]}`)
			Expect(resp).To(Equal(Response{Result: nil}))

			vmCID, diskCID := cpi.AttachDiskArgsForCall(0)
			Expect(vmCID).To(Equal(apiv1.NewVMCID("vm-cid")))
			Expect(diskCID).To(Equal(apiv1.NewDiskCID("disk-cid")))
		})

		It("errs", func() {
			cpi.AttachDiskReturns(errors.New("err"))

			resp, _ := act(`{"method":"attach_disk", "arguments":["vm-cid", "disk-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("detach_disk", func() {
		It("works", func() {
			cpi.DetachDiskReturns(nil)

			resp, _ := act(`{"method":"detach_disk", "arguments":["vm-cid", "disk-cid"]}`)
			Expect(resp).To(Equal(Response{Result: nil}))

			vmCID, diskCID := cpi.DetachDiskArgsForCall(0)
			Expect(vmCID).To(Equal(apiv1.NewVMCID("vm-cid")))
			Expect(diskCID).To(Equal(apiv1.NewDiskCID("disk-cid")))
		})

		It("errs", func() {
			cpi.DetachDiskReturns(errors.New("err"))

			resp, _ := act(`{"method":"detach_disk", "arguments":["vm-cid", "disk-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("has_disk", func() {
		It("works", func() {
			cpi.HasDiskReturns(true, nil)

			resp, _ := act(`{"method":"has_disk", "arguments":["disk-cid"]}`)
			Expect(resp).To(Equal(Response{Result: true}))

			cid := cpi.HasDiskArgsForCall(0)
			Expect(cid).To(Equal(apiv1.NewDiskCID("disk-cid")))
		})

		It("errs", func() {
			cpi.HasDiskReturns(false, errors.New("err"))

			resp, _ := act(`{"method":"has_disk", "arguments":["disk-cid"]}`)
			Expect(resp).To(Equal(Response{Error: &ResponseError{Type: "Bosh::Clouds::CloudError", Message: "err"}}))
		})
	})

	Describe("unknown methods", func() {
		It("errs", func() {
			resp, _ := act(`{"method":"unknown", "arguments":[]}`)
			Expect(resp).To(Equal(Response{
				Error: &ResponseError{
					Type:    "Bosh::Clouds::NotImplemented",
					Message: "Must call implemented method",
				},
			}))
		})
	})
})
