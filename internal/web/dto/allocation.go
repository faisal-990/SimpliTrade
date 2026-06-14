package dto

// AllocateRequest opens a capped copy sub-account that mirrors an investor.
type AllocateRequest struct {
	InvestorID string  `json:"investor_id" binding:"required"`
	Capital    float64 `json:"capital" binding:"required,gt=0"`
}

// AllocationDTO is a user's copy sub-account: how much was committed to an
// investor and what it's currently worth.
type AllocationDTO struct {
	ID           string  `json:"id"`
	InvestorID   string  `json:"investor_id"`
	InvestorName string  `json:"investor_name"`
	Strategy     string  `json:"strategy"`
	Capital      float64 `json:"capital"`
	Cash         float64 `json:"cash"`
	MarketValue  float64 `json:"market_value"`
	ReturnPct    float64 `json:"return_pct"`
	IsActive     bool    `json:"is_active"`
}

// AllocationHoldingDTO is one position the bot opened with allocated capital.
type AllocationHoldingDTO struct {
	Symbol          string  `json:"symbol"`
	Quantity        float64 `json:"quantity"`
	AvgPrice        float64 `json:"avg_price"`
	CurrentPrice    float64 `json:"current_price"`
	MarketValue     float64 `json:"market_value"`
	UnrealizedPL    float64 `json:"unrealized_pl"`
	UnrealizedPLPct float64 `json:"unrealized_pl_pct"`
}

// AllocationTradeDTO is one order the bot executed with allocated capital.
type AllocationTradeDTO struct {
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	TotalValue float64 `json:"total_value"`
	ExecutedAt int64   `json:"executed_at"`
}

// AllocationDetailDTO is an allocation plus exactly what the bot has done with
// the capital: its positions and recent orders.
type AllocationDetailDTO struct {
	AllocationDTO
	Holdings []AllocationHoldingDTO `json:"holdings"`
	Trades   []AllocationTradeDTO   `json:"trades"`
}
