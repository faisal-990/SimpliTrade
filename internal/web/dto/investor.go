package dto

// InvestorDTO is a bot investor as shown on the leaderboard and profile.
type InvestorDTO struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Bio       string  `json:"bio"`
	Strategy  string  `json:"strategy"`
	ROI       float64 `json:"roi"`
	Rank      int     `json:"rank"`
	Followers int     `json:"followers"`
	CreatedBy string  `json:"created_by"` // creator's name for user-built investors; empty for presets
}

// CreateInvestorRequest is the curated "build your own investor" form. The
// service expands it into a full engine strategy config.
type CreateInvestorRequest struct {
	Name         string `json:"name" binding:"required,min=2,max=100"`
	Philosophy   string `json:"philosophy" binding:"max=500"`
	Approach     string `json:"approach" binding:"required,oneof=value quality growth momentum"`
	MaxPositions int    `json:"max_positions"`

	// Approach-specific (0 = skip that gate).
	PEMax              float64 `json:"pe_max"`
	PBMax              float64 `json:"pb_max"`
	ROEMin             float64 `json:"roe_min"`
	OperatingMarginMin float64 `json:"operating_margin_min"`
	RevenueGrowthMin   float64 `json:"revenue_growth_min"`
	EPSGrowthMin       float64 `json:"eps_growth_min"`
	Return6MMin        float64 `json:"return_6m_min"`

	// Sell + risk.
	StopLossPct           float64 `json:"stop_loss_pct"`
	TakeProfitVsIntrinsic float64 `json:"take_profit_vs_intrinsic"`
	MaxPositionSize       float64 `json:"max_position_size"`
	CashBufferMin         float64 `json:"cash_buffer_min"`
	PositionSizing        string  `json:"position_sizing"`
}

// FeedItem is one trade in the aggregated feed of followed investors, annotated
// with which investor made it.
type FeedItem struct {
	InvestorID   string  `json:"investor_id"`
	InvestorName string  `json:"investor_name"`
	Symbol       string  `json:"symbol"`
	Side         string  `json:"side"`
	Quantity     float64 `json:"quantity"`
	Price        float64 `json:"price"`
	ExecutedAt   int64   `json:"executed_at"`
}
