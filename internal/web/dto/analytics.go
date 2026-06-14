package dto

// AnalyticsPoint is one point on a portfolio/benchmark time series.
type AnalyticsPoint struct {
	Date  int64   `json:"date"` // unix seconds
	Value float64 `json:"value"`
}

// SectorSliceDTO is one sector's share of the current portfolio.
type SectorSliceDTO struct {
	Sector string  `json:"sector"`
	Value  float64 `json:"value"`
	Pct    float64 `json:"pct"`
}

// AnalyticsDTO is the portfolio performance + risk summary.
type AnalyticsDTO struct {
	StartValue    float64          `json:"start_value"`
	CurrentValue  float64          `json:"current_value"`
	ROI           float64          `json:"roi"`
	MaxDrawdown   float64          `json:"max_drawdown"`
	Volatility    float64          `json:"volatility"` // annualized
	Sharpe        float64          `json:"sharpe"`
	WinRate       float64          `json:"win_rate"`
	TradeCount    int              `json:"trade_count"`
	BestDay       float64          `json:"best_day"`
	WorstDay      float64          `json:"worst_day"`
	Equity        []AnalyticsPoint `json:"equity"`
	Benchmark     []AnalyticsPoint `json:"benchmark"`
	Sectors       []SectorSliceDTO `json:"sectors"`
	BenchmarkName string           `json:"benchmark_name"`
}
