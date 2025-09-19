package Payment

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func ProcessPayment(ctx context.Context, provider string, amount string) (string, error) {
	// simulate async provider; return provider payment id
	if provider == "noop" {
		return uuid.New().String(), nil
	}
	return "", errors.New("provider not implemented")
}