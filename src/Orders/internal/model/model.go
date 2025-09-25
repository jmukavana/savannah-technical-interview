package Orders

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Order represents the main order entity.
type Order struct {
	ID         uuid.UUID          `db:"id" json:"id"`
	CustomerID *uuid.UUID         `db:"customer_id" json:"customer_id,omitempty"`
	Status     string             `db:"status" json:"status"`
	Subtotal   decimal.Decimal    `db:"subtotal" json:"subtotal"`
	Tax        decimal.Decimal    `db:"tax" json:"tax"`
	Shipping   decimal.Decimal    `db:"shipping" json:"shipping"`
	Total      decimal.Decimal    `db:"total" json:"total"`
	Currency   string             `db:"currency" json:"currency"`
	CreatedAt  time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `db:"updated_at" json:"updated_at"`
	DeletedAt  *time.Time         `db:"deleted_at" json:"deleted_at,omitempty"`
	Version    int                `db:"version" json:"version"`
	Metadata   map[string]string  `db:"metadata" json:"metadata,omitempty"`
	Warehouse  string             `db:"warehouse" json:"warehouse"`
	Addresses  []OrderAddress     `db:"-" json:"addresses,omitempty"`
}

// OrderItem represents an individual item within an order.
type OrderItem struct {
	ID        uuid.UUID       `db:"id" json:"id"`
	OrderID   uuid.UUID       `db:"order_id" json:"order_id"`
	ProductID *uuid.UUID      `db:"product_id" json:"product_id,omitempty"`
	SKU       *string         `db:"sku" json:"sku,omitempty"`
	Name      *string         `db:"name" json:"name,omitempty"`
	UnitPrice decimal.Decimal `db:"unit_price" json:"unit_price"`
	Quantity  int             `db:"quantity" json:"quantity"`
	LineTotal decimal.Decimal `db:"line_total" json:"line_total"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
	Version   int             `db:"version" json:"version"`
}

// OrderAddress represents shipping or billing addresses.
type OrderAddress struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	OrderID     uuid.UUID  `db:"order_id" json:"order_id"`
	AddressType string     `db:"address_type" json:"address_type"` // shipping, billing
	FirstName   string     `db:"first_name" json:"first_name"`
	LastName    string     `db:"last_name" json:"last_name"`
	Company     *string    `db:"company" json:"company,omitempty"`
	Address1    string     `db:"address1" json:"address1"`
	Address2    *string    `db:"address2" json:"address2,omitempty"`
	City        string     `db:"city" json:"city"`
	State       string     `db:"state" json:"state"`
	PostalCode  string     `db:"postal_code" json:"postal_code"`
	Country     string     `db:"country" json:"country"`
	Phone       *string    `db:"phone" json:"phone,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// OrderEvent represents order status changes and important events.
type OrderEvent struct {
	ID        uuid.UUID              `db:"id" json:"id"`
	OrderID   uuid.UUID              `db:"order_id" json:"order_id"`
	EventType string                 `db:"event_type" json:"event_type"`
	Data      map[string]interface{} `db:"data" json:"data"`
	CreatedAt time.Time              `db:"created_at" json:"created_at"`
	CreatedBy *uuid.UUID             `db:"created_by" json:"created_by,omitempty"`
}

// Valid status constants.
const (
	StatusCreated    = "CREATED"
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusShipped    = "SHIPPED"
	StatusDelivered  = "DELIVERED"
	StatusCancelled  = "CANCELLED"
	StatusRefunded   = "REFUNDED"
)

// ValidStatusTransitions defines allowed state transitions.
var ValidStatusTransitions = map[string][]string{
	StatusCreated:    {StatusPending, StatusCancelled},
	StatusPending:    {StatusProcessing, StatusCancelled},
	StatusProcessing: {StatusShipped, StatusCancelled},
	StatusShipped:    {StatusDelivered, StatusRefunded},
	StatusDelivered:  {StatusRefunded},
	StatusCancelled:  {},
	StatusRefunded:   {},
}

// IsValidStatus checks if the status is valid.
func (o *Order) IsValidStatus(status string) bool {
	validStatuses := []string{StatusCreated, StatusPending, StatusProcessing, StatusShipped, StatusDelivered, StatusCancelled, StatusRefunded}
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// CanTransitionTo checks if the order can transition to the new status.
func (o *Order) CanTransitionTo(newStatus string) bool {
	transitions, exists := ValidStatusTransitions[o.Status]
	if !exists {
		return false
	}
	for _, validTransition := range transitions {
		if validTransition == newStatus {
			return true
		}
	}
	return false
}

// IsCancellable checks if the order can be cancelled.
func (o *Order) IsCancellable() bool {
	return o.Status == StatusCreated || o.Status == StatusPending || o.Status == StatusProcessing
}

// IsRefundable checks if the order can be refunded.
func (o *Order) IsRefundable() bool {
	return o.Status == StatusDelivered || o.Status == StatusShipped
}

// CalculateTotal recalculates the order total.
func (o *Order) CalculateTotal() decimal.Decimal {
	return o.Subtotal.Add(o.Tax).Add(o.Shipping)
}

// CalculateLineTotal calculates the total for an order item.
func (oi *OrderItem) CalculateLineTotal() decimal.Decimal {
	return oi.UnitPrice.Mul(decimal.NewFromInt(int64(oi.Quantity)))
}

// HasCustomer checks if the order has a customer ID.
func (o *Order) HasCustomer() bool {
	return o.CustomerID != nil
}

// IsGuest checks if the order is a guest order.
func (o *Order) IsGuest() bool {
	return o.CustomerID == nil
}

// GetStatusDisplayName returns a human-readable status name.
func (o *Order) GetStatusDisplayName() string {
	statusNames := map[string]string{
		StatusCreated:    "Order Created",
		StatusPending:    "Payment Pending",
		StatusProcessing: "Processing",
		StatusShipped:    "Shipped",
		StatusDelivered:  "Delivered",
		StatusCancelled:  "Cancelled",
		StatusRefunded:   "Refunded",
	}
	if displayName, exists := statusNames[o.Status]; exists {
		return displayName
	}
	return o.Status
}