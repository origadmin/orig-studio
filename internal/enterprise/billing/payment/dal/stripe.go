package dal

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"origadmin/application/origstudio/internal/enterprise/billing/payment/biz"
	"origadmin/application/origstudio/internal/enterprise/billing/payment/dto"
)

const (
	stripeAPIBase     = "https://api.stripe.com/v1"
	defaultTimeout    = 30 * time.Second
	maxBodyReadBytes  = 65536
)

type StripeGateway struct {
	apiKey        string
	webhookSecret string
	baseURL       string
	currency      string
	httpClient    *http.Client
	log           *log.Helper
}

func NewStripeGatewayWithConfig(cfg *biz.GatewayConfig, logger log.Logger) (*StripeGateway, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gateway config is nil")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("stripe api key is required")
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = stripeAPIBase
	}
	currency := cfg.Currency
	if currency == "" {
		currency = "usd"
	}
	return &StripeGateway{
		apiKey:        cfg.APIKey,
		webhookSecret: cfg.WebhookSecret,
		baseURL:       strings.TrimRight(baseURL, "/"),
		currency:      currency,
		httpClient:    &http.Client{Timeout: defaultTimeout},
		log:           log.NewHelper(log.With(logger, "module", "enterprise/payment.stripe")),
	}, nil
}

func (g *StripeGateway) Channel() dto.PaymentChannel {
	return dto.ChannelStripe
}

func (g *StripeGateway) CreatePayment(ctx context.Context, intent *dto.PaymentIntent) (*dto.PaymentIntent, error) {
	if intent == nil {
		return nil, fmt.Errorf("payment intent is nil")
	}
	currency := intent.Currency
	if currency == "" {
		currency = g.currency
	}
	amountCents := int64(intent.Amount * 100)

	params := map[string]interface{}{
		"line_items": []map[string]interface{}{
			{
				"price_data": map[string]interface{}{
					"currency":     currency,
					"product_data": map[string]interface{}{"name": fmt.Sprintf("Order %s", intent.OrderNo)},
					"unit_amount":  amountCents,
				},
				"quantity": 1,
			},
		},
		"mode": "payment",
		"metadata": map[string]interface{}{
			"order_no": intent.OrderNo,
		},
	}

	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal checkout params: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/checkout/sessions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(g.apiKey, "")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stripe api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyReadBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		g.log.Errorf("stripe checkout error: status=%d body=%s", resp.StatusCode, string(respBody))
		return nil, fmt.Errorf("stripe api error: status=%d", resp.StatusCode)
	}

	var result struct {
		ID           string `json:"id"`
		URL          string `json:"url"`
		Status       string `json:"status"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal stripe response: %w", err)
	}

	intent.ID = result.ID
	intent.PaymentURL = result.URL
	intent.ClientSecret = result.ClientSecret
	intent.Status = result.Status
	intent.CreatedAt = time.Now()
	intent.UpdatedAt = time.Now()

	return intent, nil
}

func (g *StripeGateway) VerifyPayment(ctx context.Context, transactionID string) (*dto.PaymentIntent, error) {
	if transactionID == "" {
		return nil, fmt.Errorf("transaction id is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.baseURL+"/payment_intents/"+transactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth(g.apiKey, "")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stripe api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyReadBytes))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("stripe api error: status=%d", resp.StatusCode)
	}

	var result struct {
		ID            string  `json:"id"`
		Amount        int64   `json:"amount"`
		Currency      string  `json:"currency"`
		Status        string  `json:"status"`
		ClientSecret  string  `json:"client_secret"`
		Metadata      map[string]interface{} `json:"metadata"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal stripe response: %w", err)
	}

	intent := &dto.PaymentIntent{
		ID:            result.ID,
		Amount:        float64(result.Amount) / 100,
		Currency:      result.Currency,
		Status:        result.Status,
		ClientSecret:  result.ClientSecret,
		TransactionID: result.ID,
		Metadata:      result.Metadata,
		Channel:       dto.ChannelStripe,
	}

	if metaOrderNo, ok := result.Metadata["order_no"]; ok {
		if s, ok := metaOrderNo.(string); ok {
			intent.OrderNo = s
		}
	}

	return intent, nil
}

func (g *StripeGateway) HandleWebhook(ctx context.Context, event *dto.WebhookEvent) (*dto.WebhookEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("webhook event is nil")
	}

	if g.webhookSecret != "" && !g.verifyWebhookSignature(event.RawBody, event.Signature) {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	var stripeEvent struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID            string  `json:"id"`
				Amount        int64   `json:"amount"`
				Currency      string  `json:"currency"`
				Status        string  `json:"status"`
				ClientSecret  string  `json:"client_secret"`
				Metadata      map[string]interface{} `json:"metadata"`
				PaymentIntent string  `json:"payment_intent"`
			} `json:"object"`
		} `json:"data"`
	}

	if event.RawBody != "" {
		if err := json.Unmarshal([]byte(event.RawBody), &stripeEvent); err != nil {
			return nil, fmt.Errorf("unmarshal webhook body: %w", err)
		}
	}

	event.Type = stripeEvent.Type
	event.ID = stripeEvent.ID

	switch stripeEvent.Type {
	case "checkout.session.completed":
		event.Status = "success"
		event.Amount = float64(stripeEvent.Data.Object.Amount) / 100
		event.Currency = stripeEvent.Data.Object.Currency
		event.TransactionID = stripeEvent.Data.Object.PaymentIntent
	case "payment_intent.succeeded":
		event.Status = "success"
		event.Amount = float64(stripeEvent.Data.Object.Amount) / 100
		event.Currency = stripeEvent.Data.Object.Currency
		event.TransactionID = stripeEvent.Data.Object.ID
	case "payment_intent.payment_failed":
		event.Status = "failed"
	default:
		event.Status = "unknown"
	}

	event.Metadata = stripeEvent.Data.Object.Metadata
	if metaOrderNo, ok := stripeEvent.Data.Object.Metadata["order_no"]; ok {
		if s, ok := metaOrderNo.(string); ok {
			event.OrderNo = s
		}
	}

	return event, nil
}

func (g *StripeGateway) verifyWebhookSignature(payload, sigHeader string) bool {
	if sigHeader == "" || payload == "" {
		return false
	}

	parts := strings.Split(sigHeader, ",")
	var timestamp, signature string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			signature = strings.TrimPrefix(part, "v1=")
		}
	}

	if timestamp == "" || signature == "" {
		return false
	}

	signedPayload := fmt.Sprintf("%s.%s", timestamp, payload)
	mac := hmac.New(sha256.New, []byte(g.webhookSecret))
	mac.Write([]byte(signedPayload))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}