package logger

import (
	"log/slog"
	"os"
)

// InitLogger initializes and sets the default structured logger using slog.
func InitLogger() *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
