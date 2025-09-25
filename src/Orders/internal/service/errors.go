package Orders

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// Domain-specific errors.
var (
	ErrOrderNotFound        = errors.New("order not found")
	ErrInvalidOrderStatus   = errors.New("invalid order status")
	ErrVersionConflict      = errors.New("version conflict - order was modified by another process")
	ErrInsufficientInventory = errors.New("insufficient inventory")
	ErrInvalidCustomerID    = errors.New("invalid customer ID")
	ErrEmptyOrderItems      = errors.New("order must contain at least one item")
	ErrInvalidQuantity      = errors.New("quantity must be greater than zero")
	ErrInvalidPrice         = errors.New("price must be greater than zero")
	ErrProductNotFound      = errors.New("product not found")
	ErrOrderAlreadyProcessed = errors.New("order has already been processed")
	ErrCannotCancelOrder    = errors.New("order cannot be cancelled in current status")
	ErrInvalidWarehouse     = errors.New("invalid warehouse specified")
)

// OrderError represents a domain-specific error with HTTP status code.
type OrderError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// Error implements the error interface.
func (e OrderError) Error() string {
	return e.Message
}

// NewOrderError creates a new domain-specific error.
func NewOrderError(code, message string, status int) OrderError {
	return OrderError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// ErrOrderNotFoundWithID creates an error for a specific order ID.
func ErrOrderNotFoundWithID(id uuid.UUID) OrderError {
	return NewOrderError("ORDER_NOT_FOUND", fmt.Sprintf("order with ID %s not found", id), http.StatusNotFound)
}

// ErrInvalidOrderStatusTransition creates an error for invalid status transitions.
func ErrInvalidOrderStatusTransition(from, to string) OrderError {
	return NewOrderError("INVALID_STATUS_TRANSITION", fmt.Sprintf("cannot transition from %s to %s", from, to), http.StatusBadRequest)
}