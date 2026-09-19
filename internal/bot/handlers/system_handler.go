package handlers

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot"
	"hs1-bot/internal/service"
)

// SystemHandler handles system inspection commands like /status.
type SystemHandler struct {
	systemService service.SystemService
	serverName    string
}

// NewSystemHandler creates a new SystemHandler.
func NewSystemHandler(sys service.SystemService, serverName string) *SystemHandler {
	return &SystemHandler{
		systemService: sys,
		serverName:    serverName,
	}
}

// HandleStatus handles the /status command.
func (h *SystemHandler) HandleStatus(ctx *bot.Context) error {
	status, err := h.systemService.GetSystemStatus(context.Background())
	if err != nil {
		return ctx.Reply(fmt.Sprintf("❌ Gagal mengambil status server: %v", err))
	}

	msg := fmt.Sprintf("*Server Status (%s)*:\n```\n%s\n```", h.serverName, status)
	return ctx.ReplyMarkdown(msg)
}
