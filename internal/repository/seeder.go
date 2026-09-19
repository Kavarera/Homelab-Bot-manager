package repository

import (
	"context"
	"database/sql"
	"fmt"
	"hs1-bot/internal/domain"
	"log/slog"
)

// SeedDefaultData seeds initial client and product data if tables are empty.
func SeedDefaultData(db *sql.DB) error {
	ctx := context.Background()
	clientRepo := NewClientRepository(db)
	productRepo := NewProductRepository(db)

	clients, err := clientRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing clients: %w", err)
	}

	if len(clients) == 0 {
		defaultClient := &domain.Client{
			PICName:        "Darmawati Hartono",
			CompanyName:    "PT. KavaLabs Indonesia",
			CompanyAddress: "WTC Matahari, Ruko Royal Serpong Village, Jl Raya Serpong No.835, Pondok Jagung, Kec Serpong Utara, Kota Tangerang Selatan, Banten 15326, Indonesia",
			CompanyEmail:   "rafli.030715@gmail.com",
			ProductOrdered: "PayrollPro License (Month)",
		}
		if err := clientRepo.Create(ctx, defaultClient); err != nil {
			return fmt.Errorf("failed to seed default client: %w", err)
		}
		slog.Info("Default client seeded", slog.String("company", defaultClient.CompanyName))
	}

	products, err := productRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to check existing products: %w", err)
	}

	if len(products) == 0 {
		defaultProduct := &domain.Product{
			Name:  "PayrollPro License (Month)",
			Price: 400000.0,
		}
		if err := productRepo.Create(ctx, defaultProduct); err != nil {
			return fmt.Errorf("failed to seed default product: %w", err)
		}
		slog.Info("Default product seeded", slog.String("name", defaultProduct.Name), slog.Float64("price", defaultProduct.Price))
	}

	return nil
}
