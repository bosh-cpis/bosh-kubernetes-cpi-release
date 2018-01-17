package apiv1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cppforlife/bosh-cpi-go/apiv1"
)

var _ = Describe("AgentOptions", func() {
	var (
		opts AgentOptions
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			opts = AgentOptions{
				Mbus: "fake-mbus",
				NTP:  []string{},

				Blobstore: BlobstoreOptions{
					Type: "fake-blobstore-type",
				},
			}
		})

		It("does not return error if all fields are valid", func() {
			err := opts.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if Mbus is empty", func() {
			opts.Mbus = ""

			err := opts.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty Mbus"))
		})

		It("returns error if blobstore section is not valid", func() {
			opts.Blobstore.Type = ""

			err := opts.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Validating Blobstore configuration"))
		})
	})
})

var _ = Describe("BlobstoreOptions", func() {
	var (
		opts BlobstoreOptions
	)

	Describe("Validate", func() {
		BeforeEach(func() {
			opts = BlobstoreOptions{
				Type:    "fake-type",
				Options: map[string]interface{}{"fake-key": "fake-value"},
			}
		})

		It("does not return error if all fields are valid", func() {
			err := opts.Validate()
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if Type is empty", func() {
			opts.Type = ""

			err := opts.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Must provide non-empty Type"))
		})
	})
})
