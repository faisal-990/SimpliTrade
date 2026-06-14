package marketdata

import (
	"context"
	"sync"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// CachingProvider wraps a Provider with per-symbol TTL caches. Under a free-tier
// quota this is essential: fundamentals change daily, quotes every few seconds,
// so we cache each accordingly and only call the upstream API on a miss. Batch
// quote requests fetch only the symbols whose cache has expired.
type CachingProvider struct {
	inner     Provider
	quoteTTL  time.Duration
	fundTTL   time.Duration
	candleTTL time.Duration
	now       func() time.Time

	mu      sync.Mutex
	quotes  map[string]cacheEntry[Quote]
	funds   map[string]cacheEntry[models.Fundamentals]
	candles map[string]cacheEntry[[]Candle]
}

type cacheEntry[T any] struct {
	val T
	at  time.Time
}

// NewCachingProvider wraps inner with the given TTLs.
func NewCachingProvider(inner Provider, quoteTTL, fundTTL, candleTTL time.Duration) *CachingProvider {
	return &CachingProvider{
		inner: inner, quoteTTL: quoteTTL, fundTTL: fundTTL, candleTTL: candleTTL,
		now:     time.Now,
		quotes:  map[string]cacheEntry[Quote]{},
		funds:   map[string]cacheEntry[models.Fundamentals]{},
		candles: map[string]cacheEntry[[]Candle]{},
	}
}

func (c *CachingProvider) fresh(at time.Time, ttl time.Duration) bool {
	return c.now().Sub(at) < ttl
}

func (c *CachingProvider) Quote(ctx context.Context, symbol string) (Quote, error) {
	c.mu.Lock()
	if e, ok := c.quotes[symbol]; ok && c.fresh(e.at, c.quoteTTL) {
		c.mu.Unlock()
		return e.val, nil
	}
	c.mu.Unlock()

	q, err := c.inner.Quote(ctx, symbol)
	if err != nil {
		return Quote{}, err
	}
	c.mu.Lock()
	c.quotes[symbol] = cacheEntry[Quote]{q, c.now()}
	c.mu.Unlock()
	return q, nil
}

func (c *CachingProvider) BatchQuotes(ctx context.Context, symbols []string) (map[string]Quote, error) {
	out := make(map[string]Quote, len(symbols))
	var misses []string

	c.mu.Lock()
	for _, s := range symbols {
		if e, ok := c.quotes[s]; ok && c.fresh(e.at, c.quoteTTL) {
			out[s] = e.val
		} else {
			misses = append(misses, s)
		}
	}
	c.mu.Unlock()

	if len(misses) == 0 {
		return out, nil
	}
	fetched, err := c.inner.BatchQuotes(ctx, misses)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	now := c.now()
	for s, q := range fetched {
		c.quotes[s] = cacheEntry[Quote]{q, now}
		out[s] = q
	}
	c.mu.Unlock()
	return out, nil
}

func (c *CachingProvider) Fundamentals(ctx context.Context, symbol string) (models.Fundamentals, error) {
	c.mu.Lock()
	if e, ok := c.funds[symbol]; ok && c.fresh(e.at, c.fundTTL) {
		c.mu.Unlock()
		return e.val, nil
	}
	c.mu.Unlock()

	f, err := c.inner.Fundamentals(ctx, symbol)
	if err != nil {
		return models.Fundamentals{}, err
	}
	c.mu.Lock()
	c.funds[symbol] = cacheEntry[models.Fundamentals]{f, c.now()}
	c.mu.Unlock()
	return f, nil
}

func (c *CachingProvider) Candles(ctx context.Context, symbol, interval string, from, to time.Time) ([]Candle, error) {
	key := symbol + "|" + interval
	c.mu.Lock()
	if e, ok := c.candles[key]; ok && c.fresh(e.at, c.candleTTL) {
		c.mu.Unlock()
		return e.val, nil
	}
	c.mu.Unlock()

	cs, err := c.inner.Candles(ctx, symbol, interval, from, to)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.candles[key] = cacheEntry[[]Candle]{cs, c.now()}
	c.mu.Unlock()
	return cs, nil
}
