package cpi

import (
	"crypto/tls"
	"net/http"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	dkrclient "github.com/docker/engine-api/client"
	dkrtlsconfig "github.com/docker/go-connections/tlsconfig"
)

type DockerClientFactory struct {
	opts DockerClientOpts
}

func (f DockerClientFactory) Build() (*dkrclient.Client, error) {
	if !f.opts.IsPresent() {
		return nil, nil // no options, no client
	}

	httpClient, err := f.httpClient(f.opts)
	if err != nil {
		return nil, err
	}

	dkrClient, err := dkrclient.NewClient(f.opts.Host, f.opts.APIVersion, httpClient, nil)
	if err != nil {
		return nil, err
	}

	return dkrClient, nil
}

func (DockerClientFactory) httpClient(opts DockerClientOpts) (*http.Client, error) {
	if !opts.RequiresTLS() {
		return nil, nil
	}

	certPool, err := dkrtlsconfig.SystemCertPool()
	if err != nil {
		return nil, bosherr.WrapError(err, "Adding system CA certs")
	}

	if !certPool.AppendCertsFromPEM([]byte(opts.TLS.CA)) {
		return nil, bosherr.WrapError(err, "Appending configured CA certs")
	}

	tlsConfig := dkrtlsconfig.ClientDefault()
	tlsConfig.InsecureSkipVerify = false
	tlsConfig.RootCAs = certPool

	tlsCert, err := tls.X509KeyPair([]byte(opts.TLS.Certificate), []byte(opts.TLS.PrivateKey))
	if err != nil {
		return nil, bosherr.WrapError(err, "Loading X509 key pair (make sure the key is not encrypted)")
	}

	tlsConfig.Certificates = []tls.Certificate{tlsCert}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}
