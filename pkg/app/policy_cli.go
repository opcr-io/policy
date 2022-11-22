package app

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/aserto-dev/clui"
	"github.com/opcr-io/policy/pkg/cc/config"
)

// PolicyApp represents the policy CLI.
type PolicyApp struct {
	Context       context.Context
	Cancel        context.CancelFunc
	Logger        *zerolog.Logger
	Configuration *config.Config
	UI            *clui.UI
}

func (c *PolicyApp) SaveServerCreds(server string, creds config.ServerCredentials) error {
	defer c.Cancel()
	if server == "" {
		server = c.Configuration.DefaultDomain
	}

	if c.Configuration.Servers == nil {
		c.Configuration.Servers = map[string]config.ServerCredentials{}
	}

	c.Configuration.Servers[server] = creds

	err := c.Configuration.SaveCreds()
	if err != nil {
		return errors.Wrap(err, "failed to save server credentials")
	}

	return nil
}

func (c *PolicyApp) RemoveServerCreds(server string) error {
	defer c.Cancel()
	if server == "" {
		server = c.Configuration.DefaultDomain
	}

	if c.Configuration.Servers == nil {
		c.Configuration.Servers = map[string]config.ServerCredentials{}
	}

	c.Configuration.Servers[server] = config.ServerCredentials{}

	err := c.Configuration.SaveCreds()
	if err != nil {
		return errors.Wrap(err, "failed to save server credentials")
	}

	return nil
}
