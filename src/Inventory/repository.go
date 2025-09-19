package Inventory

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	GetByProductAndWarehouse(ctx context.Context, productID uuid.UUID, warehouse string) (*Inventory, error)
	UpsertInventory(ctx context.Context, inv *Inventory) error
	AdjustInventory(ctx context.Context, inventoryID uuid.UUID, change int, reason, reference string) error
}

type repository struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRepository(db *sqlx.DB, log *zap.Logger) Repository { return &repository{db: db, log: log} }

func (r *repository) GetByProductAndWarehouse(ctx context.Context, productID uuid.UUID, warehouse string) (*Inventory, error) {
	var inv Inventory
	if err := r.db.GetContext(ctx, &inv, `SELECT id,product_id,warehouse,quantity,reserved,created_at,updated_at FROM inventory WHERE product_id=$1 AND warehouse=$2`, productID, warehouse); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &inv, nil
}

func (r *repository) UpsertInventory(ctx context.Context, inv *Inventory) error {
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
		inv.CreatedAt = time.Now().UTC()
	}
	inv.UpdatedAt = time.Now().UTC()
	_, err := r.db.ExecContext(ctx, `INSERT INTO inventory (id,product_id,warehouse,quantity,reserved,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT (product_id,warehouse) DO UPDATE SET quantity=EXCLUDED.quantity, reserved=EXCLUDED.reserved, updated_at=EXCLUDED.updated_at`, inv.ID, inv.ProductID, inv.Warehouse, inv.Quantity, inv.Reserved, inv.CreatedAt, inv.UpdatedAt)
	return err
}

func (r *repository) AdjustInventory(ctx context.Context, inventoryID uuid.UUID, change int, reason, reference string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE inventory SET quantity = quantity + $1, updated_at = NOW() WHERE id=$2`, change, inventoryID)
	if err != nil {
		return err
	}
	st := &StockTransaction{ID: uuid.New(), InventoryID: inventoryID, Change: change, Reason: reason, Reference: &reference, CreatedAt: time.Now().UTC()}
	_, err = r.db.NamedExecContext(ctx, `INSERT INTO stock_transactions (id,inventory_id,change,reason,reference,created_at) VALUES (:id,:inventory_id,:change,:reason,:reference,:created_at)`, st)
	return err
}
