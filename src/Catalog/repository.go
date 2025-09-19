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
	ListProducts(ctx context.Context, limit, offset int, search string) ([]Product, error)
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
	c.CreatedAt = time.Now().Local().UTC()
	c.UpdatedAt = c.CreatedAt
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO categories (id,name,slug,description,parent_id,created_at,updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		c.ID, c.Name, c.Slug, c.Description, c.ParentID, c.CreatedAt, c.UpdatedAt)
	return err
}

// CreateProduct implements Repository.
func (r *repository) CreateProduct(ctx context.Context, p *Product) error {
	p.ID = uuid.New()
	p.CreatedAt = time.Now().Local().UTC()
	p.UpdatedAt = p.CreatedAt
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO products (id,sku,name,description,category_id,price,currency,created_at,updated)
		VALUES($1,$2,$3,$4,$5,$6,$7,&8,&9)`,
		p.ID, p.SKU, p.Name, p.Description, p.CategoryID, p.Price, p.Currency, p.CreatedAt, p.UpdatedAt)

	return err
}

// GetCategory implements Repository.ows }

func (r *repository) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	var category Category
	err := r.db.GetContext(
		ctx,
		&category,
		`SELECT id,name,slug,description,parent_id,created_at,updated_at 
		FROM categories WHERE id=$1`, id)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return &category, err
}

// GetProduct implements Repository.
func (r *repository) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	var product Product
	err := r.db.GetContext(
		ctx,
		&product,
		`SELECT id,sku,name,description,category_id,price,currency,created_at,updated_at
		FROM products WHERE id=$1`, id)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return &product, err
}

// ListProducts implements Repository.
func (r *repository) ListProducts(ctx context.Context, limit, offset int, search string) ([]Product, error) {
	query := `SELECT id,sku,name,description,category_id,price,currency,created_at,updated_at FROM products WHERE 1=1`
	args := []interface{}{}
	idx := 1
	if search != "" {
		query += ` AND (name ILIKE $` + itoa(idx) + ` OR sku ILIKE $` + itoa(idx+1) + ` )`
		args = append(args, "%"+search+"%", "%"+search+"%")
		idx += 2
	}
	query += ` ORDER BY created_at DESC LIMIT $` + itoa(idx) + ` OFFSET $` + itoa(idx+1)
	args = append(args, limit, offset)
	var res []Product
	if err := r.db.SelectContext(ctx, &res, query, args...); err != nil {
		return nil, err
	}
	return res, nil
}

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
