package middleware

import (
	"fmt"
	"hs1-bot/internal/bot"
	"log/slog"
	"runtime/debug"
)

// Recover creates a middleware that catches panics and prevents the bot from crashing.
func Recover() bot.MiddlewareFunc {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx *bot.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					slog.Error("recovered from panic in handler",
						slog.Any("panic", r),
						slog.String("stack", stack),
					)
					err = fmt.Errorf("panic occurred: %v", r)
					_ = ctx.Reply("❌ Terjadi kesalahan internal pada server (panic).")
				}
			}()
			return next(ctx)
		}
	}
}
