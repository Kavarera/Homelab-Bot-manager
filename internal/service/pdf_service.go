package service

import (
	"bytes"
	"fmt"
	"hs1-bot/internal/assets"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/service/template"

	"github.com/go-pdf/fpdf"
)

// PDFService defines invoice PDF generation operations.
type PDFService interface {
	GenerateInvoicePDF(invoice *domain.Invoice, client *domain.Client, products []domain.Product) ([]byte, error)
}

type invoicePDFService struct {
	logoPath string
	sigPath  string
}

// NewPDFService creates a new PDFService.
func NewPDFService(logoPath string, sigPath string) PDFService {
	return &invoicePDFService{
		logoPath: logoPath,
		sigPath:  sigPath,
	}
}

// GenerateInvoicePDF renders a pixel-perfect, clean, modern A4 invoice PDF supporting multiple items.
func (s *invoicePDFService) GenerateInvoicePDF(invoice *domain.Invoice, client *domain.Client, products []domain.Product) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	// 1. Header: "INVOICE" and Logo Image
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(100, 12, "INVOICE", "", 0, "L", false, 0, "")

	logoBytes := assets.GetLogo(s.logoPath)
	if len(logoBytes) > 0 {
		pdf.RegisterImageOptionsReader("company_logo", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(logoBytes))
		pdf.ImageOptions("company_logo", 160, 15, 25, 0, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}
	pdf.Ln(16)

	// 2. Metadata: Invoice No, Date, Due Date, Status
	pdf.SetDrawColor(231, 229, 228)
	pdf.SetLineWidth(0.3)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(4)

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(30, 6, "INVOICE NO:", "", 0, "L", false, 0, "")
	pdf.SetFont("Courier", "B", 10)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(50, 6, invoice.InvoiceNumber, "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(25, 6, "DATE:", "", 0, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(65, 6, invoice.IssueDate.Format("02.01.2006"), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(30, 6, "DUE DATE:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(50, 6, invoice.DueDate.Format("02.01.2006"), "", 0, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(25, 6, "STATUS:", "", 0, "R", false, 0, "")
	pdf.SetTextColor(16, 149, 108)
	pdf.CellFormat(65, 6, "SENT / UNPAID", "", 1, "R", false, 0, "")

	pdf.Ln(4)
	pdf.Line(20, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(6)

	// 3. Issued To Section
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(0, 5, "ISSUED TO:", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(0, 5, client.PICName, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(68, 64, 60)
	pdf.CellFormat(0, 5, client.CompanyName, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(120, 113, 108)
	if client.CompanyAddress != "" {
		// MultiCell wraps long addresses cleanly so they never overflow to the right
		pdf.MultiCell(105, 4.5, client.CompanyAddress, "", "L", false)
	}
	if client.CompanyEmail != "" {
		pdf.CellFormat(0, 5, fmt.Sprintf("Email: %s", client.CompanyEmail), "", 1, "L", false, 0, "")
	}

	pdf.Ln(6)

	// 4. Items Table
	// Table Header
	pdf.SetFillColor(245, 245, 244)
	pdf.SetDrawColor(28, 25, 23)
	pdf.SetLineWidth(0.5)

	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(85, 8, " DESCRIPTION", "B", 0, "L", true, 0, "")
	pdf.CellFormat(30, 8, "RATE ", "B", 0, "C", true, 0, "")
	pdf.CellFormat(20, 8, "QTY ", "B", 0, "C", true, 0, "")
	pdf.CellFormat(35, 8, "TOTAL ", "B", 1, "R", true, 0, "")

	prodMap := make(map[int64]domain.Product)
	for _, p := range products {
		prodMap[p.ID] = p
	}

	for _, item := range invoice.Items {
		name := fmt.Sprintf("Item #%d", item.ProductID)
		if prod, ok := prodMap[item.ProductID]; ok {
			name = prod.Name
		}
		rateStr := fmt.Sprintf("IDR %s", template.FormatCurrencyShort(item.UnitPrice))
		totalItemStr := fmt.Sprintf("IDR %s", template.FormatCurrencyShort(item.Subtotal))

		pdf.SetLineWidth(0.2)
		pdf.SetDrawColor(231, 229, 228)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(28, 25, 23)
		pdf.CellFormat(85, 10, fmt.Sprintf(" %s", name), "B", 0, "L", false, 0, "")
		pdf.SetFont("Courier", "", 9)
		pdf.CellFormat(30, 10, rateStr, "B", 0, "C", false, 0, "")
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(20, 10, fmt.Sprintf("%d", item.Qty), "B", 0, "C", false, 0, "")
		pdf.SetFont("Arial", "B", 9)
		pdf.CellFormat(35, 10, fmt.Sprintf("%s ", totalItemStr), "B", 1, "R", false, 0, "")
	}

	pdf.Ln(4)

	// 5. Total Section
	totalPriceStr := fmt.Sprintf("IDR %s", template.FormatCurrencyShort(invoice.TotalPrice))
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(135, 6, "SUBTOTAL: ", "", 0, "R", false, 0, "")
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(35, 6, fmt.Sprintf("%s ", totalPriceStr), "", 1, "R", false, 0, "")

	pdf.SetDrawColor(28, 25, 23)
	pdf.SetLineWidth(0.5)
	pdf.Line(120, pdf.GetY(), 190, pdf.GetY())
	pdf.Ln(2)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(12, 10, 9)
	pdf.CellFormat(135, 8, "TOTAL (NETT): ", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 8, fmt.Sprintf("%s ", totalPriceStr), "", 1, "R", false, 0, "")

	pdf.Ln(10)

	// 6. Payment Info (Left) & Signature (Right)
	startY := pdf.GetY()

	// Payment Info
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(90, 5, "PAYMENT INFO:", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(90, 5, "Bank BLU (BCA DIGITAL)", "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(87, 83, 78)
	pdf.CellFormat(90, 5, "Account Name: Rafli Iskandar Kavarera", "", 1, "L", false, 0, "")

	pdf.SetFont("Courier", "B", 9)
	pdf.SetTextColor(28, 25, 23)
	pdf.CellFormat(90, 5, "Account No: 0024 8934 0640", "", 1, "L", false, 0, "")

	// Signature Image on the right
	sigBytes := assets.GetSignature(s.sigPath)
	if len(sigBytes) > 0 {
		pdf.RegisterImageOptionsReader("director_signature", fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(sigBytes))
		pdf.ImageOptions("director_signature", 135, startY, 50, 0, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	pdf.SetY(startY + 20)
	pdf.SetX(135)
	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(168, 162, 158)
	pdf.CellFormat(50, 4, "AUTHORIZED SIGNATORY", "", 1, "C", false, 0, "")

	// 7. Footnote
	pdf.SetY(260)
	pdf.SetDrawColor(231, 229, 228)
	pdf.SetLineWidth(0.2)
	pdf.Line(20, 258, 190, 258)

	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(120, 113, 108)
	pdf.CellFormat(0, 5, "*Mohon lampirkan bukti transfer dan bukti potong pajak.", "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate invoice PDF buffer: %w", err)
	}

	return buf.Bytes(), nil
}
