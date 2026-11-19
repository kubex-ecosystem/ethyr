package flow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kubex-ecosystem/ethyr/tools/flow/control"
)

type MockPayload struct {
	ID int
}

func TestEngine_FIMExplosion(t *testing.T) {
	// Teste de carga extrema (100.000 payloads simultâneos)
	numPayloads := 100000
	workers := 100
	bufferSize := 1000

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	engine := NewEngine[MockPayload](bufferSize)

	var completedCount atomic.Int32

	// Definindo Passes (Esquadrões) simulando operações e mutações de bitflags
	pass1 := PassFunc[MockPayload](func(ctx context.Context, tctx *TranspileContext[MockPayload]) error {
		// Simula processamento leve
		tctx.State.Store(control.StepPreHooks)
		return nil
	})

	pass2 := PassFunc[MockPayload](func(ctx context.Context, tctx *TranspileContext[MockPayload]) error {
		tctx.State.Store(control.StepDone)
		completedCount.Add(1)
		tctx.Complete() // Finaliza o contexto
		return nil
	})

	engine.Register(ctx, "Pass1", control.StepValidate, pass1, workers)
	engine.Register(ctx, "Pass2", control.StepPreHooks, pass2, workers)

	engine.Start(ctx)

	// Injetando "FIM Explosion"
	var wg sync.WaitGroup
	wg.Add(numPayloads)

	for i := 0; i < numPayloads; i++ {
		go func(id int) {
			defer wg.Done()
			tctx := NewTranspileContext(ctx, MockPayload{ID: id})
			tctx.State.Store(control.StepValidate)
			engine.Submit(ctx, tctx)
		}(i)
	}

	wg.Wait()

	// Aguardando processamento da engine
	// Como a engine é assíncrona, precisamos esperar os contextos serem finalizados
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for FIM Explosion to complete. Completed: %d", completedCount.Load())
		case <-ticker.C:
			if completedCount.Load() == int32(numPayloads) {
				fmt.Printf("FIM Explosion completed successfully. Processed %d payloads.\n", numPayloads)
				return // Success
			}
		}
	}
}
