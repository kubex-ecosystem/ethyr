package genkit

import (
	"sync"
	"time"
)

// JobPayload isola os dados transitados, protegido pela memória compartilhada
type JobPayload struct {
	SystemPrompt string
	Data         any
	ExpectedJSON string // Schema JSON
	Result       any
	Err          error

	// Telemetria (Cold Path)
	StartedAt  time.Time
	FinishedAt time.Time
}

// DataPlane atua como memória compartilhada e controlada.
type DataPlane struct {
	mu    sync.RWMutex
	store map[uint32]*JobPayload
}

// NewDataPlane cria uma nova instância de Registry
func NewDataPlane() *DataPlane {
	return &DataPlane{
		store: make(map[uint32]*JobPayload),
	}
}

// Get obtém um payload em segurança
func (dp *DataPlane) Get(id uint32) (*JobPayload, bool) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	val, ok := dp.store[id]
	return val, ok
}

// Set injeta um payload
func (dp *DataPlane) Set(id uint32, p *JobPayload) {
	dp.mu.Lock()
	defer dp.mu.Unlock()
	dp.store[id] = p
}
