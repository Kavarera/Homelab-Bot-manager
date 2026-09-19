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

func TestTambahProdukFlow_Success(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	productRepo := repository.NewProductRepository(db)

	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)
	engine.Register(flow.NewTambahProdukFlow(productRepo))

	const userID int64 = 33333

	// 1. Start Flow
	_ = engine.StartFlow(createTestBotContext(userID, ui.ButtonTambahProduk), flow.TambahProdukFlowID)

	// 2. Input Product Name
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "Server Hosting Basic"))

	// 3. Input Price with "k" suffix (e.g. 250k)
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "250k"))

	// 4. Confirm save
	handled, err := engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonConfirm))
	if err != nil || !handled {
		t.Fatalf("expected product save to be handled: %v", err)
	}

	// Verify in DB
	products, _ := productRepo.List(context.Background())
	if len(products) != 1 {
		t.Fatalf("expected 1 product in DB, got %d", len(products))
	}

	p := products[0]
	if p.Name != "Server Hosting Basic" || p.Price != 250000.0 {
		t.Errorf("unexpected product data: %+v", p)
	}
}

func TestEditProdukFlow_Success(t *testing.T) {
	db, err := repository.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	productRepo := repository.NewProductRepository(db)
	prod := &domain.Product{Name: "Old Product Name", Price: 100000.0}
	_ = productRepo.Create(context.Background(), prod)

	store := session.NewMemoryStore(1 * time.Hour)
	engine := flow.NewEngine(store, 1*time.Hour)
	engine.Register(flow.NewEditProdukFlow(productRepo))

	const userID int64 = 44444

	// 1. Start Flow
	_ = engine.StartFlow(createTestBotContext(userID, ui.ButtonEditProduk), flow.EditProdukFlowID)

	// 2. Select Product
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "Old Product Name"))

	// 3. Edit Name
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "New Product Name"))

	// 4. Edit Price with "M" suffix (1.2M)
	_, _ = engine.HandleActiveFlow(createTestBotContext(userID, "1.2M"))

	// 5. Confirm Save
	handled, err := engine.HandleActiveFlow(createTestBotContext(userID, ui.ButtonConfirm))
	if err != nil || !handled {
		t.Fatalf("expected product edit to be handled: %v", err)
	}

	// Verify in DB
	updated, _ := productRepo.GetByID(context.Background(), prod.ID)
	if updated.Name != "New Product Name" || updated.Price != 1200000.0 {
		t.Errorf("unexpected updated product data: %+v", updated)
	}
}
