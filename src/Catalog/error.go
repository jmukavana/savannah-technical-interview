package Catalog

import "errors"

// Product related errors
var (
	ErrProductNotFound      = errors.New("product not found")
	ErrProductInvalidPayload = errors.New("invalid product payload")
	ErrProductConflict      = errors.New("product already exists")
	ErrProductSKUExists     = errors.New("product SKU already exists")
)

// Category related errors
var (
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryInvalidPayload = errors.New("invalid category payload")
	ErrCategoryConflict      = errors.New("category already exists")
	ErrCategorySlugExists    = errors.New("category slug already exists")
	ErrCategoryCircularRef   = errors.New("circular reference detected in category hierarchy")
)
// Validation errors
var (
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidUUID      = errors.New("invalid UUID format")
)

