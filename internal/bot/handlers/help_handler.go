package handlers

import (
	"fmt"
	"hs1-bot/internal/bot"
	"hs1-bot/internal/bot/ui"
)

// HelpHandler provides help and fallback responses.
type HelpHandler struct {
	serverName string
}

// NewHelpHandler creates a new HelpHandler.
func NewHelpHandler(serverName string) *HelpHandler {
	return &HelpHandler{
		serverName: serverName,
	}
}

// HandleStart handles the /start command, greeting user and showing the main menu keyboard.
func (h *HelpHandler) HandleStart(ctx *bot.Context) error {
	msg := fmt.Sprintf(`🤖 *Selamat Datang di HS1 Server Bot* (%s)

Silakan gunakan tombol menu di bawah atau ketik perintah yang diinginkan.
Ketik /help untuk panduan lengkap.`, h.serverName)

	return ctx.ReplyWithReplyKeyboard(msg, ui.MainMenuKeyboard())
}

// HandleMenu handles the /menu command to display the main menu keyboard.
func (h *HelpHandler) HandleMenu(ctx *bot.Context) error {
	msg := fmt.Sprintf("📋 *Menu Utama* (%s)\nSilakan pilih opsi menu di bawah:", h.serverName)
	return ctx.ReplyWithReplyKeyboard(msg, ui.MainMenuKeyboard())
}

// HandleHelp handles the /help command.
func (h *HelpHandler) HandleHelp(ctx *bot.Context) error {
	msg := fmt.Sprintf(`🤖 *HS1 Server Bot* (%s)

Perintah yang tersedia:
• /menu - Menampilkan menu utama
• /status - Cek status dan uptime server
• /help - Menampilkan panduan ini
`, h.serverName)

	return ctx.ReplyWithReplyKeyboard(msg, ui.MainMenuKeyboard())
}

// HandleUnknown handles unknown commands or unrecognized plain messages.
func (h *HelpHandler) HandleUnknown(ctx *bot.Context) error {
	return ctx.Reply("Perintah tidak dikenali. Ketik /help atau /menu untuk melihat daftar perintah.")
}
