package biz

import (
	"context"

	"origadmin/application/origstudio/internal/enterprise/billing/payment/dto"
)

// GatewayConfig holds payment gateway configuration.
type GatewayConfig struct {
	APIKey        string
	WebhookSecret string
	BaseURL       string
	Currency      string
}

// PaymentGateway defines the interface for payment gateway operations.
type PaymentGateway interface {
	CreatePayment(ctx context.Context, intent *dto.PaymentIntent) (*dto.PaymentIntent, error)
	VerifyPayment(ctx context.Context, transactionID string) (*dto.PaymentIntent, error)
	HandleWebhook(ctx context.Context, event *dto.WebhookEvent) (*dto.WebhookEvent, error)
	Channel() dto.PaymentChannel
}