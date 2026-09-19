package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"hs1-bot/config"
	"hs1-bot/internal/assets"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/service/template"
	"io"
	"strings"

	"gopkg.in/gomail.v2"
)

// EmailService defines operations for sending emails.
type EmailService interface {
	SendInvoiceEmail(ctx context.Context, toEmail string, invoice *domain.Invoice, client *domain.Client, products []domain.Product) ([]byte, error)
}

type gomailService struct {
	cfg        *config.Config
	logoPath   string
	sigPath    string
	dialer     *gomail.Dialer
	pdfService PDFService
}

// NewEmailService creates a new EmailService using gomail and attaches PDF invoices.
func NewEmailService(cfg *config.Config, logoPath string, sigPath string) EmailService {
	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	if cfg.SMTPSSL {
		dialer.SSL = true
		dialer.TLSConfig = &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         cfg.SMTPHost,
		}
	}

	return &gomailService{
		cfg:        cfg,
		logoPath:   logoPath,
		sigPath:    sigPath,
		dialer:     dialer,
		pdfService: NewPDFService(logoPath, sigPath),
	}
}

// SendInvoiceEmail renders the HTML card email body, generates PDF document, attaches the PDF, and sends via SMTP.
func (s *gomailService) SendInvoiceEmail(ctx context.Context, toEmail string, invoice *domain.Invoice, client *domain.Client, products []domain.Product) ([]byte, error) {
	m := gomail.NewMessage()

	fromHeader := s.cfg.SMTPSenderEmail
	if s.cfg.SMTPSenderName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", s.cfg.SMTPSenderName, s.cfg.SMTPSenderEmail)
	}

	m.SetHeader("From", fromHeader)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", fmt.Sprintf("Invoice %s - %s", invoice.InvoiceNumber, client.CompanyName))

	logoBytes := assets.GetLogo(s.logoPath)
	logoCID := ""
	if len(logoBytes) > 0 {
		m.Embed("logo.png", gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(logoBytes)
			return err
		}))
		logoCID = "logo.png"
	}

	sigBytes := assets.GetSignature(s.sigPath)
	sigCID := ""
	if len(sigBytes) > 0 {
		m.Embed("signature.png", gomail.SetCopyFunc(func(w io.Writer) error {
			_, err := w.Write(sigBytes)
			return err
		}))
		sigCID = "signature.png"
	}

	data := template.BuildTemplateData(invoice, client, products, logoCID, sigCID)
	htmlBytes, err := template.RenderInvoiceHTML(data)
	if err != nil {
		return nil, fmt.Errorf("failed to render invoice html: %w", err)
	}

	m.SetBody("text/html", string(htmlBytes))

	// Generate PDF document
	pdfBytes, err := s.pdfService.GenerateInvoicePDF(invoice, client, products)
	if err != nil {
		return nil, fmt.Errorf("failed to generate invoice pdf: %w", err)
	}

	// Attach PDF document
	safeInvName := strings.ReplaceAll(invoice.InvoiceNumber, "/", "-")
	attachmentFilename := fmt.Sprintf("Invoice-%s.pdf", safeInvName)
	m.Attach(attachmentFilename, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(pdfBytes)
		return err
	}))

	if err := s.dialer.DialAndSend(m); err != nil {
		return pdfBytes, fmt.Errorf("failed to send invoice email via SMTP (%s:%d): %w", s.cfg.SMTPHost, s.cfg.SMTPPort, err)
	}

	return pdfBytes, nil
}
