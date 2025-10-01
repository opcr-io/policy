package app

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
	"runtime"
)

func (c *PolicyApp) TransportWithTrustedCAs() *http.Transport {
	if c.Configuration.Insecure {
		return &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}} //nolint:gosec // feature used for debugging
	}
	// Get the SystemCertPool, continue with an empty pool on error
	var (
		rootCAs *x509.CertPool
		err     error
	)
	if runtime.GOOS != `windows` {
		rootCAs, err = x509.SystemCertPool()
		if err != nil {
			c.UI.Problem().WithErr(err).WithEnd(1).Msg("Failed to load system cert pool.")
		}
	} else {
		// remove runtime check when updating to go1.18 https://github.com/deviceinsight/kafkactl/issues/108.
		if len(c.Configuration.CA) > 0 {
			c.UI.Exclamation().Msg("Cannot use custom CAs on Windows. Please configure your system store to trust your CAs.")
		}

		return http.DefaultTransport.(*http.Transport) //nolint:forcetypeassert
	}

	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read in the cert files
	for _, localCertFile := range c.Configuration.CA {
		certs, err := os.ReadFile(localCertFile)
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
