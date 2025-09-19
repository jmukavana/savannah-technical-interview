package Orders

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	CustomerID *uuid.UUID      `db:"customer_id" json:"customer_id,omitempty"`
	Status     string          `db:"status" json:"status"`
	Subtotal   decimal.Decimal `db:"subtotal" json:"subtotal"`
	Tax        decimal.Decimal `db:"tax" json:"tax"`
	Shipping   decimal.Decimal `db:"shipping" json:"shipping"`
	Total      decimal.Decimal `db:"total" json:"total"`
	Currency   string          `db:"currency" json:"currency"`
	CreatedAt  time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time       `db:"updated_at" json:"updated_at"`
	Version    int             `db:"version" json:"version"`
}

type OrderItem struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	OrderID   uuid.UUID       `db:"order_id" json:"order_id"`
	ProductID *uuid.UUID      `db:"product_id" json:"product_id,omitempty"`
	SKU       *string         `db:"sku" json:"sku,omitempty"`
	Name      *string         `db:"name" json:"name,omitempty"`
	UnitPrice decimal.Decimal `db:"unit_price" json:"unit_price"`
	Quantity  int             `db:"quantity" json:"quantity"`
	LineTotal decimal.Decimal `db:"line_total" json:"line_total"`
}
