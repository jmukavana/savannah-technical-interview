package Customer

import "errors"

var (
	ErrorNotFound       = errors.New("Customer Not Found!")
	ErrorConflict       = errors.New("Customer Already Exist!")
	ErrorInvalidPayload = errors.New("Invalid Paylod!")
)
