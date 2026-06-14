package dto

import "github.com/faisal-990/ProjectInvestApp/internal/platform/models"

// StockSummaryDTO is a row in the fundamentals listing.
type StockSummaryDTO struct {
	Symbol       string              `json:"symbol"`
	Name         string              `json:"name"`
	Sector       string              `json:"sector"`
	AssetClass   string              `json:"asset_class"`
	CurrentPrice float64             `json:"current_price"`
	Fundamentals models.Fundamentals `json:"fundamentals"`
}

// CandleDTO is one OHLCV bar for charting.
type CandleDTO struct {
	Time   int64   `json:"time"` // unix seconds
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

// StockDetailDTO is the full stock page payload: identity, fundamentals, and
// recent price history for the chart.
type StockDetailDTO struct {
	Symbol       string              `json:"symbol"`
	Name         string              `json:"name"`
	Sector       string              `json:"sector"`
	Exchange     string              `json:"exchange"`
	AssetClass   string              `json:"asset_class"`
	CurrentPrice float64             `json:"current_price"`
	Fundamentals models.Fundamentals `json:"fundamentals"`
	Candles      []CandleDTO         `json:"candles"`
}
