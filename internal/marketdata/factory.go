package marketdata

import "time"

// Provider TTLs for the caching wrapper around a real provider. Fundamentals
// barely move day to day; quotes are cached briefly to collapse bursts without
// going stale; candle history changes once a day.
const (
	quoteCacheTTL  = 15 * time.Second
	fundCacheTTL   = 24 * time.Hour
	candleCacheTTL = 6 * time.Hour
)

// NewProvider is the single switch point for choosing a market-data source.
// Everything else in the app depends on the Provider interface, not a concrete
// vendor — so adding a new provider is one new Provider implementation plus a
// case here, with no other code changes.
//
// Real providers are wrapped in a CachingProvider to respect free-tier quotas.
// If a real provider is selected without a key, we fall back to the deterministic
// FakeProvider so the app always boots.
func NewProvider(name, apiKey string) Provider {
	switch name {
	case "twelvedata":
		if apiKey == "" {
			return NewFakeProvider()
		}
		real := NewTwelveDataProvider(apiKey, "", nil)
		return NewCachingProvider(real, quoteCacheTTL, fundCacheTTL, candleCacheTTL)
	// case "finnhub": return NewCachingProvider(NewFinnhubProvider(apiKey, "", nil), ...)
	// case "fmp":     return NewCachingProvider(NewFMPProvider(apiKey, "", nil), ...)
	default:
		return NewFakeProvider()
	}
}
