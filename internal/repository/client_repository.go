package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hs1-bot/internal/domain"
	"time"
)

// ClientRepository defines database operations for Client entities.
type ClientRepository interface {
	Create(ctx context.Context, client *domain.Client) error
	GetByID(ctx context.Context, id int64) (*domain.Client, error)
	List(ctx context.Context) ([]domain.Client, error)
	Update(ctx context.Context, client *domain.Client) error
	Delete(ctx context.Context, id int64) error
}

type sqliteClientRepository struct {
	db *sql.DB
}

// NewClientRepository creates a new SQLite ClientRepository.
func NewClientRepository(db *sql.DB) ClientRepository {
	return &sqliteClientRepository{db: db}
}

func (r *sqliteClientRepository) Create(ctx context.Context, client *domain.Client) error {
	query := `
		INSERT INTO clients (pic_name, company_name, company_address, company_email, product_ordered, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	client.CreatedAt = now
	client.UpdatedAt = now

	res, err := r.db.ExecContext(ctx, query,
		client.PICName,
		client.CompanyName,
		client.CompanyAddress,
		client.CompanyEmail,
		client.ProductOrdered,
		client.CreatedAt,
		client.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert client: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	client.ID = id
	return nil
}

func (r *sqliteClientRepository) GetByID(ctx context.Context, id int64) (*domain.Client, error) {
	query := `
		SELECT id, pic_name, company_name, company_address, company_email, product_ordered, created_at, updated_at, deleted_at
		FROM clients WHERE id = ? AND deleted_at IS NULL
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var c domain.Client
	err := row.Scan(
		&c.ID,
		&c.PICName,
		&c.CompanyName,
		&c.CompanyAddress,
		&c.CompanyEmail,
		&c.ProductOrdered,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query client by id: %w", err)
	}
	return &c, nil
}

func (r *sqliteClientRepository) List(ctx context.Context) ([]domain.Client, error) {
	query := `
		SELECT id, pic_name, company_name, company_address, company_email, product_ordered, created_at, updated_at, deleted_at
		FROM clients WHERE deleted_at IS NULL ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query clients: %w", err)
	}
	defer rows.Close()

	var clients []domain.Client
	for rows.Next() {
		var c domain.Client
		if err := rows.Scan(
			&c.ID,
			&c.PICName,
			&c.CompanyName,
			&c.CompanyAddress,
			&c.CompanyEmail,
			&c.ProductOrdered,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}

	return clients, rows.Err()
}

func (r *sqliteClientRepository) Update(ctx context.Context, client *domain.Client) error {
	query := `
		UPDATE clients
		SET pic_name = ?, company_name = ?, company_address = ?, company_email = ?, product_ordered = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`
	client.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query,
		client.PICName,
		client.CompanyName,
		client.CompanyAddress,
		client.CompanyEmail,
		client.ProductOrdered,
		client.UpdatedAt,
		client.ID,
	)
	return err
}

func (r *sqliteClientRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now()
	query := `UPDATE clients SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, now, now, id)
	return err
}
