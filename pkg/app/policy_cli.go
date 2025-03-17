package app

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/aserto-dev/clui"
	"github.com/opcr-io/policy/pkg/cc/config"

	"github.com/docker/cli/cli/config/types"
)

// PolicyApp represents the policy CLI.
type PolicyApp struct {
	Context       context.Context
	Cancel        context.CancelFunc
	Logger        *zerolog.Logger
	Configuration *config.Config
	UI            *clui.UI
}

func (c *PolicyApp) SaveServerCreds(creds *types.AuthConfig) error {
	defer c.Cancel()

	if creds == nil {
		return errors.New("could not save nil credentials")
	}

	err := c.Configuration.CredentialsStore.Store(*creds)
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

	err := c.Configuration.CredentialsStore.Erase(server)
	if err != nil {
		return errors.Wrap(err, "failed to save server credentials")
	}

	return nil
}
