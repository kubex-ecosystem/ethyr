package genkit

import (
	"sync"
	"testing"
)

func TestDataPlaneConcurrency(t *testing.T) {
	dp := NewDataPlane()
	var wg sync.WaitGroup

	// Testa concorrência extrema: múltiplos leitores e escritores simultâneos
	numWorkers := 100
	wg.Add(numWorkers * 2)

	// Escritores
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			dp.Set(uint32(id), &JobPayload{SystemPrompt: "Test"})
		}(i)
	}

	// Leitores
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			// Pode falhar se a escrita ainda não tiver ocorrido, mas o mapa não deve causar panic (thread-safe)
			dp.Get(uint32(id))
		}(i)
	}

	wg.Wait()

	// Valida se os 100 payloads foram escritos
	for i := 0; i < numWorkers; i++ {
		payload, ok := dp.Get(uint32(i))
		if !ok || payload.SystemPrompt != "Test" {
			t.Fatalf("Falha ao recuperar payload id %d", i)
		}
	}
}
