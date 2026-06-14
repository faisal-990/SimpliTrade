package marketdata

import (
	"context"
	"sync"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// ThrottledProvider enforces a minimum interval between upstream calls so we stay
// within a provider's free-tier rate limit (e.g. 8 req/min). It serializes calls
// and spaces them out; it wraps the real provider *inside* the cache, so cache
// hits never wait — only genuine upstream calls are throttled.
type ThrottledProvider struct {
	inner       Provider
	minInterval time.Duration
	mu          sync.Mutex
	last        time.Time
}

// NewThrottledProvider limits inner to one call per minInterval.
func NewThrottledProvider(inner Provider, minInterval time.Duration) *ThrottledProvider {
	return &ThrottledProvider{inner: inner, minInterval: minInterval}
}

// wait blocks until the next call is allowed, or the context is cancelled.
func (t *ThrottledProvider) wait(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.last.IsZero() {
		if d := t.minInterval - time.Since(t.last); d > 0 {
			select {
			case <-time.After(d):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	t.last = time.Now()
	return nil
}

func (t *ThrottledProvider) Quote(ctx context.Context, symbol string) (Quote, error) {
	if err := t.wait(ctx); err != nil {
		return Quote{}, err
	}
	return t.inner.Quote(ctx, symbol)
}

func (t *ThrottledProvider) BatchQuotes(ctx context.Context, symbols []string) (map[string]Quote, error) {
	if err := t.wait(ctx); err != nil {
		return nil, err
	}
	return t.inner.BatchQuotes(ctx, symbols)
}

func (t *ThrottledProvider) Fundamentals(ctx context.Context, symbol string) (models.Fundamentals, error) {
	if err := t.wait(ctx); err != nil {
		return models.Fundamentals{}, err
	}
	return t.inner.Fundamentals(ctx, symbol)
}

func (t *ThrottledProvider) Candles(ctx context.Context, symbol, interval string, from, to time.Time) ([]Candle, error) {
	if err := t.wait(ctx); err != nil {
		return nil, err
	}
	return t.inner.Candles(ctx, symbol, interval, from, to)
}
