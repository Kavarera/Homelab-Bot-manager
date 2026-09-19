package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Sender abstracts telegram-bot-api interactions for testability and loose coupling.
type Sender interface {
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
	Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error)
}

// TelegramBotAdapter adapts tgbotapi.BotAPI to the Sender interface.
type TelegramBotAdapter struct {
	bot *tgbotapi.BotAPI
}

// NewTelegramBotAdapter creates an adapter around a concrete tgbotapi.BotAPI instance.
func NewTelegramBotAdapter(bot *tgbotapi.BotAPI) *TelegramBotAdapter {
	return &TelegramBotAdapter{bot: bot}
}

// Send sends a message or chattable object to Telegram.
func (a *TelegramBotAdapter) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return a.bot.Send(c)
}

// Request sends an API request to Telegram (e.g. answer callback).
func (a *TelegramBotAdapter) Request(c tgbotapi.Chattable) (*tgbotapi.APIResponse, error) {
	return a.bot.Request(c)
}
