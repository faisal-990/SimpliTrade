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
// A real provider is wrapped throttle→cache: the throttle (ratePerMin) keeps
// upstream calls under the free-tier limit; the cache (outer) serves repeats
// without waiting. If a real provider is selected without a key, we fall back to
// the deterministic FakeProvider so the app always boots.
func NewProvider(name, apiKey string, ratePerMin int) Provider {
	switch name {
	case "twelvedata":
		if apiKey == "" {
			return NewFakeProvider()
		}
		var inner Provider = NewTwelveDataProvider(apiKey, "", nil)
		if ratePerMin > 0 {
			inner = NewThrottledProvider(inner, time.Minute/time.Duration(ratePerMin))
		}
		return NewCachingProvider(inner, quoteCacheTTL, fundCacheTTL, candleCacheTTL)
	// case "finnhub": ...
	// case "fmp":     ...
	default:
		return NewFakeProvider()
	}
}

// WithFundamentals layers a dedicated fundamentals source over a price provider
// when the price feed doesn't supply fundamentals (e.g. Twelve Data's free tier).
// It returns base unchanged when no fundamentals provider/key is configured, so
// it's safe to call unconditionally. Fundamentals are cached aggressively (they
// move slowly) to stay within the source's free-tier budget.
func WithFundamentals(base Provider, name, apiKey string) Provider {
	if apiKey == "" {
		return base
	}
	switch name {
	case "finnhub":
		return NewCompositeProvider(base, NewFinnhubProvider(apiKey))
	default:
		return base
	}
}
