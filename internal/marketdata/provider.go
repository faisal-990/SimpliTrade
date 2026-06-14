// Package marketdata is the boundary to market data (quotes, fundamentals,
// candles). Everything downstream depends on the Provider interface, never on a
// concrete source — so the deterministic FakeProvider used in development and
// tests can be swapped for a real API (Finnhub/AlphaVantage) at T9 with no
// changes to callers. Only the engine (Tower 2) ever calls a Provider, keeping
// user traffic decoupled from external API quotas.
package marketdata

import (
	"context"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// Quote is a point-in-time price snapshot for a symbol.
type Quote struct {
	Symbol string    `json:"symbol"`
	Price  float64   `json:"price"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Volume int64     `json:"volume"`
	Time   time.Time `json:"time"`
}

// Candle is a single OHLCV bar for a given interval.
type Candle struct {
	Time     time.Time `json:"time"`
	Open     float64   `json:"open"`
	High     float64   `json:"high"`
	Low      float64   `json:"low"`
	Close    float64   `json:"close"`
	Volume   int64     `json:"volume"`
	Interval string    `json:"interval"`
}

// Security is a tradable instrument's static identity (the universe entry).
type Security struct {
	Symbol   string
	Name     string
	Sector   string
	Exchange string
}

// Provider supplies market data. BatchQuotes exists because real providers bill
// and rate-limit per request: the engine fetches all symbols in one call rather
// than N, which is essential under free-tier quotas.
type Provider interface {
	Quote(ctx context.Context, symbol string) (Quote, error)
	BatchQuotes(ctx context.Context, symbols []string) (map[string]Quote, error)
	Fundamentals(ctx context.Context, symbol string) (models.Fundamentals, error)
	Candles(ctx context.Context, symbol, interval string, from, to time.Time) ([]Candle, error)
}
