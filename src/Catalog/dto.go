package Catalog

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateProductRequest struct {
	Name        string          `json:"name" validate:"required,min=2,max=100"`
	Description *string         `json:"description,omitempty"`
	CategoryID  *uuid.UUID      `json:"category_id" validate:"required"`
	Price       decimal.Decimal `json:"price" validate:"required"`
	Currency    string          `json:"currency"`
}
type CreateCategoryRequest struct{
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	Slug        string     `json:"slug" validate:"required,min=2,max=100"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
}

type ProductResponse struct{
	ID uuid.UUID `json:"id"`
	Name        string          `json:"name"`
	Description *string         `json:"description"`
	CategoryID  *uuid.UUID      `json:"category_id"`
	Price       decimal.Decimal `json:"price"`
	Currency    string          `json:"currency"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type ListProductsQuery struct {
	Limit  int    `schema:"limit"`
	Offset int    `schema:"offset"`
	Search string `schema:"search"`	
}
