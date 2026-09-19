package main

import (
	"context"
	"hs1-bot/config"
	"hs1-bot/internal/bot"
	"hs1-bot/internal/bot/flow"
	"hs1-bot/internal/bot/handlers"
	"hs1-bot/internal/bot/middleware"
	"hs1-bot/internal/bot/session"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/executor"
	"hs1-bot/internal/repository"
	"hs1-bot/internal/service"
	"hs1-bot/pkg/logger"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// 1. Initialize structured logger
	logger.InitLogger()

	// 2. Setup Context for Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Load & validate configuration (auto-discovers .env from current or parent directories)
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Starting bot",
		slog.String("server_name", cfg.ServerName),
		slog.Int64("allowed_user_id", cfg.AllowedUserID),
		slog.String("ssh_host", cfg.SSHHost),
		slog.String("database_path", cfg.DatabasePath),
	)

	// 4. Initialize SQLite Database & Repositories
	db, err := repository.InitDB(cfg.DatabasePath)
	if err != nil {
		slog.Error("Failed to initialize database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	if err := repository.SeedDefaultData(db); err != nil {
		slog.Warn("Failed to seed initial data", slog.Any("error", err))
	}

	clientRepo := repository.NewClientRepository(db)
	productRepo := repository.NewProductRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)

	// 5. Initialize Telegram Bot API Client
	telegramBot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		slog.Error("Failed to connect to Telegram API", slog.Any("error", err))
		os.Exit(1)
	}
	botAdapter := bot.NewTelegramBotAdapter(telegramBot)

	// 6. Initialize Infrastructure / Executors
	localExec := executor.NewLocalExecutor()
	sshExec := executor.NewSSHExecutor(localExec, executor.SSHConfig{
		Host:           cfg.SSHHost,
		PrivateKeyPath: cfg.SSHPrivateKeyPath,
	})

	// 7. Initialize Domain Services (Business Logic)
	sysService := service.NewSystemService(localExec)
	emailService := service.NewEmailService(cfg, "assets/logo.png", "assets/signature.png")
	backupService := service.NewBackupService(sshExec)

	// 8. Initialize Stateful Session Store & Flow Engine (1 Hour TTL)
	sessionStore := session.NewMemoryStore(1 * time.Hour)
	sessionStore.StartAutoCleanup(ctx, 10*time.Minute)

	flowEngine := flow.NewEngine(sessionStore, 1*time.Hour)

	// Register All Flows
	flowEngine.Register(flow.NewInvoiceFlow(clientRepo, productRepo, invoiceRepo, emailService))
	flowEngine.Register(flow.NewTambahClientFlow(clientRepo, productRepo))
	flowEngine.Register(flow.NewEditClientFlow(clientRepo, productRepo))
	flowEngine.Register(flow.NewHapusClientFlow(clientRepo))
	flowEngine.Register(flow.NewTambahProdukFlow(productRepo))
	flowEngine.Register(flow.NewEditProdukFlow(productRepo))
	flowEngine.Register(flow.NewHapusProdukFlow(productRepo))
	flowEngine.Register(flow.NewBackupFlow(backupService, cfg.BackupBasePath))

	// 9. Initialize Bot Handlers
	sysHandler := handlers.NewSystemHandler(sysService, cfg.ServerName)
	helpHandler := handlers.NewHelpHandler(cfg.ServerName)

	// 10. Initialize Router & Configure Middlewares
	router := bot.NewRouter(botAdapter)
	router.SetFlowDispatcher(flowEngine)
	router.Use(middleware.Recover())
	router.Use(middleware.Logger())
	router.Use(middleware.Auth(cfg.AllowedUserID))

	// 11. Register Commands
	router.RegisterCommand("start", helpHandler.HandleStart)
	router.RegisterCommand("menu", helpHandler.HandleMenu)
	router.RegisterCommand("help", helpHandler.HandleHelp)
	router.RegisterCommand("status", sysHandler.HandleStatus)
	router.RegisterCommand("backup", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.BackupFlowID)
	})

	// 12. Register Reply Keyboard Button Triggers
	// Kirim Invoice
	router.RegisterText(ui.ButtonKirimInvoice, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.InvoiceFlowID)
	})
	router.RegisterText("Kirim Invoice", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.InvoiceFlowID)
	})

	// Tambah Client
	router.RegisterText(ui.ButtonTambahClient, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.TambahClientFlowID)
	})
	router.RegisterText("Tambah Client", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.TambahClientFlowID)
	})

	// Edit Client
	router.RegisterText(ui.ButtonEditClient, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.EditClientFlowID)
	})
	router.RegisterText("Edit Client", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.EditClientFlowID)
	})

	// Hapus Client
	router.RegisterText(ui.ButtonHapusClient, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.HapusClientFlowID)
	})
	router.RegisterText("Hapus Client", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.HapusClientFlowID)
	})

	// Tambah Produk
	router.RegisterText(ui.ButtonTambahProduk, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.TambahProdukFlowID)
	})
	router.RegisterText("Tambah Produk", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.TambahProdukFlowID)
	})

	// Edit Produk
	router.RegisterText(ui.ButtonEditProduk, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.EditProdukFlowID)
	})
	router.RegisterText("Edit Produk", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.EditProdukFlowID)
	})

	// Hapus Produk
	router.RegisterText(ui.ButtonHapusProduk, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.HapusProdukFlowID)
	})
	router.RegisterText("Hapus Produk", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.HapusProdukFlowID)
	})

	// Backup Database
	router.RegisterText(ui.ButtonBackupDB, func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.BackupFlowID)
	})
	router.RegisterText("Backup Database", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.BackupFlowID)
	})
	router.RegisterText("Backup DB", func(c *bot.Context) error {
		return flowEngine.StartFlow(c, flow.BackupFlowID)
	})

	// Callback: Delete Encryption Key Message Immediately
	router.RegisterCallbackPrefix("delete_key:", func(c *bot.Context) error {
		parts := strings.Split(c.CallbackData, ":")
		if len(parts) == 2 {
			if msgID, err := strconv.Atoi(parts[1]); err == nil {
				delReq := tgbotapi.NewDeleteMessage(c.ChatID, msgID)
				_, _ = c.Sender.Request(delReq)
			}
		}
		_ = c.AnswerCallback("🔒 Kunci enkripsi berhasil dihapus dari chat demi keamanan.")
		return nil
	})

	// Tutup Menu
	router.RegisterText(ui.ButtonHideMenu, func(c *bot.Context) error {
		msg := tgbotapi.NewMessage(c.ChatID, "🔽 _Menu ditutup. Ketik /menu atau /start untuk membuka kembali._")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = ui.RemoveKeyboard()
		_, err := c.Sender.Send(msg)
		return err
	})
	router.RegisterText("Tutup Menu", func(c *bot.Context) error {
		msg := tgbotapi.NewMessage(c.ChatID, "🔽 _Menu ditutup. Ketik /menu atau /start untuk membuka kembali._")
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = ui.RemoveKeyboard()
		_, err := c.Sender.Send(msg)
		return err
	})

	router.SetDefaultHandler(helpHandler.HandleUnknown)

	// 13. Start Poller
	poller := bot.NewPoller(telegramBot, router)
	if err := poller.Start(ctx); err != nil {
		slog.Error("Poller encountered an error", slog.Any("error", err))
	}

	slog.Info("Bot shutdown completed cleanly.")
}
