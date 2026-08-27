package payment

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"

	"origadmin/application/origstudio/internal/enterprise/billing/payment/biz"
	"origadmin/application/origstudio/internal/enterprise/billing/payment/dal"
	"origadmin/application/origstudio/internal/enterprise/billing/payment/dto"
	"origadmin/application/origstudio/internal/enterprise/billing/payment/service"
)

var ProviderSet = wire.NewSet(
	dto.ProviderSet,
	NewGatewayConfig,
	NewStripeGateway,
	wire.Bind(new(biz.PaymentGateway), new(*dal.StripeGateway)),
	wire.Bind(new(biz.Repo), new(*dal.Repo)),
	dal.ProviderSet,
	biz.ProviderSet,
	service.ProviderSet,
)

func NewGatewayConfig() *biz.GatewayConfig {
	return &biz.GatewayConfig{
		APIKey:        os.Getenv("STRIPE_API_KEY"),
		WebhookSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
		BaseURL:       os.Getenv("STRIPE_API_BASE"),
		Currency:      os.Getenv("STRIPE_CURRENCY"),
	}
}

func NewStripeGateway(cfg *biz.GatewayConfig, logger log.Logger) (*dal.StripeGateway, error) {
	return dal.NewStripeGatewayWithConfig(cfg, logger)
}