package dto

// InvestorDTO is a bot investor as shown on the leaderboard and profile.
type InvestorDTO struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Bio      string  `json:"bio"`
	Strategy string  `json:"strategy"`
	ROI      float64 `json:"roi"`
	Rank     int     `json:"rank"`
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
