package flow_test

import (
	"context"
	"hs1-bot/internal/bot/flow"
	"hs1-bot/internal/bot/session"
	"hs1-bot/internal/bot/ui"
	"hs1-bot/internal/domain"
	"hs1-bot/internal/repository"
	"testing"
	"time"
)

func TestTambahClientFlow_Success(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	clientRepo := repository.NewClientRepository(db)
	productRepo := repository.NewProductRepository(db)

	_ = productRepo.Create(context.Background(), &domain.Product{Name: "PayrollPro License (Month)", Price: 400000})
	_ = productRepo.Create(context.Background(), &domain.Product{Name: "Cloud Hosting Pro", Price: 150000})

	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)
	engine.Register(flow.NewTambahClientFlow(clientRepo, productRepo))

	const userID int64 = 11111

	// 1. Start Flow
	startCtx := createTestBotContext(userID, ui.ButtonTambahClient)
	_ = engine.StartFlow(startCtx, flow.TambahClientFlowID)

	// 2. Input Company Name
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "PT. Maju Bersama"))

	// 3. Input PIC Name
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "Budi Santoso"))

	// 4. Input Address
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "Jl. Sudirman No. 12, Jakarta"))

	// 5. Input Email
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "budi@majubersama.co.id"))

	// 6. Select Product 1
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "PayrollPro License (Month)"))

	// 7. Select Product 2
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "Cloud Hosting Pro"))

	// 8. Confirm product selection
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonConfirm))

	// 9. Confirm save client to DB
	handled, err := engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonConfirm))
	if err != nil || !handled {
		t.Fatalf("expected client save to be handled: %v", err)
	}

	// Verify client exists in DB
	clients, _ := clientRepo.List(context.Background())
	if len(clients) != 1 {
		t.Fatalf("expected 1 client in DB, got %d", len(clients))
	}

	c := clients[0]
	if c.CompanyName != "PT. Maju Bersama" || c.PICName != "Budi Santoso" {
		t.Errorf("unexpected client data: %+v", c)
	}
	if !c.HasProduct("PayrollPro License (Month)") || !c.HasProduct("Cloud Hosting Pro") {
		t.Errorf("expected both products in client, got: %s", c.ProductOrdered)
	}
}

func TestEditClientFlow_SuccessWithSkipAndProductModification(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	clientRepo := repository.NewClientRepository(db)
	productRepo := repository.NewProductRepository(db)

	_ = productRepo.Create(context.Background(), &domain.Product{Name: "PayrollPro License (Month)", Price: 400000})
	_ = productRepo.Create(context.Background(), &domain.Product{Name: "VPN Dedicated", Price: 200000})

	client := &domain.Client{
		CompanyName:    "PT. KavaLabs Indonesia",
		PICName:        "Darmawati",
		CompanyAddress: "WTC Matahari",
		CompanyEmail:   "finance@kavalabs.id",
		ProductOrdered: "PayrollPro License (Month)",
	}
	_ = clientRepo.Create(context.Background(), client)

	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)
	engine.Register(flow.NewEditClientFlow(clientRepo, productRepo))

	const userID int64 = 22222

	// 1. Start Edit Flow
	startCtx := createTestBotContext(userID, ui.ButtonEditClient)
	_ = engine.StartFlow(startCtx, flow.EditClientFlowID)

	// 2. Select Client
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "PT. KavaLabs Indonesia"))

	// 3. Skip Company Name
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonSkip))

	// 4. Update PIC Name
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "Darmawati Hartono"))

	// 5. Skip Address
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonSkip))

	// 6. Skip Email
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonSkip))

	// 7. In Product Menu: Click "➕ Tambah Produk"
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonOptionTambah))

	// 8. Select "VPN Dedicated" to add
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "VPN Dedicated"))

	// 9. Confirm product addition
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonConfirm))

	// 10. Confirm final save to DB
	handled, err := engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonConfirm))
	if err != nil || !handled {
		t.Fatalf("expected update to be handled: %v", err)
	}

	// Verify updated data in DB
	updated, _ := clientRepo.GetByID(context.Background(), client.ID)
	if updated.PICName != "Darmawati Hartono" {
		t.Errorf("expected PIC to be updated to Darmawati Hartono, got %s", updated.PICName)
	}
	if !updated.HasProduct("VPN Dedicated") || !updated.HasProduct("PayrollPro License (Month)") {
		t.Errorf("expected both products after edit, got %s", updated.ProductOrdered)
	}
}

func TestHapusClientFlow_Success(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	clientRepo := repository.NewClientRepository(db)

	client := &domain.Client{
		CompanyName:    "PT. KavaLabs Indonesia",
		PICName:        "Darmawati",
		CompanyAddress: "WTC Matahari",
		CompanyEmail:   "finance@kavalabs.id",
		ProductOrdered: "PayrollPro License (Month)",
	}
	_ = clientRepo.Create(context.Background(), client)

	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)
	engine.Register(flow.NewHapusClientFlow(clientRepo))

	const userID int64 = 55555

	// 1. Start Hapus Client Flow
	startCtx := createTestBotContext(userID, ui.ButtonHapusClient)
	_ = engine.StartFlow(startCtx, flow.HapusClientFlowID)

	// 2. Select Client "PT. KavaLabs Indonesia"
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "PT. KavaLabs Indonesia"))

	// 3. Confirm Delete
	handled, err := engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonConfirm))
	if err != nil || !handled {
		t.Fatalf("expected delete to be handled: %v", err)
	}

	// Verify client is soft-deleted
	activeClients, err := clientRepo.List(context.Background())
	if err != nil {
		t.Fatalf("failed to list clients: %v", err)
	}
	if len(activeClients) != 0 {
		t.Errorf("expected 0 active clients, got %d", len(activeClients))
	}
}
