package genkit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestEngineProcessamentoNormal valida se a engine consome os vetores e atualiza as flags de saída
func TestEngineProcessamentoNormal(t *testing.T) {
	dp := NewDataPlane()
	engine := NewEngine(dp, 5)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)

	// Injeta payload e envia 10 vetores
	for i := 1; i <= 10; i++ {
		dp.Set(uint32(i), &JobPayload{Data: "Mock"})
		var v Vector = Vector(i)
		engine.IngressChan() <- v.WithFlag(FlagPending)
	}

	// Consome 10 resultados
	successCount := 0
	timeout := time.After(2 * time.Second)

	for i := 0; i < 10; i++ {
		select {
		case outVector := <-engine.EgressChan():
			if outVector.HasFlag(FlagSuccess) {
				successCount++
			} else {
				t.Errorf("Esperado FlagSuccess, recebido Vetor: %v", outVector)
			}
		case <-timeout:
			t.Fatalf("Timeout aguardando respostas da Engine")
		}
	}

	if successCount != 10 {
		t.Fatalf("Esperado 10 sucessos, obtido %d", successCount)
	}
}

// BenchmarkEngine faz o teste de carga massiva no pool de workers
// Execute com: go test -bench . -benchmem
func BenchmarkEngine(b *testing.B) {
	dp := NewDataPlane()
	// Simulando um pool robusto para processamento pesado
	engine := NewEngine(dp, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)

	// Pre-popula um payload base para todos
	dp.Set(1, &JobPayload{Data: "Bench Data"})

	ingress := engine.IngressChan()
	egress := engine.EgressChan()

	// Inicia rotina para drenar o egress sem bloquear
	var processed uint64
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-egress:
				atomic.AddUint64(&processed, 1)
			}
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	// Dispara o benchmark real
	for i := 0; i < b.N; i++ {
		// O Registry ID será 1 para reaproveitar o payload, JobContext 10
		var v Vector = 1
		v = v | (10 << 32)
		ingress <- v.WithFlag(FlagPending)
	}

	// Aguarda processamento de todos (spin wait simples com timeout por segurança)
	timeout := time.Now().Add(30 * time.Second)
	for atomic.LoadUint64(&processed) < uint64(b.N) {
		if time.Now().After(timeout) {
			b.Fatalf("Benchmark Timeout: processados %d de %d", atomic.LoadUint64(&processed), b.N)
		}
		time.Sleep(1 * time.Millisecond)
	}
}
