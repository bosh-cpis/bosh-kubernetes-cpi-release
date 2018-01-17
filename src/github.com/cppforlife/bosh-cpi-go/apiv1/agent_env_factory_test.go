package apiv1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("AgentEnvFactory", func() {
	Describe("ForVM/FromBytes", func() {
		It("works", func() {
			net1 := NewNetwork(NetworkOpts{
				Type: "fake-type",

				IP:      "fake-ip",
				Netmask: "fake-netmask",
				Gateway: "fake-gateway",

				DNS:     []string{"fake-dns"},
				Default: []string{"fake-default"},
			})

			net1.SetMAC("fake-mac")
			net1.SetPreconfigured()

			networks := Networks{"fake-net-name": net1}

			env := NewVMEnv(map[string]interface{}{"fake-env-key": "fake-env-value"})

			agentOptions := AgentOptions{
				Mbus: "fake-mbus",
				NTP:  []string{"fake-ntp"},

				Blobstore: BlobstoreOptions{
					Type: "fake-blobstore-type",
					Options: map[string]interface{}{
						"fake-blobstore-key": "fake-blobstore-value",
					},
				},
			}

			agentEnv1 := AgentEnvFactory{}.ForVM(
				NewAgentID("fake-agent-id"), NewVMCID("fake-vm-id"), networks, env, agentOptions)

			agentEnv1.AttachSystemDisk("fake-system-path")
			agentEnv1.AttachEphemeralDisk("fake-ephemeral-path")
			agentEnv1.AttachPersistentDisk(NewDiskCID("fake-persistent-id1"), "fake-persistent-path1")
			agentEnv1.AttachPersistentDisk(NewDiskCID("fake-persistent-id2"), "fake-persistent-path2")

			agentEnvJSON := `{
        "agent_id": "fake-agent-id",

        "vm": {
          "name": "fake-vm-id",
          "id": "fake-vm-id"
        },

        "mbus": "fake-mbus",
        "ntp": ["fake-ntp"],

        "blobstore": {
          "provider": "fake-blobstore-type",
          "options": {
            "fake-blobstore-key": "fake-blobstore-value"
          }
        },

        "networks": {
          "fake-net-name": {
            "type":    "fake-type",
            "ip":      "fake-ip",
            "netmask": "fake-netmask",
            "gateway": "fake-gateway",

            "dns":     ["fake-dns"],
            "default": ["fake-default"],

            "mac": "fake-mac",
            "preconfigured": true,

            "cloud_properties": {"fake-cp-key": "fake-cp-value"}
          }
        },

        "disks": {
          "system": "fake-system-path",
          "ephemeral": "fake-ephemeral-path",
          "persistent": {
            "fake-persistent-id1": "fake-persistent-path1",
            "fake-persistent-id2": "fake-persistent-path2"
          }
        },

        "env": {"fake-env-key": "fake-env-value"}
      }`

			agentEnv2, err := AgentEnvFactory{}.FromBytes([]byte(agentEnvJSON))
			Expect(err).ToNot(HaveOccurred())
			Expect(agentEnv1).To(Equal(agentEnv2))
		})
	})

	Describe("FromBytes", func() {
		It("returns error when json is not valid", func() {
			_, err := AgentEnvFactory{}.FromBytes([]byte(`-`))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid character"))
		})
	})
})
