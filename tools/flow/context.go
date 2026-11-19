package flow

import (
	"context"
	"sync"

	"github.com/kubex-ecosystem/ethyr/tools/flow/control"
)

// TranspileContext carries the payload and atomic state through the autonomous passes.
type TranspileContext[T any] struct {
	// Context allows cancellation and timeouts for this specific payload.
	Ctx context.Context

	// Payload is the generic data being processed (e.g., LogEntry, Genkit flow input).
	Payload T

	// State tracks the item's specific state atomically using bit flags.
	State *control.FlagReg32[control.StepFlag]

	// Errs collects any errors that occur during the passes.
	Errs chan error

	// Done is closed when the item has completed all passes or failed terminally.
	Done chan struct{}

	once sync.Once
}

// NewTranspileContext creates a new reactive context for a payload.
func NewTranspileContext[T any](ctx context.Context, payload T) *TranspileContext[T] {
	tctx := &TranspileContext[T]{
		Ctx:     ctx,
		Payload: payload,
		State:   &control.FlagReg32[control.StepFlag]{},
		Errs:    make(chan error, 10), // Buffered to prevent blocking on multiple quick errors
		Done:    make(chan struct{}),
	}
	return tctx
}

// Complete gracefully marks this context as fully processed.
func (t *TranspileContext[T]) Complete() {
	t.once.Do(func() {
		close(t.Done)
		close(t.Errs)
	})
}
