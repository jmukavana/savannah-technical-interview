package Customer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	Create(ctx context.Context, dto CreateCustomerRequest) (*Customer, error)
	Get(ctx context.Context, id uuid.UUID) (*Customer, error)
	List(ctx context.Context, q ListCustomersQuery) ([]Customer, error)
	Update(ctx context.Context, id uuid.UUID, dto UpdateCustomerRequest) (*Customer, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
	log  *zap.Logger
}

func NewService(r Repository, log *zap.Logger) Service {
	return &service{repo: r, log: log}
}

func (s *service) Create(ctx context.Context, dto CreateCustomerRequest) (*Customer, error) {
	c := &Customer{
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Email:     dto.Email,
		Phone:     dto.Phone,
		Status:    "ACTIVE",
	}
	if err := s.repo.Create(ctx, c); err != nil {
		s.log.Error("create customer", zap.Error(err))
		return nil, err
	}
	return c, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*Customer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *service) List(ctx context.Context, q ListCustomersQuery) ([]Customer, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}
	return s.repo.List(ctx, q)
}
func (s *service) Update(ctx context.Context, id uuid.UUID, dto UpdateCustomerRequest) (*Customer, error) {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// optimistic lock check
	if dto.Version != c.Version {
		return nil, ErrorConflict
	}
	if dto.FirstName != nil {
		c.FirstName = *dto.FirstName
	}
	if dto.LastName != nil {
		c.LastName = *dto.LastName
	}
	if dto.Email != nil {
		c.Email = *dto.Email
	}
	if dto.Phone != nil {
		c.Phone = *dto.Phone
	}
	c.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	// return updated record
	return s.repo.GetByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
