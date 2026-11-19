package tools

import (
	"fmt"
	"os"
	"sync"
)

// GoGroup is a simple wrapper around sync.WaitGroup to manage goroutines.
type GoGroup struct{ wg sync.WaitGroup }

// GoFunc defines a function that returns a value of type T.
type GoFunc[T any] func() T

// GoWG defines the interface for managing goroutines.
type GoWG[T any] interface {
	Go(fn GoFunc[T], wait bool, cb func(any)) T
	Wait(cb func(any))
}

// Go starts a new goroutine and tracks it in the WaitGroup.
func (g *GoGroup) Go(fn GoFunc[any], cb func(any), wait bool) *any {
	r := make(chan any, 1)

	if g != nil {
		g.wg.Go(func() {
			defer func(cb func(any)) {
				if r := recover(); r != nil {
					if cb != nil {
						cb(r)
					} else {
						fmt.Fprintf(os.Stderr, "error: %v\n", r)
					}
					// Here we assume that if we're here, maybe we need to call .Done. Maybe not... We'll see. kkkk
					g.wg.Done()
				}
				// Here we assume that if we're here, no error occurred, so doesn't need to call .Done, or anything else.

				if fn != nil {
					r <- fn()
				}
			}(cb)
		})
	}

	if wait {
		g.Wait(cb)
		if r := <-r; r != nil {
			return &r
		}
	}

	return nil
}

// Wait waits for all goroutines to complete.
func (g *GoGroup) Wait(cb func(any)) {
	if g != nil {
		g.wg.Wait()
		if cb != nil {
			cb(nil)
		}
	}
}
