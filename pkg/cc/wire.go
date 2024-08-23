//go:build wireinject
// +build wireinject

package cc

import (
	"github.com/aserto-dev/logger"
	"github.com/google/wire"

	runtimeLogger "github.com/aserto-dev/runtime/logger"
	"github.com/opcr-io/policy/pkg/cc/config"
	cc_context "github.com/opcr-io/policy/pkg/cc/context"
)

var (
	commonSet = wire.NewSet(
		config.NewConfig,
		config.NewLoggerConfig,
		runtimeLogger.NewLogger,
		wire.FieldsOf(new(config.Config), "Logging"),
		wire.FieldsOf(new(*cc_context.ErrGroupAndContext), "Ctx", "ErrGroup", "Cancel"),

		wire.Struct(new(CC), "*"),
	)

	ccSet = wire.NewSet(
		commonSet,
		cc_context.NewContext,
	)

	ccTestSet = wire.NewSet(
		commonSet,
		cc_context.NewTestContext,
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
