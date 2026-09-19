package flow

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/repository"
	"strings"
)

const (
	TambahClientFlowID = "tambah_client"
	EditClientFlowID   = "edit_client"
)

// NewTambahClientFlow creates the multi-step flow to register a new client.
func NewTambahClientFlow(clientRepo repository.ClientRepository, productRepo repository.ProductRepository) Flow {
	return NewBuilder(TambahClientFlowID).
		// Step 1: Input Company Name
		InitialStep("input_company", func(c *Context) (Action, error) {
			_ = c.ReplyWithCancelKeyboard("🏢 *Tambah Client Baru (1/5)*\n\nSilakan masukkan *Nama Perusahaan / Organisasi*:")
			return c.Next("input_pic"), nil
		}).

		// Step 2: Input PIC Name
		Step("input_pic", func(c *Context) (Action, error) {
			companyName := strings.TrimSpace(c.RawText)
			if companyName == "" {
				_ = c.ReplyWithCancelKeyboard("❌ Nama Perusahaan tidak boleh kosong. Silakan masukkan nama perusahaan:")
				return c.Stay(), nil
			}
			c.Set("company_name", companyName)

			_ = c.ReplyWithCancelKeyboard("👤 *Tambah Client Baru (2/5)*\n\nSilakan masukkan *Nama PIC* (Person in Charge):")
			return c.Next("input_address"), nil
		}).

		// Step 3: Input Address
		Step("input_address", func(c *Context) (Action, error) {
			picName := strings.TrimSpace(c.RawText)
			if picName == "" {
				_ = c.ReplyWithCancelKeyboard("❌ Nama PIC tidak boleh kosong. Silakan masukkan nama PIC:")
				return c.Stay(), nil
			}
			c.Set("pic_name", picName)

			_ = c.ReplyWithCancelKeyboard("📍 *Tambah Client Baru (3/5)*\n\nSilakan masukkan *Alamat Lengkap Perusahaan*:")
			return c.Next("input_email"), nil
		}).

		// Step 4: Input Email
		Step("input_email", func(c *Context) (Action, error) {
			address := strings.TrimSpace(c.RawText)
			if address == "" {
				_ = c.ReplyWithCancelKeyboard("❌ Alamat tidak boleh kosong. Silakan masukkan alamat:")
				return c.Stay(), nil
			}
			c.Set("company_address", address)

			_ = c.ReplyWithCancelKeyboard("✉️ *Tambah Client Baru (4/5)*\n\nSilakan masukkan *Alamat Email Perusahaan* (untuk pengiriman invoice):")
			return c.Next("pilih_produk"), nil
		}).

		// Step 5: Multi-select Products
		Step("pilih_produk", func(c *Context) (Action, error) {
			ctx := context.Background()
			email := strings.TrimSpace(c.RawText)
			if email == "" || !strings.Contains(email, "@") {
				_ = c.ReplyWithCancelKeyboard("❌ Format email tidak valid. Silakan masukkan alamat email yang benar:")
				return c.Stay(), nil
			}
			c.Set("company_email", email)

			products, err := productRepo.List(ctx)
			if err != nil || len(products) == 0 {
				_ = c.ReplyAndRemoveKeyboard("❌ Belum ada produk di database. Silakan tambahkan produk terlebih dahulu.")
				return c.Complete(), nil
			}

			var availableNames []string
			for _, p := range products {
				availableNames = append(availableNames, p.Name)
			}
			c.Set("available_products", strings.Join(availableNames, "|||"))
			c.Set("selected_products", "")

			kb := ui.BuildMultiSelectKeyboard(availableNames, nil, ui.ButtonConfirm)
			prompt := fmt.Sprintf(
				"📦 *Tambah Client Baru (5/5) - Pilih Produk Langganan:*\n\n"+
					"_Klik nama produk untuk memilih/membatalkan, lalu klik \"%s\" jika sudah selesai._",
				ui.ButtonConfirm,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, kb)
			return c.Next("proses_pilih_produk"), nil
		}).

		// Step 5 Handler: Process product toggles or finish
		Step("proses_pilih_produk", func(c *Context) (Action, error) {
			rawInput := strings.TrimSpace(c.RawText)

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

			if ui.IsConfirmMessage(rawInput) {
				companyName, _ := c.GetString("company_name")
				picName, _ := c.GetString("pic_name")
				address, _ := c.GetString("company_address")
				email, _ := c.GetString("company_email")

				var prodStr string
				if len(selectedList) > 0 {
					prodStr = strings.Join(selectedList, ", ")
				} else {
					prodStr = "(Tidak ada produk)"
				}

				summary := fmt.Sprintf(
					"📋 *Konfirmasi Data Client Baru:*\n\n"+
						"🏢 Perusahaan: *%s*\n"+
						"👤 PIC: *%s*\n"+
						"📍 Alamat: *%s*\n"+
						"✉️ Email: *%s*\n"+
						"📦 Produk Dipesan: *%s*\n\n"+
						"Apakah data di atas sudah benar? Klik \"%s\" untuk menyimpan.",
					companyName, picName, address, email, prodStr, ui.ButtonConfirm,
				)

				kb := ui.BuildStepKeyboardFlat(ui.ButtonConfirm)
				_ = c.ReplyWithReplyKeyboard(summary, kb)
				return c.Next("simpan_client"), nil
			}

			// Toggle product
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
				kb := ui.BuildMultiSelectKeyboard(availableProductNames, selectedList, ui.ButtonConfirm)
				_ = c.ReplyWithReplyKeyboard("❌ Pilihan produk tidak dikenali. Silakan klik tombol produk:", kb)
				return c.Stay(), nil
			}

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

			var selSummary []string
			for _, s := range updatedList {
				selSummary = append(selSummary, fmt.Sprintf("• %s", s))
			}
			if len(selSummary) == 0 {
				selSummary = append(selSummary, "_Belum ada produk dipilih_")
			}

			kb := ui.BuildMultiSelectKeyboard(availableProductNames, updatedList, ui.ButtonConfirm)
			prompt := fmt.Sprintf(
				"📦 *Produk Terpilih (%d):*\n%s\n\n_Klik produk lain untuk memilih/membatalkan, atau klik \"%s\" jika sudah selesai._",
				len(updatedList),
				strings.Join(selSummary, "\n"),
				ui.ButtonConfirm,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, kb)
			return c.Stay(), nil
		}).

		// Step 6: Save to DB
		Step("simpan_client", func(c *Context) (Action, error) {
			if !ui.IsConfirmMessage(c.RawText) {
				_ = c.ReplyWithReplyKeyboard("Silakan klik '"+ui.ButtonConfirm+"' untuk menyimpan, atau '"+ui.DefaultCancelButton+"' untuk membatalkan.", ui.BuildStepKeyboardFlat(ui.ButtonConfirm))
				return c.Stay(), nil
			}

			companyName, _ := c.GetString("company_name")
			picName, _ := c.GetString("pic_name")
			address, _ := c.GetString("company_address")
			email, _ := c.GetString("company_email")
			selectedRaw, _ := c.GetString("selected_products")

			var prodOrdered string
			if selectedRaw != "" {
				prodOrdered = strings.Join(strings.Split(selectedRaw, "|||"), ", ")
			}

			newClient := &domain.Client{
				CompanyName:    companyName,
				PICName:        picName,
				CompanyAddress: address,
				CompanyEmail:   email,
				ProductOrdered: prodOrdered,
			}

			ctx := context.Background()
			if err := clientRepo.Create(ctx, newClient); err != nil {
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ Gagal menyimpan client ke database: %v", err), ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			successMsg := fmt.Sprintf(
				"✅ *Client Berhasil Ditambahkan!*\n\n"+
					"🏢 ID: `%d`\n"+
					"🏢 Perusahaan: *%s*\n"+
					"👤 PIC: *%s*\n"+
					"📍 Alamat: *%s*\n"+
					"✉️ Email: *%s*\n"+
					"📦 Produk: *%s*",
				newClient.ID, newClient.CompanyName, newClient.PICName, newClient.CompanyAddress, newClient.CompanyEmail, newClient.ProductOrdered,
			)
			_ = c.ReplyWithReplyKeyboard(successMsg, ui.MainMenuKeyboard())
			return c.Complete(), nil
		}).
		Build()
}

// NewEditClientFlow creates the multi-step flow to edit an existing client with Skip and Add/Remove product options.
func NewEditClientFlow(clientRepo repository.ClientRepository, productRepo repository.ProductRepository) Flow {
	return NewBuilder(EditClientFlowID).
		// Step 1: Select Client
		InitialStep("pilih_client", func(c *Context) (Action, error) {
			ctx := context.Background()
			clients, err := clientRepo.List(ctx)
			if err != nil || len(clients) == 0 {
				_ = c.ReplyAndRemoveKeyboard("❌ Belum ada client yang terdaftar untuk diedit.")
				return c.Complete(), nil
			}

			var clientButtons []string
			for _, client := range clients {
				clientButtons = append(clientButtons, client.CompanyName)
			}

			_ = c.ReplyWithStepKeyboard("✏️ *Edit Client - Pilih Client:*\n\nSilakan pilih client yang ingin diubah:", clientButtons...)
			return c.Next("proses_pilih_client"), nil
		}).

		// Step 2: Edit Company Name
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
				_ = c.ReplyWithStepKeyboard("❌ Client tidak ditemukan. Silakan pilih dari tombol:", clientButtons...)
				return c.Stay(), nil
			}

			c.Set("client_id", selectedClient.ID)
			c.Set("company_name", selectedClient.CompanyName)
			c.Set("pic_name", selectedClient.PICName)
			c.Set("company_address", selectedClient.CompanyAddress)
			c.Set("company_email", selectedClient.CompanyEmail)
			c.Set("product_ordered", selectedClient.ProductOrdered)

			prompt := fmt.Sprintf(
				"🏢 *Edit Client: %s (1/4)*\n\n"+
					"Nama Perusahaan saat ini: *%s*\n\n"+
					"Ketik nama perusahaan baru, atau klik \"%s\" untuk tidak mengubah:",
				selectedClient.CompanyName,
				selectedClient.CompanyName,
				ui.ButtonSkip,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, ui.BuildStepKeyboardWithSkip())
			return c.Next("edit_pic"), nil
		}).

		// Step 3: Edit PIC Name
		Step("edit_pic", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)
			if !ui.IsSkipMessage(raw) && raw != "" {
				c.Set("company_name", raw)
			}

			currentPIC, _ := c.GetString("pic_name")
			prompt := fmt.Sprintf(
				"👤 *Edit Client (2/4)*\n\n"+
					"Nama PIC saat ini: *%s*\n\n"+
					"Ketik nama PIC baru, atau klik \"%s\" untuk tidak mengubah:",
				currentPIC,
				ui.ButtonSkip,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, ui.BuildStepKeyboardWithSkip())
			return c.Next("edit_address"), nil
		}).

		// Step 4: Edit Address
		Step("edit_address", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)
			if !ui.IsSkipMessage(raw) && raw != "" {
				c.Set("pic_name", raw)
			}

			currentAddress, _ := c.GetString("company_address")
			prompt := fmt.Sprintf(
				"📍 *Edit Client (3/4)*\n\n"+
					"Alamat saat ini:\n*%s*\n\n"+
					"Ketik alamat baru, atau klik \"%s\" untuk tidak mengubah:",
				currentAddress,
				ui.ButtonSkip,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, ui.BuildStepKeyboardWithSkip())
			return c.Next("edit_email"), nil
		}).

		// Step 5: Edit Email
		Step("edit_email", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)
			if !ui.IsSkipMessage(raw) && raw != "" {
				c.Set("company_address", raw)
			}

			currentEmail, _ := c.GetString("company_email")
			prompt := fmt.Sprintf(
				"✉️ *Edit Client (4/4)*\n\n"+
					"Email saat ini: *%s*\n\n"+
					"Ketik email baru, atau klik \"%s\" untuk tidak mengubah:",
				currentEmail,
				ui.ButtonSkip,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, ui.BuildStepKeyboardWithSkip())
			return c.Next("menu_produk"), nil
		}).

		// Step 6: Product Management Menu (Tambah / Hapus / Skip)
		Step("menu_produk", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)
			if !ui.IsSkipMessage(raw) && raw != "" {
				if !strings.Contains(raw, "@") {
					_ = c.ReplyWithReplyKeyboard("❌ Format email tidak valid. Masukkan email yang benar atau klik '"+ui.ButtonSkip+"':", ui.BuildStepKeyboardWithSkip())
					return c.Stay(), nil
				}
				c.Set("company_email", raw)
			}

			prodOrdered, _ := c.GetString("product_ordered")
			tempClient := &domain.Client{ProductOrdered: prodOrdered}
			currentProducts := tempClient.GetProductList()

			var prodDisplay string
			if len(currentProducts) > 0 {
				prodDisplay = strings.Join(currentProducts, "\n• ")
				prodDisplay = "• " + prodDisplay
			} else {
				prodDisplay = "_(Tidak ada produk berlangganan)_"
			}

			prompt := fmt.Sprintf(
				"📦 *Kelola Produk Client:*\n\n"+
					"Produk berlangganan saat ini:\n%s\n\n"+
					"Pilih opsi di bawah untuk menambah atau menghapus produk, atau klik \"%s\" jika sudah selesai:",
				prodDisplay,
				ui.ButtonSkip,
			)

			kb := ui.BuildStepKeyboard([][]string{
				{ui.ButtonOptionTambah, ui.ButtonOptionHapus},
				{ui.ButtonSkip},
			})
			_ = c.ReplyWithReplyKeyboard(prompt, kb)
			return c.Next("proses_menu_produk"), nil
		}).

		// Step 6.1: Handle Product Menu Options
		Step("proses_menu_produk", func(c *Context) (Action, error) {
			ctx := context.Background()
			raw := strings.TrimSpace(c.RawText)

			if ui.IsSkipMessage(raw) || ui.IsConfirmMessage(raw) {
				// Proceed to Final Confirmation
				return showClientFinalSummary(c)
			}

			prodOrdered, _ := c.GetString("product_ordered")
			tempClient := &domain.Client{ProductOrdered: prodOrdered}
			currentProducts := tempClient.GetProductList()

			if strings.EqualFold(raw, ui.ButtonOptionHapus) {
				// Show list of client's current products to remove
				if len(currentProducts) == 0 {
					_ = c.ReplyWithReplyKeyboard("❌ Client belum memiliki produk untuk dihapus. Silakan pilih opsi lain:", ui.BuildStepKeyboard([][]string{
						{ui.ButtonOptionTambah},
						{ui.ButtonSkip},
					}))
					return c.Stay(), nil
				}

				c.Set("subflow_action", "hapus")
				c.Set("staged_products", prodOrdered)

				kb := ui.BuildStepKeyboard([][]string{
					currentProducts,
					{ui.ButtonConfirm, ui.ButtonSkip},
				})
				prompt := "➖ *Hapus Produk:*\n\nKlik produk yang ingin dihapus, lalu klik \"" + ui.ButtonConfirm + "\":"
				_ = c.ReplyWithReplyKeyboard(prompt, kb)
				return c.Next("proses_hapus_produk"), nil
			}

			if strings.EqualFold(raw, ui.ButtonOptionTambah) {
				// Show list of products from DB not currently subscribed
				allProducts, err := productRepo.List(ctx)
				if err != nil || len(allProducts) == 0 {
					_ = c.ReplyWithReplyKeyboard("❌ Belum ada produk di database.", ui.BuildStepKeyboard([][]string{{ui.ButtonSkip}}))
					return c.Stay(), nil
				}

				var unsubscribed []string
				for _, p := range allProducts {
					if !tempClient.HasProduct(p.Name) {
						unsubscribed = append(unsubscribed, p.Name)
					}
				}

				if len(unsubscribed) == 0 {
					_ = c.ReplyWithReplyKeyboard("ℹ️ Client sudah berlangganan semua produk yang tersedia.", ui.BuildStepKeyboard([][]string{{ui.ButtonSkip}}))
					return c.Stay(), nil
				}

				c.Set("subflow_action", "tambah")
				c.Set("unsubscribed_available", strings.Join(unsubscribed, "|||"))
				c.Set("tambah_selected", "")

				kb := ui.BuildMultiSelectKeyboard(unsubscribed, nil, ui.ButtonConfirm)
				prompt := "➕ *Tambah Produk:*\n\nKlik produk yang ingin ditambahkan, lalu klik \"" + ui.ButtonConfirm + "\":"
				_ = c.ReplyWithReplyKeyboard(prompt, kb)
				return c.Next("proses_tambah_produk"), nil
			}

			_ = c.ReplyWithReplyKeyboard("Pilihan tidak dikenali. Silakan pilih tombol opsi:", ui.BuildStepKeyboard([][]string{
				{ui.ButtonOptionTambah, ui.ButtonOptionHapus},
				{ui.ButtonSkip},
			}))
			return c.Stay(), nil
		}).

		// Step 6.2: Subflow Hapus Produk
		Step("proses_hapus_produk", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)

			if ui.IsConfirmMessage(raw) || ui.IsSkipMessage(raw) {
				return showClientFinalSummary(c)
			}

			prodOrdered, _ := c.GetString("product_ordered")
			tempClient := &domain.Client{ProductOrdered: prodOrdered}
			if tempClient.HasProduct(raw) {
				tempClient.RemoveProduct(raw)
				c.Set("product_ordered", tempClient.ProductOrdered)

				currentList := tempClient.GetProductList()
				if len(currentList) == 0 {
					_ = c.Reply("Produk berhasil dihapus. Semua produk telah dihapus dari client.")
					return showClientFinalSummary(c)
				}

				kb := ui.BuildStepKeyboard([][]string{
					currentList,
					{ui.ButtonConfirm, ui.ButtonSkip},
				})
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("🗑 Dihapus: *%s*\n\nKlik produk lain untuk dihapus atau klik '%s':", raw, ui.ButtonConfirm), kb)
				return c.Stay(), nil
			}

			_ = c.Reply("Produk tidak ditemukan dalam daftar langganan.")
			return c.Stay(), nil
		}).

		// Step 6.3: Subflow Tambah Produk
		Step("proses_tambah_produk", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)

			unsubRaw, _ := c.GetString("unsubscribed_available")
			var unsubList []string
			if unsubRaw != "" {
				unsubList = strings.Split(unsubRaw, "|||")
			}

			selRaw, _ := c.GetString("tambah_selected")
			var selectedList []string
			if selRaw != "" {
				selectedList = strings.Split(selRaw, "|||")
			}

			if ui.IsConfirmMessage(raw) {
				// Apply added products to product_ordered
				prodOrdered, _ := c.GetString("product_ordered")
				tempClient := &domain.Client{ProductOrdered: prodOrdered}
				for _, p := range selectedList {
					tempClient.AddProduct(p)
				}
				c.Set("product_ordered", tempClient.ProductOrdered)
				return showClientFinalSummary(c)
			}

			if ui.IsSkipMessage(raw) {
				return showClientFinalSummary(c)
			}

			// Toggle selection
			clean := strings.TrimPrefix(raw, "✅ ")
			clean = strings.TrimSpace(clean)

			var matched string
			for _, u := range unsubList {
				if strings.EqualFold(u, clean) {
					matched = u
					break
				}
			}

			if matched == "" {
				kb := ui.BuildMultiSelectKeyboard(unsubList, selectedList, ui.ButtonConfirm)
				_ = c.ReplyWithReplyKeyboard("Pilihan produk tidak dikenali.", kb)
				return c.Stay(), nil
			}

			isSel := false
			var updated []string
			for _, s := range selectedList {
				if strings.EqualFold(s, matched) {
					isSel = true
				} else {
					updated = append(updated, s)
				}
			}
			if !isSel {
				updated = append(updated, matched)
			}
			c.Set("tambah_selected", strings.Join(updated, "|||"))

			kb := ui.BuildMultiSelectKeyboard(unsubList, updated, ui.ButtonConfirm)
			_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("Pilihan diperbarui (%d produk dipilih). Klik '%s' jika sudah selesai.", len(updated), ui.ButtonConfirm), kb)
			return c.Stay(), nil
		}).

		// Step 7: Final Confirmation and Save to DB
		Step("konfirmasi_update_client", func(c *Context) (Action, error) {
			if !ui.IsConfirmMessage(c.RawText) {
				_ = c.ReplyWithReplyKeyboard("Klik '"+ui.ButtonConfirm+"' untuk menyimpan perubahan atau '"+ui.DefaultCancelButton+"' untuk membatalkan.", ui.BuildStepKeyboardFlat(ui.ButtonConfirm))
				return c.Stay(), nil
			}

			clientIDInt, _ := c.GetInt("client_id")
			companyName, _ := c.GetString("company_name")
			picName, _ := c.GetString("pic_name")
			address, _ := c.GetString("company_address")
			email, _ := c.GetString("company_email")
			prodOrdered, _ := c.GetString("product_ordered")

			updatedClient := &domain.Client{
				ID:             int64(clientIDInt),
				CompanyName:    companyName,
				PICName:        picName,
				CompanyAddress: address,
				CompanyEmail:   email,
				ProductOrdered: prodOrdered,
			}

			ctx := context.Background()
			if err := clientRepo.Update(ctx, updatedClient); err != nil {
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ Gagal mengupdate data client: %v", err), ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			successMsg := fmt.Sprintf(
				"✅ *Data Client Berhasil Diperbarui!*\n\n"+
					"🏢 ID: `%d`\n"+
					"🏢 Perusahaan: *%s*\n"+
					"👤 PIC: *%s*\n"+
					"📍 Alamat: *%s*\n"+
					"✉️ Email: *%s*\n"+
					"📦 Produk: *%s*",
				updatedClient.ID, updatedClient.CompanyName, updatedClient.PICName, updatedClient.CompanyAddress, updatedClient.CompanyEmail, updatedClient.ProductOrdered,
			)
			_ = c.ReplyWithReplyKeyboard(successMsg, ui.MainMenuKeyboard())
			return c.Complete(), nil
		}).
		Build()
}

func showClientFinalSummary(c *Context) (Action, error) {
	companyName, _ := c.GetString("company_name")
	picName, _ := c.GetString("pic_name")
	address, _ := c.GetString("company_address")
	email, _ := c.GetString("company_email")
	prodOrdered, _ := c.GetString("product_ordered")

	if strings.TrimSpace(prodOrdered) == "" {
		prodOrdered = "(Tidak ada)"
	}

	summary := fmt.Sprintf(
		"📋 *Ringkasan Perubahan Client:*\n\n"+
			"🏢 Perusahaan: *%s*\n"+
			"👤 PIC: *%s*\n"+
			"📍 Alamat: *%s*\n"+
			"✉️ Email: *%s*\n"+
			"📦 Produk: *%s*\n\n"+
			"Klik \"%s\" untuk menyimpan perubahan ke database.",
		companyName, picName, address, email, prodOrdered, ui.ButtonConfirm,
	)

	kb := ui.BuildStepKeyboardFlat(ui.ButtonConfirm)
	_ = c.ReplyWithReplyKeyboard(summary, kb)
	return c.Next("konfirmasi_update_client"), nil
}
