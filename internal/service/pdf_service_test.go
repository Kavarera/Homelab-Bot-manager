package service_test

import (
	"bytes"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/service"
	"testing"
	"time"
)

func TestPDFService_GenerateInvoicePDF(t *testing.T) {
	pdfService := service.NewPDFService("assets/logo.png", "assets/signature.png")

	invoice := &domain.Invoice{
		ID:            1,
		ClientID:      1,
		InvoiceNumber: "INV/1/2026/09/19/1",
		TotalPrice:    400000.0,
		IssueDate:     time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC),
		DueDate:       time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		Status:        "SENT",
		Items: []domain.InvoiceItem{
			{
				ProductID: 1,
				Qty:       1,
				UnitPrice: 400000.0,
				Subtotal:  400000.0,
			},
		},
	}

	client := &domain.Client{
		ID:             1,
		PICName:        "Darmawati Hartono",
		CompanyName:    "PT. KavaLabs Indonesia",
		CompanyAddress: "WTC Matahari",
		CompanyEmail:   "rafli.030715@gmail.com",
		ProductOrdered: "PayrollPro License (Month)",
	}

	product := &domain.Product{
		ID:    1,
		Name:  "PayrollPro License (Month)",
		Price: 400000.0,
	}

	pdfBytes, err := pdfService.GenerateInvoicePDF(invoice, client, []domain.Product{*product})
	if err != nil {
		t.Fatalf("failed to generate invoice PDF: %v", err)
	}

	if len(pdfBytes) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}

	// Verify standard PDF header (%PDF-)
	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Errorf("expected PDF header %%PDF-, got %s", string(pdfBytes[:10]))
	}
}
