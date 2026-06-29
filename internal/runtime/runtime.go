package runtime

import (
	"context"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/plugins"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/rs/zerolog"
)

type Runtime struct {
	Logger *zerolog.Logger
	Config *Config

	plugins          map[string]plugins.Factory
	builtins1        map[*rego.Function]rego.Builtin1
	builtins2        map[*rego.Function]rego.Builtin2
	builtins3        map[*rego.Function]rego.Builtin3
	builtins4        map[*rego.Function]rego.Builtin4
	builtinsDyn      map[*rego.Function]rego.BuiltinDyn
	compilerBuiltins map[string]*ast.Builtin
}

func New(ctx context.Context) (*Runtime, error) {
	newLogger := zerolog.Ctx(ctx).With().Str("component", "runtime").Logger()

	runtime := &Runtime{
		Logger:           &newLogger,
		Config:           &Config{Config: OPAConfig{}},
		plugins:          make(map[string]plugins.Factory),
		builtins1:        make(map[*rego.Function]rego.Builtin1),
		builtins2:        make(map[*rego.Function]rego.Builtin2),
		builtins3:        make(map[*rego.Function]rego.Builtin3),
		builtins4:        make(map[*rego.Function]rego.Builtin4),
		builtinsDyn:      make(map[*rego.Function]rego.BuiltinDyn),
		compilerBuiltins: make(map[string]*ast.Builtin),
	}

	return runtime, nil
}
