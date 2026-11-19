package genkit

import (
	"testing"
)

func TestVectorBitwise(t *testing.T) {
	// 1. Cria um Vector com RegistryID 42
	var v Vector = 42

	if v.RegistryID() != 42 {
		t.Fatalf("Expected RegistryID 42, got %d", v.RegistryID())
	}

	// 2. Define o Contexto (JobType)
	// Vamos jogar o valor 7 para os bits 32-55
	v = v | (7 << 32)
	if v.JobContext() != 7 {
		t.Fatalf("Expected JobContext 7, got %d", v.JobContext())
	}
	if v.RegistryID() != 42 {
		t.Fatalf("RegistryID foi corrompido, esperado 42, recebido %d", v.RegistryID())
	}

	// 3. Testa manipulador de Flags
	v = v.WithFlag(FlagPending)
	if !v.HasFlag(FlagPending) {
		t.Fatalf("Expected FlagPending to be set")
	}

	// 4. Limpa as flags antigas e define a nova
	v = v.WithFlag(FlagSuccess)
	if v.HasFlag(FlagPending) {
		t.Fatalf("Expected FlagPending to be cleared")
	}
	if !v.HasFlag(FlagSuccess) {
		t.Fatalf("Expected FlagSuccess to be set")
	}

	// 5. Garante que as máscaras não corromperam o contexto e o ID
	if v.RegistryID() != 42 {
		t.Fatalf("RegistryID foi corrompido pós-flag, esperado 42, recebido %d", v.RegistryID())
	}
	if v.JobContext() != 7 {
		t.Fatalf("JobContext foi corrompido pós-flag, esperado 7, recebido %d", v.JobContext())
	}
}
