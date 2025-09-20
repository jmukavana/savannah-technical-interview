package Catalog

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
	CreateCategory(ctx context.Context, c *Category) error
	GetCategory(ctx context.Context, id uuid.UUID) (*Category, error)

	CreateProduct(ctx context.Context, p *Product) error

	GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)
	ListProducts(ctx context.Context, q ListProductsQuery) ([]Product, error)
}

type repository struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRepository(db *sqlx.DB, log *zap.Logger) Repository {
	return &repository{db: db, log: log}
}

func (r *repository) CreateCategory(ctx context.Context, c *Category) error {
	c.ID = uuid.New()
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	c.Version = 1
	query := fmt.Sprintf(
		`INSERT INTO %s (id,name,slug,description,parent_id,created_at,updated_at,version)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, CategoryName)
	_, err := r.db.ExecContext(
		ctx,
		query,
		c.ID, c.Name, c.Slug, c.Description, c.ParentID, c.CreatedAt, c.UpdatedAt, c.Version)
	return err
}

// CreateProduct implements Repository.
func (r *repository) CreateProduct(ctx context.Context, p *Product) error {
	p.ID = uuid.New()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	p.Version = 1

	query := fmt.Sprintf(`
	INSERT INTO %s 
	(id, sku, name, description, category_id, price, currency, created_at, updated_at, version)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, ProductName)

	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.SKU, p.Name, p.Description, p.CategoryID,
		p.Price, p.Currency, p.CreatedAt, p.UpdatedAt, p.Version,
	)
	return err
}

// GetCategory implements Repository.ows }

func (r *repository) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	var category Category
	query := fmt.Sprintf(`SELECT id,name,slug,description,parent_id,created_at,updated_at,version 
		FROM %s WHERE id=$1`, CategoryName)

	err := r.db.GetContext(
		ctx,
		&category,
		query, id)
	if err == sql.ErrNoRows {
		return nil, CategoryErrorNotFound
	}
	return &category, err
}

// GetProduct implements Repository.
func (r *repository) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	var product Product
	query := fmt.Sprintf(`SELECT id,sku,name,description,category_id,price,currency,created_at,updated_at,version
		FROM %s WHERE id=$1`, ProductName)
	err := r.db.GetContext(
		ctx,
		&product,
		query, id)
	if err == sql.ErrNoRows {
		return nil, ProductErrorNotFound
	}
	return &product, err
}

// ListProducts implements Repository.
func (r *repository) ListProducts(ctx context.Context, q ListProductsQuery) ([]Product, error) {
	base := fmt.Sprintf(`SELECT id,sku,name,description,category_id,price,currency,created_at,updated_at,version FROM %s WHERE 1=1`, ProductName)
	args := []interface{}{}
	idx := 1
	if q.Search != "" {
		base += fmt.Sprintf(" AND (name ILIKE $%d OR sku ILIKE $%d)", idx, idx+1)
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
		idx += 2
	}
	base += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, q.Limit, q.Offset)

	products := []Product{}
	err := r.db.SelectContext(ctx, &products, base, args...)
	return products, err
}
