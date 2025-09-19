package Orders

import (
	"context"
	"errors"
	

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type InventoryService interface {
	Reserve(ctx context.Context, productID uuid.UUID, qty int, warehouse string) error
	Release(ctx context.Context, productID uuid.UUID, qty int, warehouse string) error
}

type service struct {
	repo Repository
	db   *sqlx.DB
	inv  InventoryService
	log  *zap.Logger
}

func NewService(r Repository, db *sqlx.DB, inv InventoryService, log *zap.Logger) *service {
	return &service{repo: r, db: db, inv: inv, log: log}
}

func (s *service) Create(ctx context.Context, customerID *uuid.UUID, items []OrderItem, warehouse string) (*Order, error) {
	// calculate totals
	sub := decimal.NewFromInt(0)
	for i := range items {
		sub = sub.Add(items[i].LineTotal)
	}
	tax := decimal.NewFromFloat(0)
	shipping := decimal.NewFromFloat(0)
	total := sub.Add(tax).Add(shipping)
	order := &Order{CustomerID: customerID, Status: "CREATED", Subtotal: sub, Tax: tax, Shipping: shipping, Total: total, Currency: "USD", Version: 1}

	// begin tx
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// reserve inventory for each item
	for _, it := range items {
		if it.ProductID == nil {
			err = errors.New("product_id required")
			return nil, err
		}
		if perr := s.inv.Reserve(ctx, *it.ProductID, it.Quantity, warehouse); perr != nil {
			s.log.Error("reserve failed", zap.Error(perr))
			err = perr
			return nil, err
		}
	}

	if err = s.repo.CreateOrderTx(ctx, tx, order, items); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*Order, []OrderItem, error) {
	return s.repo.GetOrder(ctx, id)
}

func (s *service) UpdateStatus(ctx context.Context, id uuid.UUID, status string, version int) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if err = s.repo.UpdateOrderStatusTx(ctx, tx, id, status, version); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
