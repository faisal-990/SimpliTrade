package marketdata

import (
	"context"
	"hash/fnv"
	"math"
	"math/rand"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// FakeProvider is a fully deterministic Provider for development and tests: it
// needs no network or API key. Fundamentals are a stable function of the symbol;
// quotes/candles evolve with time but are reproducible given the same clock, so
// tests can assert exact values by injecting a fixed clock.
type FakeProvider struct {
	now func() time.Time
}

// NewFakeProvider returns a FakeProvider driven by the wall clock.
func NewFakeProvider() *FakeProvider {
	return &FakeProvider{now: time.Now}
}

// WithClock overrides the time source (tests inject a fixed clock for
// deterministic quotes). Returns the receiver for chaining.
func (f *FakeProvider) WithClock(now func() time.Time) *FakeProvider {
	f.now = now
	return f
}

// seed derives a stable 64-bit seed from a symbol.
func seed(symbol string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(symbol))
	return h.Sum64()
}

// basePrice maps a symbol to a stable price in roughly [20, 500].
func basePrice(symbol string) float64 {
	return 20 + float64(seed(symbol)%48000)/100.0
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func (f *FakeProvider) Quote(_ context.Context, symbol string) (Quote, error) {
	base := basePrice(symbol)
	now := f.now()

	// Deterministic intraday oscillation: a slow sine keyed by the symbol so
	// every symbol moves on its own phase. Given a fixed clock this is exact.
	phase := float64(now.Unix()) / 300.0 // ~5-minute period scaling
	wiggle := 1 + 0.03*math.Sin(phase+float64(seed(symbol)%360))
	price := round2(base * wiggle)

	open := round2(base)
	high := round2(math.Max(price, open) * 1.004)
	low := round2(math.Min(price, open) * 0.996)
	volume := int64(100_000 + seed(symbol)%5_000_000)

	return Quote{
		Symbol: symbol,
		Price:  price,
		Open:   open,
		High:   high,
		Low:    low,
		Volume: volume,
		Time:   now,
	}, nil
}

func (f *FakeProvider) BatchQuotes(ctx context.Context, symbols []string) (map[string]Quote, error) {
	out := make(map[string]Quote, len(symbols))
	for _, s := range symbols {
		q, err := f.Quote(ctx, s)
		if err != nil {
			return nil, err
		}
		out[s] = q
	}
	return out, nil
}

func (f *FakeProvider) Fundamentals(_ context.Context, symbol string) (models.Fundamentals, error) {
	// A fresh RNG seeded by the symbol, read in a fixed field order, yields the
	// same fundamentals on every call for a given symbol.
	r := rand.New(rand.NewSource(int64(seed(symbol)))) //nolint:gosec // deterministic test data, not security
	price := basePrice(symbol)

	eps := price / (5 + r.Float64()*35) // implies P/E 5–40
	bvps := price / (0.5 + r.Float64()*4)
	pe := round2(price / eps)
	pb := round2(price / bvps)
	revGrowth := round4(-0.10 + r.Float64()*0.60)
	epsGrowth := round4(-0.10 + r.Float64()*0.60)

	return models.Fundamentals{
		PE:            pe,
		ForwardPE:     round2(pe * (0.8 + r.Float64()*0.4)),
		PB:            pb,
		PS:            round2(0.5 + r.Float64()*9.5),
		PEG:           round2(pe / (math.Max(epsGrowth, 0.01) * 100)),
		EVEBITDA:      round2(4 + r.Float64()*16),
		EarningsYield: round4(1 / pe),
		FCFYield:      round4(0.01 + r.Float64()*0.09),
		DividendYield: round4(r.Float64() * 0.05),
		EPSTTM:        round2(eps),
		BVPS:          round2(bvps),

		ROE:             round4(0.05 + r.Float64()*0.30),
		ROIC:            round4(0.04 + r.Float64()*0.26),
		GrossMargin:     round4(0.20 + r.Float64()*0.60),
		OperatingMargin: round4(0.05 + r.Float64()*0.35),
		NetMargin:       round4(0.02 + r.Float64()*0.28),
		DebtToEquity:    round2(r.Float64() * 2.0),
		CurrentRatio:    round2(0.5 + r.Float64()*3.0),
		InterestCover:   round2(1 + r.Float64()*14),
		FCFPositive:     r.Float64() > 0.2,

		RevenueGrowthYoY: revGrowth,
		EPSGrowthYoY:     epsGrowth,
		RevenueCAGR3Y:    round4(-0.05 + r.Float64()*0.45),
		EPSGrowth5Y:      round4(-0.05 + r.Float64()*0.55),

		EPSPositiveYears: r.Intn(16),
		DividendYears:    r.Intn(26),
		Beta:             round2(0.5 + r.Float64()*1.5),
		MarketCap:        round2(1e9 * math.Pow(10, r.Float64()*3.3)), // ~$1B–$2T
	}, nil
}

func (f *FakeProvider) Candles(_ context.Context, symbol, interval string, from, to time.Time) ([]Candle, error) {
	if !to.After(from) {
		return nil, nil
	}
	base := basePrice(symbol)
	s := seed(symbol)

	var candles []Candle
	prevClose := base
	for day := from; day.Before(to); day = day.AddDate(0, 0, 1) {
		dayIdx := day.Unix() / 86400
		// Slow deterministic drift plus a small per-day jitter.
		drift := math.Sin(float64(dayIdx)/15.0 + float64(s%360))
		jitter := math.Sin(float64(dayIdx) * float64(s%7+1))
		closePx := round2(base * (1 + 0.12*drift + 0.01*jitter))
		open := round2(prevClose)
		high := round2(math.Max(open, closePx) * 1.01)
		low := round2(math.Min(open, closePx) * 0.99)

		candles = append(candles, Candle{
			Time:     day,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    closePx,
			Volume:   int64(100_000 + (uint64(dayIdx)^s)%5_000_000),
			Interval: interval,
		})
		prevClose = closePx
	}
	return candles, nil
}

func round4(v float64) float64 { return math.Round(v*10000) / 10000 }
