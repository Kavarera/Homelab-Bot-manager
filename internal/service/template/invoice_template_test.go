package template_test

import (
	"hs1-bot/internal/domain"
	"hs1-bot/internal/service/template"
	"strings"
	"testing"
	"time"
)

func TestFormatCurrencyShort(t *testing.T) {
	if res := template.FormatCurrencyShort(400000); res != "400K" {
		t.Errorf("expected 400K, got %s", res)
	}
	if res := template.FormatCurrencyShort(15000000); res != "15M" {
		t.Errorf("expected 15M, got %s", res)
	}
	if res := template.FormatCurrencyShort(250); res != "250" {
		t.Errorf("expected 250, got %s", res)
	}
}

func TestRenderInvoiceHTML(t *testing.T) {
	issueDate := time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC)
	dueDate := domain.CalculateDueDate(issueDate)

	client := &domain.Client{
		ID:             1,
		PICName:        "Darmawati Hartono",
		CompanyName:    "PT. KavaLabs Indonesia",
		CompanyAddress: "WTC Matahari, Ruko Royal Serpong Village",
		CompanyEmail:   "finance@kavalabs.id",
	}

	product := &domain.Product{
		ID:    1,
		Name:  "PayrollPro License (Month)",
		Price: 400000,
	}

	invoice := &domain.Invoice{
		ID:            100,
		ClientID:      client.ID,
		InvoiceNumber: "INV/1/2026/08/03/100",
		TotalPrice:    400000,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		Status:        "SENT",
		Items: []domain.InvoiceItem{
			{
				ProductID: product.ID,
				Qty:       1,
				UnitPrice: product.Price,
				Subtotal:  product.Price,
			},
		},
	}

	data := template.BuildTemplateData(invoice, client, []domain.Product{*product}, "company-logo", "signature-img")
	htmlBytes, err := template.RenderInvoiceHTML(data)
	if err != nil {
		t.Fatalf("failed to render invoice html: %v", err)
	}

	htmlStr := string(htmlBytes)

	// Verify crucial fields are present
	if !strings.Contains(htmlStr, "INV/1/2026/08/03/100") {
		t.Error("invoice number missing from rendered HTML")
	}
	if !strings.Contains(htmlStr, "Darmawati Hartono") {
		t.Error("PIC name missing from rendered HTML")
	}
	if !strings.Contains(htmlStr, "PT. KavaLabs Indonesia") {
		t.Error("Company name missing from rendered HTML")
	}
	if !strings.Contains(htmlStr, "PayrollPro License (Month)") {
		t.Error("Product name missing from rendered HTML")
	}
	if !strings.Contains(htmlStr, "IDR 400K") {
		t.Error("Total price missing from rendered HTML")
	}
	if !strings.Contains(htmlStr, "cid:company-logo") {
		t.Error("Logo CID missing from rendered HTML")
	}
	if !strings.Contains(htmlStr, "cid:signature-img") {
		t.Error("Signature CID missing from rendered HTML")
	}
}
