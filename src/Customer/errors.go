package Customer

import "errors"

var (
	ErrorNotFound       = errors.New("customer not found")
	ErrorConflict       = errors.New("customer already exist")
	ErrorInvalidPayload = errors.New("invalid payload")
)
