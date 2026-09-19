package flow

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot"
	"hs1-bot/internal/bot/session"
	"hs1-bot/internal/bot/ui"
	"log/slog"
	"sync"
	"time"
)

// Engine is the central orchestrator that manages flows, active sessions, and step transitions.
type Engine struct {
	store session.Store
	flows map[string]Flow
	ttl   time.Duration
	mu    sync.RWMutex
}

// NewEngine creates a new Flow Engine instance.
func NewEngine(store session.Store, ttl time.Duration) *Engine {
	if ttl <= 0 {
		ttl = 1 * time.Hour
	}
	return &Engine{
		store: store,
		flows: make(map[string]Flow),
		ttl:   ttl,
	}
}

// Register registers one or more Flows into the engine.
func (e *Engine) Register(flows ...Flow) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, f := range flows {
		e.flows[f.ID()] = f
	}
}

// GetFlow retrieves a registered Flow by its ID.
func (e *Engine) GetFlow(flowID string) (Flow, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	f, exists := e.flows[flowID]
	return f, exists
}

// StartFlow initiates a flow for the user, executing its initial step immediately.
func (e *Engine) StartFlow(botCtx *bot.Context, flowID string) error {
	f, exists := e.GetFlow(flowID)
	if !exists {
		return fmt.Errorf("flow %q is not registered", flowID)
	}

	ctx := context.Background()
	sess := session.NewSession(botCtx.UserID, flowID, f.InitialStep(), e.ttl)
	if err := e.store.Save(ctx, sess); err != nil {
		return fmt.Errorf("failed to save initial flow session: %w", err)
	}

	flowCtx := NewContext(botCtx, sess)
	action, err := f.HandleStep(flowCtx, f.InitialStep())
	if err != nil {
		_ = e.store.Delete(ctx, botCtx.UserID)
		return err
	}

	return e.applyAction(ctx, f, flowCtx, sess, action)
}

// HandleActiveFlow checks if the user has an ongoing session and routes the message to the active step.
// Returns (handled = true) if the message was consumed by a flow, or (handled = false) if not.
func (e *Engine) HandleActiveFlow(botCtx *bot.Context) (bool, error) {
	ctx := context.Background()
	sess, err := e.store.Get(ctx, botCtx.UserID)
	if err != nil || sess == nil {
		return false, nil
	}

	f, exists := e.GetFlow(sess.FlowID)
	if !exists {
		// Flow no longer exists, clear invalid session
		_ = e.store.Delete(ctx, botCtx.UserID)
		return false, nil
	}

	flowCtx := NewContext(botCtx, sess)

	// Check if user clicked the Cancel button (e.g. "❌ Batal")
	if ui.IsCancelMessage(botCtx.RawText) {
		_ = f.OnCancel(flowCtx)
		_ = e.store.Delete(ctx, botCtx.UserID)
		slog.Info("flow cancelled via cancel button",
			slog.Int64("user_id", botCtx.UserID),
			slog.String("flow_id", sess.FlowID),
		)
		return true, nil
	}

	// Execute the current active step handler
	action, err := f.HandleStep(flowCtx, sess.CurrentStep)
	if err != nil {
		slog.Error("error handling flow step",
			slog.String("flow_id", sess.FlowID),
			slog.String("step", sess.CurrentStep),
			slog.Any("error", err),
		)
		return true, err
	}

	if applyErr := e.applyAction(ctx, f, flowCtx, sess, action); applyErr != nil {
		return true, applyErr
	}

	return true, nil
}

func (e *Engine) applyAction(ctx context.Context, f Flow, flowCtx *Context, sess *session.Session, action Action) error {
	switch action.Type {
	case ActionNext:
		sess.CurrentStep = action.NextStep
		sess.Touch(e.ttl)
		return e.store.Save(ctx, sess)

	case ActionStay:
		sess.Touch(e.ttl)
		return e.store.Save(ctx, sess)

	case ActionComplete:
		slog.Info("flow completed successfully",
			slog.Int64("user_id", sess.UserID),
			slog.String("flow_id", sess.FlowID),
		)
		return e.store.Delete(ctx, sess.UserID)

	case ActionCancel:
		_ = f.OnCancel(flowCtx)
		return e.store.Delete(ctx, sess.UserID)

	default:
		return nil
	}
}
