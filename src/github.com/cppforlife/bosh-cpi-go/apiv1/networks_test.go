package apiv1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("Network", func() {
	var (
		networks                     Networks
		networkA, networkB, networkC Network
	)

	BeforeEach(func() {
		networks = map[string]Network{}
		networkA = NewNetwork(NetworkOpts{Type: "A"})
		networkB = NewNetwork(NetworkOpts{Type: "B"})
		networkC = NewNetwork(NetworkOpts{Type: "C"})
	})

	Describe("Default", func() {
		Context("when there are no networks defined", func() {
			It("returns and empty network", func() {
				Expect(networks.Default()).To(Equal(NewNetwork(NetworkOpts{})))
			})
		})

		Context("when is only one network defined", func() {
			It("returns the deinfed network", func() {
				networks["A"] = networkA
				Expect(networks.Default()).To(Equal(networkA))
			})
		})

		Context("when are multiple networks defined", func() {
			var allNetworks []Network

			It("returns the one that has the default gateway", func() {
				networkB = NewNetwork(NetworkOpts{Type: "B", Default: []string{"gateway"}})
				networks["A"] = networkA
				networks["B"] = networkB
				networks["C"] = networkC
				Expect(networks.Default()).To(Equal(networkB))
			})

			It("returns the one that has the default gateway even with others have other defaults", func() {
				networkA = NewNetwork(NetworkOpts{Type: "A", Default: []string{"dns"}})
				networkB = NewNetwork(NetworkOpts{Type: "B", Default: []string{"other"}})
				networkC = NewNetwork(NetworkOpts{Type: "C", Default: []string{"gateway"}})
				networks["A"] = networkA
				networks["B"] = networkB
				networks["C"] = networkC
				Expect(networks.Default()).To(Equal(networkC))
			})

			It("returns one of the networks when none have the default gateway set", func() {
				networks["A"] = networkA
				networks["B"] = networkB
				networks["C"] = networkC
				for _, net := range networks {
					allNetworks = append(allNetworks, net)
				}
				Expect(allNetworks).To(ContainElement(networks.Default()))
			})

			It("returns one of the networks when none have the default gateway set but have other defaults", func() {
				networkA = NewNetwork(NetworkOpts{Type: "A", Default: []string{"dns", "foo"}})
				networkB = NewNetwork(NetworkOpts{Type: "B", Default: []string{"bar"}})
				networks["A"] = networkA
				networks["B"] = networkB
				networks["C"] = networkC
				for _, net := range networks {
					allNetworks = append(allNetworks, net)
				}
				Expect(allNetworks).To(ContainElement(networks.Default()))
			})
		})
	})

	Describe("BackfillDefaultDNS", func() {
		Context("when default network for DNS already has DNS servers", func() {
			It("sets DNS servers on that network", func() {
				networkA = NewNetwork(NetworkOpts{Type: "A", Default: []string{"dns", "foo"}})
				networks["A"] = networkA

				networkB = NewNetwork(NetworkOpts{Type: "B", Default: []string{"gateway"}})
				networks["B"] = networkB

				networks.BackfillDefaultDNS([]string{"8.8.8.8", "4.4.4.4"})

				Expect(networks).To(Equal(Networks{
					"A": NewNetwork(NetworkOpts{
						Type:    "A",
						Default: []string{"dns", "foo"},
						DNS:     []string{"8.8.8.8", "4.4.4.4"},
					}),
					"B": NewNetwork(NetworkOpts{
						Type:    "B",
						Default: []string{"gateway"},
					}),
				}))
			})
		})

		Context("when default network for DNS already has DNS servers", func() {
			It("keeps already set DNS servers", func() {
				networkA = NewNetwork(NetworkOpts{Type: "A", Default: []string{"gateway"}})
				networks["A"] = networkA

				networkB = NewNetwork(NetworkOpts{
					Type:    "B",
					Default: []string{"dns", "foo"},
					DNS:     []string{"127.0.0.1"},
				})
				networks["B"] = networkB

				networks.BackfillDefaultDNS([]string{"8.8.8.8", "4.4.4.4"})

				Expect(networks).To(Equal(Networks{
					"A": NewNetwork(NetworkOpts{
						Type:    "A",
						Default: []string{"gateway"},
					}),
					"B": NewNetwork(NetworkOpts{
						Type:    "B",
						Default: []string{"dns", "foo"},
						DNS:     []string{"127.0.0.1"},
					}),
				}))
			})
		})

		Context("when there is no default network for DNS", func() {
			It("does not do anything", func() {
				networkA = NewNetwork(NetworkOpts{Type: "A", Default: []string{"foo"}})
				networks["A"] = networkA

				networkB = NewNetwork(NetworkOpts{
					Type:    "B",
					Default: []string{"gateway"},
					DNS:     []string{"127.0.0.1"},
				})
				networks["B"] = networkB

				networks.BackfillDefaultDNS([]string{"8.8.8.8", "4.4.4.4"})

				Expect(networks).To(Equal(Networks{
					"A": NewNetwork(NetworkOpts{
						Type:    "A",
						Default: []string{"foo"},
					}),
					"B": NewNetwork(NetworkOpts{
						Type:    "B",
						Default: []string{"gateway"},
						DNS:     []string{"127.0.0.1"},
					}),
				}))
			})
		})
	})

	Describe("IsDynamic", func() {
		It("returns true if the type is 'dynamic'", func() {
			Expect(NewNetwork(NetworkOpts{Type: "A"}).IsDynamic()).To(BeFalse())
			Expect(NewNetwork(NetworkOpts{Type: "manual"}).IsDynamic()).To(BeFalse())
			Expect(NewNetwork(NetworkOpts{Type: "Dynamic"}).IsDynamic()).To(BeFalse())
			Expect(NewNetwork(NetworkOpts{Type: "dynamic"}).IsDynamic()).To(BeTrue())
		})
	})

	Describe("IPWithSubnetMask", func() {
		It("returns 12.18.3.4/24 when IP is 12.18.3.4 and netmask is 255.255.255.0", func() {
			net := NewNetwork(NetworkOpts{IP: "12.18.3.4", Netmask: "255.255.255.0"})
			Expect(net.IPWithSubnetMask()).To(Equal("12.18.3.4/24"))
		})

		It("returns fd7e:964d:32c6:777c:0000:0000:0000:0006/64 when IP is "+
			"fd7e:964d:32c6:777c:0000:0000:0000:0006 and netmask is ffff:ffff:ffff:ffff:0000:0000:0000:0000", func() {
			net := NewNetwork(NetworkOpts{
				IP:      "fd7e:964d:32c6:777c:0000:0000:0000:0006",
				Netmask: "ffff:ffff:ffff:ffff:0000:0000:0000:0000",
			})
			Expect(net.IPWithSubnetMask()).To(Equal("fd7e:964d:32c6:777c:0000:0000:0000:0006/64"))
		})
	})
})
