package Billing

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Invoice struct {
	ID            uuid.UUID       `db:"id" json:"id"`
	OrderID       uuid.UUID       `db:"order_id" json:"order_id"`
	InvoiceNumber string          `db:"invoice_number" json:"invoice_number"`
	Status        string          `db:"status" json:"status"`
	Amount        decimal.Decimal `db:"amount" json:"amount"`
	Currency      string          `db:"currency" json:"currency"`
	IssuedAt      time.Time       `db:"issued_at" json:"issued_at"`
	DueAt         *time.Time      `db:"due_at" json:"due_at,omitempty"`
	PaidAt        *time.Time      `db:"paid_at" json:"paid_at,omitempty"`
}

type Payment struct {
	ID                uuid.UUID       `db:"id" json:"id"`
	InvoiceID         uuid.UUID       `db:"invoice_id" json:"invoice_id"`
	Provider          string          `db:"provider" json:"provider"`
	ProviderPaymentID *string         `db:"provider_payment_id" json:"provider_payment_id,omitempty"`
	Amount            decimal.Decimal `db:"amount" json:"amount"`
	Currency          string          `db:"currency" json:"currency"`
	Status            string          `db:"status" json:"status"`
	Metadata          []byte          `db:"metadata" json:"metadata"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
}
