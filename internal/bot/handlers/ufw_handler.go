package handlers

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/service"
	"strings"
)

const requiredSecurityKeyword = "payrollpro"

// UFWHandler handles firewall management commands like /ufw53.
type UFWHandler struct {
	firewallService service.FirewallService
	serverName      string
}

// NewUFWHandler creates a new UFWHandler.
func NewUFWHandler(fw service.FirewallService, serverName string) *UFWHandler {
	return &UFWHandler{
		firewallService: fw,
		serverName:      serverName,
	}
}

// HandleTogglePort53 handles the /ufw53 command.
func (h *UFWHandler) HandleTogglePort53(ctx *bot.Context) error {
	// Command syntax: /ufw53 <on|off> <payrollpro>
	if len(ctx.CommandArgs) < 2 {
		return ctx.ReplyMarkdown("❌ Format salah. Gunakan format:\n`/ufw53 on payrollpro` atau `/ufw53 off payrollpro`")
	}

	actionStr := strings.ToLower(ctx.CommandArgs[0])
	keyword := strings.ToLower(ctx.CommandArgs[1])

	if keyword != requiredSecurityKeyword {
		return ctx.Reply("❌ Keyword / identifier server salah!")
	}

	action, err := domain.ParseUFWAction(actionStr)
	if err != nil {
		return ctx.Reply("❌ Aksi harus menggunakan 'on' atau 'off'.")
	}

	if err := h.firewallService.TogglePort53(context.Background(), action); err != nil {
		return ctx.Reply(fmt.Sprintf("❌ Gagal mengubah UFW port 53: %v", err))
	}

	statusText := "DIAKTIFKAN (Allow)"
	if action == domain.UFWActionDeny {
		statusText = "DINONAKTIFKAN (Deny)"
	}

	reply := fmt.Sprintf("✅ *UFW Port 53 (TCP/UDP)* berhasil *%s* untuk server *%s*.", statusText, h.serverName)
	return ctx.ReplyMarkdown(reply)
}
