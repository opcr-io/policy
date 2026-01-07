package cc

import (
	"context"
	"sync"

	"github.com/aserto-dev/logger"
	logger2 "github.com/aserto-dev/runtime/logger"
	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// CC contains dependencies that are cross cutting and are needed in most
// of the providers that make up this application.
type CC struct {
	Context    context.Context
	Config     *config.Config
	Log        *zerolog.Logger
	ErrGroup   *errgroup.Group
	CancelFunc context.CancelFunc
}

var (
	once         sync.Once
	cc           *CC
	cleanup      func()
	errSingleton error
)

// ErrGroupAndContext wraps a context and an error group.
type ErrGroupAndContext struct {
	Ctx      context.Context
	ErrGroup *errgroup.Group
	Cancel   context.CancelFunc
}

// NewContext creates a context that responds to user signals.
func NewContext() *ErrGroupAndContext {
	errGroup, ctxErr := errgroup.WithContext(signals.SetupSignalHandler())
	ctx, cancelFunc := context.WithCancel(ctxErr)

	return &ErrGroupAndContext{
		Ctx:      ctx,
		ErrGroup: errGroup,
		Cancel:   cancelFunc,
	}
}

// NewCC creates a singleton CC.
func NewCC(logOutput logger.Writer, errOutput logger.ErrWriter, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	once.Do(func() {
		cc, cleanup, errSingleton = buildCC(logOutput, errOutput, configPath, overrides)
	})

	return cc, func() {
		cleanup()

		once = sync.Once{}
	}, errSingleton
}

func buildCC(
	logOutput logger.Writer,
	errOutput logger.ErrWriter,
	configPath config.Path,
	overrides config.Overrider,
) (
	*CC, func(), error,
) {
	errGroupAndContext := NewContext()
	contextContext := errGroupAndContext.Ctx

	loggerConfig, err := config.NewLoggerConfig(configPath, overrides)
	if err != nil {
		return nil, nil, err
	}

	zerologLogger, err := logger2.NewLogger(logOutput, errOutput, loggerConfig)
	if err != nil {
		return nil, nil, err
	}

	configConfig, err := config.NewConfig(configPath, zerologLogger, overrides)
	if err != nil {
		return nil, nil, err
	}

	group := errGroupAndContext.ErrGroup

	cancelFunc := errGroupAndContext.Cancel

	ccCC := &CC{
		Context:    contextContext,
		Config:     configConfig,
		Log:        zerologLogger,
		ErrGroup:   group,
		CancelFunc: cancelFunc,
	}

	return ccCC, func() {
	}, nil
}
