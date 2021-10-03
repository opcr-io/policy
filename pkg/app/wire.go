//go:build wireinject
// +build wireinject

package app

import (
	"io"

	"github.com/google/wire"

	"github.com/aserto-dev/clui"
	"github.com/opcr-io/policy/pkg/cc"
	"github.com/opcr-io/policy/pkg/cc/config"
)

var (
	policyAppSet = wire.NewSet(
		cc.NewCC,
		clui.NewUI,

		wire.FieldsOf(new(*cc.CC), "Config", "Log", "Context", "ErrGroup", "CancelFunc"),
	)

	policyAppTestSet = wire.NewSet(
		// Test
		cc.NewTestCC,

		// Normal
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
