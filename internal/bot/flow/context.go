package flow

import (
	"hs1-bot/internal/bot"
	"hs1-bot/internal/bot/session"
	"hs1-bot/internal/bot/ui"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context wraps the Telegram bot context and the user's active conversation session.
type Context struct {
	*bot.Context
	Session *session.Session
}

// NewContext creates a new Flow Context.
func NewContext(botCtx *bot.Context, sess *session.Session) *Context {
	return &Context{
		Context: botCtx,
		Session: sess,
	}
}

// Set stores a key-value pair in the user's active session.
func (c *Context) Set(key string, val any) {
	if c.Session != nil {
		c.Session.Set(key, val)
	}
}

// Get retrieves a value from the active session.
func (c *Context) Get(key string) (any, bool) {
	if c.Session == nil {
		return nil, false
	}
	return c.Session.Get(key)
}

// GetString retrieves a string value from the active session.
func (c *Context) GetString(key string) (string, bool) {
	if c.Session == nil {
		return "", false
	}
	return c.Session.GetString(key)
}

// GetInt retrieves an integer value from the active session.
func (c *Context) GetInt(key string) (int, bool) {
	if c.Session == nil {
		return 0, false
	}
	return c.Session.GetInt(key)
}

// ReplyWithStepKeyboard sends a prompt message with the given buttons plus the cancel button.
func (c *Context) ReplyWithStepKeyboard(text string, buttons ...string) error {
	keyboard := ui.BuildStepKeyboardFlat(buttons...)
	msg := tgbotapi.NewMessage(c.ChatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	_, err := c.Sender.Send(msg)
	return err
}

// ReplyWithStepKeyboardGrid sends a prompt message with a grid of buttons plus the cancel button.
func (c *Context) ReplyWithStepKeyboardGrid(text string, rows [][]string) error {
	keyboard := ui.BuildStepKeyboard(rows)
	msg := tgbotapi.NewMessage(c.ChatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	_, err := c.Sender.Send(msg)
	return err
}

// ReplyWithCancelKeyboard sends a prompt with only the cancel button (for free-text inputs).
func (c *Context) ReplyWithCancelKeyboard(text string) error {
	keyboard := ui.CancelOnlyKeyboard()
	msg := tgbotapi.NewMessage(c.ChatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	_, err := c.Sender.Send(msg)
	return err
}

// ReplyAndRemoveKeyboard sends a final message and hides/removes the reply keyboard.
func (c *Context) ReplyAndRemoveKeyboard(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID, text)
	msg.ReplyMarkup = ui.RemoveKeyboard()
	msg.ParseMode = "Markdown"
	_, err := c.Sender.Send(msg)
	return err
}

// Next transitions to the specified next step.
func (c *Context) Next(nextStep string) Action {
	return Next(nextStep)
}

// Stay remains on the current step.
func (c *Context) Stay() Action {
	return Stay()
}

// Complete completes the flow and clears session.
func (c *Context) Complete() Action {
	return Complete()
}

// Cancel cancels the flow and clears session.
func (c *Context) Cancel() Action {
	return Cancel()
}
