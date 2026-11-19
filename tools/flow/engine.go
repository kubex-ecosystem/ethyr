package flow

import (
	"context"

	"github.com/kubex-ecosystem/ethyr/tools/flow/control"
)

// Route maps a specific StepFlag state to a Pass input channel.
type Route[T any] struct {
	State control.StepFlag
	In    chan<- *TranspileContext[T]
}

// Engine is the central orchestrator that routes TranspileContexts.
// It acts as the "FSM/Manager" dictating the flow.
type Engine[T any] struct {
	// router is the central channel where Passes send contexts back after execution.
	router chan *TranspileContext[T]

	// routes map a StepFlag to the input channel of a specific PassWorkerPool.
	routes []Route[T]
}

// NewEngine creates a new Engine with a given buffer capacity to prevent FIM Explosion bottlenecks.
func NewEngine[T any](bufferSize int) *Engine[T] {
	return &Engine[T]{
		router: make(chan *TranspileContext[T], bufferSize),
	}
}

// Register adds a pass to the engine, creating its worker pool and configuring routing.
func (e *Engine[T]) Register(ctx context.Context, name string, state control.StepFlag, pass Pass[T], workers int) {
	// Using the same buffer size as the main router to balance the load.
	inChan := make(chan *TranspileContext[T], cap(e.router))

	pool := &PassWorkerPool[T]{
		Name:    name,
		Pass:    pass,
		In:      inChan,
		Out:     e.router,
		Workers: workers,
	}
	pool.Start(ctx)

	e.routes = append(e.routes, Route[T]{State: state, In: inChan})
}

// Start launches the Engine's central reactive routing goroutine.
func (e *Engine[T]) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case tctx := <-e.router:
				// Route the context based on its current State bit.
				// Autonomous Passes mutate this bit, and the Engine routes accordingly.
				routed := false
				for _, r := range e.routes {
					if tctx.State.Has(r.State) {
						go func(route Route[T], t *TranspileContext[T]) {
							select {
							case route.In <- t:
							case <-ctx.Done():
							}
						}(r, tctx)
						routed = true
						break
					}
				}

				// If no route matches (e.g. terminal state), we complete the context.
				if !routed {
					tctx.Complete()
				}
			}
		}
	}()
}

// Submit injects a new context into the engine router.
func (e *Engine[T]) Submit(ctx context.Context, tctx *TranspileContext[T]) {
	select {
	case e.router <- tctx:
	case <-ctx.Done():
	}
}
