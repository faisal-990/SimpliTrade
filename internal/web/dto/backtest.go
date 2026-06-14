package dto

// BacktestPoint is one point on the equity curve.
type BacktestPoint struct {
	Date  int64   `json:"date"` // unix seconds
	Value float64 `json:"value"`
}

// BacktestTrade is one order executed during the replay.
type BacktestTrade struct {
	Date       int64   `json:"date"`
	Side       string  `json:"side"`
	Symbol     string  `json:"symbol"`
	Quantity   float64 `json:"quantity"`
	Price      float64 `json:"price"`
	TotalValue float64 `json:"total_value"`
}

// BacktestHolding is a position the strategy still held at the end of the run.
type BacktestHolding struct {
	Symbol       string  `json:"symbol"`
	Quantity     float64 `json:"quantity"`
	AvgPrice     float64 `json:"avg_price"`
	LastPrice    float64 `json:"last_price"`
	MarketValue  float64 `json:"market_value"`
	UnrealizedPL float64 `json:"unrealized_pl"`
}

// BacktestResultDTO is the outcome of replaying an investor's strategy over
// historical prices.
type BacktestResultDTO struct {
	InvestorID   string            `json:"investor_id"`
	InvestorName string            `json:"investor_name"`
	Strategy     string            `json:"strategy"`
	StartCash    float64           `json:"start_cash"`
	FinalValue   float64           `json:"final_value"`
	EndCash      float64           `json:"end_cash"`
	ROI          float64           `json:"roi"`
	MaxDrawdown  float64           `json:"max_drawdown"`
	WinRate      float64           `json:"win_rate"`
	TradeCount   int               `json:"trade_count"`
	BuyCount     int               `json:"buy_count"`
	SellCount    int               `json:"sell_count"`
	HeldCount    int               `json:"held_count"`
	StartDate    int64             `json:"start_date"`
	EndDate      int64             `json:"end_date"`
	Equity       []BacktestPoint   `json:"equity"`
	Trades       []BacktestTrade   `json:"trades"`
	Holdings     []BacktestHolding `json:"holdings"`
	Note         string            `json:"note"`
}
