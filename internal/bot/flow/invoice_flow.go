package flow

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/repository"
	"hs1-bot/internal/service"
	"hs1-bot/internal/service/template"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const InvoiceFlowID = "kirim_invoice"

// NewInvoiceFlow creates the multi-product "kirim_invoice" conversation flow.
func NewInvoiceFlow(
	clientRepo repository.ClientRepository,
	productRepo repository.ProductRepository,
	invoiceRepo repository.InvoiceRepository,
	emailService service.EmailService,
) Flow {
	return NewBuilder(InvoiceFlowID).
		// Step 1: Query & display list of clients
		InitialStep("pilih_client", func(c *Context) (Action, error) {
			ctx := context.Background()
			clients, err := clientRepo.List(ctx)
			if err != nil {
				_ = c.ReplyAndRemoveKeyboard(fmt.Sprintf("❌ Gagal mengambil data client: %v", err))
				return c.Complete(), nil
			}

			if len(clients) == 0 {
				_ = c.ReplyAndRemoveKeyboard("❌ Belum ada client yang terdaftar di database.")
				return c.Complete(), nil
			}

			var clientButtons []string
			for _, client := range clients {
				clientButtons = append(clientButtons, client.CompanyName)
			}

			_ = c.ReplyWithStepKeyboard("🏢 *Pilih Client Tujuan:*", clientButtons...)
			return c.Next("proses_pilih_client"), nil
		}).

		// Step 2: Receive selected client & initialize multi-select product keyboard
		Step("proses_pilih_client", func(c *Context) (Action, error) {
			ctx := context.Background()
			clients, err := clientRepo.List(ctx)
			if err != nil {
				_ = c.ReplyAndRemoveKeyboard(fmt.Sprintf("❌ Gagal membaca data client: %v", err))
				return c.Complete(), nil
			}

			selectedText := strings.TrimSpace(c.RawText)
			var selectedClient *domain.Client

			for i := range clients {
				if strings.EqualFold(strings.TrimSpace(clients[i].CompanyName), selectedText) {
					selectedClient = &clients[i]
					break
				}
			}

			if selectedClient == nil {
				var clientButtons []string
				for _, client := range clients {
					clientButtons = append(clientButtons, client.CompanyName)
				}
				_ = c.ReplyWithStepKeyboard("❌ Client tidak ditemukan. Silakan pilih dari tombol yang tersedia:", clientButtons...)
				return c.Stay(), nil
			}

			c.Set("client_id", selectedClient.ID)
			c.Set("client_company", selectedClient.CompanyName)

			products, err := productRepo.List(ctx)
			if err != nil || len(products) == 0 {
				_ = c.ReplyAndRemoveKeyboard("❌ Belum ada produk yang terdaftar di database.")
				return c.Complete(), nil
			}

			// Determine available product names for this client
			var availableProductNames []string
			clientOrdered := selectedClient.GetProductList()
			if len(clientOrdered) > 0 {
				availableProductNames = clientOrdered
			} else {
				for _, p := range products {
					availableProductNames = append(availableProductNames, p.Name)
				}
			}

			c.Set("available_products", strings.Join(availableProductNames, "|||"))
			c.Set("selected_products", "")

			kb := ui.BuildMultiSelectKeyboard(availableProductNames, nil, ui.ButtonSubmit)
			prompt := fmt.Sprintf(
				"🏢 Client: *%s*\n👤 PIC: *%s*\n\n"+
					"📦 *Pilih Produk untuk Invoice:*\n"+
					"_Klik nama produk untuk memilih/membatalkan pilihan, lalu klik \"%s\" jika sudah selesai._",
				selectedClient.CompanyName,
				selectedClient.PICName,
				ui.ButtonSubmit,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, kb)
			return c.Next("proses_pilih_product"), nil
		}).

		// Step 3: Handle multi-product selection toggles and submit action
		Step("proses_pilih_product", func(c *Context) (Action, error) {
			ctx := context.Background()
			rawInput := strings.TrimSpace(c.RawText)

			clientIDInt, _ := c.GetInt("client_id")
			clientID := int64(clientIDInt)

			client, err := clientRepo.GetByID(ctx, clientID)
			if err != nil || client == nil {
				_ = c.ReplyAndRemoveKeyboard("❌ Data client tidak ditemukan.")
				return c.Complete(), nil
			}

			availRaw, _ := c.GetString("available_products")
			var availableProductNames []string
			if availRaw != "" {
				availableProductNames = strings.Split(availRaw, "|||")
			}

			selectedRaw, _ := c.GetString("selected_products")
			var selectedList []string
			if selectedRaw != "" {
				selectedList = strings.Split(selectedRaw, "|||")
			}

			// If user clicked Submit / Konfirmasi button
			if ui.IsSubmitMessage(rawInput) {
				if len(selectedList) == 0 {
					kb := ui.BuildMultiSelectKeyboard(availableProductNames, selectedList, ui.ButtonSubmit)
					_ = c.ReplyWithReplyKeyboard("❌ Anda belum memilih produk. Silakan klik minimal 1 produk lalu klik "+ui.ButtonSubmit, kb)
					return c.Stay(), nil
				}

				// Load all products to build invoice
				allProducts, err := productRepo.List(ctx)
				if err != nil {
					_ = c.ReplyAndRemoveKeyboard(fmt.Sprintf("❌ Gagal mengambil produk: %v", err))
					return c.Complete(), nil
				}

				prodMap := make(map[string]domain.Product)
				for _, p := range allProducts {
					prodMap[strings.ToLower(strings.TrimSpace(p.Name))] = p
				}

				now := time.Now()
				var selectedProducts []domain.Product
				var invoiceItems []domain.InvoiceItem
				var totalPrice float64

				// Validation: Check monthly duplicate for every selected product
				for _, pName := range selectedList {
					prod, ok := prodMap[strings.ToLower(strings.TrimSpace(pName))]
					if !ok {
						continue
					}

					existingInv, err := invoiceRepo.CheckMonthlyInvoiceExists(ctx, client.ID, prod.ID, now.Year(), int(now.Month()))
					if err != nil {
						_ = c.ReplyAndRemoveKeyboard(fmt.Sprintf("❌ Gagal memvalidasi invoice bulanan: %v", err))
						return c.Complete(), nil
					}

					if existingInv != nil {
						rejectionMsg := fmt.Sprintf(
							"⚠️ *[DUPLIKASI INVOICE DITOLAK]*\n\n"+
								"Invoice untuk client *%s* dengan produk *%s* sudah pernah diterbitkan pada bulan ini (*%s %d*).\n\n"+
								"📄 Nomor Invoice: `%s`\n"+
								"📅 Tanggal Terbit: `%s`\n"+
								"💵 Total: *IDR %s*",
							client.CompanyName,
							prod.Name,
							now.Month().String(),
							now.Year(),
							existingInv.InvoiceNumber,
							existingInv.IssueDate.Format("02.01.2006"),
							template.FormatCurrencyShort(existingInv.TotalPrice),
						)
						_ = c.ReplyWithReplyKeyboard(rejectionMsg, ui.MainMenuKeyboard())
						return c.Complete(), nil
					}

					selectedProducts = append(selectedProducts, prod)
					invoiceItems = append(invoiceItems, domain.InvoiceItem{
						ProductID: prod.ID,
						Qty:       1,
						UnitPrice: prod.Price,
						Subtotal:  prod.Price,
					})
					totalPrice += prod.Price
				}

				if len(selectedProducts) == 0 {
					_ = c.ReplyAndRemoveKeyboard("❌ Tidak ada produk valid yang dipilih.")
					return c.Complete(), nil
				}

				dueDate := domain.CalculateDueDate(now)
				invoiceNumber := domain.GenerateInvoiceNumber(client.ID, now)

				invoice := &domain.Invoice{
					ClientID:      client.ID,
					InvoiceNumber: invoiceNumber,
					TotalPrice:    totalPrice,
					IssueDate:     now,
					DueDate:       dueDate,
					Status:        "SENT",
					Items:         invoiceItems,
				}

				_ = c.Reply("⏳ Sedang memproses dan mengirim email invoice...")

				// Send Email via Gomail first (with PDF attachment)
				pdfBytes, err := emailService.SendInvoiceEmail(ctx, client.CompanyEmail, invoice, client, selectedProducts)
				if err != nil {
					_ = c.ReplyWithReplyKeyboard(
						fmt.Sprintf("❌ *Gagal Mengirim Email Invoice!*\n\nDetail Error: `%v`\n\n_Pembuatan invoice dibatalkan dan tidak disimpan ke database._", err),
						ui.MainMenuKeyboard(),
					)
					return c.Complete(), nil
				}

				// Persist invoice & items to database only after email successfully sent
				if err := invoiceRepo.Create(ctx, invoice); err != nil {
					_ = c.ReplyWithReplyKeyboard(
						fmt.Sprintf("⚠️ Email berhasil terkirim ke `%s`, namun gagal menyimpan invoice ke database: %v", client.CompanyEmail, err),
						ui.MainMenuKeyboard(),
					)
					return c.Complete(), nil
				}

				// Send generated PDF document to user via Telegram
				safeFilename := fmt.Sprintf("Invoice-%s.pdf", strings.ReplaceAll(invoice.InvoiceNumber, "/", "-"))
				docMsg := tgbotapi.NewDocument(c.ChatID, tgbotapi.FileBytes{
					Name:  safeFilename,
					Bytes: pdfBytes,
				})
				docMsg.Caption = fmt.Sprintf("📄 Dokumen Invoice PDF: %s", invoice.InvoiceNumber)
				_, _ = c.Sender.Send(docMsg)

				// Product list summary string
				var prodSummary []string
				for _, p := range selectedProducts {
					prodSummary = append(prodSummary, fmt.Sprintf("  • %s (IDR %s)", p.Name, template.FormatCurrencyShort(p.Price)))
				}

				// Reply success summary
				successMsg := fmt.Sprintf(
					"✅ *Invoice Berhasil Diterbitkan & Terkirim!*\n\n"+
						"📄 Nomor Invoice: `%s`\n"+
						"🏢 Client: *%s*\n"+
						"👤 PIC: *%s*\n"+
						"✉️ Email Tujuan: `%s`\n"+
						"📦 *Produk (%d item):*\n%s\n\n"+
						"💵 Total: *IDR %s (NETT)*\n"+
						"📅 Jatuh Tempo: `%s`",
					invoice.InvoiceNumber,
					client.CompanyName,
					client.PICName,
					client.CompanyEmail,
					len(selectedProducts),
					strings.Join(prodSummary, "\n"),
					template.FormatCurrencyShort(invoice.TotalPrice),
					invoice.DueDate.Format("02.01.2006"),
				)

				_ = c.ReplyWithReplyKeyboard(successMsg, ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			// Handle Toggle product selection
			cleanName := strings.TrimPrefix(rawInput, "✅ ")
			cleanName = strings.TrimSpace(cleanName)

			var matched string
			for _, av := range availableProductNames {
				if strings.EqualFold(strings.TrimSpace(av), cleanName) {
					matched = av
					break
				}
			}

			if matched == "" {
				kb := ui.BuildMultiSelectKeyboard(availableProductNames, selectedList, ui.ButtonSubmit)
				_ = c.ReplyWithReplyKeyboard("❌ Pilihan tidak dikenali. Silakan klik tombol produk yang tersedia:", kb)
				return c.Stay(), nil
			}

			// Toggle presence
			isAlreadySelected := false
			var updatedList []string
			for _, s := range selectedList {
				if strings.EqualFold(s, matched) {
					isAlreadySelected = true
				} else {
					updatedList = append(updatedList, s)
				}
			}

			if !isAlreadySelected {
				updatedList = append(updatedList, matched)
			}

			c.Set("selected_products", strings.Join(updatedList, "|||"))

			var statusText string
			if isAlreadySelected {
				statusText = fmt.Sprintf("➖ Dihapus dari pilihan: *%s*", matched)
			} else {
				statusText = fmt.Sprintf("➕ Dipilih: *%s*", matched)
			}

			var selSummary []string
			for _, s := range updatedList {
				selSummary = append(selSummary, fmt.Sprintf("• %s", s))
			}
			if len(selSummary) == 0 {
				selSummary = append(selSummary, "_Belum ada produk yang dipilih_")
			}

			kb := ui.BuildMultiSelectKeyboard(availableProductNames, updatedList, ui.ButtonSubmit)
			prompt := fmt.Sprintf(
				"%s\n\n📦 *Produk Terpilih (%d):*\n%s\n\n_Klik produk lain untuk memilih/membatalkan, atau klik \"%s\" jika sudah selesai._",
				statusText,
				len(updatedList),
				strings.Join(selSummary, "\n"),
				ui.ButtonSubmit,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, kb)
			return c.Stay(), nil
		}).
		Build()
}
