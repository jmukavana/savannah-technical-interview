package Orders

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/jmoiron/sqlx"

)

// Service defines the order service interface.
type Service interface {
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*model.OrderResponse, error)
	GetOrderByID(ctx context.Context, id uuid.UUID) (*model.OrderResponse, error)
	ListOrders(ctx context.Context, req *model.OrderListRequest) (*model.OrderListResponse, error)
	UpdateOrderStatus(ctx context.Context, id uuid.UUID, status string, version int) error
	CancelOrder(ctx context.Context, id uuid.UUID, version int, reason string) error
	RefundOrder(ctx context.Context, id uuid.UUID, version int, reason string) error
	GetOrderItems(ctx context.Context, orderID uuid.UUID) ([]model.OrderItemResponse, error)
	ProcessOrder(ctx context.Context, id uuid.UUID) error
	GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]model.OrderResponse, error)
	GetOrdersByStatus(ctx context.Context, status string, limit, offset int) ([]model.OrderResponse, error)
	GetOrderStatistics(ctx context.Context, start, end time.Time) (*model.OrderStatistics, error)
	BulkUpdateOrderStatus(ctx context.Context, orderIDs []uuid.UUID, status string) error
}

// Repository defines the data access layer interface.
type Repository interface {
	CreateOrderTx(ctx context.Context, tx *sqlx.Tx, o *model.Order, items []model.OrderItem, addresses []model.OrderAddress) error
	GetOrder(ctx context.Context, id uuid.UUID) (*model.Order, []model.OrderItem, []model.OrderAddress, error)
	ListOrders(ctx context.Context, req *model.OrderListRequest) ([]model.Order, int, error)
	UpdateOrderStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status string, version int) error
	GetOrdersByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]model.Order, error)
	GetOrdersByStatus(ctx context.Context, status string, limit, offset int) ([]model.Order, error)
	GetOrderStatistics(ctx context.Context, start, end time.Time) (*model.OrderStatistics, error)
	BulkUpdateOrderStatus(ctx context.Context, orderIDs []uuid.UUID, status string) error
	SoftDeleteOrderTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) error
}

// InventoryService defines inventory management operations.
type InventoryService interface {
	Reserve(ctx context.Context, productID uuid.UUID, quantity int, warehouse string) error
	Release(ctx context.Context, productID uuid.UUID, quantity int, warehouse string) error
}

// TaxCalculator calculates taxes for an order.
type TaxCalculator interface {
	Calculate(ctx context.Context, subtotal decimal.Decimal, customerID *uuid.UUID) (decimal.Decimal, error)
}

// ShippingCalculator calculates shipping costs for an order.
type ShippingCalculator interface {
	Calculate(ctx context.Context, items []model.CreateOrderItem, customerID *uuid.UUID) (decimal.Decimal, error)
}

// NotificationService handles order-related notifications.
type NotificationService interface {
	SendOrderConfirmation(ctx context.Context, order *model.Order) error
	SendOrderStatusUpdate(ctx context.Context, order *model.Order, previousStatus string) error
	SendOrderCancellation(ctx context.Context, order *model.Order, reason string) error
}

// PaymentService handles payment processing and refunds.
type PaymentService interface {
	ProcessPayment(ctx context.Context, orderID uuid.UUID, amount decimal.Decimal) error
	RefundPayment(ctx context.Context, orderID uuid.UUID, amount decimal.Decimal) error
}

// AuditService logs order-related events.
type AuditService interface {
	LogOrderEvent(ctx context.Context, orderID uuid.UUID, event string, data map[string]interface{}) error
}