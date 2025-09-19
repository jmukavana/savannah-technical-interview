package Inventory

import (
	"time"

	"github.com/google/uuid"
)

type Inventory struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ProductID uuid.UUID `db:"product_id" json:"product_id"`
	Warehouse string    `db:"warehouse" json:"warehouse"`
	Quantity  int       `db:"quantity" json:"quantity"`
	Reserved  int       `db:"reserved" json:"reserved"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type StockTransaction struct {
	ID          uuid.UUID `db:"id" json:"id"`
	InventoryID uuid.UUID `db:"inventory_id" json:"inventory_id"`
	Change      int       `db:"change" json:"change"`
	Reason      string    `db:"reason" json:"reason"`
	Reference   *string   `db:"reference" json:"reference,omitempty"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
