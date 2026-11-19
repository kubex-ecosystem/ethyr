package genkit

import (
	"context"
	"time"
)

// PassRequest define os parâmetros para uma execução cognitiva (IA/LLM) via Genkit.
type PassRequest struct {
	JobID        uint32         `json:"job_id"`
	Provider     string         `json:"provider,omitempty"` // gemini, openai, etc.
	Model        string         `json:"model,omitempty"`
	SystemPrompt string         `json:"system_prompt,omitempty"`
	UserPrompt   string         `json:"user_prompt"`
	Temperature  float64        `json:"temperature,omitempty"`
	MaxTokens    int            `json:"max_tokens,omitempty"`
	SchemaName   string         `json:"schema_name,omitempty"`  // Nome do schema esperado
	JSONSchema   string         `json:"json_schema,omitempty"`  // JSON Schema para estruturação (opcional)
	ContextData  map[string]any `json:"context_data,omitempty"` // Metadados adicionais do contexto
}

// PassResult representa a resposta da execução cognitiva estruturada.
type PassResult struct {
	JobID        uint32        `json:"job_id"`
	RawContent   string        `json:"raw_content"`
	Structured   any           `json:"structured,omitempty"` // Objeto decodificado caso haja schema
	Success      bool          `json:"success"`
	Latency      time.Duration `json:"latency"`
	TokenUsage   TokenUsage    `json:"token_usage"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// TokenUsage detalha o consumo do modelo cognitivo.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CognitivePass define o contrato para squads/passos que realizam processamento cognitivo inteligente.
type CognitivePass interface {
	// ExecuteCognitive realiza a chamada cognitiva baseada no request e retorna o resultado correspondente.
	ExecuteCognitive(ctx context.Context, req *PassRequest) (*PassResult, error)
}
