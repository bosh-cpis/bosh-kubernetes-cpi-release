package integration_test

import (
	"strings"
)

type CPIProps struct{}

func NewCPIProps() CPIProps {
	return CPIProps{}
}

func (p CPIProps) DummyImage() string {
	return "gcr.io/google_containers/pause-amd64:3.0"
}

func (p CPIProps) StemcellImage() string {
	// tests will wait for it to properly start
	return "bosh.io.invalid/stemcells:img-b917ab90-6020-4334-58f6-17a1c7d2e816"
}

func (p CPIProps) IPStr(ip string) string { return strings.Replace(ip, ".", "-", -1) }

func (p CPIProps) RandomNetwork() map[string]interface{} {
	return map[string]interface{}{
		"private": p.DynamicNetwork(),
	}
}

func (p CPIProps) DynamicNetwork() map[string]interface{} {
	return map[string]interface{}{
		"type":             "dynamic",
		"dns":              []string{"8.8.8.8"},
		"default":          []string{"dns", "gateway"},
		"cloud_properties": nil,
	}
}

func (p CPIProps) ManualNetwork1IP() string { return "10.96.0.50" }
func (p CPIProps) ManualNetwork2IP() string { return "10.96.0.51" }

func (p CPIProps) ManualNetwork1() map[string]interface{} {
	return map[string]interface{}{
		"type":             "manual",
		"ip":               p.ManualNetwork1IP(),
		"gateway":          "unused",
		"netmask":          "unused",
		"dns":              []string{"8.8.8.8"},
		"default":          []string{"dns", "gateway"},
		"cloud_properties": nil,
	}
}

func (p CPIProps) ManualNetwork2() map[string]interface{} {
	return map[string]interface{}{
		"type":             "manual",
		"ip":               p.ManualNetwork2IP(),
		"gateway":          "unused",
		"netmask":          "unused",
		"dns":              []string{"8.8.8.8"},
		"default":          []string{"dns", "gateway"},
		"cloud_properties": nil,
	}
}

func (p CPIProps) VIPNetwork1IP() string { return "10.96.0.52" }
func (p CPIProps) VIPNetwork2IP() string { return "10.96.0.53" }

func (p CPIProps) VIPNetwork1() map[string]interface{} {
	return map[string]interface{}{
		"type":             "vip",
		"ip":               p.VIPNetwork1IP(),
		"dns":              []string{"8.8.8.8"},
		"default":          []string{},
		"cloud_properties": nil,
	}
}

func (p CPIProps) VIPNetwork2() map[string]interface{} {
	return map[string]interface{}{
		"type":             "vip",
		"ip":               p.VIPNetwork2IP(),
		"dns":              []string{"8.8.8.8"},
		"default":          []string{},
		"cloud_properties": nil,
	}
}
