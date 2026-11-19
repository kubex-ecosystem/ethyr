package flow

import (
	"context"
	"fmt"
)

// Pass represents an autonomous squad that performs a specific stage of execution.
// It mutates the state or the payload via bitwise flags and contexts.
type Pass[T any] interface {
	// Execute performs the pass logic.
	Execute(ctx context.Context, tctx *TranspileContext[T]) error
}

// PassFunc allows using plain functions as Passes.
type PassFunc[T any] func(ctx context.Context, tctx *TranspileContext[T]) error

func (f PassFunc[T]) Execute(ctx context.Context, tctx *TranspileContext[T]) error {
	return f(ctx, tctx)
}

// PassWorkerPool manages a group of goroutines running a specific Pass reactively.
type PassWorkerPool[T any] struct {
	Name    string
	Pass    Pass[T]
	In      <-chan *TranspileContext[T]
	Out     chan<- *TranspileContext[T]
	Workers int
}

// Start launches the goroutines for this squad, making it fully event-driven.
func (p *PassWorkerPool[T]) Start(ctx context.Context) {
	if p.Workers <= 0 {
		p.Workers = 1
	}

	for i := 0; i < p.Workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return // Global context cancelled
				case tctx, ok := <-p.In:
					if !ok {
						return // Input channel closed
					}

					// Execute the pass
					err := p.Pass.Execute(ctx, tctx)
					if err != nil {
						// Send to context errors without blocking
						select {
						case tctx.Errs <- fmt.Errorf("[%s] pass failed: %w", p.Name, err):
						default:
						}
					}

					// Forward back to the central Engine router
					select {
					case p.Out <- tctx:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}
}
