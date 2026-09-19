package flow

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/repository"
	"hs1-bot/internal/service/template"
	"strings"
)

const (
	TambahProdukFlowID = "tambah_produk"
	EditProdukFlowID   = "edit_produk"
	HapusProdukFlowID  = "hapus_produk"
)

// NewTambahProdukFlow creates the multi-step flow to add a new product with price parsing.
func NewTambahProdukFlow(productRepo repository.ProductRepository) Flow {
	return NewBuilder(TambahProdukFlowID).
		// Step 1: Input Product Name
		InitialStep("input_nama_produk", func(c *Context) (Action, error) {
			_ = c.ReplyWithCancelKeyboard("📦 *Tambah Produk Baru (1/2)*\n\nSilakan masukkan *Nama Produk / Layanan*:")
			return c.Next("input_harga_produk"), nil
		}).

		// Step 2: Input Product Price
		Step("input_harga_produk", func(c *Context) (Action, error) {
			name := strings.TrimSpace(c.RawText)
			if name == "" {
				_ = c.ReplyWithCancelKeyboard("❌ Nama produk tidak boleh kosong. Silakan masukkan nama produk:")
				return c.Stay(), nil
			}
			c.Set("product_name", name)

			prompt := fmt.Sprintf(
				"📦 *Tambah Produk Baru (2/2)*\n\n"+
					"Produk: *%s*\n\n"+
					"Silakan masukkan *Harga Produk*:\n"+
					"_(Mendukung format angka, misal `400000`, `400k`, `1.5M`, `Rp 400.000`)_",
				name,
			)
			_ = c.ReplyWithCancelKeyboard(prompt)
			return c.Next("konfirmasi_tambah_produk"), nil
		}).

		// Step 3: Parse price, show summary and confirm
		Step("konfirmasi_tambah_produk", func(c *Context) (Action, error) {
			rawPrice := strings.TrimSpace(c.RawText)
			price, err := domain.ParsePrice(rawPrice)
			if err != nil {
				_ = c.ReplyWithCancelKeyboard(fmt.Sprintf("❌ %v\n\nSilakan masukkan nominal harga yang valid (contoh `400000` atau `400k`):", err))
				return c.Stay(), nil
			}

			name, _ := c.GetString("product_name")
			c.Set("product_price", fmt.Sprintf("%.0f", price))

			priceShort := template.FormatCurrencyShort(price)
			summary := fmt.Sprintf(
				"📋 *Konfirmasi Tambah Produk:*\n\n"+
					"📦 Nama Produk: *%s*\n"+
					"💵 Harga: *IDR %s* (Rp %.0f)\n\n"+
					"Klik \"%s\" untuk menyimpan ke database.",
				name, priceShort, price, ui.ButtonConfirm,
			)

			kb := ui.BuildStepKeyboardFlat(ui.ButtonConfirm)
			_ = c.ReplyWithReplyKeyboard(summary, kb)
			return c.Next("simpan_produk"), nil
		}).

		// Step 4: Save to Database
		Step("simpan_produk", func(c *Context) (Action, error) {
			if !ui.IsConfirmMessage(c.RawText) {
				_ = c.ReplyWithReplyKeyboard("Klik '"+ui.ButtonConfirm+"' untuk menyimpan atau '"+ui.DefaultCancelButton+"' untuk membatalkan.", ui.BuildStepKeyboardFlat(ui.ButtonConfirm))
				return c.Stay(), nil
			}

			name, _ := c.GetString("product_name")
			priceStr, _ := c.GetString("product_price")
			price, _ := domain.ParsePrice(priceStr)

			newProduct := &domain.Product{
				Name:  name,
				Price: price,
			}

			ctx := context.Background()
			if err := productRepo.Create(ctx, newProduct); err != nil {
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ Gagal menyimpan produk: %v", err), ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			successMsg := fmt.Sprintf(
				"✅ *Produk Berhasil Ditambahkan!*\n\n"+
					"📦 ID: `%d`\n"+
					"📦 Nama: *%s*\n"+
					"💵 Harga: *IDR %s*",
				newProduct.ID, newProduct.Name, template.FormatCurrencyShort(newProduct.Price),
			)
			_ = c.ReplyWithReplyKeyboard(successMsg, ui.MainMenuKeyboard())
			return c.Complete(), nil
		}).
		Build()
}

// NewEditProdukFlow creates the multi-step flow to edit an existing product.
func NewEditProdukFlow(productRepo repository.ProductRepository) Flow {
	return NewBuilder(EditProdukFlowID).
		// Step 1: Select Product
		InitialStep("pilih_produk", func(c *Context) (Action, error) {
			ctx := context.Background()
			products, err := productRepo.List(ctx)
			if err != nil || len(products) == 0 {
				_ = c.ReplyAndRemoveKeyboard("❌ Belum ada produk yang terdaftar untuk diedit.")
				return c.Complete(), nil
			}

			var productButtons []string
			for _, p := range products {
				productButtons = append(productButtons, p.Name)
			}

			_ = c.ReplyWithStepKeyboard("⚙️ *Edit Produk - Pilih Produk:*\n\nSilakan pilih produk yang ingin diubah:", productButtons...)
			return c.Next("proses_pilih_produk"), nil
		}).

		// Step 2: Edit Product Name
		Step("proses_pilih_produk", func(c *Context) (Action, error) {
			ctx := context.Background()
			products, err := productRepo.List(ctx)
			if err != nil {
				_ = c.ReplyAndRemoveKeyboard(fmt.Sprintf("❌ Gagal membaca produk: %v", err))
				return c.Complete(), nil
			}

			selectedText := strings.TrimSpace(c.RawText)
			var selectedProduct *domain.Product

			for i := range products {
				if strings.EqualFold(strings.TrimSpace(products[i].Name), selectedText) {
					selectedProduct = &products[i]
					break
				}
			}

			if selectedProduct == nil {
				var productButtons []string
				for _, p := range products {
					productButtons = append(productButtons, p.Name)
				}
				_ = c.ReplyWithStepKeyboard("❌ Produk tidak ditemukan. Silakan pilih dari tombol:", productButtons...)
				return c.Stay(), nil
			}

			c.Set("product_id", selectedProduct.ID)
			c.Set("product_name", selectedProduct.Name)
			c.Set("product_price", fmt.Sprintf("%.0f", selectedProduct.Price))

			prompt := fmt.Sprintf(
				"⚙️ *Edit Produk: %s (1/2)*\n\n"+
					"Nama saat ini: *%s*\n\n"+
					"Ketik nama produk baru, atau klik \"%s\" untuk tidak mengubah:",
				selectedProduct.Name,
				selectedProduct.Name,
				ui.ButtonSkip,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, ui.BuildStepKeyboardWithSkip())
			return c.Next("edit_harga"), nil
		}).

		// Step 3: Edit Product Price
		Step("edit_harga", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)
			if !ui.IsSkipMessage(raw) && raw != "" {
				c.Set("product_name", raw)
			}

			priceStr, _ := c.GetString("product_price")
			price, _ := domain.ParsePrice(priceStr)
			currentPriceShort := template.FormatCurrencyShort(price)

			prompt := fmt.Sprintf(
				"💵 *Edit Produk (2/2)*\n\n"+
					"Harga saat ini: *IDR %s* (Rp %.0f)\n\n"+
					"Ketik harga baru (contoh `500k` / `500000`), atau klik \"%s\" untuk tidak mengubah:",
				currentPriceShort,
				price,
				ui.ButtonSkip,
			)
			_ = c.ReplyWithReplyKeyboard(prompt, ui.BuildStepKeyboardWithSkip())
			return c.Next("konfirmasi_edit_produk"), nil
		}).

		// Step 4: Show summary & confirm
		Step("konfirmasi_edit_produk", func(c *Context) (Action, error) {
			raw := strings.TrimSpace(c.RawText)
			if !ui.IsSkipMessage(raw) && raw != "" {
				newPrice, err := domain.ParsePrice(raw)
				if err != nil {
					_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ %v\n\nMasukkan nominal harga yang valid atau klik '%s':", err, ui.ButtonSkip), ui.BuildStepKeyboardWithSkip())
					return c.Stay(), nil
				}
				c.Set("product_price", fmt.Sprintf("%.0f", newPrice))
			}

			name, _ := c.GetString("product_name")
			priceStr, _ := c.GetString("product_price")
			price, _ := domain.ParsePrice(priceStr)
			priceShort := template.FormatCurrencyShort(price)

			summary := fmt.Sprintf(
				"📋 *Ringkasan Perubahan Produk:*\n\n"+
					"📦 Nama Produk: *%s*\n"+
					"💵 Harga: *IDR %s* (Rp %.0f)\n\n"+
					"Klik \"%s\" untuk menyimpan perubahan.",
				name, priceShort, price, ui.ButtonConfirm,
			)

			kb := ui.BuildStepKeyboardFlat(ui.ButtonConfirm)
			_ = c.ReplyWithReplyKeyboard(summary, kb)
			return c.Next("simpan_edit_produk"), nil
		}).

		// Step 5: Save update to database
		Step("simpan_edit_produk", func(c *Context) (Action, error) {
			if !ui.IsConfirmMessage(c.RawText) {
				_ = c.ReplyWithReplyKeyboard("Klik '"+ui.ButtonConfirm+"' untuk menyimpan atau '"+ui.DefaultCancelButton+"' untuk membatalkan.", ui.BuildStepKeyboardFlat(ui.ButtonConfirm))
				return c.Stay(), nil
			}

			prodIDInt, _ := c.GetInt("product_id")
			name, _ := c.GetString("product_name")
			priceStr, _ := c.GetString("product_price")
			price, _ := domain.ParsePrice(priceStr)

			updatedProduct := &domain.Product{
				ID:    int64(prodIDInt),
				Name:  name,
				Price: price,
			}

			ctx := context.Background()
			if err := productRepo.Update(ctx, updatedProduct); err != nil {
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ Gagal mengupdate produk: %v", err), ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			successMsg := fmt.Sprintf(
				"✅ *Produk Berhasil Diperbarui!*\n\n"+
					"📦 ID: `%d`\n"+
					"📦 Nama: *%s*\n"+
					"💵 Harga: *IDR %s*",
				updatedProduct.ID, updatedProduct.Name, template.FormatCurrencyShort(updatedProduct.Price),
			)
			_ = c.ReplyWithReplyKeyboard(successMsg, ui.MainMenuKeyboard())
			return c.Complete(), nil
		}).
		Build()
}

// NewHapusProdukFlow creates the multi-step flow to soft-delete a product.
func NewHapusProdukFlow(productRepo repository.ProductRepository) Flow {
	return NewBuilder(HapusProdukFlowID).
		// Step 1: Select product to delete
		InitialStep("pilih_produk", func(c *Context) (Action, error) {
			ctx := context.Background()
			products, err := productRepo.List(ctx)
			if err != nil || len(products) == 0 {
				_ = c.ReplyAndRemoveKeyboard("❌ Belum ada produk yang terdaftar untuk dihapus.")
				return c.Complete(), nil
			}

			var productButtons []string
			for _, p := range products {
				productButtons = append(productButtons, p.Name)
			}

			_ = c.ReplyWithStepKeyboard("🗑️ *Hapus Produk - Pilih Produk:*\n\nSilakan pilih produk yang ingin dihapus:", productButtons...)
			return c.Next("proses_pilih_produk"), nil
		}).

		// Step 2: Validate selected product and show confirmation
		Step("proses_pilih_produk", func(c *Context) (Action, error) {
			ctx := context.Background()
			products, err := productRepo.List(ctx)
			if err != nil {
				_ = c.ReplyAndRemoveKeyboard(fmt.Sprintf("❌ Gagal membaca data produk: %v", err))
				return c.Complete(), nil
			}

			selectedText := strings.TrimSpace(c.RawText)
			var selectedProduct *domain.Product

			for i := range products {
				if strings.EqualFold(strings.TrimSpace(products[i].Name), selectedText) {
					selectedProduct = &products[i]
					break
				}
			}

			if selectedProduct == nil {
				var productButtons []string
				for _, p := range products {
					productButtons = append(productButtons, p.Name)
				}
				_ = c.ReplyWithStepKeyboard("❌ Produk tidak ditemukan. Silakan pilih dari tombol:", productButtons...)
				return c.Stay(), nil
			}

			c.Set("product_id", selectedProduct.ID)
			c.Set("product_name", selectedProduct.Name)
			c.Set("product_price", fmt.Sprintf("%.0f", selectedProduct.Price))

			priceShort := template.FormatCurrencyShort(selectedProduct.Price)
			prompt := fmt.Sprintf(
				"⚠️ *Konfirmasi Hapus Produk:*\n\n"+
					"Apakah Anda yakin ingin menghapus produk berikut?\n\n"+
					"📦 Nama Produk: *%s*\n"+
					"💵 Harga: *IDR %s* (Rp %.0f)\n\n"+
					"Klik \"%s\" untuk menghapus atau \"%s\" untuk membatalkan.",
				selectedProduct.Name,
				priceShort,
				selectedProduct.Price,
				ui.ButtonConfirm,
				ui.DefaultCancelButton,
			)

			kb := ui.BuildStepKeyboardFlat(ui.ButtonConfirm)
			_ = c.ReplyWithReplyKeyboard(prompt, kb)
			return c.Next("eksekusi_hapus_produk"), nil
		}).

		// Step 3: Execute soft delete on confirmation
		Step("eksekusi_hapus_produk", func(c *Context) (Action, error) {
			if !ui.IsConfirmMessage(c.RawText) {
				_ = c.ReplyWithReplyKeyboard("Klik '"+ui.ButtonConfirm+"' untuk menghapus atau '"+ui.DefaultCancelButton+"' untuk membatalkan.", ui.BuildStepKeyboardFlat(ui.ButtonConfirm))
				return c.Stay(), nil
			}

			prodIDInt, _ := c.GetInt("product_id")
			prodName, _ := c.GetString("product_name")

			ctx := context.Background()
			if err := productRepo.Delete(ctx, int64(prodIDInt)); err != nil {
				_ = c.ReplyWithReplyKeyboard(fmt.Sprintf("❌ Gagal menghapus produk: %v", err), ui.MainMenuKeyboard())
				return c.Complete(), nil
			}

			successMsg := fmt.Sprintf(
				"✅ *Produk Berhasil Dihapus (Soft Delete)!*\n\n"+
					"📦 Produk: *%s*\n"+
					"Produk telah dinonaktifkan dari sistem.",
				prodName,
			)
			_ = c.ReplyWithReplyKeyboard(successMsg, ui.MainMenuKeyboard())
			return c.Complete(), nil
		}).
		Build()
}
