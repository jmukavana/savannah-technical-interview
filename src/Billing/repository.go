package Billing

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	CreateInvoice(ctx context.Context, inv *Invoice) error
	GetInvoiceByOrder(ctx context.Context, orderID uuid.UUID) (*Invoice, error)
	CreatePayment(ctx context.Context, p *Payment) error
	UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status string, paidAt *time.Time) error
}

type repository struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRepository(db *sqlx.DB, log *zap.Logger) Repository { return &repository{db: db, log: log} }

func (r *repository) CreateInvoice(ctx context.Context, inv *Invoice) error {
	inv.ID = uuid.New()
	inv.IssuedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `INSERT INTO invoices (id,order_id,invoice_number,status,amount,currency,issued_at,due_at,paid_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, inv.ID, inv.OrderID, inv.InvoiceNumber, inv.Status, inv.Amount, inv.Currency, inv.IssuedAt, inv.DueAt, inv.PaidAt)
	return err
}

func (r *repository) GetInvoiceByOrder(ctx context.Context, orderID uuid.UUID) (*Invoice, error) {
	var inv Invoice
	if err := r.db.GetContext(ctx, &inv, `SELECT id,order_id,invoice_number,status,amount,currency,issued_at,due_at,paid_at FROM invoices WHERE order_id=$1`, orderID); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &inv, nil
}

func (r *repository) CreatePayment(ctx context.Context, p *Payment) error {
	p.ID = uuid.New()
	p.CreatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `INSERT INTO payments (id,invoice_id,provider,provider_payment_id,amount,currency,status,metadata,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, p.ID, p.InvoiceID, p.Provider, p.ProviderPaymentID, p.Amount, p.Currency, p.Status, p.Metadata, p.CreatedAt)
	return err
}

func (r *repository) UpdateInvoiceStatus(ctx context.Context, id uuid.UUID, status string, paidAt *time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE invoices SET status=$1, paid_at=$2 WHERE id=$3`, status, paidAt, id)
	return err
}
