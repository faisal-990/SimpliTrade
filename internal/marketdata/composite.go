package marketdata

import (
	"context"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// CompositeProvider serves prices and candles from one provider and fundamentals
// from another. It exists because no single free tier covers everything: Twelve
// Data gives real-time prices but gates fundamentals, while Finnhub gives
// fundamentals. The composite presents both as one Provider, so the seeder and
// engine stay vendor-agnostic.
type CompositeProvider struct {
	prices       Provider
	fundamentals FundamentalsSource
}

func NewCompositeProvider(prices Provider, fundamentals FundamentalsSource) *CompositeProvider {
	return &CompositeProvider{prices: prices, fundamentals: fundamentals}
}

func (c *CompositeProvider) Quote(ctx context.Context, symbol string) (Quote, error) {
	return c.prices.Quote(ctx, symbol)
}

func (c *CompositeProvider) BatchQuotes(ctx context.Context, symbols []string) (map[string]Quote, error) {
	return c.prices.BatchQuotes(ctx, symbols)
}

func (c *CompositeProvider) Candles(ctx context.Context, symbol, interval string, from, to time.Time) ([]Candle, error) {
	return c.prices.Candles(ctx, symbol, interval, from, to)
}

func (c *CompositeProvider) Fundamentals(ctx context.Context, symbol string) (models.Fundamentals, error) {
	return c.fundamentals.Fundamentals(ctx, symbol)
}
