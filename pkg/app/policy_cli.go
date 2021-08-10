package app

import (
	"context"

	"github.com/rs/zerolog"

	runtime "github.com/aserto-dev/aserto-runtime"
	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/policy/pkg/cc/config"
)

// PolicyApp represents the policy CLI
type PolicyApp struct {
	Context       context.Context
	Cancel        context.CancelFunc
	Logger        *zerolog.Logger
	Configuration *config.Config
	Runtime       *runtime.Runtime
	UI            *clui.UI
}
