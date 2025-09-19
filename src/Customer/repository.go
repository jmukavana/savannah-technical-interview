package Customer

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
	Create(ctx context.Context, c *Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*Customer, error)
	List(ctx context.Context, q ListCustomersQuery) ([]Customer, error)
	Update(ctx context.Context, c *Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRepository(db *sqlx.DB, log *zap.Logger) Repository {
	return &repository{db: db, log: log}
}

func (r *repository) Create(ctx context.Context, c *Customer) error {
	c.ID = uuid.New()
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	c.Version = 1

	query := fmt.Sprintf(`INSERT INTO %s (id, first_name, last_name, email, phone, status, created_at, updated_at, version) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, TableName)
	_, err := r.db.ExecContext(ctx, query, c.ID, c.FirstName, c.LastName, c.Email, c.Phone, c.Status, c.CreatedAt, c.UpdatedAt, c.Version)
	return err
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Customer, error) {
	var c Customer
	query := fmt.Sprintf(`SELECT id, first_name, last_name, email, phone, status, created_at, updated_at, version FROM %s WHERE id=$1 AND status <> 'DELETED'`, TableName)
	err := r.db.GetContext(ctx, &c, query, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &c, err
}
func (r *repository) List(ctx context.Context, q ListCustomersQuery) ([]Customer, error) {
	base := fmt.Sprintf(`SELECT id, first_name, last_name, email, phone, status, created_at, updated_at, version FROM %s WHERE status <> 'DELETED'`, TableName)
	args := []interface{}{}
	idx := 1
	if q.Search != "" {
		base += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d)", idx, idx+1, idx+2)
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%", "%"+q.Search+"%")
		idx += 3
	}
	if q.Status != "" {
		base += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, q.Status)
		idx++
	}
	base += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, q.Limit, q.Offset)

	customers := []Customer{}
	err := r.db.SelectContext(ctx, &customers, base, args...)
	return customers, err
}
