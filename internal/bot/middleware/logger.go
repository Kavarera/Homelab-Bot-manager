package middleware

import (
	"hs1-bot/internal/bot"
	"log/slog"
	"time"
)

// Logger creates a middleware that logs each incoming interaction with execution time.
func Logger() bot.MiddlewareFunc {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx *bot.Context) error {
			start := time.Now()

			slog.Info("processing update",
				slog.Int64("chat_id", ctx.ChatID),
				slog.Int64("user_id", ctx.UserID),
				slog.String("command", ctx.Command),
				slog.String("callback", ctx.CallbackData),
			)

			err := next(ctx)

			duration := time.Since(start)
			if err != nil {
				slog.Error("update processed with error",
					slog.Duration("latency", duration),
					slog.Any("error", err),
				)
			} else {
				slog.Info("update processed successfully",
					slog.Duration("latency", duration),
				)
			}

			return err
		}
	}
}
