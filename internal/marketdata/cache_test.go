package marketdata

import (
	"context"
	"testing"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// countingProvider counts upstream calls so we can prove the cache absorbs them.
type countingProvider struct {
	quotes, batches, funds, candles int
}

func (c *countingProvider) Quote(context.Context, string) (Quote, error) {
	c.quotes++
	return Quote{Symbol: "AAPL", Price: 100}, nil
}
func (c *countingProvider) BatchQuotes(_ context.Context, syms []string) (map[string]Quote, error) {
	c.batches++
	out := map[string]Quote{}
	for _, s := range syms {
		out[s] = Quote{Symbol: s, Price: 100}
	}
	return out, nil
}
func (c *countingProvider) Fundamentals(context.Context, string) (models.Fundamentals, error) {
	c.funds++
	return models.Fundamentals{PE: 10}, nil
}
func (c *countingProvider) Candles(context.Context, string, string, time.Time, time.Time) ([]Candle, error) {
	c.candles++
	return []Candle{{Close: 100}}, nil
}

func TestCache_QuoteHitsUpstreamOnce(t *testing.T) {
	inner := &countingProvider{}
	c := NewCachingProvider(inner, time.Minute, time.Hour, time.Hour)

	for range 5 {
		if _, err := c.Quote(context.Background(), "AAPL"); err != nil {
			t.Fatal(err)
		}
	}
	if inner.quotes != 1 {
		t.Errorf("upstream Quote called %d times, want 1 (cached)", inner.quotes)
	}
}

func TestCache_ExpiryRefetches(t *testing.T) {
	inner := &countingProvider{}
	c := NewCachingProvider(inner, time.Minute, time.Hour, time.Hour)
	now := time.Unix(0, 0)
	c.now = func() time.Time { return now }

	_, _ = c.Quote(context.Background(), "AAPL")
	now = now.Add(2 * time.Minute) // past the quote TTL
	_, _ = c.Quote(context.Background(), "AAPL")

	if inner.quotes != 2 {
		t.Errorf("upstream Quote called %d times, want 2 (expired then refetched)", inner.quotes)
	}
}

func TestCache_BatchOnlyFetchesMisses(t *testing.T) {
	inner := &countingProvider{}
	c := NewCachingProvider(inner, time.Minute, time.Hour, time.Hour)
	ctx := context.Background()

	// Warm AAPL.
	_, _ = c.Quote(ctx, "AAPL")
	// Batch of AAPL (cached) + MSFT (miss): one upstream batch for just the miss.
	out, err := c.BatchQuotes(ctx, []string{"AAPL", "MSFT"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("batch returned %d, want 2", len(out))
	}
	if inner.batches != 1 {
		t.Errorf("upstream BatchQuotes called %d times, want 1", inner.batches)
	}
}

func TestCache_FundamentalsAndCandlesCached(t *testing.T) {
	inner := &countingProvider{}
	c := NewCachingProvider(inner, time.Minute, time.Hour, time.Hour)
	ctx := context.Background()
	for range 3 {
		_, _ = c.Fundamentals(ctx, "AAPL")
		_, _ = c.Candles(ctx, "AAPL", "1d", time.Unix(0, 0), time.Unix(100000, 0))
	}
	if inner.funds != 1 {
		t.Errorf("Fundamentals upstream %d, want 1", inner.funds)
	}
	if inner.candles != 1 {
		t.Errorf("Candles upstream %d, want 1", inner.candles)
	}
}
