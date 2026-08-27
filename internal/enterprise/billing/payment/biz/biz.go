package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/enterprise/billing/payment/dto"
)

type Repo interface {
	ListSubscriptionPlans(ctx context.Context) ([]*dto.SubscriptionPlanDTO, error)
	GetSubscriptionPlanByID(ctx context.Context, id string) (*dto.SubscriptionPlanDTO, error)
	CreateSubscriptionPlan(ctx context.Context, p *dto.SubscriptionPlanDTO) (*dto.SubscriptionPlanDTO, error)
	UpdateSubscriptionPlan(ctx context.Context, p *dto.SubscriptionPlanDTO) (*dto.SubscriptionPlanDTO, error)
	DeleteSubscriptionPlan(ctx context.Context, id string) error
	ListActiveSubscriptionPlans(ctx context.Context) ([]*dto.SubscriptionPlanDTO, error)

	ListOrders(ctx context.Context, page, pageSize int) ([]*dto.OrderDTO, int, error)
	GetOrderByID(ctx context.Context, id string) (*dto.OrderDTO, error)
	GetOrderByOrderNo(ctx context.Context, orderNo string) (*dto.OrderDTO, error)
	CreateOrder(ctx context.Context, o *dto.OrderDTO) (*dto.OrderDTO, error)
	UpdateOrderStatus(ctx context.Context, orderNo string, status string, paidAt time.Time) error

	CreatePayment(ctx context.Context, p *dto.PaymentDTO) (*dto.PaymentDTO, error)
	UpdatePaymentStatus(ctx context.Context, transactionID string, status string, paidAt time.Time) error

	ListWallets(ctx context.Context, page, pageSize int) ([]*dto.WalletDTO, int, error)
	GetWalletByUserID(ctx context.Context, userID string) (*dto.WalletDTO, error)
	CreateWallet(ctx context.Context, w *dto.WalletDTO) (*dto.WalletDTO, error)
	UpdateWalletBalance(ctx context.Context, userID string, amount float64) error
	CreateWalletTransaction(ctx context.Context, walletID string, amount float64, txType string, orderNo string) error
}

type UseCase struct {
	repo    Repo
	gateway PaymentGateway
	log     *log.Helper
}

func NewUseCase(repo Repo, gateway PaymentGateway, logger log.Logger) *UseCase {
	return &UseCase{
		repo:    repo,
		gateway: gateway,
		log:     log.NewHelper(log.With(logger, "module", "enterprise/payment.biz")),
	}
}

func (uc *UseCase) ListSubscriptionPlans(ctx context.Context) ([]*dto.SubscriptionPlanDTO, error) {
	return uc.repo.ListSubscriptionPlans(ctx)
}

func (uc *UseCase) GetSubscriptionPlanByID(ctx context.Context, id string) (*dto.SubscriptionPlanDTO, error) {
	return uc.repo.GetSubscriptionPlanByID(ctx, id)
}

func (uc *UseCase) CreateSubscriptionPlan(ctx context.Context, p *dto.SubscriptionPlanDTO) (*dto.SubscriptionPlanDTO, error) {
	return uc.repo.CreateSubscriptionPlan(ctx, p)
}

func (uc *UseCase) UpdateSubscriptionPlan(ctx context.Context, p *dto.SubscriptionPlanDTO) (*dto.SubscriptionPlanDTO, error) {
	return uc.repo.UpdateSubscriptionPlan(ctx, p)
}

func (uc *UseCase) DeleteSubscriptionPlan(ctx context.Context, id string) error {
	return uc.repo.DeleteSubscriptionPlan(ctx, id)
}

func (uc *UseCase) ListActiveSubscriptionPlans(ctx context.Context) ([]*dto.SubscriptionPlanDTO, error) {
	return uc.repo.ListActiveSubscriptionPlans(ctx)
}

func (uc *UseCase) ListOrders(ctx context.Context, page, pageSize int) ([]*dto.OrderDTO, int, error) {
	return uc.repo.ListOrders(ctx, page, pageSize)
}

func (uc *UseCase) GetOrderByID(ctx context.Context, id string) (*dto.OrderDTO, error) {
	return uc.repo.GetOrderByID(ctx, id)
}

func (uc *UseCase) ListWallets(ctx context.Context, page, pageSize int) ([]*dto.WalletDTO, int, error) {
	return uc.repo.ListWallets(ctx, page, pageSize)
}

func (uc *UseCase) Checkout(ctx context.Context, userID string, req *dto.CheckoutRequest) (*dto.CheckoutResponse, error) {
	plan, err := uc.repo.GetSubscriptionPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if !plan.IsActive {
		return nil, fmt.Errorf("plan is not active")
	}

	orderNo := generateOrderNo()
	order := &dto.OrderDTO{
		OrderNo:       orderNo,
		Amount:        plan.Price,
		Currency:      plan.Currency,
		Status:        "pending",
		PaymentMethod: string(req.Channel),
		UserID:        userID,
		PlanID:        req.PlanID,
		CreateTime:    time.Now(),
	}

	created, err := uc.repo.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	if uc.gateway == nil {
		return &dto.CheckoutResponse{
			OrderNo: created.OrderNo,
			Status:  "pending",
		}, nil
	}

	intent := &dto.PaymentIntent{
		Amount:   plan.Price,
		Currency: plan.Currency,
		OrderNo:  orderNo,
		Channel:  req.Channel,
	}

	result, err := uc.gateway.CreatePayment(ctx, intent)
	if err != nil {
		uc.log.Errorf("gateway create payment: %v", err)
		return &dto.CheckoutResponse{
			OrderNo: created.OrderNo,
			Status:  "pending",
		}, nil
	}

	payment := &dto.PaymentDTO{
		Channel:       string(req.Channel),
		TransactionID: result.TransactionID,
		Amount:        plan.Price,
		Currency:      plan.Currency,
		Status:        "pending",
		OrderID:       created.ID,
		CreateTime:    time.Now(),
	}
	if _, err := uc.repo.CreatePayment(ctx, payment); err != nil {
		uc.log.Errorf("create payment record: %v", err)
	}

	return &dto.CheckoutResponse{
		OrderNo:      created.OrderNo,
		PaymentURL:   result.PaymentURL,
		ClientSecret: result.ClientSecret,
		Status:       "pending",
	}, nil
}

func (uc *UseCase) HandleWebhook(ctx context.Context, event *dto.WebhookEvent) error {
	if uc.gateway == nil {
		return fmt.Errorf("no gateway configured")
	}

	processed, err := uc.gateway.HandleWebhook(ctx, event)
	if err != nil {
		return fmt.Errorf("process webhook: %w", err)
	}

	if processed.OrderNo == "" {
		uc.log.Warnf("webhook event has no order_no: type=%s id=%s", processed.Type, processed.ID)
		return nil
	}

	order, err := uc.repo.GetOrderByOrderNo(ctx, processed.OrderNo)
	if err != nil {
		return fmt.Errorf("get order by order_no: %w", err)
	}

	switch processed.Status {
	case "success":
		paidAt := time.Now()
		if err := uc.repo.UpdateOrderStatus(ctx, processed.OrderNo, "paid", paidAt); err != nil {
			return fmt.Errorf("update order status: %w", err)
		}
		if processed.TransactionID != "" {
			if err := uc.repo.UpdatePaymentStatus(ctx, processed.TransactionID, "success", paidAt); err != nil {
				uc.log.Errorf("update payment status: %v", err)
			}
		}
		uc.log.Infof("order %s payment success, order_id=%s", processed.OrderNo, order.ID)
	case "failed":
		if err := uc.repo.UpdateOrderStatus(ctx, processed.OrderNo, "cancelled", time.Time{}); err != nil {
			return fmt.Errorf("update order status: %w", err)
		}
		if processed.TransactionID != "" {
			if err := uc.repo.UpdatePaymentStatus(ctx, processed.TransactionID, "failed", time.Time{}); err != nil {
				uc.log.Errorf("update payment status: %v", err)
			}
		}
		uc.log.Infof("order %s payment failed, order_id=%s", processed.OrderNo, order.ID)
	}

	return nil
}

func (uc *UseCase) WalletRecharge(ctx context.Context, userID string, amount float64, currency string) (*dto.WalletDTO, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if currency == "" {
		currency = "CNY"
	}

	wallet, err := uc.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		wallet = &dto.WalletDTO{
			UserID:   userID,
			Balance:  0,
			Frozen:   0,
			Currency: currency,
		}
		wallet, err = uc.repo.CreateWallet(ctx, wallet)
		if err != nil {
			return nil, fmt.Errorf("create wallet: %w", err)
		}
	}

	if err := uc.repo.UpdateWalletBalance(ctx, userID, amount); err != nil {
		return nil, fmt.Errorf("update wallet balance: %w", err)
	}

	orderNo := generateOrderNo()
	if err := uc.repo.CreateWalletTransaction(ctx, wallet.ID, amount, "recharge", orderNo); err != nil {
		uc.log.Errorf("create wallet transaction: %v", err)
	}

	wallet.Balance += amount
	return wallet, nil
}

func (uc *UseCase) GetWalletByUserID(ctx context.Context, userID string) (*dto.WalletDTO, error) {
	return uc.repo.GetWalletByUserID(ctx, userID)
}

func generateOrderNo() string {
	return fmt.Sprintf("ORD%d%06d", time.Now().UnixMilli(), time.Now().Nanosecond()%1000000)
}