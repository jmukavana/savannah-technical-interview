package Billing

import (
	"context"

	"github.com/google/uuid"
)

type NoopProvider struct{}

func (n *NoopProvider) Charge(ctx context.Context, provider string, amount decimal.Decimal, currency string, metadata map[string]interface{}) (string, error) {
	// immediate success with generated id
	return "noop-" + uuid.New().String(), nil
}