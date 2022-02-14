//go:build wireinject
// +build wireinject

package cc

import (
	"github.com/aserto-dev/go-utils/certs"
	"github.com/aserto-dev/go-utils/logger"
	"github.com/google/wire"

	"github.com/opcr-io/policy/pkg/cc/config"
	cc_context "github.com/opcr-io/policy/pkg/cc/context"
)

var (
	ccSet = wire.NewSet(
		cc_context.NewContext,
		config.NewConfig,
		config.NewLoggerConfig,
		logger.NewLogger,
		certs.NewGenerator,
		wire.FieldsOf(new(config.Config), "Logging"),
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup", "Cancel"),

		wire.Struct(new(CC), "*"),
	)

	ccTestSet = wire.NewSet(
		// Test
		cc_context.NewTestContext,

		// Normal
		config.NewConfig,
		config.NewLoggerConfig,
		logger.NewLogger,
		certs.NewGenerator,
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup", "Cancel"),

		wire.Struct(new(CC), "*"),
	)
)

// buildCC sets up the CC struct that contains all dependencies that
// are cross cutting
func buildCC(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	wire.Build(ccSet)
	return &CC{}, func() {}, nil
}

func buildTestCC(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	wire.Build(ccTestSet)
	return &CC{}, func() {}, nil
}
