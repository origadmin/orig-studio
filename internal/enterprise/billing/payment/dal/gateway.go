package dal

type WalletRechargeDTO struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	OrderNo  string  `json:"order_no"`
}