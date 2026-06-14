package dto

// TradeRequest is the body for POST /trade/buy and /trade/sell.
type TradeRequest struct {
	Symbol   string  `json:"symbol" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	// IdempotencyKey is optional; if set, retrying the same key returns the
	// original trade instead of executing twice.
	IdempotencyKey string `json:"idempotency_key"`
}

// TradeResponse is returned after a successful buy/sell.
type TradeResponse struct {
	TradeID    string  `json:"trade_id"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	TotalValue float64 `json:"total_value"`
	ExecutedAt int64   `json:"executed_at"` // unix seconds
}

// TradeHistoryItem is one row of an account's trade history.
type TradeHistoryItem struct {
	TradeID    string  `json:"trade_id"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	TotalValue float64 `json:"total_value"`
	Status     string  `json:"status"`
	ExecutedAt int64   `json:"executed_at"`
	Reason     string  `json:"reason"` // strategy rationale; empty for manual trades
}
