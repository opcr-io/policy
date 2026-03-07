package app

import (
	"io"

	"github.com/opcr-io/policy/pkg/cc"
	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/opcr-io/policy/pkg/clui"
)

func BuildPolicyApp(
	logOutput io.Writer,
	errOutput io.Writer,
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
