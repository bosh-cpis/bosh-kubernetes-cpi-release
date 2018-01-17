package apiv1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("AgentEnv", func() {
	Describe("AttachPersistentDisk", func() {
		It("sets persistent disk path for given disk id", func() {
			agentEnv1JSON := `{
        "disks": {
          "system": "0",
          "ephemeral": "fake-ephemeral-path",
          "persistent": {
            "fake-persistent-id1": "fake-persistent-path1"
          }
        }
      }`

			agentEnv2JSON := `{
        "disks": {
          "system": "0",
          "ephemeral": "fake-ephemeral-path",
          "persistent": {
            "fake-persistent-id1": "fake-persistent-path1",
            "fake-persistent-id2": "fake-persistent-path2"
          }
        }
      }`

			agentEnv1, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv1JSON))
			Expect(err).ToNot(HaveOccurred())

			agentEnv1.AttachPersistentDisk(NewDiskCID("fake-persistent-id2"), "fake-persistent-path2")

			agentEnv2, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv2JSON))
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnv1).To(Equal(agentEnv2))
		})

		It("sets persistent disk path for given disk id on an empty agent env", func() {
			agentEnv1JSON := `{ "disks": {} }`

			agentEnv2JSON := `{
        "disks": { "persistent": { "fake-persistent-id1": "fake-persistent-path1" } }
      }`

			agentEnv1, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv1JSON))
			Expect(err).ToNot(HaveOccurred())

			agentEnv1.AttachPersistentDisk(NewDiskCID("fake-persistent-id1"), "fake-persistent-path1")

			agentEnv2, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv2JSON))
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnv1).To(Equal(agentEnv2))
		})

		It("overwrites persistent disk path for given disk id", func() {
			agentEnv1JSON := `{
        "disks": { "persistent": { "fake-persistent-id1": "fake-persistent-path1" } }
      }`

			agentEnv2JSON := `{
        "disks": { "persistent": { "fake-persistent-id1": "fake-persistent-path2" } }
      }`

			agentEnv1, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv1JSON))
			Expect(err).ToNot(HaveOccurred())

			agentEnv1.AttachPersistentDisk(NewDiskCID("fake-persistent-id1"), "fake-persistent-path2")

			agentEnv2, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv2JSON))
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnv1).To(Equal(agentEnv2))
		})
	})

	Describe("DetachPersistentDisk", func() {
		It("unsets persistent disk path if previously set", func() {
			agentEnv1JSON := `{
        "disks": { "persistent": { "fake-persistent-id1": "fake-persistent-path1" } }
      }`

			agentEnv2JSON := `{
        "disks": { "persistent": { } }
      }`

			agentEnv1, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv1JSON))
			Expect(err).ToNot(HaveOccurred())

			agentEnv1.DetachPersistentDisk(NewDiskCID("fake-persistent-id1"))

			agentEnv2, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv2JSON))
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnv1).To(Equal(agentEnv2))
		})

		It("does not change anything if persistent disk was not set", func() {
			agentEnv1JSON := `{
        "disks": { "persistent": { "fake-persistent-id1": "fake-persistent-path1" } }
      }`

			agentEnv1, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv1JSON))
			Expect(err).ToNot(HaveOccurred())

			agentEnv1.DetachPersistentDisk(NewDiskCID("fake-persistent-id2"))

			agentEnv2, err := AgentEnvFactory{}.FromBytes([]byte(agentEnv1JSON))
			Expect(err).ToNot(HaveOccurred())

			Expect(agentEnv1).To(Equal(agentEnv2))
		})
	})

	Describe("AsBytes", func() {
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
            "preconfigured": true
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

			agentEnv1 := AgentEnvFactory{}.ForVM(
				NewAgentID("fake-agent-id"), NewVMCID("fake-vm-id"), networks, env, agentOptions)

			agentEnv1.AttachSystemDisk("fake-system-path")
			agentEnv1.AttachEphemeralDisk("fake-ephemeral-path")
			agentEnv1.AttachPersistentDisk(NewDiskCID("fake-persistent-id1"), "fake-persistent-path1")
			agentEnv1.AttachPersistentDisk(NewDiskCID("fake-persistent-id2"), "fake-persistent-path2")

			agentEnv1JSON, err := agentEnv1.AsBytes()
			Expect(err).ToNot(HaveOccurred())

			var serialized1, serialized2 interface{}

			err = json.Unmarshal([]byte(agentEnv1JSON), &serialized1)
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal([]byte(agentEnvJSON), &serialized2)
			Expect(err).ToNot(HaveOccurred())

			Expect(serialized1).To(Equal(serialized2))
		})
	})
})
