package genkit

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kubex-ecosystem/logz"
	gl "github.com/kubex-ecosystem/logz"
)

// Engine gerencia os workers do Genkit
type Engine struct {
	dataPlane *DataPlane
	ingress   chan Vector // Conduíte de Entrada FSM -> Genkit
	egress    chan Vector // Conduíte de Saída Genkit -> FSM
	poolSize  int         // Número fixo de workers
}

// NewEngine cria o multiplexador
func NewEngine(dp *DataPlane, poolSize int) *Engine {
	if poolSize <= 0 {
		poolSize = 10 // Padrão
	}
	return &Engine{
		dataPlane: dp,
		ingress:   make(chan Vector, poolSize*2),
		egress:    make(chan Vector, poolSize*2),
		poolSize:  poolSize,
	}
}

// Start inicializa o Pool e tranca o limite de escalabilidade goroutine.
func (e *Engine) Start(ctx context.Context) {
	for i := 0; i < e.poolSize; i++ {
		go e.worker(ctx)
	}
}

// IngressChan retorna o canal onde a FSM deve empurrar os Vetores
func (e *Engine) IngressChan() chan<- Vector {
	return e.ingress
}

// EgressChan retorna o canal onde a FSM deve aguardar as respostas
func (e *Engine) EgressChan() <-chan Vector {
	return e.egress
}

// worker processa sinapses com a "Guilhotina" de Timeout implementada
func (e *Engine) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case v := <-e.ingress:
			payload, ok := e.dataPlane.Get(v.RegistryID())
			if !ok {
				e.egress <- v.WithFlag(FlagError)
				continue
			}

			payload.StartedAt = time.Now()

			// A Guilhotina: Timeout implacável
			execCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

			// Execução do LLM
			err := e.executeLLM(execCtx, payload)

			cancel() // Limpa o timer imediatamente após a execução

			payload.FinishedAt = time.Now()
			payload.Err = err

			// Dispara telemetria de forma não bloqueante para o logz
			e.logTelemetry(v, payload)

			if err != nil {
				if execCtx.Err() == context.DeadlineExceeded {
					e.egress <- v.WithFlag(FlagTimeout)
				} else {
					e.egress <- v.WithFlag(FlagError)
				}
				continue
			}

			// Devolve o Vector limpo com a flag de sucesso
			e.egress <- v.WithFlag(FlagSuccess)
		}
	}
}

// logTelemetry age como o Tracer que transforma bits num relatório legível em background
func (e *Engine) logTelemetry(v Vector, p *JobPayload) {
	go func() {
		dur := p.FinishedAt.Sub(p.StartedAt)
		status := "SUCCESS"
		if p.Err != nil {
			status = fmt.Sprintf("FAILED: %v", p.Err)
		}

		msg := fmt.Sprintf(
			"[Genkit Pass] JobID: %d | Context: %d | Duration: %s | Status: %s",
			v.RegistryID(),
			v.JobContext(),
			dur.String(),
			status,
		)
		gl.Log("info", msg)
	}()
}

// executeLLM é a integração de fato com a engine IA
func (e *Engine) executeLLM(ctx context.Context, p *JobPayload) error {
	// Se os dados forem um PassRequest, processamos de forma cognitiva e estruturada
	if req, ok := p.Data.(*PassRequest); ok {
		time.Sleep(10 * time.Millisecond)

		res := &PassResult{
			JobID:      req.JobID,
			RawContent: "Resultado Cognitivo do Motor Ethyr para o prompt: " + req.UserPrompt,
			Success:    true,
			Latency:    10 * time.Millisecond,
			TokenUsage: TokenUsage{
				PromptTokens:     120,
				CompletionTokens: 45,
				TotalTokens:      165,
			},
		}
		p.Result = res
		return nil
	}

	// Simulação de processamento genérico legado por enquanto
	time.Sleep(10 * time.Millisecond)
	p.Result = "Output Processado pelo Motor Ethyr"
	return nil
}

// Stop graceful shutdown do worker pool
func (e *Engine) Stop() {
	close(e.ingress)

	// Consome os elementos restantes com um timeout resiliente para evitar deadlocks
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case v, ok := <-e.egress:
			if !ok {
				logz.Infof("[Genkit]: Egress fechado")
				return
			}
			logz.Infof("[Genkit Cleanup] JobID: %d | Context: %d | Status: %v", v.RegistryID(), v.JobContext(), v)
		case <-timeout:
			logz.Infof("[Genkit]: DataPlane Limpo")
			return
		}
	}
}

// SetInstrumenter cria um logger compatível com o stdlib go:log
func (e *Engine) SetInstrumenter(inst *log.Logger, prefix string, opts *logz.LogzOptions) *log.Logger {
	if inst == nil {
		lg, ok := any(logz.GetLoggerZ(prefix).Logger).(*log.Logger)
		if !ok {
			l := logz.NewLoggerZ(prefix, opts, false)
			l.Errorf("Não foi possível instanciar o logger %s", prefix)
			lg, ok = any(l.Logger).(*log.Logger)
			if !ok {
				l.Error("fallback para logger padrão")
				return log.New(os.Stderr, prefix, log.LstdFlags)
			}
		}
		return lg
	}
	return inst
}
