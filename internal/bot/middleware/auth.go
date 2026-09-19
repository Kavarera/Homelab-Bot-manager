package middleware

import (
	"fmt"
	"hs1-bot/internal/bot"
	"hs1-bot/internal/domain"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Auth creates a middleware that restricts bot access to only the allowed user ID.
func Auth(allowedUserID int64) bot.MiddlewareFunc {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx *bot.Context) error {
			if ctx.UserID != allowedUserID {
				slog.Warn("unauthorized access attempt",
					slog.Int64("user_id", ctx.UserID),
					slog.String("user_name", ctx.UserName),
					slog.String("command", ctx.Command),
					slog.String("raw_text", ctx.RawText),
				)

				// If callback query, send an alert popup
				if ctx.CallbackData != "" {
					_ = ctx.AnswerCallback("Akses ditolak!")
					return domain.ErrUnauthorized
				}

				// Send alert notification to the authorized owner
				warning := fmt.Sprintf("⚠️ *[UNAUTHORIZED ACCESS]*\nUser: @%s (`%d`)\nMessage: %s",
					ctx.UserName,
					ctx.UserID,
					ctx.RawText,
				)

				alertMsg := tgbotapi.NewMessage(allowedUserID, warning)
				alertMsg.ParseMode = "Markdown"
				_, _ = ctx.Sender.Send(alertMsg)

				return domain.ErrUnauthorized
			}

			return next(ctx)
		}
	}
}
