package integration_test

type CPIProps struct{}

func NewCPIProps() CPIProps {
	return CPIProps{}
}

func (p CPIProps) DummyImage() string {
	return "gcr.io/google_containers/pause-amd64:3.0"
}

func (p CPIProps) StemcellImage() string {
	// tests will wait for it to properly start
	return "localhost:5000/bosh.io/stemcells"
}

func (p CPIProps) RandomNetwork() map[string]interface{} {
	return map[string]interface{}{
		"private": map[string]interface{}{
			"type":             "manual",
			"netmask":          "255.255.255.0",
			"gateway":          "10.230.13.1",
			"ip":               "10.230.13.6",
			"dns":              []string{"8.8.8.8"},
			"default":          []string{"dns", "gateway"},
			"cloud_properties": nil,
		},
	}
}
