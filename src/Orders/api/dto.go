package Orders

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)
// CreateOrderRequest represents the request to create a new order
type CreateOrderRequest struct {
	CustomerID *uuid.UUID         `json:"customer_id,omitempty" validate:"omitempty,uuid"`
	Items      []CreateOrderItem  `json:"items" validate:"required,min=1,dive"`
	Warehouse  string             `json:"warehouse" validate:"required"`
	Currency   string             `json:"currency" validate:"required,len=3"`
	Metadata   map[string]string  `json:"metadata,omitempty"`
}

// CreateOrderItem represents an item in the order creation request
type CreateOrderItem struct {
	ProductID *uuid.UUID      `json:"product_id,omitempty" validate:"omitempty,uuid"`
	SKU       *string         `json:"sku,omitempty" validate:"omitempty"`
	Name      *string         `json:"name,omitempty" validate:"omitempty"`
	UnitPrice decimal.Decimal `json:"unit_price" validate:"required,gt=0"`
	Quantity  int             `json:"quantity" validate:"required,gt=0"`
}

// OrderResponse represents the response when returning order data
type OrderResponse struct {
	ID         uuid.UUID         `json:"id"`
	CustomerID *uuid.UUID        `json:"customer_id,omitempty"`
	Status     string            `json:"status"`
	Subtotal   decimal.Decimal   `json:"subtotal"`
	Tax        decimal.Decimal   `json:"tax"`
	Shipping   decimal.Decimal   `json:"shipping"`
	Total      decimal.Decimal   `json:"total"`
	Currency   string            `json:"currency"`
	Items      []OrderItemResponse `json:"items"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Version    int               `json:"version"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// OrderItemResponse represents an order item in responses
type OrderItemResponse struct {
	ID        uuid.UUID       `json:"id"`
	ProductID *uuid.UUID      `json:"product_id,omitempty"`
	SKU       *string         `json:"sku,omitempty"`
	Name      *string         `json:"name,omitempty"`
	UnitPrice decimal.Decimal `json:"unit_price"`
	Quantity  int             `json:"quantity"`
	LineTotal decimal.Decimal `json:"line_total"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status  string `json:"status" validate:"required"`
	Version int    `json:"version" validate:"required,gte=1"`
}

// OrderListRequest represents the request for listing orders
type OrderListRequest struct {
	CustomerID *uuid.UUID `form:"customer_id" validate:"omitempty,uuid"`
	Status     string     `form:"status" validate:"omitempty"`
	Limit      int        `form:"limit" validate:"omitempty,min=1,max=100"`
	Offset     int        `form:"offset" validate:"omitempty,min=0"`
	SortBy     string     `form:"sort_by" validate:"omitempty,oneof=created_at updated_at total"`
	SortDir    string     `form:"sort_dir" validate:"omitempty,oneof=asc desc"`
}

// OrderListResponse represents the response for order listing
type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// CancelOrderRequest represents the request to cancel an order
type CancelOrderRequest struct {
	Version int    `json:"version" validate:"required,gte=1"`
	Reason  string `json:"reason" validate:"required"`
}