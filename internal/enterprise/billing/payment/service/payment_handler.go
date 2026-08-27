package service

import (
	"strconv"


	http2 "origadmin/application/origstudio/internal/pkg/http"
	authbiz "origadmin/application/origstudio/internal/features/auth/biz"
	"origadmin/application/origstudio/internal/infra/auth"
	"origadmin/application/origstudio/internal/server"
	"origadmin/application/origstudio/internal/enterprise/billing/payment/biz"
	"origadmin/application/origstudio/internal/enterprise/billing/payment/dto"
)

type Handler struct {
	uc          *biz.UseCase
	jwt         *auth.Manager
	permChecker authbiz.PermissionChecker
}

func NewHandler(uc *biz.UseCase, jwt *auth.Manager, permChecker authbiz.PermissionChecker) *Handler {
	return &Handler{uc: uc, jwt: jwt, permChecker: permChecker}
}

func (h *Handler) RegisterRoutes(r http2.Router) {
	adminPlans := r.Group("/admin/subscription-plans")
	adminPlans.Use(server.JWTMiddlewareCtx(h.jwt), server.RequirePermissionCtx(h.permChecker, "payment:read"))
	{
		adminPlans.GET("", h.listSubscriptionPlans())
		adminPlans.POST("", server.RequirePermissionCtx(h.permChecker, "payment:write")(h.createSubscriptionPlan()))
		adminPlans.PUT("/:id", server.RequirePermissionCtx(h.permChecker, "payment:write")(h.updateSubscriptionPlan()))
		adminPlans.DELETE("/:id", server.RequirePermissionCtx(h.permChecker, "payment:delete")(h.deleteSubscriptionPlan()))
	}

	adminOrders := r.Group("/admin/orders")
	adminOrders.Use(server.JWTMiddlewareCtx(h.jwt), server.RequirePermissionCtx(h.permChecker, "payment:read"))
	{
		adminOrders.GET("", h.listOrders())
		adminOrders.GET("/:id", h.getOrderByID())
	}

	adminWallets := r.Group("/admin/wallets")
	adminWallets.Use(server.JWTMiddlewareCtx(h.jwt), server.RequirePermissionCtx(h.permChecker, "payment:read"))
	{
		adminWallets.GET("", h.listWallets())
	}

	publicPlans := r.Group("/subscription-plans")
	{
		publicPlans.GET("", h.listActiveSubscriptionPlans())
	}

	userPayment := r.Group("/payment")
	userPayment.Use(server.JWTMiddlewareCtx(h.jwt))
	{
		userPayment.POST("/checkout", h.checkout())
		userPayment.POST("/wallet/recharge", h.walletRecharge())
		userPayment.GET("/wallet", h.getWallet())
	}

	webhook := r.Group("/webhook")
	{
		webhook.POST("/stripe", h.stripeWebhook())
	}
}

func (h *Handler) listSubscriptionPlans() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListSubscriptionPlans(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) createSubscriptionPlan() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		var input struct {
			Name         string                 `json:"name" binding:"required"`
			Description  string                 `json:"description"`
			Price        float64                `json:"price" binding:"required"`
			Currency     string                 `json:"currency"`
			DurationDays int                    `json:"duration_days" binding:"required"`
			Features     map[string]interface{} `json:"features"`
			IsActive     *bool                  `json:"is_active"`
			SortOrder    int                    `json:"sort_order"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}
		currency := "CNY"
		if input.Currency != "" {
			currency = input.Currency
		}
		p := &dto.SubscriptionPlanDTO{
			Name: input.Name, Description: input.Description, Price: input.Price,
			Currency: currency, DurationDays: input.DurationDays, Features: input.Features,
			IsActive: isActive, SortOrder: input.SortOrder,
		}
		created, err := h.uc.CreateSubscriptionPlan(ctx, p)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, created)
		return nil
	}
}

func (h *Handler) updateSubscriptionPlan() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		var input struct {
			Name         string                 `json:"name"`
			Description  string                 `json:"description"`
			Price        *float64               `json:"price"`
			Currency     string                 `json:"currency"`
			DurationDays *int                   `json:"duration_days"`
			Features     map[string]interface{} `json:"features"`
			IsActive     *bool                  `json:"is_active"`
			SortOrder    *int                   `json:"sort_order"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		existing, err := h.uc.GetSubscriptionPlanByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "plan not found")
			return nil
		}
		p := &dto.SubscriptionPlanDTO{ID: id}
		if input.Name != "" {
			p.Name = input.Name
		} else {
			p.Name = existing.Name
		}
		p.Description = input.Description
		if input.Price != nil {
			p.Price = *input.Price
		} else {
			p.Price = existing.Price
		}
		if input.Currency != "" {
			p.Currency = input.Currency
		} else {
			p.Currency = existing.Currency
		}
		if input.DurationDays != nil {
			p.DurationDays = *input.DurationDays
		} else {
			p.DurationDays = existing.DurationDays
		}
		if input.Features != nil {
			p.Features = input.Features
		} else {
			p.Features = existing.Features
		}
		if input.IsActive != nil {
			p.IsActive = *input.IsActive
		} else {
			p.IsActive = existing.IsActive
		}
		if input.SortOrder != nil {
			p.SortOrder = *input.SortOrder
		} else {
			p.SortOrder = existing.SortOrder
		}
		updated, err := h.uc.UpdateSubscriptionPlan(ctx, p)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, updated)
		return nil
	}
}

func (h *Handler) deleteSubscriptionPlan() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		if err := h.uc.DeleteSubscriptionPlan(ctx, id); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, nil)
		return nil
	}
}

func (h *Handler) listActiveSubscriptionPlans() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		items, err := h.uc.ListActiveSubscriptionPlans(ctx)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, items)
		return nil
	}
}

func (h *Handler) listOrders() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}
		items, total, err := h.uc.ListOrders(ctx, page, pageSize)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": total, "page": page, "page_size": pageSize})
		return nil
	}
}

func (h *Handler) checkout() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
		}
		userID := claims.GetUserID()

		var req dto.CheckoutRequest
		if err := ctx.BindJSON(&req); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		if req.PlanID == "" {
			http2.Fail(ctx, server.ErrBadRequest, "plan_id is required")
			return nil
		}
		if req.Channel == "" {
			req.Channel = dto.ChannelStripe
		}

		resp, err := h.uc.Checkout(ctx, userID, &req)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, resp)
		return nil
	}
}

func (h *Handler) stripeWebhook() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		body, err := ctx.GetRawData()
		if err != nil {
			http2.Fail(ctx, server.ErrBadRequest, "read body failed")
			return nil
		}
		signature := ctx.GetHeader("Stripe-Signature")

		event := &dto.WebhookEvent{
			RawBody:   string(body),
			Signature: signature,
		}

		if err := h.uc.HandleWebhook(ctx, event); err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"status": "ok"})
		return nil
	}
}

func (h *Handler) walletRecharge() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
		}
		userID := claims.GetUserID()

		var input struct {
			Amount   float64 `json:"amount" binding:"required"`
			Currency string  `json:"currency"`
		}
		if err := ctx.BindJSON(&input); err != nil {
			http2.Fail(ctx, server.ErrBadRequest, err.Error())
			return nil
		}
		if input.Amount <= 0 {
			http2.Fail(ctx, server.ErrBadRequest, "amount must be positive")
			return nil
		}

		wallet, err := h.uc.WalletRecharge(ctx, userID, input.Amount, input.Currency)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, wallet)
		return nil
	}
}

func (h *Handler) getWallet() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		claims, ok := server.GetClaimsCtx(ctx)
		if !ok {
			http2.Fail(ctx, server.ErrUnauthorized, "unauthorized")
			return nil
		}
		userID := claims.GetUserID()

		wallet, err := h.uc.GetWalletByUserID(ctx, userID)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "wallet not found")
			return nil
		}
		http2.OK(ctx, wallet)
		return nil
	}
}

func (h *Handler) getOrderByID() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		id := ctx.Var("id")
		if id == "" {
			http2.Fail(ctx, server.ErrBadRequest, "id is required")
			return nil
		}
		item, err := h.uc.GetOrderByID(ctx, id)
		if err != nil {
			http2.Fail(ctx, server.ErrNotFound, "order not found")
			return nil
		}
		http2.OK(ctx, item)
		return nil
	}
}

func (h *Handler) listWallets() http2.HandlerFunc {
	return func(ctx http2.Context) error {
		page, _ := strconv.Atoi(ctx.QueryVarDefault("page", "1"))
		pageSize, _ := strconv.Atoi(ctx.QueryVarDefault("page_size", "20"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}
		items, total, err := h.uc.ListWallets(ctx, page, pageSize)
		if err != nil {
			http2.Fail(ctx, server.ErrInternal, err.Error())
			return nil
		}
		http2.OK(ctx, map[string]interface{}{"items": items, "total": total, "page": page, "page_size": pageSize})
		return nil
	}
}
