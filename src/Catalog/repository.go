package Catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)
type Repository interface {
	// Category operations
	CreateCategory(ctx context.Context, c *Category) error
	GetCategory(ctx context.Context, id uuid.UUID) (*Category, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	UpdateCategory(ctx context.Context, c *Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	ListCategories(ctx context.Context, q ListCategoriesQuery) ([]Category, int64, error)
	
	// Product operations
	CreateProduct(ctx context.Context, p *Product) error
	GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*Product, error)
	UpdateProduct(ctx context.Context, p *Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	ListProducts(ctx context.Context, q ListProductsQuery) ([]Product, int64, error)
}

type repository struct {
	db  *sqlx.DB
	log *zap.Logger
}



func NewRepository(db *sqlx.DB, log *zap.Logger) Repository {
	return &repository{db: db, log: log}
}

// Category operations
func (r *repository) CreateCategory(ctx context.Context, c *Category) error {
	c.ID = uuid.New()
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	c.Version = 1

	query := fmt.Sprintf(`
		INSERT INTO %s (id, name, slug, description, parent_id, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		CategoryTableName)

	_, err := r.db.ExecContext(ctx, query,
		c.ID, c.Name, c.Slug, c.Description, c.ParentID, c.CreatedAt, c.UpdatedAt, c.Version)
	
	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return ErrCategorySlugExists
	}
	return err
}

func (r *repository) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	var category Category
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, parent_id, created_at, updated_at, version 
		FROM %s WHERE id = $1`, CategoryTableName)

	err := r.db.GetContext(ctx, &category, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrCategoryNotFound
	}
	return &category, err
}

func (r *repository) GetCategoryBySlug(ctx context.Context, slug string) (*Category, error) {
	var category Category
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, parent_id, created_at, updated_at, version 
		FROM %s WHERE slug = $1`, CategoryTableName)

	err := r.db.GetContext(ctx, &category, query, slug)
	if err == sql.ErrNoRows {
		return nil, ErrCategoryNotFound
	}
	return &category, err
}

func (r *repository) UpdateCategory(ctx context.Context, c *Category) error {
	c.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`
		UPDATE %s 
		SET name = $1, slug = $2, description = $3, parent_id = $4, updated_at = $5, version = version + 1
		WHERE id = $6 AND version = $7`, CategoryTableName)

	res, err := r.db.ExecContext(ctx, query,
		c.Name, c.Slug, c.Description, c.ParentID, c.UpdatedAt, c.ID, c.Version)
	
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return ErrCategorySlugExists
		}
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrCategoryConflict
	}
	return nil
}

func (r *repository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, CategoryTableName)
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrCategoryNotFound
	}
	return nil
}

func (r *repository) ListCategories(ctx context.Context, q ListCategoriesQuery) ([]Category, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR slug ILIKE $%d)", idx, idx+1))
		searchTerm := "%" + q.Search + "%"
		args = append(args, searchTerm, searchTerm)
		idx += 2
	}

	if q.ParentID != nil {
		conditions = append(conditions, fmt.Sprintf("parent_id = $%d", idx))
		args = append(args, *q.ParentID)
		idx++
	}

	whereClause := strings.Join(conditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, CategoryTableName, whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT id, name, slug, description, parent_id, created_at, updated_at, version
		FROM %s WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, CategoryTableName, whereClause, idx, idx+1)
	
	args = append(args, q.Limit, q.Offset)

	categories := []Category{}
	err = r.db.SelectContext(ctx, &categories, dataQuery, args...)
	return categories, total, err
}

// Product operations
func (r *repository) CreateProduct(ctx context.Context, p *Product) error {
	p.ID = uuid.New()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	p.Version = 1
	
	// Generate SKU if not provided
	if p.SKU == "" {
		p.SKU = generateSKU(p.Name)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (id, sku, name, description, category_id, price, currency, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`, ProductTableName)

	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.SKU, p.Name, p.Description, p.CategoryID,
		p.Price, p.Currency, p.CreatedAt, p.UpdatedAt, p.Version)

	if err != nil && strings.Contains(err.Error(), "duplicate") {
		return ErrProductSKUExists
	}
	return err
}

func (r *repository) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	var product Product
	query := fmt.Sprintf(`
		SELECT id, sku, name, description, category_id, price, currency, created_at, updated_at, version
		FROM %s WHERE id = $1`, ProductTableName)

	err := r.db.GetContext(ctx, &product, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

func (r *repository) GetProductBySKU(ctx context.Context, sku string) (*Product, error) {
	var product Product
	query := fmt.Sprintf(`
		SELECT id, sku, name, description, category_id, price, currency, created_at, updated_at, version
		FROM %s WHERE sku = $1`, ProductTableName)

	err := r.db.GetContext(ctx, &product, query, sku)
	if err == sql.ErrNoRows {
		return nil, ErrProductNotFound
	}
	return &product, err
}

func (r *repository) UpdateProduct(ctx context.Context, p *Product) error {
	p.UpdatedAt = time.Now().UTC()
	query := fmt.Sprintf(`
		UPDATE %s 
		SET sku = $1, name = $2, description = $3, category_id = $4, price = $5, currency = $6, updated_at = $7, version = version + 1
		WHERE id = $8 AND version = $9`, ProductTableName)

	res, err := r.db.ExecContext(ctx, query,
		p.SKU, p.Name, p.Description, p.CategoryID,
		p.Price, p.Currency, p.UpdatedAt, p.ID, p.Version)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return ErrProductSKUExists
		}
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrProductConflict
	}
	return nil
}

func (r *repository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, ProductTableName)
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *repository) ListProducts(ctx context.Context, q ListProductsQuery) ([]Product, int64, error) {
	conditions := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if q.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR sku ILIKE $%d)", idx, idx+1))
		searchTerm := "%" + q.Search + "%"
		args = append(args, searchTerm, searchTerm)
		idx += 2
	}

	if q.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", idx))
		args = append(args, *q.CategoryID)
		idx++
	}

	if q.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", idx))
		args = append(args, *q.MinPrice)
		idx++
	}

	if q.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", idx))
		args = append(args, *q.MaxPrice)
		idx++
	}

	whereClause := strings.Join(conditions, " AND ")
	
	// Count query
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, ProductTableName, whereClause)
	var total int64
	err := r.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	dataQuery := fmt.Sprintf(`
		SELECT id, sku, name, description, category_id, price, currency, created_at, updated_at, version
		FROM %s WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, ProductTableName, whereClause, idx, idx+1)
	
	args = append(args, q.Limit, q.Offset)

	products := []Product{}
	err = r.db.SelectContext(ctx, &products, dataQuery, args...)
	return products, total, err
}

// generateSKU creates a simple SKU from product name
func generateSKU(name string) string {
	// Simple SKU generation - in production you might want something more sophisticated
	cleaned := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	if len(cleaned) > 8 {
		cleaned = cleaned[:8]
	}
	return fmt.Sprintf("%s-%d", cleaned, time.Now().Unix()%10000)
}
