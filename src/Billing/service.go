package Billing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type Provider interface {
	Charge(ctx context.Context, provider string, amount decimal.Decimal, currency string, metadata map[string]interface{}) (string, error)
}

type service struct {
	repo     Repository
	provider Provider
	log      *zap.Logger
}

func NewService(r Repository, p Provider, log *zap.Logger) *service {
	return &service{repo: r, provider: p, log: log}
}

func (s *service) IssueInvoice(ctx context.Context, orderID uuid.UUID, amount decimal.Decimal, currency string, dueInDays int) (*Invoice, error) {
	inv := &Invoice{OrderID: orderID, InvoiceNumber: uuid.New().String(), Status: "UNPAID", Amount: amount, Currency: currency}
	if dueInDays > 0 {
		d := time.Now().UTC().AddDate(0, 0, dueInDays)
		inv.DueAt = &d
	}
	if err := s.repo.CreateInvoice(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (s *service) PayInvoice(ctx context.Context, invoiceID uuid.UUID, provider string, metadata map[string]interface{}) (*Payment, error) {
	inv, err := s.repo.GetInvoiceByOrder(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if inv.Status == "PAID" {
		return nil, errors.New("invoice already paid")
	}
	// call provider
	amount := inv.Amount
	ppid, perr := s.provider.Charge(ctx, provider, amount, inv.Currency, metadata)
	if perr != nil {
		return nil, perr
	}
	p := &Payment{InvoiceID: inv.ID, Provider: provider, ProviderPaymentID: &ppid, Amount: amount, Currency: inv.Currency, Status: "SUCCESS", Metadata: nil}
	if err := s.repo.CreatePayment(ctx, p); err != nil {
		return nil, err
	}
	paidAt := time.Now().UTC()
	if err := s.repo.UpdateInvoiceStatus(ctx, inv.ID, "PAID", &paidAt); err != nil {
		return nil, err
	}
	return p, nil
}
