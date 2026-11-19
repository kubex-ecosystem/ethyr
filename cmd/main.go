package main

import (
	"context"

	"github.com/kubex-ecosystem/ethyr/tools"
	"github.com/kubex-ecosystem/ethyr/tools/flow"
	"github.com/kubex-ecosystem/ethyr/tools/flow/fsm"
	"github.com/kubex-ecosystem/ethyr/tools/flow/queue"
)

// MMin is the main function of the application.
func MMin() {
	kernel := NewKernel()
	_ = kernel.fsm.Trigger(EvStart)
}

// Estados e Eventos do Kernel (Atômicos)
const (
	// StateIdle state when the kernel is idle
	StateIdle fsm.State = 1 << iota
	// StateRunning state when the kernel is running
	StateRunning
	// StateStopped state when the kernel is stopped
	StateStopped
)

const (
	// EvStart event to start the kernel
	EvStart fsm.Event = 1 << iota
	// EvStop event to stop the kernel
	EvStop
	// EvReset event to reset the kernel
	EvReset
)

// Kernel is a simple example of a kernel with an atomic state machine, a queue, and a retryer.
type Kernel struct {
	fsm   *fsm.FSM
	queue *queue.Queue
	retry *tools.Retryer
	flow  *flow.Flow
}

// NewKernel creates a new Kernel with the given initial state and transitions.
func NewKernel() *Kernel {
	transitions := []fsm.Transition{
		{From: StateIdle, Event: EvStart, To: StateRunning},
		{From: StateRunning, Event: EvStop, To: StateStopped},
		{From: StateStopped, Event: EvReset, To: StateIdle},
	}

	fsm := fsm.NewFSM(StateIdle, transitions)
	queue := queue.NewQueue(100)

	retry := tools.NewRetryer(tools.RetryConfig{
		Retries:       5,
		BackoffFactor: 2.0,
		Timeout:       0,
	})

	return &Kernel{
		fsm:   fsm,
		queue: queue,
		retry: retry,
		flow:  flow.NewFlow(),
	}
}

// Execute demonstra a submissão de uma tarefa atômica
func (k *Kernel) Execute(ctx context.Context, task func(context.Context) error) error {
	return k.flow.Submit(ctx, task)
}
