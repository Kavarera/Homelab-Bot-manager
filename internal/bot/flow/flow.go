package flow

import (
	"fmt"
)

// StepHandler defines the function signature for executing a single step in a flow.
type StepHandler func(ctx *Context) (Action, error)

// Flow defines the interface that all conversation flows must implement.
type Flow interface {
	// ID returns the unique identifier for the flow.
	ID() string
	// InitialStep returns the name of the first step when starting this flow.
	InitialStep() string
	// HandleStep routes execution to the handler of the specified step.
	HandleStep(ctx *Context, stepName string) (Action, error)
	// OnCancel is invoked when the flow is cancelled via cancel button.
	OnCancel(ctx *Context) error
}

// StepFlow is a standard implementation of Flow with declarative step mapping.
type StepFlow struct {
	id          string
	initialStep string
	steps       map[string]StepHandler
	cancelFunc  func(ctx *Context) error
}

// ID returns the unique identifier for this flow.
func (f *StepFlow) ID() string {
	return f.id
}

// InitialStep returns the name of the starting step.
func (f *StepFlow) InitialStep() string {
	return f.initialStep
}

// HandleStep executes the handler mapped to the given step name.
func (f *StepFlow) HandleStep(ctx *Context, stepName string) (Action, error) {
	handler, exists := f.steps[stepName]
	if !exists {
		return Stay(), fmt.Errorf("step %q not found in flow %q", stepName, f.id)
	}
	return handler(ctx)
}

// OnCancel is called when user cancels the flow.
func (f *StepFlow) OnCancel(ctx *Context) error {
	if f.cancelFunc != nil {
		return f.cancelFunc(ctx)
	}
	return ctx.ReplyAndRemoveKeyboard("❌ Operasi telah dibatalkan.")
}

// Builder helps create StepFlow instances declaratively.
type Builder struct {
	flow *StepFlow
}

// NewBuilder creates a new declarative Flow builder.
func NewBuilder(id string) *Builder {
	return &Builder{
		flow: &StepFlow{
			id:    id,
			steps: make(map[string]StepHandler),
		},
	}
}

// InitialStep registers the entry step of the flow.
func (b *Builder) InitialStep(name string, handler StepHandler) *Builder {
	b.flow.initialStep = name
	b.flow.steps[name] = handler
	return b
}

// Step registers a subsequent step in the flow.
func (b *Builder) Step(name string, handler StepHandler) *Builder {
	b.flow.steps[name] = handler
	return b
}

// OnCancel registers a custom cancellation handler.
func (b *Builder) OnCancel(handler func(ctx *Context) error) *Builder {
	b.flow.cancelFunc = handler
	return b
}

// Build creates and returns the Flow instance.
func (b *Builder) Build() Flow {
	return b.flow
}
