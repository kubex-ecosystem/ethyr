package genkit

import (
	"context"
	"os"
	"runtime/trace"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kubex-ecosystem/ethyr/tools/flow/fsm"
)

// TestIntegrationFlowGenkitTrace simula a interação hiper-realista entre múltiplas
// instâncias de FSM (Control Plane) e o Engine do Genkit (Worker Pool).
// Ele gera um `integration_trace.out` que prova o quão assíncrono e leve é o processo,
// podendo ser inspecionado com `go tool trace integration_trace.out`.
func TestIntegrationFlowGenkitTrace(t *testing.T) {
	f, err := os.Create("integration_trace.out")
	if err != nil {
		t.Fatalf("falha ao criar arquivo de trace: %v", err)
	}
	defer f.Close()
	ctx := context.Background()
	var task *trace.Task

	if !trace.IsEnabled() {
		t.Log("Trace não habilitado, iniciando trace")
		if err := trace.Start(f); err != nil {
			t.Fatalf("falha ao iniciar trace: %v", err)
		}
		defer trace.Stop()
	}

	ctx, task = trace.NewTask(ctx, "Integration")
	defer task.End()

	// 1. Inicializa o Genkit Engine com 50 Workers Fixos
	dp := NewDataPlane()
	engine := NewEngine(dp, 50)

	// Inicia o worker pool do engine
	engine.Start(ctx)
	defer engine.Stop()

	// 2. Definindo a topologia da FSM
	const (
		StateIdle fsm.State = 1 << iota
		StateAIProcessing
		StateDone
	)
	const (
		EvSendToAI fsm.Event = iota
		EvAIResponse
	)

	// Simulação de 2000 operações independentes (cada uma com sua própria FSM)
	numJobs := 2000
	var successCount atomic.Int32

	ingress := engine.IngressChan()
	egress := engine.EgressChan()

	// 3. O Loop Central de Saída (FSM Consumindo)
	// Na vida real, o engine.go do ethyr/flow faria esse papel de ler do egress
	// e injetar o EvAIResponse nas FSMs correspondentes
	go func() {
		for i := 0; i < numJobs; i++ {
			v := <-egress
			if v.HasFlag(FlagSuccess) {
				// Simula a FSM correspondente transitando para Done a partir da resposta
				successCount.Add(1)
			}
		}
	}()

	// 4. Disparo Massivo (FSM Produzindo)
	for i := 1; i <= numJobs; i++ {
		go func(jobID uint32) {
			machine := fsm.NewFSM(StateIdle, []fsm.Transition{
				{From: StateIdle, Event: EvSendToAI, To: StateAIProcessing},
				{From: StateAIProcessing, Event: EvAIResponse, To: StateDone},
			})

			// 4.1 FSM decide que precisa de IA e transita
			if machine.Trigger(EvSendToAI) {
				// 4.2 O Payload vai pro DataPlane (Cold Path)
				dp.Set(jobID, &JobPayload{Data: "Conteúdo a ser processado pela IA"})

				// 4.3 A Identidade vai pro Control Plane via Bitwise Vector (Hot Path)
				var v Vector = Vector(jobID)
				ingress <- v.WithFlag(FlagPending)
			}
		}(uint32(i))
	}

	// 5. Guilhotina de Segurança do Teste
	timeout := time.After(10 * time.Second)
	for successCount.Load() < int32(numJobs) {
		select {
		case <-timeout:
			t.Fatalf("Trace Timeout! Processados %d de %d", successCount.Load(), numJobs)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
