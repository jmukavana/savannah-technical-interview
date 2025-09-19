package Inventory

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type Service interface {
	Reserve(ctx context.Context, productID uuid.UUID, qty int, warehouse string) error
	Release(ctx context.Context, productID uuid.UUID, qty int, warehouse string) error
	GetAvailable(ctx context.Context, productID uuid.UUID, warehouse string) (int, error)
}

type service struct {
	repo Repository
	db   *sqlx.DB
	log  *zap.Logger
}

func NewService(r Repository, db *sqlx.DB, log *zap.Logger) Service {
	return &service{repo: r, db: db, log: log}
}

func (s *service) Reserve(ctx context.Context, productID uuid.UUID, qty int, warehouse string) error {
	// simple strategy: single inventory row per product+warehouse; use transaction + row lock
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var inv Inventory
	if err := tx.GetContext(ctx, &inv, `SELECT id,product_id,warehouse,quantity,reserved FROM inventory WHERE product_id=$1 AND warehouse=$2 FOR UPDATE`, productID, warehouse); err != nil {
		return err
	}
	available := inv.Quantity - inv.Reserved
	if available < qty {
		return errors.New("insufficient stock")
	}
	inv.Reserved += qty
	inv.UpdatedAt = time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE inventory SET reserved=$1, updated_at=$2 WHERE id=$3`, inv.Reserved, inv.UpdatedAt, inv.ID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO stock_transactions (id,inventory_id,change,reason,created_at) VALUES ($1,$2,$3,$4,$5)`, uuid.New(), inv.ID, -qty, "reserve", time.Now().UTC()); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *service) Release(ctx context.Context, productID uuid.UUID, qty int, warehouse string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var inv Inventory
	if err := tx.GetContext(ctx, &inv, `SELECT id,product_id,warehouse,quantity,reserved FROM inventory WHERE product_id=$1 AND warehouse=$2 FOR UPDATE`, productID, warehouse); err != nil {
		return err
	}
	if inv.Reserved < qty {
		return errors.New("release quantity exceeds reserved")
	}
	inv.Reserved -= qty
	inv.UpdatedAt = time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `UPDATE inventory SET reserved=$1, updated_at=$2 WHERE id=$3`, inv.Reserved, inv.UpdatedAt, inv.ID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO stock_transactions (id,inventory_id,change,reason,created_at) VALUES ($1,$2,$3,$4,$5)`, uuid.New(), inv.ID, qty, "release", time.Now().UTC()); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *service) GetAvailable(ctx context.Context, productID uuid.UUID, warehouse string) (int, error) {
	inv, err := s.repo.GetByProductAndWarehouse(ctx, productID, warehouse)
	if err != nil {
		return 0, err
	}
	return inv.Quantity - inv.Reserved, nil
}
