package Orders

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Repository interface {
	CreateOrderTx(ctx context.Context, tx *sqlx.Tx, o *Order, items []OrderItem) error
	GetOrder(ctx context.Context, id uuid.UUID) (*Order, []OrderItem, error)
	UpdateOrderStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status string, version int) error
}

type repository struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRepository(db *sqlx.DB, log *zap.Logger) Repository { return &repository{db: db, log: log} }

func (r *repository) CreateOrderTx(ctx context.Context, tx *sqlx.Tx, o *Order, items []OrderItem) error {
	o.ID = uuid.New()
	now := time.Now().UTC()
	o.CreatedAt = now
	o.UpdatedAt = now
	_, err := tx.ExecContext(ctx, `INSERT INTO orders (id,customer_id,status,subtotal,tax,shipping,total,currency,created_at,updated_at,version) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, o.ID, o.CustomerID, o.Status, o.Subtotal, o.Tax, o.Shipping, o.Total, o.Currency, o.CreatedAt, o.UpdatedAt, o.Version)
	if err != nil {
		return err
	}
	for i := range items {
		items[i].ID = uuid.New()
		items[i].OrderID = o.ID
		if _, err := tx.ExecContext(ctx, `INSERT INTO order_items (id,order_id,product_id,sku,name,unit_price,quantity,line_total) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, items[i].ID, items[i].OrderID, items[i].ProductID, items[i].SKU, items[i].Name, items[i].UnitPrice, items[i].Quantity, items[i].LineTotal); err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) GetOrder(ctx context.Context, id uuid.UUID) (*Order, []OrderItem, error) {
	var o Order
	if err := r.db.GetContext(ctx, &o, `SELECT id,customer_id,status,subtotal,tax,shipping,total,currency,created_at,updated_at,version FROM orders WHERE id=$1`, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, sql.ErrNoRows
		}
		return nil, nil, err
	}
	var items []OrderItem
	if err := r.db.SelectContext(ctx, &items, `SELECT id,order_id,product_id,sku,name,unit_price,quantity,line_total FROM order_items WHERE order_id=$1`, id); err != nil {
		return &o, nil, err
	}
	return &o, items, nil
}

func (r *repository) UpdateOrderStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status string, version int) error {
	res, err := tx.ExecContext(ctx, `UPDATE orders SET status=$1, version=version+1, updated_at=NOW() WHERE id=$2 AND version=$3`, status, id, version)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("version conflict")
	}
	return nil
}
