package app

import (
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/pkg/errors"
)

func (c *PolicyApp) SetVisibility(server, policy string, public bool) error {
	defer c.Cancel()

	creds := c.Configuration.Servers[server]

	xClient, err := extendedregistry.GetExtendedClient(server,
		c.Logger,
		&extendedregistry.Config{
			Address:  "https://" + server,
			Username: creds.Username,
			Password: creds.Password,
		},
		c.TransportWithTrustedCAs())

	// If the server doesn't support list APIs, print a message and return.
	if err != nil {
		return errors.Wrap(err, "failed to get extended client")
	}

	err = xClient.SetVisibility(policy, public)
	if err != nil {
		return err
	}

	c.UI.Normal().Msgf("OK.")

	return nil
}
