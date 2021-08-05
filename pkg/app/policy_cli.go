package app

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/aserto-dev/policy-cli/pkg/cc/config"
)

// PolicyCLI represents the policy CLI
type PolicyCLI struct {
	Context       context.Context
	Logger        *zerolog.Logger
	Configuration *config.Config
}
