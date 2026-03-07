package cc

import (
	"context"
	"io"
	"os/signal"
	"sync"
	"syscall"

	"github.com/opcr-io/policy/internal/logger"
	"github.com/opcr-io/policy/pkg/cc/config"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
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
	ctxNotify, cancelFunc := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	errGroup, ctx := errgroup.WithContext(ctxNotify)

	return &ErrGroupAndContext{
		Ctx:      ctx,
		ErrGroup: errGroup,
		Cancel:   cancelFunc,
	}
}

// NewCC creates a singleton CC.
func NewCC(logOutput io.Writer, errOutput io.Writer, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	once.Do(func() {
		cc, cleanup, errSingleton = buildCC(logOutput, errOutput, configPath, overrides)
	})

	return cc, func() {
		cleanup()

		once = sync.Once{}
	}, errSingleton
}

func buildCC(
	logOutput io.Writer,
	errOutput io.Writer,
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

	zerologLogger, err := logger.NewLogger(logOutput, errOutput, loggerConfig)
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
