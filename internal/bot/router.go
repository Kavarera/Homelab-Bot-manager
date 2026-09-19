package bot

import (
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerFunc defines the function signature for bot update handlers.
type HandlerFunc func(ctx *Context) error

// MiddlewareFunc defines the function signature for router middlewares.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// FlowDispatcher defines the interface for handling active multi-step conversation flows.
type FlowDispatcher interface {
	HandleActiveFlow(ctx *Context) (bool, error)
}

// Router manages routing of Telegram commands, callbacks, text buttons, and conversation flows.
type Router struct {
	sender           Sender
	middlewares      []MiddlewareFunc
	commands         map[string]HandlerFunc
	textHandlers     map[string]HandlerFunc
	callbackPrefixes map[string]HandlerFunc
	defaultHandler   HandlerFunc
	flowDispatcher   FlowDispatcher
}

// NewRouter creates a new Router instance.
func NewRouter(sender Sender) *Router {
	return &Router{
		sender:           sender,
		commands:         make(map[string]HandlerFunc),
		textHandlers:     make(map[string]HandlerFunc),
		callbackPrefixes: make(map[string]HandlerFunc),
	}
}

// SetFlowDispatcher sets the flow engine to handle active multi-step conversations.
func (r *Router) SetFlowDispatcher(dispatcher FlowDispatcher) {
	r.flowDispatcher = dispatcher
}

// Use adds one or more middlewares to the router pipeline.
func (r *Router) Use(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

// RegisterCommand registers a handler for a specific Telegram slash command (without the leading slash).
func (r *Router) RegisterCommand(cmd string, handler HandlerFunc) {
	r.commands[strings.ToLower(strings.TrimPrefix(cmd, "/"))] = handler
}

// RegisterText registers a handler for specific plain text or reply button label.
func (r *Router) RegisterText(text string, handler HandlerFunc) {
	r.textHandlers[strings.ToLower(strings.TrimSpace(text))] = handler
}

// RegisterCallbackPrefix registers a handler for callback query data starting with the given prefix.
func (r *Router) RegisterCallbackPrefix(prefix string, handler HandlerFunc) {
	r.callbackPrefixes[prefix] = handler
}

// SetDefaultHandler sets the fallback handler for unknown commands or messages.
func (r *Router) SetDefaultHandler(handler HandlerFunc) {
	r.defaultHandler = handler
}

// HandleUpdate processes an incoming Telegram update through the middleware pipeline and routes it.
func (r *Router) HandleUpdate(update tgbotapi.Update) {
	ctx := NewContext(update, r.sender)

	// If update is neither message nor callback query, skip
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	handler := r.resolveHandler(ctx)
	if handler == nil {
		return
	}

	// Wrap handler in middleware pipeline
	pipeline := r.buildPipeline(handler)

	if err := pipeline(ctx); err != nil {
		slog.Error("error handling update",
			slog.Int64("chat_id", ctx.ChatID),
			slog.Int64("user_id", ctx.UserID),
			slog.String("command", ctx.Command),
			slog.String("raw_text", ctx.RawText),
			slog.String("callback", ctx.CallbackData),
			slog.Any("error", err),
		)
	}
}

func (r *Router) resolveHandler(ctx *Context) HandlerFunc {
	// 1. Handle Callback Queries
	if ctx.CallbackData != "" {
		for prefix, handler := range r.callbackPrefixes {
			if strings.HasPrefix(ctx.CallbackData, prefix) {
				return handler
			}
		}
		// Fallback for callback: answer empty to remove loading state
		return func(c *Context) error {
			return c.AnswerCallback("")
		}
	}

	// 2. Wrap message handling with Flow checking
	return func(c *Context) error {
		// Check if user has an active conversation flow
		if r.flowDispatcher != nil {
			handled, err := r.flowDispatcher.HandleActiveFlow(c)
			if err != nil {
				return err
			}
			if handled {
				return nil
			}
		}

		// 3. Handle Slash Commands
		if c.Command != "" {
			if handler, ok := r.commands[strings.ToLower(c.Command)]; ok {
				return handler(c)
			}
		}

		// 4. Handle Registered Plain Text / Reply Buttons
		if c.RawText != "" {
			if handler, ok := r.textHandlers[strings.ToLower(strings.TrimSpace(c.RawText))]; ok {
				return handler(c)
			}
		}

		// 5. Fallback / Default Handler
		if r.defaultHandler != nil {
			return r.defaultHandler(c)
		}
		return nil
	}
}

func (r *Router) buildPipeline(handler HandlerFunc) HandlerFunc {
	current := handler
	// Wrap in reverse order so middlewares execute in the order they were added
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		current = r.middlewares[i](current)
	}
	return current
}
