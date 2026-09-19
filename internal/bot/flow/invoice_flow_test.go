package flow_test

import (
	"context"
	"fmt"
	"hs1-bot/internal/bot/flow"
	"hs1-bot/internal/bot/session"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/repository"
	"testing"
	"time"
)

type mockEmailService struct {
	sendCalled bool
	lastTo     string
}

func (m *mockEmailService) SendInvoiceEmail(ctx context.Context, toEmail string, invoice *domain.Invoice, client *domain.Client, products []domain.Product) ([]byte, error) {
	m.sendCalled = true
	m.lastTo = toEmail
	return []byte("%PDF-1.4 Mock Invoice PDF"), nil
}

func TestInvoiceFlow_FullCycleAndDuplicateRejection(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	clientRepo := repository.NewClientRepository(db)
	productRepo := repository.NewProductRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	mockEmail := &mockEmailService{}

	// Seed client and product
	client := &domain.Client{
		PICName:        "Darmawati Hartono",
		CompanyName:    "PT. KavaLabs Indonesia",
		CompanyAddress: "WTC Matahari",
		CompanyEmail:   "rafli.030715@gmail.com",
		ProductOrdered: "PayrollPro License (Month)",
	}
	_ = clientRepo.Create(context.Background(), client)

	product := &domain.Product{
		Name:  "PayrollPro License (Month)",
		Price: 400000.0,
	}
	_ = productRepo.Create(context.Background(), product)

	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)

	invoiceFlow := flow.NewInvoiceFlow(clientRepo, productRepo, invoiceRepo, mockEmail)
	engine.Register(invoiceFlow)

	const userID int64 = 123456

	// 1. User starts Kirim Invoice Flow
	startCtx := createTestBotContext(userID, "📄 Kirim Invoice")
	err = engine.StartFlow(startCtx, flow.InvoiceFlowID)
	if err != nil {
		t.Fatalf("failed to start invoice flow: %v", err)
	}

	// 2. User selects client "PT. KavaLabs Indonesia"
	clientCtx := createTestBotContext(userID, "PT. KavaLabs Indonesia")
	handled, err := engine.HandleActiveFlow(clientCtx)
	if err != nil || !handled {
		t.Fatalf("expected client selection to be handled, handled=%v, err=%v", handled, err)
	}

	// 3. User selects product "PayrollPro License (Month)"
	productCtx := createTestBotContext(userID, "PayrollPro License (Month)")
	handled, err = engine.HandleActiveFlow(productCtx)
	if err != nil || !handled {
		t.Fatalf("expected product selection to be handled, handled=%v, err=%v", handled, err)
	}

	// 4. User clicks Submit
	submitCtx := createTestBotContext(userID, ui.ButtonSubmit)
	handled, err = engine.HandleActiveFlow(submitCtx)
	if err != nil || !handled {
		t.Fatalf("expected submit to be handled, handled=%v, err=%v", handled, err)
	}

	if !mockEmail.sendCalled {
		t.Error("expected email to be sent")
	}
	if mockEmail.lastTo != "rafli.030715@gmail.com" {
		t.Errorf("expected email to be sent to rafli.030715@gmail.com, got %s", mockEmail.lastTo)
	}

	// Verify invoice exists in DB
	invoices, _ := invoiceRepo.List(context.Background())
	if len(invoices) != 1 {
		t.Fatalf("expected 1 invoice in DB, got %d", len(invoices))
	}
	expectedInvNum := domain.GenerateInvoiceNumber(client.ID, time.Now(), invoices[0].ID)
	if invoices[0].InvoiceNumber != expectedInvNum {
		t.Errorf("expected invoice number %q, got %q", expectedInvNum, invoices[0].InvoiceNumber)
	}

	// 5. Test Duplicate Rejection in the same month
	mockEmail.sendCalled = false

	// Start flow again
	_ = engine.StartFlow(startCtx, flow.InvoiceFlowID)
	_, _ = engine.HandleActiveFlow(clientCtx)
	_, _ = engine.HandleActiveFlow(productCtx)
	handled, err = engine.HandleActiveFlow(submitCtx)
	if err != nil || !handled {
		t.Fatalf("expected duplicate attempt to be handled: %v", err)
	}

	// Email should NOT be called again because duplicate was rejected
	if mockEmail.sendCalled {
		t.Error("expected email NOT to be sent for duplicate monthly invoice")
	}

	// Invoice count in DB should still be 1
	invoicesAfter, _ := invoiceRepo.List(context.Background())
	if len(invoicesAfter) != 1 {
		t.Errorf("expected invoice count to remain 1, got %d", len(invoicesAfter))
	}
}

type failingEmailService struct{}

func (f *failingEmailService) SendInvoiceEmail(ctx context.Context, toEmail string, invoice *domain.Invoice, client *domain.Client, products []domain.Product) ([]byte, error) {
	return nil, fmt.Errorf("simulated SMTP connection failure")
}

func TestInvoiceFlow_EmailFailureDoesNotSaveToDB(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init test db: %v", err)
	}
	defer db.Close()

	clientRepo := repository.NewClientRepository(db)
	productRepo := repository.NewProductRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	failingEmail := &failingEmailService{}

	client := &domain.Client{
		CompanyName:    "PT. KavaLabs Indonesia",
		CompanyEmail:   "rafli.030715@gmail.com",
		ProductOrdered: "PayrollPro License (Month)",
	}
	_ = clientRepo.Create(context.Background(), client)

	product := &domain.Product{
		Name:  "PayrollPro License (Month)",
		Price: 400000.0,
	}
	_ = productRepo.Create(context.Background(), product)

	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)

	invoiceFlow := flow.NewInvoiceFlow(clientRepo, productRepo, invoiceRepo, failingEmail)
	engine.Register(invoiceFlow)

	const userID int64 = 99999
	startCtx := createTestBotContext(userID, "📄 Kirim Invoice")
	_ = engine.StartFlow(startCtx, flow.InvoiceFlowID)

	clientCtx := createTestBotContext(userID, "PT. KavaLabs Indonesia")
	_, _ = engine.HandleActiveFlow(clientCtx)

	productCtx := createTestBotContext(userID, "PayrollPro License (Month)")
	_, _ = engine.HandleActiveFlow(productCtx)

	submitCtx := createTestBotContext(userID, ui.ButtonSubmit)
	handled, err := engine.HandleActiveFlow(submitCtx)
	if err != nil || !handled {
		t.Fatalf("expected submit to be handled, handled=%v, err=%v", handled, err)
	}

	// Verify that NO invoice was saved in DB because email failed
	invoices, err := invoiceRepo.List(context.Background())
	if err != nil {
		t.Fatalf("failed to list invoices: %v", err)
	}
	if len(invoices) != 0 {
		t.Errorf("expected 0 invoices in DB on email failure, got %d", len(invoices))
	}
}
