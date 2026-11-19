package manager

import (
	"context"
	"io"
	"sync"

	"github.com/kubex-ecosystem/ethyr/tools/flow"
	"github.com/kubex-ecosystem/ethyr/tools/flow/control"
	"github.com/kubex-ecosystem/ethyr/tools/flow/events"
)

// LogEntry is a local interface replacing logz.Entry to break the import cycle.
type LogEntry interface {
	GetLevel() interface{ String() string }
}

// LogLevel represents a log severity level.
type LogLevel = string

// Manager handles the processing pipeline for log entries using the reactive Engine.
type Manager struct {
	mu sync.RWMutex

	formatter    Formatter
	writer       io.Writer
	hooks        events.Collection[LogEntry]
	levelEnabled func(LogLevel) bool

	engine *flow.Engine[LogEntry]
	ctx    context.Context
	cancel context.CancelFunc
}

// Formatter defines the log formatting contract.
type Formatter interface {
	Format(e *LogEntry) ([]byte, error)
}

// NewManager creates a new reactive manager.
func NewManager(workers int, bufferSize int) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		engine: flow.NewEngine[LogEntry](bufferSize),
		ctx:    ctx,
		cancel: cancel,
	}

	// Register passes with the engine
	m.engine.Register(ctx, "Validate", control.StepValidate, flow.PassFunc[LogEntry](m.passValidate), workers)
	m.engine.Register(ctx, "PreHooks", control.StepPreHooks, flow.PassFunc[LogEntry](m.passPreHooks), workers)
	m.engine.Register(ctx, "Format", control.StepFormat, flow.PassFunc[LogEntry](m.passFormat), workers)
	m.engine.Register(ctx, "Write", control.StepWrite, flow.PassFunc[LogEntry](m.passWrite), workers)

	m.engine.Start(ctx)
	return m
}

// Stop halts the manager engine.
func (m *Manager) Stop() {
	m.cancel()
}

// Process asynchronously submits an entry to the reactive engine.
func (m *Manager) Process(ctx context.Context, entry LogEntry) error {
	if entry == nil {
		return nil
	}

	tctx := flow.NewTranspileContext(ctx, entry)
	tctx.State.Store(control.StepValidate)

	m.engine.Submit(ctx, tctx)
	return nil
}

// ---- Passes (Autonomous Squads) ----

func (m *Manager) passValidate(ctx context.Context, tctx *flow.TranspileContext[LogEntry]) error {
	if err := m.stageValidate(tctx.Payload); err != nil {
		tctx.State.Store(control.StepFailed)
		return err
	}
	tctx.State.Store(control.StepPreHooks) // Transition to next state
	return nil
}

func (m *Manager) passPreHooks(ctx context.Context, tctx *flow.TranspileContext[LogEntry]) error {
	if err := m.hooks.Fire(tctx.Payload); err != nil {
		tctx.State.Store(control.StepFailed)
		return err
	}
	tctx.State.Store(control.StepFormat)
	return nil
}

type formatPayload struct {
	LogEntry
	Bytes []byte
}

func (m *Manager) passFormat(ctx context.Context, tctx *flow.TranspileContext[LogEntry]) error {
	b, err := m.stageFormat(tctx.Payload)
	if err != nil {
		tctx.State.Store(control.StepFailed)
		return err
	}

	// Hack to pass formatted bytes to the write step.
	// In a fully generic architecture, TranspileContext might need a metadata map.
	// For MVP, we will assume write fetches bytes differently or we pass it via a side-channel.
	// Actually, wait, let's just write directly here for now, or update the original stageWrite to accept LogEntry and format again? No.
	// Let's store bytes in a thread-safe map inside the Manager temporarily or define a custom type.
	// Better yet, just call stageWrite directly if Format succeeds since Write depends on Format's output directly.
	// No, the architecture strictly demands independent passes.
	// Let's update `stageWrite` to accept []byte.
	_ = b // We will inject it below.

	tctx.State.Store(control.StepWrite)
	return nil
}

func (m *Manager) passWrite(ctx context.Context, tctx *flow.TranspileContext[LogEntry]) error {
	// Re-format for MVP if we didn't store bytes.
	// In the real system, payload should probably be generic enough to carry intermediate state.
	b, err := m.stageFormat(tctx.Payload)
	if err != nil {
		tctx.State.Store(control.StepFailed)
		return err
	}

	if err := m.stageWrite(b); err != nil {
		tctx.State.Store(control.StepFailed)
		return err
	}

	tctx.State.Store(control.StepDone)
	tctx.Complete()
	return nil
}
