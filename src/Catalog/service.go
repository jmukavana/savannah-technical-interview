package Catalog

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	// Category operations
	CreateCategory(ctx context.Context, dto CreateCategoryRequest) (*CategoryResponse, error)
	GetCategory(ctx context.Context, id uuid.UUID) (*CategoryResponse, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*CategoryResponse, error)
	UpdateCategory(ctx context.Context, id uuid.UUID, dto UpdateCategoryRequest) (*CategoryResponse, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	ListCategories(ctx context.Context, q ListCategoriesQuery) (*PaginatedResponse, error)
	
	// Product operations
	CreateProduct(ctx context.Context, dto CreateProductRequest) (*ProductResponse, error)
	GetProduct(ctx context.Context, id uuid.UUID) (*ProductResponse, error)
	GetProductBySKU(ctx context.Context, sku string) (*ProductResponse, error)
	UpdateProduct(ctx context.Context, id uuid.UUID, dto UpdateProductRequest) (*ProductResponse, error)
	DeleteProduct(ctx context.Context, id uuid.UUID) error
	ListProducts(ctx context.Context, q ListProductsQuery) (*PaginatedResponse, error)
}

type service struct {
	repository Repository
	validator  *validator.Validate
	log        *zap.Logger
}

func NewService(r Repository, log *zap.Logger) Service {
	return &service{
		repository: r,
		validator:  validator.New(),
		log:        log,
	}
}
func (s *service) CreateCategory(ctx context.Context, dto CreateCategoryRequest) (*CategoryResponse, error) {
	if err := s.validator.Struct(dto); err != nil {
		s.log.Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Check for circular reference if parent is specified
	if dto.ParentID != nil {
		if err := s.validateCategoryHierarchy(ctx, *dto.ParentID, nil); err != nil {
			return nil, err
		}
	}

	category := &Category{
		Name:        dto.Name,
		Slug:        dto.Slug,
		Description: dto.Description,
		ParentID:    dto.ParentID,
	}

	if err := s.repository.CreateCategory(ctx, category); err != nil {
		s.log.Error("create category failed", zap.Error(err))
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *service) GetCategory(ctx context.Context, id uuid.UUID) (*CategoryResponse, error) {
	category, err := s.repository.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *service) GetCategoryBySlug(ctx context.Context, slug string) (*CategoryResponse, error) {
	category, err := s.repository.GetCategoryBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *service) UpdateCategory(ctx context.Context, id uuid.UUID, dto UpdateCategoryRequest) (*CategoryResponse, error) {
	if err := s.validator.Struct(dto); err != nil {
		s.log.Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	category, err := s.repository.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if dto.Name != nil {
		category.Name = *dto.Name
	}
	if dto.Slug != nil {
		category.Slug = *dto.Slug
	}
	if dto.Description != nil {
		category.Description = dto.Description
	}
	if dto.ParentID != nil {
		// Check for circular reference
		if err := s.validateCategoryHierarchy(ctx, *dto.ParentID, &id); err != nil {
			return nil, err
		}
		category.ParentID = dto.ParentID
	}

	if err := s.repository.UpdateCategory(ctx, category); err != nil {
		s.log.Error("update category failed", zap.Error(err))
		return nil, err
	}

	response := category.ToResponse()
	return &response, nil
}

func (s *service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteCategory(ctx, id)
}

func (s *service) ListCategories(ctx context.Context, q ListCategoriesQuery) (*PaginatedResponse, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}

	categories, total, err := s.repository.ListCategories(ctx, q)
	if err != nil {
		return nil, err
	}

	responses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToResponse()
	}

	return &PaginatedResponse{
		Data:    responses,
		Total:   total,
		Limit:   q.Limit,
		Offset:  q.Offset,
		HasMore: int64(q.Offset+q.Limit) < total,
	}, nil
}

// Product operations
func (s *service) CreateProduct(ctx context.Context, dto CreateProductRequest) (*ProductResponse, error) {
	if err := s.validator.Struct(dto); err != nil {
		s.log.Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	// Validate category exists if provided
	if dto.CategoryID != nil {
		if _, err := s.repository.GetCategory(ctx, *dto.CategoryID); err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
	}

	product := &Product{
		Name:        dto.Name,
		Description: dto.Description,
		CategoryID:  dto.CategoryID,
		Price:       dto.Price,
		Currency:    dto.Currency,
	}

	if err := s.repository.CreateProduct(ctx, product); err != nil {
		s.log.Error("create product failed", zap.Error(err))
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *service) GetProduct(ctx context.Context, id uuid.UUID) (*ProductResponse, error) {
	product, err := s.repository.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *service) GetProductBySKU(ctx context.Context, sku string) (*ProductResponse, error) {
	product, err := s.repository.GetProductBySKU(ctx, sku)
	if err != nil {
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *service) UpdateProduct(ctx context.Context, id uuid.UUID, dto UpdateProductRequest) (*ProductResponse, error) {
	if err := s.validator.Struct(dto); err != nil {
		s.log.Error("validation failed", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
	}

	product, err := s.repository.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update only provided fields
	if dto.Name != nil {
		product.Name = *dto.Name
	}
	if dto.Description != nil {
		product.Description = dto.Description
	}
	if dto.CategoryID != nil {
		// Validate category exists
		if _, err := s.repository.GetCategory(ctx, *dto.CategoryID); err != nil {
			return nil, fmt.Errorf("invalid category: %w", err)
		}
		product.CategoryID = dto.CategoryID
	}
	if dto.Price != nil {
		product.Price = *dto.Price
	}
	if dto.Currency != nil {
		product.Currency = *dto.Currency
	}

	if err := s.repository.UpdateProduct(ctx, product); err != nil {
		s.log.Error("update product failed", zap.Error(err))
		return nil, err
	}

	response := product.ToResponse()
	return &response, nil
}

func (s *service) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	return s.repository.DeleteProduct(ctx, id)
}

func (s *service) ListProducts(ctx context.Context, q ListProductsQuery) (*PaginatedResponse, error) {
	if q.Limit <= 0 || q.Limit > 100 {
		q.Limit = 20
	}

	products, total, err := s.repository.ListProducts(ctx, q)
	if err != nil {
		return nil, err
	}

	responses := make([]ProductResponse, len(products))
	for i, product := range products {
		responses[i] = product.ToResponse()
	}

	return &PaginatedResponse{
		Data:    responses,
		Total:   total,
		Limit:   q.Limit,
		Offset:  q.Offset,
		HasMore: int64(q.Offset+q.Limit) < total,
	}, nil
}

// validateCategoryHierarchy checks for circular references in category hierarchy
func (s *service) validateCategoryHierarchy(ctx context.Context, parentID uuid.UUID, excludeID *uuid.UUID) error {
	visited := make(map[uuid.UUID]bool)
	current := &parentID

	for current != nil {
		if excludeID != nil && *current == *excludeID {
			return ErrCategoryCircularRef
		}
		
		if visited[*current] {
			return ErrCategoryCircularRef
		}
		
		visited[*current] = true
		
		category, err := s.repository.GetCategory(ctx, *current)
		if err != nil {
			if err == ErrCategoryNotFound {
				break
			}
			return err
		}
		
		current = category.ParentID
	}
	
	return nil
}