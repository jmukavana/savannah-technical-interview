package Catalog

import (
	"context"
	

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	CreateCategory(ctx context.Context, dto CreateCategoryRequest) (*Category, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*Category, error)
	CreateProduct(ctx context.Context, dto CreateProductRequest) (*Product, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*Product, error)
	ListProducts(ctx context.Context, q ListProductsQuery) ([]Product, error)
}

type service struct {
	repository Repository
	log        *zap.Logger
}



func NewService(r Repository, log *zap.Logger) Service {
	return &service{repository: r, log: log}
}

// CreateCategory implements Service.
func (s *service) CreateCategory(ctx context.Context, dto CreateCategoryRequest) (*Category, error) {
	category:=&Category{
		Name: dto.Name,
		Slug: dto.Slug,
		Description: dto.Description,
		ParentID: dto.ParentID,
	}
	if err:=s.repository.CreateCategory(ctx,category);err!=nil{
		s.log.Error("Create category",zap.Error(err))
		return nil,err
	}
	return category,nil
}

// CreateProduct implements Service.
func (s *service) CreateProduct(ctx context.Context, dto CreateProductRequest) (*Product, error) {
	product:=&Product{
		Name: dto.Name,
		Description: dto.Description,
		CategoryID:dto.CategoryID,
		Price: dto.Price,
		Currency: dto.Currency,
	}
	if err:=s.repository.CreateProduct(ctx,product);err != nil {
		s.log.Error("Create Product")
		return nil, err
	}
	return product,nil
}

// GetCategory implements Service.
func (s *service) GetCategory(ctx context.Context, id uuid.UUID) (*Category, error) {
	return s.repository.GetCategory(ctx,id)
}

// GetProduct implements Service.
func (s *service) GetProduct(ctx context.Context, id uuid.UUID) (*Product, error) {
	return s.repository.GetProduct(ctx,id)
}

// ListProducts implements Service.
func (s *service) ListProducts(ctx context.Context, q ListProductsQuery) ([]Product, error) {
	if q.Limit<=0 || q.Limit>100 {
		q.Limit=20		
	}
	return s.repository.ListProducts(ctx,q)
}
