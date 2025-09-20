package Catalog

import "errors"

// Product related errors
var (
	ProductErrorNotFound     = errors.New("product not found")
	ProductErrorInvalidPayload = errors.New("invalid product payload")
)

// Category related errors
var (
	CategoryErrorNotFound     = errors.New("category not found")
	CategoryErrorInvalidPayload = errors.New("invalid category payload")
)
