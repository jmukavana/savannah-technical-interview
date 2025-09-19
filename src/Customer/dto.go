package Customer

import "github.com/google/uuid"

// client requests to create
type CreateCustomerRequest struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Phone     string `json:"phone" validate:"required,e164"`
}

// update customer
type UpdateCustomerRequest struct {
	FirstName *string `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,min=2,max=100"`
	Email     *string `json:"email" validate:"omitempty,email"`
	Phone     *string `json:"phone" validate:"omitempty,e164"`
	Version   int     `json:"version" validate:"required"`
}

type CustomerResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Status    string    `json:"status"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
}

// ListCustomersQuery supports pagination & filters
type ListCustomersQuery struct {
	Limit  int    `schema:"limit"`
	Offset int    `schema:"offset"`
	Search string `schema:"search"`
	Status string `schema:"status"`
}
