package genkit

import (
	"context"
	"testing"
	"time"
)

// MockCognitivePass implementa a interface CognitivePass para fins de testes.
type MockCognitivePass struct{}

func (m *MockCognitivePass) ExecuteCognitive(ctx context.Context, req *PassRequest) (*PassResult, error) {
	return &PassResult{
		JobID:      req.JobID,
		RawContent: "Resposta mock: " + req.UserPrompt,
		Success:    true,
		Latency:    5 * time.Millisecond,
		TokenUsage: TokenUsage{
			PromptTokens:     50,
			CompletionTokens: 20,
			TotalTokens:      70,
		},
	}, nil
}

func TestCognitiveInterfaceAndStructures(t *testing.T) {
	var cp CognitivePass = &MockCognitivePass{}

	req := &PassRequest{
		JobID:        99,
		Provider:     "gemini",
		Model:        "gemini-pro",
		SystemPrompt: "Você é um assistente atômico.",
		UserPrompt:   "Olá, Mundo!",
	}

	res, err := cp.ExecuteCognitive(context.Background(), req)
	if err != nil {
		t.Fatalf("falha ao executar passo cognitivo: %v", err)
	}

	if res.JobID != 99 {
		t.Errorf("esperado JobID 99, obtido %d", res.JobID)
	}

	if res.RawContent != "Resposta mock: Olá, Mundo!" {
		t.Errorf("conteúdo inesperado: %s", res.RawContent)
	}

	if !res.Success {
		t.Errorf("esperado Success=true")
	}

	if res.TokenUsage.TotalTokens != 70 {
		t.Errorf("esperado 70 total tokens, obtido %d", res.TokenUsage.TotalTokens)
	}
}

func TestEngineWithCognitivePass(t *testing.T) {
	dp := NewDataPlane()
	engine := NewEngine(dp, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)

	// Prepara a chamada cognitiva
	jobID := uint32(500)
	req := &PassRequest{
		JobID:      jobID,
		UserPrompt: "Execute o processo atômico na FSM.",
	}

	dp.Set(jobID, &JobPayload{
		Data: req,
	})

	var v Vector = Vector(jobID)
	engine.IngressChan() <- v.WithFlag(FlagPending)

	select {
	case outVector := <-engine.EgressChan():
		if !outVector.HasFlag(FlagSuccess) {
			t.Fatalf("Vetor de resposta inválido: %v", outVector)
		}

		// Valida se o payload no DataPlane foi atualizado com PassResult
		payload, ok := dp.Get(jobID)
		if !ok {
			t.Fatalf("job não encontrado no DataPlane")
		}

		result, ok := payload.Result.(*PassResult)
		if !ok {
			t.Fatalf("esperado *PassResult no resultado, recebido: %v", payload.Result)
		}

		if result.JobID != jobID {
			t.Errorf("esperado JobID %d no resultado, obtido %d", jobID, result.JobID)
		}

		if result.TokenUsage.PromptTokens != 120 {
			t.Errorf("esperado PromptTokens = 120, obtido %d", result.TokenUsage.PromptTokens)
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("Timeout aguardando processamento cognitivo")
	}
}
