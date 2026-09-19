package bot

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Poller handles long-polling Telegram updates and dispatching them to the Router.
type Poller struct {
	bot    *tgbotapi.BotAPI
	router *Router
}

// NewPoller creates a new Poller instance.
func NewPoller(bot *tgbotapi.BotAPI, router *Router) *Poller {
	return &Poller{
		bot:    bot,
		router: router,
	}
}

// Start begins long-polling updates until the provided context is cancelled.
func (p *Poller) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60 // 60 seconds long polling timeout

	updates := p.bot.GetUpdatesChan(u)
	slog.Info("Telegram bot poller started listening for updates...")

	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping Telegram bot poller...")
			p.bot.StopReceivingUpdates()
			return nil
		case update, ok := <-updates:
			if !ok {
				slog.Info("Update channel closed")
				return nil
			}
			// Run update processing in a goroutine for responsiveness
			go p.router.HandleUpdate(update)
		}
	}
}
