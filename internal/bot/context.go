package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Context encapsulates the data and sender for a Telegram update.
type Context struct {
	Update       tgbotapi.Update
	Sender       Sender
	ChatID       int64
	UserID       int64
	UserName     string
	Command      string
	CommandArgs  []string
	RawText      string
	CallbackData string
	CallbackID   string
}

// NewContext creates a new Context from a Telegram update and Sender.
func NewContext(update tgbotapi.Update, sender Sender) *Context {
	ctx := &Context{
		Update: update,
		Sender: sender,
	}

	if update.Message != nil {
		ctx.ChatID = update.Message.Chat.ID
		if update.Message.From != nil {
			ctx.UserID = update.Message.From.ID
			ctx.UserName = update.Message.From.UserName
		}
		ctx.RawText = update.Message.Text
		ctx.Command = update.Message.Command()

		// Extract command arguments
		parts := strings.Fields(update.Message.Text)
		if len(parts) > 1 {
			ctx.CommandArgs = parts[1:]
		}
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Message != nil {
			ctx.ChatID = update.CallbackQuery.Message.Chat.ID
		}
		if update.CallbackQuery.From != nil {
			ctx.UserID = update.CallbackQuery.From.ID
			ctx.UserName = update.CallbackQuery.From.UserName
		}
		ctx.CallbackData = update.CallbackQuery.Data
		ctx.CallbackID = update.CallbackQuery.ID
	}

	return ctx
}

// Reply sends a plain text message back to the current chat.
func (c *Context) Reply(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID, text)
	_, err := c.Sender.Send(msg)
	return err
}

// ReplyMarkdown sends a Markdown-formatted message back to the current chat.
func (c *Context) ReplyMarkdown(text string) error {
	msg := tgbotapi.NewMessage(c.ChatID, text)
	msg.ParseMode = "Markdown"
	_, err := c.Sender.Send(msg)
	return err
}

// ReplyWithKeyboard sends a message with an inline keyboard markup.
func (c *Context) ReplyWithKeyboard(text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(c.ChatID, text)
	msg.ReplyMarkup = keyboard
	_, err := c.Sender.Send(msg)
	return err
}

// ReplyWithReplyKeyboard sends a message with a custom ReplyKeyboardMarkup.
func (c *Context) ReplyWithReplyKeyboard(text string, keyboard tgbotapi.ReplyKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(c.ChatID, text)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"
	_, err := c.Sender.Send(msg)
	return err
}

// AnswerCallback sends a response to an inline button click (stopping the loading spinner).
func (c *Context) AnswerCallback(text string) error {
	if c.CallbackID == "" {
		return nil
	}
	resp := tgbotapi.NewCallback(c.CallbackID, text)
	_, err := c.Sender.Request(resp)
	return err
}
