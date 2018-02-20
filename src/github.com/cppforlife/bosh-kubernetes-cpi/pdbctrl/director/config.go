package director

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Config struct {
	Host string
	Port int

	TLS TLSConfig

	Client       string
	ClientSecret string
}

type TLSConfig struct {
	Cert TLSCertConfig
}

type TLSCertConfig struct {
	CA string
}

func (c Config) Validate() error {
	if len(c.Host) == 0 {
		return bosherr.Error("Missing 'Host'")
	}

	if c.Port == 0 {
		return bosherr.Error("Missing 'Port'")
	}

	if len(c.Client) == 0 {
		return bosherr.Error("Missing 'Client'")
	}

	if len(c.ClientSecret) == 0 {
		return bosherr.Error("Missing 'ClientSecret'")
	}

	return nil
}

func (c Config) AnonymousUserConfig() boshdir.FactoryConfig {
	return boshdir.FactoryConfig{
		Host: c.Host,
		Port: c.Port,

		CACert: c.TLS.Cert.CA,
	}
}

func (c Config) UserConfig() boshdir.FactoryConfig {
	return boshdir.FactoryConfig{
		Host: c.Host,
		Port: c.Port,

		CACert: c.TLS.Cert.CA,

		Client:       c.Client,
		ClientSecret: c.ClientSecret,
	}
}
