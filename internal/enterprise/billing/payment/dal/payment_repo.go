package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/dal/entity"
	"origadmin/application/origstudio/internal/dal/entity/order"
	entityPayment "origadmin/application/origstudio/internal/dal/entity/payment"
	"origadmin/application/origstudio/internal/dal/entity/subscriptionplan"
	"origadmin/application/origstudio/internal/dal/entity/wallet"
	"origadmin/application/origstudio/internal/dal/entity/wallettransaction"
	"origadmin/application/origstudio/internal/enterprise/billing/payment/dto"
)

type Repo struct {
	db  *entity.Client
	log *log.Helper
}

func NewRepo(db *entity.Client, logger log.Logger) *Repo {
	return &Repo{
		db:  db,
		log: log.NewHelper(log.With(logger, "module", "enterprise/payment.repo")),
	}
}

func entityToSubscriptionPlanDTO(e *entity.SubscriptionPlan) *dto.SubscriptionPlanDTO {
	if e == nil {
		return nil
	}
	return &dto.SubscriptionPlanDTO{
		ID:           e.ID,
		Name:         e.Name,
		Description:  e.Description,
		Price:        e.Price,
		Currency:     e.Currency,
		DurationDays: e.DurationDays,
		Features:     e.Features,
		IsActive:     e.IsActive,
		SortOrder:    e.SortOrder,
	}
}

func entityToOrderDTO(e *entity.Order) *dto.OrderDTO {
	if e == nil {
		return nil
	}
	return &dto.OrderDTO{
		ID:            e.ID,
		OrderNo:       e.OrderNo,
		Amount:        e.Amount,
		Currency:      e.Currency,
		Status:        string(e.Status),
		PaymentMethod: e.PaymentMethod,
		PaidAt:        e.PaidAt,
		UserID:        e.UserID,
		PlanID:        e.PlanID,
		CreateTime:    e.CreateTime,
	}
}

func entityToWalletDTO(e *entity.Wallet) *dto.WalletDTO {
	if e == nil {
		return nil
	}
	return &dto.WalletDTO{
		ID:       e.ID,
		Balance:  e.Balance,
		Frozen:   e.Frozen,
		Currency: e.Currency,
		UserID:   e.UserID,
	}
}

func (r *Repo) ListSubscriptionPlans(ctx context.Context) ([]*dto.SubscriptionPlanDTO, error) {
	items, err := r.db.SubscriptionPlan.Query().
		Order(entity.Asc(subscriptionplan.FieldSortOrder), entity.Asc(subscriptionplan.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list subscription plans: %w", err)
	}
	result := make([]*dto.SubscriptionPlanDTO, len(items))
	for i, item := range items {
		result[i] = entityToSubscriptionPlanDTO(item)
	}
	return result, nil
}

func (r *Repo) GetSubscriptionPlanByID(ctx context.Context, id string) (*dto.SubscriptionPlanDTO, error) {
	item, err := r.db.SubscriptionPlan.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get subscription plan: %w", err)
	}
	return entityToSubscriptionPlanDTO(item), nil
}

func (r *Repo) CreateSubscriptionPlan(ctx context.Context, p *dto.SubscriptionPlanDTO) (*dto.SubscriptionPlanDTO, error) {
	builder := r.db.SubscriptionPlan.Create().
		SetName(p.Name).
		SetPrice(p.Price).
		SetCurrency(p.Currency).
		SetDurationDays(p.DurationDays).
		SetIsActive(p.IsActive).
		SetSortOrder(p.SortOrder)
	if p.Description != "" {
		builder.SetDescription(p.Description)
	}
	if p.Features != nil {
		builder.SetFeatures(p.Features)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create subscription plan: %w", err)
	}
	return entityToSubscriptionPlanDTO(ent), nil
}

func (r *Repo) UpdateSubscriptionPlan(ctx context.Context, p *dto.SubscriptionPlanDTO) (*dto.SubscriptionPlanDTO, error) {
	builder := r.db.SubscriptionPlan.UpdateOneID(p.ID).
		SetName(p.Name).
		SetPrice(p.Price).
		SetCurrency(p.Currency).
		SetDurationDays(p.DurationDays).
		SetIsActive(p.IsActive).
		SetSortOrder(p.SortOrder)
	builder.SetDescription(p.Description)
	if p.Features != nil {
		builder.SetFeatures(p.Features)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update subscription plan: %w", err)
	}
	return entityToSubscriptionPlanDTO(ent), nil
}

func (r *Repo) DeleteSubscriptionPlan(ctx context.Context, id string) error {
	return r.db.SubscriptionPlan.DeleteOneID(id).Exec(ctx)
}

func (r *Repo) ListActiveSubscriptionPlans(ctx context.Context) ([]*dto.SubscriptionPlanDTO, error) {
	items, err := r.db.SubscriptionPlan.Query().
		Where(subscriptionplan.IsActiveEQ(true)).
		Order(entity.Asc(subscriptionplan.FieldSortOrder), entity.Asc(subscriptionplan.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active subscription plans: %w", err)
	}
	result := make([]*dto.SubscriptionPlanDTO, len(items))
	for i, item := range items {
		result[i] = entityToSubscriptionPlanDTO(item)
	}
	return result, nil
}

func (r *Repo) ListOrders(ctx context.Context, page, pageSize int) ([]*dto.OrderDTO, int, error) {
	query := r.db.Order.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}
	items, err := query.
		Order(entity.Desc(order.FieldCreateTime)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	result := make([]*dto.OrderDTO, len(items))
	for i, item := range items {
		result[i] = entityToOrderDTO(item)
	}
	return result, total, nil
}

func (r *Repo) GetOrderByID(ctx context.Context, id string) (*dto.OrderDTO, error) {
	item, err := r.db.Order.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}
	return entityToOrderDTO(item), nil
}

func (r *Repo) ListWallets(ctx context.Context, page, pageSize int) ([]*dto.WalletDTO, int, error) {
	query := r.db.Wallet.Query()
	total, err := query.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count wallets: %w", err)
	}
	items, err := query.
		Order(entity.Desc(wallet.FieldID)).
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list wallets: %w", err)
	}
	result := make([]*dto.WalletDTO, len(items))
	for i, item := range items {
		result[i] = entityToWalletDTO(item)
	}
	return result, total, nil
}

func (r *Repo) GetOrderByOrderNo(ctx context.Context, orderNo string) (*dto.OrderDTO, error) {
	item, err := r.db.Order.Query().
		Where(order.OrderNoEQ(orderNo)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get order by order_no: %w", err)
	}
	return entityToOrderDTO(item), nil
}

func (r *Repo) CreateOrder(ctx context.Context, o *dto.OrderDTO) (*dto.OrderDTO, error) {
	builder := r.db.Order.Create().
		SetOrderNo(o.OrderNo).
		SetAmount(o.Amount).
		SetCurrency(o.Currency).
		SetStatus(order.Status(o.Status)).
		SetUserID(o.UserID)
	if o.PaymentMethod != "" {
		builder.SetPaymentMethod(o.PaymentMethod)
	}
	if o.PlanID != "" {
		builder.SetPlanID(o.PlanID)
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return entityToOrderDTO(ent), nil
}

func (r *Repo) UpdateOrderStatus(ctx context.Context, orderNo string, status string, paidAt time.Time) error {
	builder := r.db.Order.Update().
		Where(order.OrderNoEQ(orderNo)).
		SetStatus(order.Status(status))
	if !paidAt.IsZero() {
		builder.SetPaidAt(paidAt)
	}
	return builder.Exec(ctx)
}

func (r *Repo) CreatePayment(ctx context.Context, p *dto.PaymentDTO) (*dto.PaymentDTO, error) {
	builder := r.db.Payment.Create().
		SetAmount(p.Amount).
		SetCurrency(p.Currency).
		SetOrderID(p.OrderID)
	if p.Channel != "" {
		builder.SetChannel(entityPayment.Channel(p.Channel))
	}
	if p.TransactionID != "" {
		builder.SetTransactionID(p.TransactionID)
	}
	if p.Status != "" {
		builder.SetStatus(entityPayment.Status(p.Status))
	}
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}
	return &dto.PaymentDTO{
		ID:            ent.ID,
		Channel:       string(ent.Channel),
		TransactionID: ent.TransactionID,
		Amount:        ent.Amount,
		Currency:      ent.Currency,
		Status:        string(ent.Status),
		OrderID:       ent.OrderID,
		CreateTime:    ent.CreateTime,
	}, nil
}

func (r *Repo) UpdatePaymentStatus(ctx context.Context, transactionID string, status string, paidAt time.Time) error {
	builder := r.db.Payment.Update().
		Where(entityPayment.TransactionIDEQ(transactionID)).
		SetStatus(entityPayment.Status(status))
	if !paidAt.IsZero() {
		builder.SetPaidAt(paidAt)
	}
	return builder.Exec(ctx)
}

func (r *Repo) GetWalletByUserID(ctx context.Context, userID string) (*dto.WalletDTO, error) {
	item, err := r.db.Wallet.Query().
		Where(wallet.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("get wallet by user_id: %w", err)
	}
	return entityToWalletDTO(item), nil
}

func (r *Repo) CreateWallet(ctx context.Context, w *dto.WalletDTO) (*dto.WalletDTO, error) {
	builder := r.db.Wallet.Create().
		SetBalance(w.Balance).
		SetFrozen(w.Frozen).
		SetCurrency(w.Currency).
		SetUserID(w.UserID)
	ent, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create wallet: %w", err)
	}
	return entityToWalletDTO(ent), nil
}

func (r *Repo) UpdateWalletBalance(ctx context.Context, userID string, amount float64) error {
	walletEnt, err := r.db.Wallet.Query().
		Where(wallet.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("get wallet: %w", err)
	}
	newBalance := walletEnt.Balance + amount
	return r.db.Wallet.UpdateOneID(walletEnt.ID).
		SetBalance(newBalance).
		Exec(ctx)
}

func (r *Repo) CreateWalletTransaction(ctx context.Context, walletID string, amount float64, txType string, orderNo string) error {
	walletEnt, err := r.db.Wallet.Get(ctx, walletID)
	if err != nil {
		return fmt.Errorf("get wallet: %w", err)
	}
	_, err = r.db.WalletTransaction.Create().
		SetType(wallettransaction.Type(txType)).
		SetAmount(amount).
		SetBalanceBefore(walletEnt.Balance).
		SetBalanceAfter(walletEnt.Balance + amount).
		SetReference(orderNo).
		SetWalletID(walletID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create wallet transaction: %w", err)
	}
	return nil
}