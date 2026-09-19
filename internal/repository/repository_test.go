package repository

import (
	"context"
	"hs1-bot/internal/domain"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) (*sqliteClientRepository, *sqliteProductRepository, *sqliteInvoiceRepository) {
	db, err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init test database: %v", err)
	}

	clientRepo := NewClientRepository(db).(*sqliteClientRepository)
	productRepo := NewProductRepository(db).(*sqliteProductRepository)
	invoiceRepo := NewInvoiceRepository(db).(*sqliteInvoiceRepository)

	return clientRepo, productRepo, invoiceRepo
}

func TestClientRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	clientRepo, _, _ := setupTestDB(t)

	client := &domain.Client{
		PICName:        "John Doe",
		CompanyName:    "Acme Corp",
		CompanyAddress: "123 Business St",
		CompanyEmail:   "finance@acme.com",
		ProductOrdered: "Payroll Pro Service",
	}

	// 1. Create
	err := clientRepo.Create(ctx, client)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	if client.ID == 0 {
		t.Fatal("expected client ID to be populated")
	}

	// 2. GetByID
	retrieved, err := clientRepo.GetByID(ctx, client.ID)
	if err != nil {
		t.Fatalf("failed to get client: %v", err)
	}
	if retrieved == nil {
		t.Fatal("expected client to exist, got nil")
	}
	if retrieved.CompanyName != "Acme Corp" || retrieved.PICName != "John Doe" {
		t.Errorf("unexpected client data: %+v", retrieved)
	}

	// 3. Update
	client.CompanyName = "Acme Global Corp"
	err = clientRepo.Update(ctx, client)
	if err != nil {
		t.Fatalf("failed to update client: %v", err)
	}
	updated, _ := clientRepo.GetByID(ctx, client.ID)
	if updated.CompanyName != "Acme Global Corp" {
		t.Errorf("expected updated company name 'Acme Global Corp', got '%s'", updated.CompanyName)
	}

	// 4. List
	list, err := clientRepo.List(ctx)
	if err != nil {
		t.Fatalf("failed to list clients: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 client in list, got %d", len(list))
	}

	// 5. Delete (Soft Delete)
	err = clientRepo.Delete(ctx, client.ID)
	if err != nil {
		t.Fatalf("failed to delete client: %v", err)
	}
	deleted, _ := clientRepo.GetByID(ctx, client.ID)
	if deleted != nil {
		t.Error("expected client to be soft-deleted (not found via GetByID)")
	}
	listAfterDelete, _ := clientRepo.List(ctx)
	if len(listAfterDelete) != 0 {
		t.Errorf("expected 0 active clients in list, got %d", len(listAfterDelete))
	}

	// Verify in DB that record physically exists with deleted_at set
	var rawDeletedAt string
	err = clientRepo.db.QueryRowContext(ctx, "SELECT deleted_at FROM clients WHERE id = ?", client.ID).Scan(&rawDeletedAt)
	if err != nil || rawDeletedAt == "" {
		t.Errorf("expected deleted_at to be populated in database, err: %v, deleted_at: %s", err, rawDeletedAt)
	}
}

func TestProductRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	_, productRepo, _ := setupTestDB(t)

	product := &domain.Product{
		Name:  "Cloud VPS Hosting",
		Price: 1500000.0,
	}

	err := productRepo.Create(ctx, product)
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}
	if product.ID == 0 {
		t.Fatal("expected product ID to be populated")
	}

	retrieved, err := productRepo.GetByID(ctx, product.ID)
	if err != nil || retrieved == nil {
		t.Fatalf("failed to get product: %v", err)
	}
	if retrieved.Name != "Cloud VPS Hosting" || retrieved.Price != 1500000.0 {
		t.Errorf("unexpected product data: %+v", retrieved)
	}

	// Delete (Soft Delete)
	err = productRepo.Delete(ctx, product.ID)
	if err != nil {
		t.Fatalf("failed to delete product: %v", err)
	}
	deletedProd, _ := productRepo.GetByID(ctx, product.ID)
	if deletedProd != nil {
		t.Error("expected product to be soft-deleted (not found via GetByID)")
	}
	prodsAfterDelete, _ := productRepo.List(ctx)
	if len(prodsAfterDelete) != 0 {
		t.Errorf("expected 0 active products in list, got %d", len(prodsAfterDelete))
	}

	// Verify in DB that record physically exists with deleted_at set
	var rawDeletedAt string
	err = productRepo.db.QueryRowContext(ctx, "SELECT deleted_at FROM products WHERE id = ?", product.ID).Scan(&rawDeletedAt)
	if err != nil || rawDeletedAt == "" {
		t.Errorf("expected deleted_at to be populated in database, err: %v, deleted_at: %s", err, rawDeletedAt)
	}
}

func TestInvoiceRepository_TransactionAndCalculation(t *testing.T) {
	ctx := context.Background()
	clientRepo, productRepo, invoiceRepo := setupTestDB(t)

	// Create Client
	client := &domain.Client{
		PICName:        "Jane PIC",
		CompanyName:    "Tech Solutions",
		CompanyAddress: "456 Tech Park",
		CompanyEmail:   "billing@tech.com",
	}
	_ = clientRepo.Create(ctx, client)

	// Create Products
	p1 := &domain.Product{Name: "Software License", Price: 5000000.0}
	p2 := &domain.Product{Name: "Maintenance Fee", Price: 1000000.0}
	_ = productRepo.Create(ctx, p1)
	_ = productRepo.Create(ctx, p2)

	issueDate := time.Date(2026, time.September, 19, 9, 0, 0, 0, time.UTC)
	dueDate := domain.CalculateDueDate(issueDate)

	nextID, err := invoiceRepo.GetNextInvoiceID(ctx)
	if err != nil {
		t.Fatalf("failed to get next invoice ID: %v", err)
	}
	if nextID != 1 {
		t.Errorf("expected next invoice ID 1, got %d", nextID)
	}

	invNumber := domain.GenerateInvoiceNumber(client.ID, issueDate, nextID)

	invoice := &domain.Invoice{
		ID:            nextID,
		ClientID:      client.ID,
		InvoiceNumber: invNumber,
		TotalPrice:    6000000.0,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		Status:        "PENDING",
		Items: []domain.InvoiceItem{
			{ProductID: p1.ID, Qty: 1, UnitPrice: 5000000.0, Subtotal: 5000000.0},
			{ProductID: p2.ID, Qty: 1, UnitPrice: 1000000.0, Subtotal: 1000000.0},
		},
	}

	// Create Invoice with Items
	err = invoiceRepo.Create(ctx, invoice)
	if err != nil {
		t.Fatalf("failed to create invoice: %v", err)
	}
	if invoice.ID != 1 {
		t.Fatalf("expected invoice ID to be 1, got %d", invoice.ID)
	}

	// Check next ID after insertion
	nextID2, err := invoiceRepo.GetNextInvoiceID(ctx)
	if err != nil {
		t.Fatalf("failed to get next invoice ID: %v", err)
	}
	if nextID2 != 2 {
		t.Errorf("expected next invoice ID 2, got %d", nextID2)
	}

	// Retrieve by ID
	retrieved, err := invoiceRepo.GetByID(ctx, invoice.ID)
	if err != nil || retrieved == nil {
		t.Fatalf("failed to get invoice: %v", err)
	}
	if retrieved.InvoiceNumber != "INV/1/2026/09/19/1" {
		t.Errorf("expected invoice number 'INV/1/2026/09/19/1', got '%s'", retrieved.InvoiceNumber)
	}
	if len(retrieved.Items) != 2 {
		t.Fatalf("expected 2 invoice items, got %d", len(retrieved.Items))
	}
	if retrieved.Items[0].Subtotal != 5000000.0 || retrieved.Items[1].Subtotal != 1000000.0 {
		t.Errorf("unexpected item subtotals: %+v", retrieved.Items)
	}

	// Retrieve by Number
	byNum, err := invoiceRepo.GetByInvoiceNumber(ctx, invNumber)
	if err != nil || byNum == nil {
		t.Fatalf("failed to get invoice by number: %v", err)
	}
	if byNum.ID != invoice.ID {
		t.Errorf("expected ID %d, got %d", invoice.ID, byNum.ID)
	}

	// Update Status
	err = invoiceRepo.UpdateStatus(ctx, invoice.ID, "PAID")
	if err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	updatedInv, _ := invoiceRepo.GetByID(ctx, invoice.ID)
	if updatedInv.Status != "PAID" {
		t.Errorf("expected status 'PAID', got '%s'", updatedInv.Status)
	}
}

func TestInvoiceRepository_CheckMonthlyInvoiceExists(t *testing.T) {
	ctx := context.Background()
	clientRepo, productRepo, invoiceRepo := setupTestDB(t)

	client := &domain.Client{
		CompanyName:  "PT. KavaLabs Indonesia",
		CompanyEmail: "rafli.030715@gmail.com",
	}
	_ = clientRepo.Create(ctx, client)

	product := &domain.Product{
		Name:  "PayrollPro License (Month)",
		Price: 400000.0,
	}
	_ = productRepo.Create(ctx, product)

	// In September 2026
	issueDate := time.Date(2026, time.September, 15, 10, 0, 0, 0, time.UTC)
	inv := &domain.Invoice{
		ClientID:      client.ID,
		InvoiceNumber: "INV/1/2026/09/15/1",
		TotalPrice:    400000.0,
		IssueDate:     issueDate,
		DueDate:       domain.CalculateDueDate(issueDate),
		Status:        "SENT",
		Items: []domain.InvoiceItem{
			{ProductID: product.ID, Qty: 1, UnitPrice: 400000.0, Subtotal: 400000.0},
		},
	}
	_ = invoiceRepo.Create(ctx, inv)

	// Check for same month (September 2026) -> Should exist
	found, err := invoiceRepo.CheckMonthlyInvoiceExists(ctx, client.ID, product.ID, 2026, 9)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if found == nil {
		t.Fatal("expected monthly invoice to be found")
	}
	if found.InvoiceNumber != "INV/1/2026/09/15/1" {
		t.Errorf("expected invoice 'INV/1/2026/09/15/1', got '%s'", found.InvoiceNumber)
	}

	// Check for different month (October 2026) -> Should NOT exist
	notFound, err := invoiceRepo.CheckMonthlyInvoiceExists(ctx, client.ID, product.ID, 2026, 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if notFound != nil {
		t.Errorf("expected nil for October 2026, got: %+v", notFound)
	}
}
