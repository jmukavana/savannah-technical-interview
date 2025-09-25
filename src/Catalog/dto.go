package Catalog

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateProductRequest struct {
	Name        string          `json:"name" validate:"required,min=2,max=100"`
	Description *string         `json:"description,omitempty"`
	CategoryID  *uuid.UUID      `json:"category_id" validate:"required"`
	Price       decimal.Decimal `json:"price" validate:"required,gt=0"`
	Currency    string          `json:"currency" validate:"required,len=3"`
}

type UpdateProductRequest struct {
	Name        *string         `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string         `json:"description,omitempty"`
	CategoryID  *uuid.UUID      `json:"category_id,omitempty"`
	Price       *decimal.Decimal `json:"price,omitempty" validate:"omitempty,gt=0"`
	Currency    *string         `json:"currency,omitempty" validate:"omitempty,len=3"`
}

type CreateCategoryRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	Slug        string     `json:"slug" validate:"required,min=2,max=100,alphanum"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

type UpdateCategoryRequest struct {
	Name        *string    `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Slug        *string    `json:"slug,omitempty" validate:"omitempty,min=2,max=100,alphanum"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

type ProductResponse struct {
	ID          uuid.UUID       `json:"id"`
	SKU         string          `json:"sku"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	CategoryID  *uuid.UUID      `json:"category_id"`
	Price       decimal.Decimal `json:"price"`
	Currency    string          `json:"currency"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
	Version     int             `json:"version"`
}

type CategoryResponse struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
	Version     int        `json:"version"`
}

type ListProductsQuery struct {
	Limit      int         `schema:"limit"`
	Offset     int         `schema:"offset"`
	Search     string      `schema:"search"`
	CategoryID *uuid.UUID  `schema:"category_id"`
	MinPrice   *decimal.Decimal `schema:"min_price"`
	MaxPrice   *decimal.Decimal `schema:"max_price"`
}

type ListCategoriesQuery struct {
	Limit    int    `schema:"limit"`
	Offset   int    `schema:"offset"`
	Search   string `schema:"search"`
	ParentID *uuid.UUID `schema:"parent_id"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Limit      int         `json:"limit"`
	Offset     int         `json:"offset"`
	HasMore    bool        `json:"has_more"`
}
