package context

import (
	"context"

	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
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

// NewTestContext creates a context that can be used for testing.
func NewTestContext() *ErrGroupAndContext {
	errGroup, ctxErr := errgroup.WithContext(context.Background())
	ctx, cancelFunc := context.WithCancel(ctxErr)

	return &ErrGroupAndContext{
		Ctx:      ctx,
		ErrGroup: errGroup,
		Cancel:   cancelFunc,
	}
}
