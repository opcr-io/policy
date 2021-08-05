//+build wireinject

package app

import (
	"io"

	"github.com/google/wire"

	"github.com/aserto-dev/policy-cli/pkg/cc"
	"github.com/aserto-dev/policy-cli/pkg/cc/config"
)

var (
	policyCLISet = wire.NewSet(
		cc.NewCC,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)

	policyCLITestSet = wire.NewSet(
		// Test
		cc.NewTestCC,

		// Normal

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup"),
	)
)

func BuildPolicyCLI(logWriter io.Writer, configPath config.Path, overrides config.Overrider) (*PolicyCLI, func(), error) {
	wire.Build(
		wire.Struct(new(PolicyCLI), "*"),
		policyCLISet,
	)
	return &PolicyCLI{}, func() {}, nil
}

func BuildTestPolicyCLI(logWriter io.Writer, configPath config.Path, overrides config.Overrider) (*PolicyCLI, func(), error) {
	wire.Build(
		wire.Struct(new(PolicyCLI), "*"),
		policyCLITestSet,
	)
	return &PolicyCLI{}, func() {}, nil
}
