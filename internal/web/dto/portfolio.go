package dto

// HoldingDTO is one valued position in the portfolio response.
type HoldingDTO struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	Quantity        float64 `json:"quantity"`
	AvgPrice        float64 `json:"avg_price"`
	CurrentPrice    float64 `json:"current_price"`
	CostBasis       float64 `json:"cost_basis"`
	MarketValue     float64 `json:"market_value"`
	UnrealizedPL    float64 `json:"unrealized_pl"`
	UnrealizedPLPct float64 `json:"unrealized_pl_pct"`
	AllocationPct   float64 `json:"allocation_pct"`
}

// PortfolioStatsDTO is the whole-account valuation summary.
type PortfolioStatsDTO struct {
	Cash            float64      `json:"cash"`
	HoldingsValue   float64      `json:"holdings_value"`
	TotalValue      float64      `json:"total_value"`
	CostBasis       float64      `json:"cost_basis"`
	UnrealizedPL    float64      `json:"unrealized_pl"`
	UnrealizedPLPct float64      `json:"unrealized_pl_pct"`
	ROI             float64      `json:"roi"`
	Holdings        []HoldingDTO `json:"holdings"`
}
