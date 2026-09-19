package template

import (
	"bytes"
	"fmt"
	"html/template"
	"hs1-bot/internal/domain"
	"strings"
)

// InvoiceItemData contains line item rendering data.
type InvoiceItemData struct {
	ProductName  string
	UnitPriceStr string
	Qty          int
	SubtotalStr  string
}

// InvoiceTemplateData contains the data needed to render an HTML invoice.
type InvoiceTemplateData struct {
	InvoiceNumber   string
	IssueDateStr    string
	DueDateStr      string
	PICName         string
	CompanyName     string
	CompanyAddress  string
	CompanyEmail    string
	Items           []InvoiceItemData
	TotalPriceStr   string
	TotalNettStr    string
	BankName        string
	BankAccount     string
	AccountName     string
	LogoCID         string
	LogoBase64      string
	SignatureCID    string
	SignatureBase64 string
}

// FormatCurrencyShort formats e.g. 400000 into "400K" or "IDR 400K".
func FormatCurrencyShort(amount float64) string {
	if amount >= 1000000 && float64(int64(amount)) == amount && int64(amount)%1000000 == 0 {
		return fmt.Sprintf("%dM", int64(amount)/1000000)
	}
	if amount >= 1000 && float64(int64(amount)) == amount && int64(amount)%1000 == 0 {
		return fmt.Sprintf("%dK", int64(amount)/1000)
	}
	return fmt.Sprintf("%.0f", amount)
}

// BuildTemplateData maps domain models into InvoiceTemplateData supporting multiple items.
func BuildTemplateData(invoice *domain.Invoice, client *domain.Client, products []domain.Product, logoCID string, signatureCID string) InvoiceTemplateData {
	prodMap := make(map[int64]domain.Product)
	for _, p := range products {
		prodMap[p.ID] = p
	}

	var itemsData []InvoiceItemData
	for _, item := range invoice.Items {
		name := fmt.Sprintf("Item #%d", item.ProductID)
		if prod, ok := prodMap[item.ProductID]; ok {
			name = prod.Name
		}
		itemsData = append(itemsData, InvoiceItemData{
			ProductName:  name,
			UnitPriceStr: FormatCurrencyShort(item.UnitPrice),
			Qty:          item.Qty,
			SubtotalStr:  FormatCurrencyShort(item.Subtotal),
		})
	}

	totalShort := FormatCurrencyShort(invoice.TotalPrice)

	return InvoiceTemplateData{
		InvoiceNumber:   invoice.InvoiceNumber,
		IssueDateStr:    invoice.IssueDate.Format("02.01.2006"),
		DueDateStr:      invoice.DueDate.Format("02.01.2006"),
		PICName:         client.PICName,
		CompanyName:     client.CompanyName,
		CompanyAddress:  client.CompanyAddress,
		CompanyEmail:    client.CompanyEmail,
		Items:           itemsData,
		TotalPriceStr:   fmt.Sprintf("IDR %s", totalShort),
		TotalNettStr:    fmt.Sprintf("IDR %s", totalShort),
		BankName:        "Bank BLU (BCA DIGITAL)",
		BankAccount:     "0024 8934 0640",
		AccountName:     "Rafli Iskandar Kavarera",
		LogoCID:         logoCID,
		SignatureCID:    signatureCID,
	}
}

const invoiceCardHTML = `
<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Invoice - {{.InvoiceNumber}}</title>
<style>
  body {
    font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    color: #1c1917;
    background-color: #f5f5f4;
    margin: 0;
    padding: 20px 10px;
  }
  .invoice-card {
    max-width: 580px;
    margin: 0 auto;
    background: #ffffff;
    border: 1px solid #e7e5e4;
    border-radius: 16px;
    padding: 32px;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.05);
  }
  .header-table {
    width: 100%;
    margin-bottom: 24px;
  }
  .title-invoice {
    font-size: 28px;
    font-weight: 900;
    letter-spacing: 2px;
    color: #171717;
    margin: 0;
  }
  .meta-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
    margin-bottom: 20px;
  }
  .meta-table td {
    padding: 4px 0;
  }
  .meta-label {
    font-weight: 700;
    color: #78716c;
    text-transform: uppercase;
    font-size: 11px;
  }
  .meta-val {
    text-align: right;
    font-weight: 600;
    color: #1c1917;
  }
  .meta-inv-no {
    font-family: monospace;
    font-size: 13px;
    font-weight: 700;
  }
  .section-divider {
    border-top: 1px solid #f0eeec;
    margin: 16px 0;
  }
  .issued-to {
    margin-bottom: 24px;
  }
  .issued-title {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #1c1917;
    margin: 0 0 4px 0;
  }
  .issued-pic {
    font-size: 14px;
    font-weight: 700;
    color: #1c1917;
    margin: 0;
  }
  .issued-company {
    font-size: 13px;
    font-weight: 700;
    color: #292524;
    margin: 2px 0;
  }
  .issued-address {
    font-size: 11px;
    color: #57534e;
    line-height: 1.4;
    margin: 4px 0 0 0;
  }
  .items-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
    margin-top: 12px;
  }
  .items-table th {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 1px;
    padding-bottom: 8px;
    border-bottom: 2px solid #1c1917;
    color: #1c1917;
  }
  .items-table td {
    padding: 12px 0;
    border-bottom: 1px solid #e7e5e4;
  }
  .total-section {
    margin-top: 16px;
    font-size: 13px;
  }
  .total-row {
    display: flex;
    justify-content: space-between;
    padding: 4px 0;
  }
  .grand-total {
    border-top: 1px dashed #d6d3d1;
    margin-top: 10px;
    padding-top: 10px;
    font-size: 15px;
    font-weight: 900;
  }
  .payment-signature-table {
    width: 100%;
    margin-top: 28px;
    padding-top: 16px;
    border-top: 1px solid #e7e5e4;
  }
  .payment-title {
    font-size: 11px;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #1c1917;
    margin-bottom: 4px;
  }
  .bank-name {
    font-size: 12px;
    font-weight: 700;
    color: #292524;
    margin: 0 0 4px 0;
  }
  .bank-detail {
    font-size: 11px;
    color: #57534e;
    margin: 2px 0;
  }
  .acc-no {
    font-family: monospace;
    font-weight: 700;
    background: #f5f5f4;
    padding: 2px 6px;
    border-radius: 4px;
    border: 1px solid #e7e5e4;
    color: #1c1917;
  }
  .signature-box {
    text-align: right;
    vertical-align: bottom;
  }
  .signature-img {
    height: 52px;
    width: auto;
    display: inline-block;
  }
  .auth-signatory {
    font-size: 9px;
    text-transform: uppercase;
    letter-spacing: 1px;
    color: #a8a29e;
    font-weight: 600;
    display: block;
    margin-top: 4px;
  }
  .footnote {
    margin-top: 24px;
    padding-top: 12px;
    border-top: 1px solid #f5f5f4;
    font-size: 10px;
    color: #78716c;
    font-style: italic;
  }
</style>
</head>
<body>
<div class="invoice-card">
  <!-- Header -->
  <table class="header-table" cellpadding="0" cellspacing="0" border="0">
    <tr>
      <td valign="middle">
        <h1 class="title-invoice">INVOICE</h1>
      </td>
      <td align="right" valign="middle">
        {{if .LogoCID}}
          <img src="cid:{{.LogoCID}}" alt="Logo" style="height: 48px; width: auto;" />
        {{else if .LogoBase64}}
          <img src="data:image/png;base64,{{.LogoBase64}}" alt="Logo" style="height: 48px; width: auto;" />
        {{end}}
      </td>
    </tr>
  </table>

  <!-- Metadata -->
  <table class="meta-table" cellpadding="0" cellspacing="0" border="0">
    <tr style="border-bottom: 1px dashed #e7e5e4;">
      <td class="meta-label">Invoice No:</td>
      <td class="meta-val meta-inv-no">{{.InvoiceNumber}}</td>
    </tr>
    <tr>
      <td class="meta-label">Date:</td>
      <td class="meta-val">{{.IssueDateStr}}</td>
    </tr>
    <tr>
      <td class="meta-label">Due Date:</td>
      <td class="meta-val">{{.DueDateStr}}</td>
    </tr>
  </table>

  <div class="section-divider"></div>

  <!-- Issued To -->
  <div class="issued-to">
    <div class="issued-title">Issued To:</div>
    <div class="issued-pic">{{.PICName}}</div>
    <div class="issued-company">{{.CompanyName}}</div>
    <div class="issued-address">{{.CompanyAddress}}</div>
  </div>

  <!-- Items Table -->
  <table class="items-table" cellpadding="0" cellspacing="0" border="0">
    <thead>
      <tr>
        <th align="left" style="width: 50%;">Description</th>
        <th align="center" style="width: 15%;">Rate</th>
        <th align="center" style="width: 15%;">Qty</th>
        <th align="right" style="width: 20%;">Total</th>
      </tr>
    </thead>
    <tbody>
      {{range .Items}}
      <tr>
        <td align="left" style="font-weight: 600; color: #1c1917;">{{.ProductName}}</td>
        <td align="center" style="color: #57534e; font-family: monospace;">{{.UnitPriceStr}}</td>
        <td align="center" style="color: #1c1917; font-weight: 600;">{{.Qty}}</td>
        <td align="right" style="color: #1c1917; font-weight: 600;">{{.SubtotalStr}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>

  <!-- Total Section -->
  <table style="width: 100%; margin-top: 14px; font-size: 12px;" cellpadding="0" cellspacing="0" border="0">
    <tr>
      <td align="left" style="font-weight: 700; text-transform: uppercase; color: #57534e; padding: 4px 0;">Subtotal</td>
      <td align="right" style="font-weight: 700; color: #1c1917; padding: 4px 0;">{{.TotalPriceStr}}</td>
    </tr>
    <tr style="border-top: 1px dashed #d6d3d1;">
      <td align="left" style="font-size: 14px; font-weight: 900; text-transform: uppercase; letter-spacing: 1px; color: #1c1917; padding-top: 10px;">Total</td>
      <td align="right" style="font-size: 15px; font-weight: 900; color: #0c0a09; padding-top: 10px;">
        {{.TotalNettStr}}
        <span style="font-size: 10px; font-weight: 600; color: #78716c; display: block;">(NETT)</span>
      </td>
    </tr>
  </table>

  <!-- Payment Info & Signature -->
  <table class="payment-signature-table" cellpadding="0" cellspacing="0" border="0">
    <tr>
      <td valign="top" style="width: 55%;">
        <div class="payment-title">Payment Info:</div>
        <div class="bank-name">{{.BankName}}</div>
        <div class="bank-detail">Account Name: <span style="font-weight: 600; color: #1c1917;">{{.AccountName}}</span></div>
        <div class="bank-detail" style="margin-top: 4px;">Account No.: <span class="acc-no">{{.BankAccount}}</span></div>
      </td>
      <td valign="bottom" align="right" class="signature-box" style="width: 45%;">
        {{if .SignatureCID}}
          <img src="cid:{{.SignatureCID}}" alt="Signature" class="signature-img" />
        {{else if .SignatureBase64}}
          <img src="data:image/png;base64,{{.SignatureBase64}}" alt="Signature" class="signature-img" />
        {{end}}
        <span class="auth-signatory">Authorized Signatory</span>
      </td>
    </tr>
  </table>

  <!-- Footnote -->
  <div class="footnote">
    *Mohon lampirkan bukti transfer dan bukti potong pajak.
  </div>
</div>
</body>
</html>
`

// RenderInvoiceHTML renders the complete HTML document for email body and attachment.
func RenderInvoiceHTML(data InvoiceTemplateData) ([]byte, error) {
	tmpl, err := template.New("invoice").Parse(strings.TrimSpace(invoiceCardHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse invoice template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute invoice template: %w", err)
	}

	return buf.Bytes(), nil
}
