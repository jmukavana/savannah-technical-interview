package Catalog

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Category struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Slug        string     `db:"slug" json:"slug"`
	Description *string    `db:"description" json:"description,omitempty"`
	ParentID    *uuid.UUID `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	Version int `db:"version" json:"version"` 
}
const CategoryName = "categories";

type Product struct {
	ID          uuid.UUID       `db:"id" json:"id"`
	SKU         string          `db:"sku" json:"sku"`
	Name        string          `db:"name" json:"name"`
	Description *string         `db:"description" json:"description,omitempty"`
	CategoryID  *uuid.UUID      `db:"category_id" json:"category_id,omitempty"`
	Price       decimal.Decimal `db:"price" json:"price"`
	Currency    string          `db:"currency" json:"currency"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`
	Version int `db:"version" json:"version"` 
}
const ProductName="products"