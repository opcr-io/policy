package app

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
)

func (c *PolicyApp) TransportWithTrustedCAs() *http.Transport {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		c.UI.Problem().WithErr(err).WithEnd(1).Msg("Failed to load system cert pool.")
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert files
	for _, localCertFile := range c.Configuration.CA {
		certs, err := ioutil.ReadFile(localCertFile)
		if err != nil {
			c.UI.Problem().WithErr(err).WithEnd(1).Msgf("Failed to append %q to RootCAs.", localCertFile)
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
			c.UI.Exclamation().Msgf("Cert %q not appended to RootCAs.", localCertFile)
		}

	}

	// Trust the augmented cert pool in our client
	config := &tls.Config{
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}
	return &http.Transport{TLSClientConfig: config}
}
