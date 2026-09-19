package ui

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Standard button texts
const (
	DefaultCancelButton = "❌ Batal"
	ButtonKirimInvoice  = "📄 Kirim Invoice"
	ButtonTambahClient  = "🏢 Tambah Client"
	ButtonEditClient    = "✏️ Edit Client"
	ButtonTambahProduk  = "📦 Tambah Produk"
	ButtonEditProduk    = "⚙️ Edit Produk"
	ButtonHideMenu      = "🔽 Tutup Menu"
	ButtonSkip          = "⏩ Skip"
	ButtonConfirm       = "✅ Konfirmasi"
	ButtonSubmit        = "✅ Submit"
	ButtonSelesai       = "✅ Selesai"
	ButtonOptionTambah  = "➕ Tambah Produk"
	ButtonOptionHapus   = "➖ Hapus Produk"
)

var defaultMainMenuRows = [][]string{
	{ButtonKirimInvoice},
	{ButtonTambahClient, ButtonEditClient},
	{ButtonTambahProduk, ButtonEditProduk},
	{ButtonHideMenu},
}

// SetMainMenuRows configures the default main menu buttons.
func SetMainMenuRows(rows [][]string) {
	defaultMainMenuRows = rows
}

// MainMenuKeyboard returns the ReplyKeyboardMarkup for the main menu (/start or /menu).
func MainMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	var keyboardRows [][]tgbotapi.KeyboardButton
	for _, row := range defaultMainMenuRows {
		var buttons []tgbotapi.KeyboardButton
		for _, btn := range row {
			btnTrimmed := strings.TrimSpace(btn)
			if btnTrimmed != "" {
				buttons = append(buttons, tgbotapi.NewKeyboardButton(btnTrimmed))
			}
		}
		if len(buttons) > 0 {
			keyboardRows = append(keyboardRows, buttons)
		}
	}
	keyboard := tgbotapi.NewReplyKeyboard(keyboardRows...)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false
	return keyboard
}

// IsCancelMessage checks if the given message text matches any cancellation phrase.
func IsCancelMessage(text string) bool {
	clean := strings.ToLower(strings.TrimSpace(text))
	return clean == "❌ batal" || clean == "batal" || clean == "cancel"
}

// IsSkipMessage checks if the text matches the Skip button.
func IsSkipMessage(text string) bool {
	clean := strings.ToLower(strings.TrimSpace(text))
	return clean == "⏩ skip" || clean == "skip" || clean == "lewati"
}

// IsConfirmMessage checks if the text matches a confirmation button.
func IsConfirmMessage(text string) bool {
	clean := strings.ToLower(strings.TrimSpace(text))
	return clean == "✅ konfirmasi" || clean == "konfirmasi" || clean == "confirm" || clean == "✅ selesai" || clean == "selesai" || clean == "✅ submit" || clean == "submit"
}

// IsSubmitMessage checks if the text matches submit/confirm.
func IsSubmitMessage(text string) bool {
	return IsConfirmMessage(text)
}

// IsHideMenuMessage checks if the text matches closing the menu.
func IsHideMenuMessage(text string) bool {
	clean := strings.ToLower(strings.TrimSpace(text))
	return clean == "🔽 tutup menu" || clean == "tutup menu" || clean == "close menu" || clean == "hide menu"
}

// BuildStepKeyboard creates a ReplyKeyboardMarkup with the provided step buttons
// and automatically appends the cancel button at the bottom row.
func BuildStepKeyboard(rows [][]string) tgbotapi.ReplyKeyboardMarkup {
	var keyboardRows [][]tgbotapi.KeyboardButton

	for _, row := range rows {
		var buttons []tgbotapi.KeyboardButton
		for _, btn := range row {
			btnTrimmed := strings.TrimSpace(btn)
			if btnTrimmed != "" {
				buttons = append(buttons, tgbotapi.NewKeyboardButton(btnTrimmed))
			}
		}
		if len(buttons) > 0 {
			keyboardRows = append(keyboardRows, buttons)
		}
	}

	// Always append the Cancel button row at the bottom
	keyboardRows = append(keyboardRows, tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(DefaultCancelButton),
	))

	keyboard := tgbotapi.NewReplyKeyboard(keyboardRows...)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false
	return keyboard
}

// BuildStepKeyboardWithSkip creates a keyboard with items, an optional Skip button, and Cancel.
func BuildStepKeyboardWithSkip(buttons ...string) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]string
	for _, b := range buttons {
		if strings.TrimSpace(b) != "" {
			rows = append(rows, []string{b})
		}
	}
	rows = append(rows, []string{ButtonSkip})
	return BuildStepKeyboard(rows)
}

// BuildStepKeyboardFlat creates a ReplyKeyboardMarkup from a list of buttons in rows
// plus the cancel button at the bottom.
func BuildStepKeyboardFlat(buttons ...string) tgbotapi.ReplyKeyboardMarkup {
	if len(buttons) == 0 {
		return CancelOnlyKeyboard()
	}
	var rows [][]string
	for _, b := range buttons {
		if strings.TrimSpace(b) != "" {
			rows = append(rows, []string{b})
		}
	}
	return BuildStepKeyboard(rows)
}

// BuildMultiSelectKeyboard creates a keyboard for selecting items, showing checked items and a finish button.
func BuildMultiSelectKeyboard(availableItems []string, selectedItems []string, finishButton string) tgbotapi.ReplyKeyboardMarkup {
	if finishButton == "" {
		finishButton = ButtonConfirm
	}

	selectedMap := make(map[string]bool)
	for _, s := range selectedItems {
		selectedMap[strings.ToLower(strings.TrimSpace(s))] = true
	}

	var rows [][]string
	for _, item := range availableItems {
		label := item
		if selectedMap[strings.ToLower(strings.TrimSpace(item))] {
			label = fmt.Sprintf("✅ %s", item)
		}
		rows = append(rows, []string{label})
	}

	// Action row: Finish button
	rows = append(rows, []string{finishButton})

	return BuildStepKeyboard(rows)
}

// CancelOnlyKeyboard creates a ReplyKeyboardMarkup containing only the cancel button.
func CancelOnlyKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(DefaultCancelButton),
		),
	)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false
	return keyboard
}

// RemoveKeyboard returns a ReplyKeyboardRemove markup to hide custom keyboards.
func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}
