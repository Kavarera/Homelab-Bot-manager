package ui

import (
	"testing"
)

func TestIsCancelMessage(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"❌ Batal", true},
		{"❌ BATAL", true},
		{"batal", true},
		{"Batal", true},
		{"cancel", true},
		{"cancel please", false},
		{"halo", false},
		{"", false},
	}

	for _, tc := range testCases {
		res := IsCancelMessage(tc.input)
		if res != tc.expected {
			t.Errorf("IsCancelMessage(%q) = %v, expected %v", tc.input, res, tc.expected)
		}
	}
}

func TestIsSkipMessage(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"⏩ Skip", true},
		{"skip", true},
		{"SKIP", true},
		{"lewati", true},
		{"random text", false},
	}

	for _, tc := range testCases {
		res := IsSkipMessage(tc.input)
		if res != tc.expected {
			t.Errorf("IsSkipMessage(%q) = %v, expected %v", tc.input, res, tc.expected)
		}
	}
}

func TestIsConfirmMessage(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"✅ Konfirmasi", true},
		{"konfirmasi", true},
		{"✅ Submit", true},
		{"submit", true},
		{"✅ Selesai", true},
		{"selesai", true},
		{"random text", false},
	}

	for _, tc := range testCases {
		res := IsConfirmMessage(tc.input)
		if res != tc.expected {
			t.Errorf("IsConfirmMessage(%q) = %v, expected %v", tc.input, res, tc.expected)
		}
	}
}

func TestIsHideMenuMessage(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"🔽 Tutup Menu", true},
		{"tutup menu", true},
		{"close menu", true},
		{"random text", false},
	}

	for _, tc := range testCases {
		res := IsHideMenuMessage(tc.input)
		if res != tc.expected {
			t.Errorf("IsHideMenuMessage(%q) = %v, expected %v", tc.input, res, tc.expected)
		}
	}
}

func TestBuildStepKeyboard_AppendsCancel(t *testing.T) {
	rows := [][]string{
		{"Opsi 1", "Opsi 2"},
		{"Opsi 3"},
	}

	kb := BuildStepKeyboard(rows)

	if len(kb.Keyboard) != 3 {
		t.Fatalf("expected 3 rows (2 data + 1 cancel), got %d", len(kb.Keyboard))
	}

	lastRow := kb.Keyboard[len(kb.Keyboard)-1]
	if len(lastRow) != 1 || lastRow[0].Text != DefaultCancelButton {
		t.Errorf("expected last row to be [%s], got %+v", DefaultCancelButton, lastRow)
	}

	if !kb.ResizeKeyboard {
		t.Error("expected ResizeKeyboard to be true")
	}
}

func TestBuildMultiSelectKeyboard(t *testing.T) {
	available := []string{"Produk A", "Produk B"}
	selected := []string{"Produk A"}

	kb := BuildMultiSelectKeyboard(available, selected, ButtonSubmit)
	// Expect row 1: "✅ Produk A"
	// Expect row 2: "Produk B"
	// Expect row 3: "✅ Submit"
	// Expect row 4: "❌ Batal"
	if len(kb.Keyboard) != 4 {
		t.Fatalf("expected 4 rows in multi-select keyboard, got %d", len(kb.Keyboard))
	}

	if kb.Keyboard[0][0].Text != "✅ Produk A" {
		t.Errorf("expected checked item '✅ Produk A', got '%s'", kb.Keyboard[0][0].Text)
	}
	if kb.Keyboard[1][0].Text != "Produk B" {
		t.Errorf("expected unchecked item 'Produk B', got '%s'", kb.Keyboard[1][0].Text)
	}
	if kb.Keyboard[2][0].Text != ButtonSubmit {
		t.Errorf("expected finish button '%s', got '%s'", ButtonSubmit, kb.Keyboard[2][0].Text)
	}
}

func TestCancelOnlyKeyboard(t *testing.T) {
	kb := CancelOnlyKeyboard()

	if len(kb.Keyboard) != 1 {
		t.Fatalf("expected 1 row, got %d", len(kb.Keyboard))
	}

	if kb.Keyboard[0][0].Text != DefaultCancelButton {
		t.Errorf("expected button %s, got %s", DefaultCancelButton, kb.Keyboard[0][0].Text)
	}
}

func TestMainMenuKeyboard(t *testing.T) {
	kb := MainMenuKeyboard()
	if len(kb.Keyboard) != 4 {
		t.Fatalf("expected 4 rows in default main menu, got %d", len(kb.Keyboard))
	}

	if kb.Keyboard[0][0].Text != ButtonKirimInvoice {
		t.Errorf("expected button '%s', got '%s'", ButtonKirimInvoice, kb.Keyboard[0][0].Text)
	}
	if kb.Keyboard[3][0].Text != ButtonHideMenu {
		t.Errorf("expected button '%s', got '%s'", ButtonHideMenu, kb.Keyboard[3][0].Text)
	}
}
