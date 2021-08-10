//+build wireinject

package app

import (
	"io"

	"github.com/google/wire"

	runtime "github.com/aserto-dev/aserto-runtime"
	runtimeconfig "github.com/aserto-dev/aserto-runtime/config"
	"github.com/aserto-dev/clui"
	eds "github.com/aserto-dev/go-eds"
	"github.com/aserto-dev/policy/pkg/cc"
	"github.com/aserto-dev/policy/pkg/cc/config"
)

var (
	policyAppSet = wire.NewSet(
		cc.NewCC,
		runtime.BuildRuntime,
		eds.NewEdgeDirectory,
		emptyEDSConfig,
		emptyOPAConfig,
		clui.NewUI,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup", "CancelFunc"),
	)

	policyAppTestSet = wire.NewSet(
		// Test
		cc.NewTestCC,

		// Normal
		runtime.BuildRuntime,
		eds.NewEdgeDirectory,
		emptyEDSConfig,
		emptyOPAConfig,
		clui.NewUI,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup", "CancelFunc"),
	)
)

func BuildPolicyApp(logWriter io.Writer, configPath config.Path, overrides config.Overrider) (*PolicyApp, func(), error) {
	wire.Build(
		wire.Struct(new(PolicyApp), "*"),
		policyAppSet,
	)
	return &PolicyApp{}, func() {}, nil
}

func BuildTestPolicyApp(logWriter io.Writer, configPath config.Path, overrides config.Overrider) (*PolicyApp, func(), error) {
	wire.Build(
		wire.Struct(new(PolicyApp), "*"),
		policyAppTestSet,
	)
	return &PolicyApp{}, func() {}, nil
}

func emptyEDSConfig() *eds.Config {
	return &eds.Config{}
}

func emptyOPAConfig() *runtimeconfig.ConfigOPA {
	return &runtimeconfig.ConfigOPA{}
}
