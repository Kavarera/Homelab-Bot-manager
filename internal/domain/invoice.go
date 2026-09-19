package domain

import (
	"fmt"
	"time"
)

// Invoice represents a billing invoice entity.
type Invoice struct {
	ID            int64         `json:"id"`
	ClientID      int64         `json:"client_id"`
	InvoiceNumber string        `json:"invoice_number"`
	TotalPrice    float64       `json:"total_price"`
	IssueDate     time.Time     `json:"issue_date"`
	DueDate       time.Time     `json:"due_date"`
	Status        string        `json:"status"`
	Items         []InvoiceItem `json:"items,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

// InvoiceItem represents a single line item in an Invoice.
type InvoiceItem struct {
	ID        int64     `json:"id"`
	InvoiceID int64     `json:"invoice_id"`
	ProductID int64     `json:"product_id"`
	Qty       int       `json:"qty"`
	UnitPrice float64   `json:"unit_price"`
	Subtotal  float64   `json:"subtotal"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GenerateInvoiceNumber formats the invoice number according to business rules:
// "INV/{id_client}/{tahun}/{bulan}/{tanggal}" (e.g. INV/1/2026/09/19).
func GenerateInvoiceNumber(clientID int64, issueDate time.Time) string {
	return fmt.Sprintf("INV/%d/%04d/%02d/%02d",
		clientID,
		issueDate.Year(),
		int(issueDate.Month()),
		issueDate.Day(),
	)
}

// CalculateDueDate calculates the invoice due date:
// Always +14 days from issueDate, and if the result falls on a weekend (Saturday or Sunday),
// it rolls forward to Monday.
func CalculateDueDate(issueDate time.Time) time.Time {
	due := issueDate.AddDate(0, 0, 14)

	switch due.Weekday() {
	case time.Saturday:
		return due.AddDate(0, 0, 2)
	case time.Sunday:
		return due.AddDate(0, 0, 1)
	default:
		return due
	}
}
