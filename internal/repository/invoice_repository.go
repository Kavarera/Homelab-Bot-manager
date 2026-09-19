package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"hs1-bot/internal/domain"
	"time"
)

// InvoiceRepository defines database operations for Invoice and InvoiceItem entities.
type InvoiceRepository interface {
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetByID(ctx context.Context, id int64) (*domain.Invoice, error)
	GetByInvoiceNumber(ctx context.Context, invNum string) (*domain.Invoice, error)
	CheckMonthlyInvoiceExists(ctx context.Context, clientID int64, productID int64, year int, month int) (*domain.Invoice, error)
	List(ctx context.Context) ([]domain.Invoice, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	Delete(ctx context.Context, id int64) error
}

type sqliteInvoiceRepository struct {
	db *sql.DB
}

// NewInvoiceRepository creates a new SQLite InvoiceRepository.
func NewInvoiceRepository(db *sql.DB) InvoiceRepository {
	return &sqliteInvoiceRepository{db: db}
}

// Create inserts an invoice and all its items inside a single atomic database transaction.
func (r *sqliteInvoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := time.Now()
	invoice.CreatedAt = now
	invoice.UpdatedAt = now

	if invoice.Status == "" {
		invoice.Status = "PENDING"
	}

	invoiceQuery := `
		INSERT INTO invoices (client_id, invoice_number, total_price, issue_date, due_date, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := tx.ExecContext(ctx, invoiceQuery,
		invoice.ClientID,
		invoice.InvoiceNumber,
		invoice.TotalPrice,
		invoice.IssueDate,
		invoice.DueDate,
		invoice.Status,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert invoice: %w", err)
	}

	invoiceID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	invoice.ID = invoiceID

	itemQuery := `
		INSERT INTO invoice_items (invoice_id, product_id, qty, unit_price, subtotal, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	for i := range invoice.Items {
		item := &invoice.Items[i]
		item.InvoiceID = invoiceID
		item.CreatedAt = now
		item.UpdatedAt = now

		itemRes, err := tx.ExecContext(ctx, itemQuery,
			item.InvoiceID,
			item.ProductID,
			item.Qty,
			item.UnitPrice,
			item.Subtotal,
			item.CreatedAt,
			item.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert invoice item for product %d: %w", item.ProductID, err)
		}

		itemID, err := itemRes.LastInsertId()
		if err != nil {
			return err
		}
		item.ID = itemID
	}

	return tx.Commit()
}

func (r *sqliteInvoiceRepository) GetByID(ctx context.Context, id int64) (*domain.Invoice, error) {
	query := `
		SELECT id, client_id, invoice_number, total_price, issue_date, due_date, status, created_at, updated_at
		FROM invoices WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var inv domain.Invoice
	err := row.Scan(
		&inv.ID,
		&inv.ClientID,
		&inv.InvoiceNumber,
		&inv.TotalPrice,
		&inv.IssueDate,
		&inv.DueDate,
		&inv.Status,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice by id: %w", err)
	}

	items, err := r.getItemsByInvoiceID(ctx, inv.ID)
	if err != nil {
		return nil, err
	}
	inv.Items = items

	return &inv, nil
}

func (r *sqliteInvoiceRepository) GetByInvoiceNumber(ctx context.Context, invNum string) (*domain.Invoice, error) {
	query := `
		SELECT id, client_id, invoice_number, total_price, issue_date, due_date, status, created_at, updated_at
		FROM invoices WHERE invoice_number = ?
	`
	row := r.db.QueryRowContext(ctx, query, invNum)

	var inv domain.Invoice
	err := row.Scan(
		&inv.ID,
		&inv.ClientID,
		&inv.InvoiceNumber,
		&inv.TotalPrice,
		&inv.IssueDate,
		&inv.DueDate,
		&inv.Status,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice by number: %w", err)
	}

	items, err := r.getItemsByInvoiceID(ctx, inv.ID)
	if err != nil {
		return nil, err
	}
	inv.Items = items

	return &inv, nil
}

// CheckMonthlyInvoiceExists checks if an invoice already exists for the client and product in the given month/year.
func (r *sqliteInvoiceRepository) CheckMonthlyInvoiceExists(ctx context.Context, clientID int64, productID int64, year int, month int) (*domain.Invoice, error) {
	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	query := `
		SELECT i.id, i.client_id, i.invoice_number, i.total_price, i.issue_date, i.due_date, i.status, i.created_at, i.updated_at
		FROM invoices i
		JOIN invoice_items ii ON i.id = ii.invoice_id
		WHERE i.client_id = ?
		  AND ii.product_id = ?
		  AND i.issue_date >= ?
		  AND i.issue_date < ?
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, clientID, productID, startOfMonth, endOfMonth)

	var inv domain.Invoice
	err := row.Scan(
		&inv.ID,
		&inv.ClientID,
		&inv.InvoiceNumber,
		&inv.TotalPrice,
		&inv.IssueDate,
		&inv.DueDate,
		&inv.Status,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check monthly invoice: %w", err)
	}

	items, err := r.getItemsByInvoiceID(ctx, inv.ID)
	if err != nil {
		return nil, err
	}
	inv.Items = items

	return &inv, nil
}

func (r *sqliteInvoiceRepository) List(ctx context.Context) ([]domain.Invoice, error) {
	query := `
		SELECT id, client_id, invoice_number, total_price, issue_date, due_date, status, created_at, updated_at
		FROM invoices ORDER BY id DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoices: %w", err)
	}
	defer rows.Close()

	var invoices []domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		if err := rows.Scan(
			&inv.ID,
			&inv.ClientID,
			&inv.InvoiceNumber,
			&inv.TotalPrice,
			&inv.IssueDate,
			&inv.DueDate,
			&inv.Status,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}

	return invoices, rows.Err()
}

func (r *sqliteInvoiceRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE invoices SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *sqliteInvoiceRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM invoices WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *sqliteInvoiceRepository) getItemsByInvoiceID(ctx context.Context, invoiceID int64) ([]domain.InvoiceItem, error) {
	query := `
		SELECT id, invoice_id, product_id, qty, unit_price, subtotal, created_at, updated_at
		FROM invoice_items WHERE invoice_id = ? ORDER BY id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invoice items: %w", err)
	}
	defer rows.Close()

	var items []domain.InvoiceItem
	for rows.Next() {
		var item domain.InvoiceItem
		if err := rows.Scan(
			&item.ID,
			&item.InvoiceID,
			&item.ProductID,
			&item.Qty,
			&item.UnitPrice,
			&item.Subtotal,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
