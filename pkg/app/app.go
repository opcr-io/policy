package app

import (
	"github.com/aserto-dev/logger"
	"github.com/opcr-io/policy/pkg/cc"
	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/opcr-io/policy/pkg/clui"
)

func BuildPolicyApp(
	logOutput logger.Writer,
	errOutput logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (
	*PolicyApp, func(), error,
) {
	ccCC, cleanup, err := cc.NewCC(logOutput, errOutput, configPath, overrides)
	if err != nil {
		return nil, nil, err
	}

	context := ccCC.Context
	cancelFunc := ccCC.CancelFunc
	zerologLogger := ccCC.Log
	configConfig := ccCC.Config

	ui := clui.NewUI()
	policyApp := &PolicyApp{
		Context:       context,
		Cancel:        cancelFunc,
		Logger:        zerologLogger,
		Configuration: configConfig,
		UI:            ui,
	}

	return policyApp, func() {
		cleanup()
	}, nil
}
