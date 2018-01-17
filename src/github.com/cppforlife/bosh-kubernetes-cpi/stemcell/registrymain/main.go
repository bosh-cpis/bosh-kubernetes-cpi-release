package main

import (
	"fmt"

	"github.com/cppforlife/bosh-kubernetes-cpi/stemcell/registry"
)

func main() {
	// todo flags?
	reg := registry.NewRegistry(registry.RegistryOpts{
		Host: "gcr.io",

		Auth: registry.RegistryAuthOpts{
			URL:      "https://gcr.io",
			Username: "_json_key",
			Password: `{ "type": "service_account", ... }`,
		},

		LogFunc: func(msg string) { fmt.Printf("-----> %s\n", msg) },
	})

	path := "/Users/pivotal/Downloads/bosh-stemcell-3468.15-warden-boshlite-ubuntu-trusty-go_agent/image.gz"

	ref, err := reg.Push(registry.NewFSTgzAsset(path), "moonlit-ceiling-407/stemcells")
	if err != nil {
		panic("Failed to push image: " + err.Error())
	}

	fmt.Printf("-----> %s\n", ref.FQ())
}
