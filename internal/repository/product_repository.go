package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hs1-bot/internal/domain"
	"time"
)

// ProductRepository defines database operations for Product entities.
type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context) ([]domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id int64) error
}

type sqliteProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new SQLite ProductRepository.
func NewProductRepository(db *sql.DB) ProductRepository {
	return &sqliteProductRepository{db: db}
}

func (r *sqliteProductRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (name, price, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	res, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Price,
		product.CreatedAt,
		product.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert product: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	product.ID = id
	return nil
}

func (r *sqliteProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at
		FROM products WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var p domain.Product
	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Price,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query product by id: %w", err)
	}
	return &p, nil
}

func (r *sqliteProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	query := `
		SELECT id, name, price, created_at, updated_at
		FROM products ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Price,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, rows.Err()
}

func (r *sqliteProductRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = ?, price = ?, updated_at = ?
		WHERE id = ?
	`
	product.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		product.Name,
		product.Price,
		product.UpdatedAt,
		product.ID,
	)
	return err
}

func (r *sqliteProductRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
