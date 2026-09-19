package flow

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/service"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const BackupFlowID = "backup_db"

// NewBackupFlow creates the multi-step conversational flow for remote PostgreSQL backups.
func NewBackupFlow(backupService service.BackupService, backupBasePath string) Flow {
	return NewBuilder(BackupFlowID).
		// Step 1: Query & display active Postgres containers from VPS via SSH
		InitialStep("pilih_container", func(c *Context) (Action, error) {
			ctx := context.Background()
			_ = c.Reply("⏳ _Mengecek container PostgreSQL aktif di VPS..._")

			containers, err := backupService.ListPostgresContainers(ctx)
			if err != nil {
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ Gagal mengambil daftar container di VPS: %v", err), ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			if len(containers) == 0 {
				_ = c.ReplyWithReplyKeyboard("❌ Tidak ditemukan container PostgreSQL yang sedang berjalan di VPS.", ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			var containerButtons []string
			var containerNames []string
			for _, ct := range containers {
				containerButtons = append(containerButtons, ct.Name)
				containerNames = append(containerNames, ct.Name)
			}

			c.Set("available_containers", strings.Join(containerNames, "|||"))

			prompt := fmt.Sprintf(
				"💾 *Backup Database PostgreSQL (VPS)*\n\n"+
					"Ditemukan *%d* container PostgreSQL aktif di VPS.\n"+
					"Silakan pilih container yang ingin dibackup:",
				len(containers),
			)
			_ = c.ReplyWithStepKeyboard(prompt, containerButtons...)
			return c.Next("proses_pilih_container"), nil
		}).

		// Step 2: Validate selected container and show confirmation prompt
		Step("proses_pilih_container", func(c *Context) (Action, error) {
			rawInput := strings.TrimSpace(c.RawText)
			availRaw, _ := c.GetString("available_containers")
			availableNames := strings.Split(availRaw, "|||")

			var matched string
			for _, name := range availableNames {
				if strings.EqualFold(strings.TrimSpace(name), rawInput) {
					matched = name
					break
				}
			}

			if matched == "" {
				var buttons []string
				for _, name := range availableNames {
					if strings.TrimSpace(name) != "" {
						buttons = append(buttons, name)
					}
				}
				_ = c.ReplyWithStepKeyboard("❌ Container tidak valid. Silakan pilih dari tombol:", buttons...)
				return c.Stay(), nil
			}

			c.Set("selected_container", matched)

			prompt := fmt.Sprintf(
				"⚠️ *Konfirmasi Backup Database VPS:*\n\n"+
					"📦 *Container:* `%s`\n"+
					"📍 *Penyimpanan Homelab:* `%s/%s/...`\n"+
					"🔒 *Keamanan:* Direct Stream (Zero Disk VPS) + Gzip + AES-256-GCM\n\n"+
					"Klik \"%s\" untuk memulai proses backup sekarang.",
				matched,
				backupBasePath,
				matched,
				ui.ButtonConfirm,
			)

			kb := ui.BuildStepKeyboardFlat(ui.ButtonConfirm)
			_ = c.ReplyWithReplyKeyboard(prompt, kb)
			return c.Next("eksekusi_backup"), nil
		}).

		// Step 3: Execute dump, compression, encryption, document sending, and key delivery
		Step("eksekusi_backup", func(c *Context) (Action, error) {
			if !ui.IsConfirmMessage(c.RawText) {
				_ = c.ReplyWithReplyKeyboard("Klik '"+ui.ButtonConfirm+"' untuk memulai backup atau '"+ui.DefaultCancelButton+"' untuk membatalkan.", ui.BuildStepKeyboardFlat(ui.ButtonConfirm))
				return c.Stay(), nil
			}

			containerName, _ := c.GetString("selected_container")
			ctx := context.Background()

			_ = c.Reply(fmt.Sprintf(
				"⏳ *Memproses Backup Database `%s`...*\n\n"+
					"1. Mengalirkan dump SQL dari VPS via Tailscale SSH...\n"+
					"2. Mengompresi stream dengan Gzip...\n"+
					"3. Mengenkripsi file dengan AES-256-GCM...\n"+
					"4. Menyimpan salinan ke server Homelab...",
				containerName,
			))

			result, err := backupService.DumpAndEncrypt(ctx, containerName, backupBasePath)
			if err != nil {
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ *Gagal melakukan backup database!*\n\nDetail: `%v`", err), ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			// 1. Send encrypted document to Telegram
			docMsg := tgbotapi.NewDocument(c.ChatID, tgbotapi.FileBytes{
				Name:  result.Filename,
				Bytes: result.EncryptedData,
			})
			docMsg.Caption = fmt.Sprintf(
				"📦 *File Backup Terenkripsi: %s*\n"+
					"• Ukuran Dump Asli: `%s`\n"+
					"• Ukuran Terenkripsi: `%s`",
				result.ContainerName,
				humanize.Bytes(uint64(result.RawSizeBytes)),
				humanize.Bytes(uint64(result.EncryptedSizeBytes)),
			)
			docMsg.ParseMode = "Markdown"
			_, err = c.Sender.Send(docMsg)
			if err != nil {
				_ = c.Reply(fmt.Sprintf("⚠️ Gagal mengirim file dokumen ke Telegram: %v", err))
			}

			// 2. Send encryption key message with self-destruct mechanism
			keyPrompt := fmt.Sprintf(
				"🔐 *KUNCI ENKRIPSI BACKUP DATABASE*\n\n"+
					"📦 *Container:* `%s`\n"+
					"📁 *File:* `%s`\n"+
					"📍 *Lokasi Homelab:* `%s`\n\n"+
					"🔑 *Kunci Enkripsi (AES-256 Hex):*\n`%s`\n\n"+
					"⚠️ *PENTING & RAHASIA:*\n"+
					"Simpan kunci ini segera di Password Manager Anda!\n"+
					"Pesan ini akan *OTOMATIS DIHAPUS DALAM 5 MENIT* atau klik tombol di bawah untuk menghapusnya sekarang.",
				result.ContainerName,
				result.Filename,
				result.LocalPath,
				result.KeyHex,
			)

			keyMsg := tgbotapi.NewMessage(c.ChatID, keyPrompt)
			keyMsg.ParseMode = "Markdown"
			sentKeyMsg, sendErr := c.Sender.Send(keyMsg)

			if sendErr == nil {
				// Attach inline delete button to the key message
				inlineKB := ui.BuildDeleteKeyInlineKeyboard(sentKeyMsg.MessageID)
				editMarkup := tgbotapi.NewEditMessageReplyMarkup(c.ChatID, sentKeyMsg.MessageID, inlineKB)
				_, _ = c.Sender.Request(editMarkup)

				// Background Goroutine timer: Auto-delete key message after 5 minutes
				keyMsgID := sentKeyMsg.MessageID
				chatID := c.ChatID
				sender := c.Sender
				time.AfterFunc(5*time.Minute, func() {
					delReq := tgbotapi.NewDeleteMessage(chatID, keyMsgID)
					_, _ = sender.Request(delReq)
				})
			}

			// 3. Final summary reply with Main Menu Keyboard
			successMsg := fmt.Sprintf(
				"✅ *Backup Database Selesai & Berhasil!*\n\n"+
					"📦 Container: *%s*\n"+
					"📁 File: `%s`\n"+
					"📍 Disimpan di Homelab: `%s`\n"+
					"🔒 Status: *Terkonversi Gzip & Terenkripsi AES-256*\n\n"+
					"_VPS tetap 100%% bersih tanpa file sementara._",
				result.ContainerName,
				result.Filename,
				result.LocalPath,
			)
			_ = c.ReplyWithReplyKeyboard(successMsg, ui.MainMenuKeyboard())
			return c.Complete(), nil
		}).
		Build()
}

// Ensure interface compatibility
var _ domain.PostgresContainer
