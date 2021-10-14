package app

import extendedclient "github.com/opcr-io/policy/pkg/extended_client"

func (c *PolicyApp) SetVisibility(server, policy string, public bool) error {
	defer c.Cancel()

	creds := c.Configuration.Servers[server]

	xClient := extendedclient.NewExtendedClient(c.Logger,
		&extendedclient.Config{
			Address:  "https://" + server,
			Username: creds.Username,
			Password: creds.Password,
		},
		c.TransportWithTrustedCAs())

	// If the server doesn't support list APIs, print a message and return.
	if !xClient.Supported() {
		c.UI.Exclamation().Msg("The registry doesn't support extended capabilities like publishing policies.")
		return nil
	}

	err := xClient.SetVisibility(policy, public)
	if err != nil {
		return err
	}

	c.UI.Normal().Msgf("OK.")

	return nil
}
