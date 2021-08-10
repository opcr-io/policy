package cc

import (
	"context"
	"io"
	"sync"

	"github.com/aserto-dev/policy/pkg/cc/config"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// CC contains dependencies that are cross cutting and are needed in most
// of the providers that make up this application
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
	singletonErr error
)

// NewCC creates a singleton CC
func NewCC(logOutput io.Writer, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	once.Do(func() {
		cc, cleanup, singletonErr = buildCC(logOutput, configPath, overrides)
	})

	return cc, func() {
		cleanup()
		once = sync.Once{}
	}, singletonErr
}

// NewTestCC creates a singleton CC to be used for testing.
// It uses a fake context (context.Background)
func NewTestCC(logOutput io.Writer, configPath config.Path, overrides config.Overrider) (*CC, func(), error) {
	once.Do(func() {
		cc, cleanup, singletonErr = buildTestCC(logOutput, configPath, overrides)
	})

	return cc, func() {
		cleanup()
		once = sync.Once{}
	}, singletonErr
}
