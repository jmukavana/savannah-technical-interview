package Customer

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID uuid.UUID `db:"id" json:"id"`
	FirstName string `db:"first_name" json:"first_name"`
	LastName string `db:"last_name" json:"last_name"`
	Email string `db:"email" json:"email"`
	Phone string `db:"phone" json:"phone"`
	Status string `db:"status" json:"status"` // ACTIVE, SUSPENDED, DELETED
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	Version int `db:"version" json:"version" // optimistic locking
}
const TableName = "customers"
