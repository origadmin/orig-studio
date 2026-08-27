package dto

import "time"

type SubscriptionPlanDTO struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Price        float64                `json:"price"`
	Currency     string                 `json:"currency"`
	DurationDays int                    `json:"duration_days"`
	Features     map[string]interface{} `json:"features"`
	IsActive     bool                   `json:"is_active"`
	SortOrder    int                    `json:"sort_order"`
}

type OrderDTO struct {
	ID            string    `json:"id"`
	OrderNo       string    `json:"order_no"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	PaidAt        time.Time `json:"paid_at"`
	UserID        string    `json:"user_id"`
	PlanID        string    `json:"plan_id"`
	CreateTime    time.Time `json:"create_time"`
}

type WalletDTO struct {
	ID       string  `json:"id"`
	Balance  float64 `json:"balance"`
	Frozen   float64 `json:"frozen"`
	Currency string  `json:"currency"`
	UserID   string  `json:"user_id"`
}

type PaymentDTO struct {
	ID            string    `json:"id"`
	Channel       string    `json:"channel"`
	TransactionID string    `json:"transaction_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PaidAt        time.Time `json:"paid_at"`
	OrderID       string    `json:"order_id"`
	CreateTime    time.Time `json:"create_time"`
}

type PaymentChannel string

const (
	ChannelStripe PaymentChannel = "stripe"
	ChannelPaypal PaymentChannel = "paypal"
	ChannelAlipay PaymentChannel = "alipay"
	ChannelWallet PaymentChannel = "wallet"
)

type PaymentIntent struct {
	ID            string                 `json:"id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Status        string                 `json:"status"`
	Channel       PaymentChannel         `json:"channel"`
	OrderNo       string                 `json:"order_no"`
	ClientSecret  string                 `json:"client_secret,omitempty"`
	PaymentURL    string                 `json:"payment_url,omitempty"`
	TransactionID string                 `json:"transaction_id,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type PaymentResult struct {
	TransactionID string
	PaymentURL    string
	ClientSecret  string
}

type WebhookEvent struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Data          map[string]interface{} `json:"data"`
	RawBody       string                 `json:"raw_body"`
	Signature     string                 `json:"signature"`
	TransactionID string                 `json:"transaction_id"`
	Status        string                 `json:"status"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	OrderNo       string                 `json:"order_no,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type WebhookResult struct {
	ID            string
	Type          string
	OrderNo       string
	TransactionID string
	Status        string
}

type CheckoutRequest struct {
	PlanID     string         `json:"plan_id"`
	Channel    PaymentChannel `json:"channel"`
	SuccessURL string         `json:"success_url,omitempty"`
	CancelURL  string         `json:"cancel_url,omitempty"`
}

type CheckoutResponse struct {
	OrderNo      string `json:"order_no"`
	PaymentURL   string `json:"payment_url,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	Status       string `json:"status"`
}