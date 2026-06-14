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
// needs no network or API key. Each symbol is assigned a believable *archetype*
// (value / quality / growth / momentum / cyclical) and gets self-consistent
// fundamentals AND a matching price path — so the real strategies actually find
// candidates and trade, rather than a uniform-random soup that passes nothing.
type FakeProvider struct {
	now func() time.Time
}

func NewFakeProvider() *FakeProvider { return &FakeProvider{now: time.Now} }

func (f *FakeProvider) WithClock(now func() time.Time) *FakeProvider {
	f.now = now
	return f
}

func seed(symbol string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(symbol))
	return h.Sum64()
}

func basePrice(symbol string) float64 {
	return 30 + float64(seed(symbol)%47000)/100.0 // ~$30–$500
}

type archetype int

const (
	archValue archetype = iota
	archQuality
	archGrowth
	archMomentum
	archCyclical
)

// archetypeOf distributes symbols ~35% value, 20% quality, 15% growth,
// 15% momentum, 15% cyclical — a spread that gives every strategy family names.
func archetypeOf(symbol string) archetype {
	switch b := seed(symbol+"|arch") % 100; {
	case b < 35:
		return archValue
	case b < 55:
		return archQuality
	case b < 70:
		return archGrowth
	case b < 85:
		return archMomentum
	default:
		return archCyclical
	}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
func round4(v float64) float64 { return math.Round(v*10000) / 10000 }

// pricePath returns a symbol's price at a point in [0,1] of a ~2-year window
// (0 = oldest, 1 = now). Uptrending archetypes peak near `progress=1`, so the
// latest quote sits near the 52-week high — which is what momentum strategies
// look for. A small deterministic wiggle gives candles texture.
func pricePath(symbol string, a archetype, progress float64) float64 {
	base := basePrice(symbol)
	s := seed(symbol)
	wig := 1 + 0.015*math.Sin(progress*18+float64(s%19))
	var mult float64
	switch a {
	case archValue:
		mult = 0.95 + 0.06*math.Sin(progress*5+float64(s%7)) // flat-ish, slightly cheap now
	case archQuality:
		mult = 0.72 + 0.40*progress // steady compounding
	case archGrowth:
		mult = 0.50 + 0.72*progress // strong rise
	case archMomentum:
		mult = 0.42 + 0.82*progress // strongest; near highs now
	case archCyclical:
		mult = 1 + 0.28*math.Sin(progress*4*math.Pi+float64(s%5))
	}
	return round2(base * mult * wig)
}

// intradayWiggle is a small (<0.5%) deterministic oscillation keyed to the time
// of day, so quotes move through the session — the underlying price paths are
// daily, this animates them intraday for a realistic live feel.
func intradayWiggle(symbol string, t time.Time) float64 {
	mins := float64(t.Hour()*60 + t.Minute())
	return 1 + 0.004*math.Sin(mins/47+float64(seed(symbol)%13))
}

func (f *FakeProvider) Quote(_ context.Context, symbol string) (Quote, error) {
	a := archetypeOf(symbol)
	now := f.now()
	price := round2(pricePath(symbol, a, 1) * intradayWiggle(symbol, now)) // "now", animated intraday
	prev := pricePath(symbol, a, 0.997)
	high := round2(math.Max(price, prev) * 1.004)
	low := round2(math.Min(price, prev) * 0.996)
	return Quote{
		Symbol: symbol, Price: price, Open: round2(prev),
		High: high, Low: low,
		Volume: int64(100_000 + seed(symbol)%5_000_000),
		Time:   f.now(),
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
	r := rand.New(rand.NewSource(int64(seed(symbol)))) //nolint:gosec // deterministic test data
	a := archetypeOf(symbol)
	price := pricePath(symbol, a, 1)

	switch a {
	case archValue:
		// Cheap, safe, long dividend record, and priced below intrinsic (passes MOS).
		eps := price / (7 + r.Float64()*5)    // P/E ~7–12
		bvps := price / (0.8 + r.Float64()*0.6) // P/B ~0.8–1.4
		return models.Fundamentals{
			PE: round2(price / eps), ForwardPE: round2(price / eps * 0.95),
			PB: round2(price / bvps), PS: round2(0.6 + r.Float64()*1.4),
			PEG: round2(0.6 + r.Float64()*0.6), EVEBITDA: round2(5 + r.Float64()*4),
			EarningsYield: round4(eps / price), FCFYield: round4(0.06 + r.Float64()*0.05),
			DividendYield: round4(0.02 + r.Float64()*0.03), EPSTTM: round2(eps), BVPS: round2(bvps),
			ROE: round4(0.10 + r.Float64()*0.12), ROIC: round4(0.12 + r.Float64()*0.14),
			GrossMargin: round4(0.30 + r.Float64()*0.20), OperatingMargin: round4(0.12 + r.Float64()*0.10),
			NetMargin: round4(0.08 + r.Float64()*0.08), DebtToEquity: round2(0.1 + r.Float64()*0.3),
			CurrentRatio: round2(2.0 + r.Float64()*1.5), InterestCover: round2(6 + r.Float64()*10),
			FCFPositive: true, RevenueGrowthYoY: round4(0.02 + r.Float64()*0.08),
			EPSGrowthYoY: round4(0.03 + r.Float64()*0.08), RevenueCAGR3Y: round4(0.03 + r.Float64()*0.06),
			EPSGrowth5Y: round4(0.04 + r.Float64()*0.06), EPSPositiveYears: 12 + r.Intn(8),
			DividendYears: 20 + r.Intn(12), Beta: round2(0.7 + r.Float64()*0.4),
			MarketCap: round2(3e9 * math.Pow(10, r.Float64()*2.5)),
		}, nil

	case archQuality:
		// High returns on capital, fat margins, fair price, comfortable intrinsic discount.
		eps := price / (16 + r.Float64()*8) // P/E ~16–24
		bvps := price / (3 + r.Float64()*4)
		return models.Fundamentals{
			PE: round2(price / eps), ForwardPE: round2(price / eps * 0.92),
			PB: round2(price / bvps), PS: round2(3 + r.Float64()*4),
			PEG: round2(1.0 + r.Float64()*0.8), EVEBITDA: round2(12 + r.Float64()*6),
			EarningsYield: round4(eps / price), FCFYield: round4(0.035 + r.Float64()*0.03),
			DividendYield: round4(r.Float64() * 0.02), EPSTTM: round2(eps), BVPS: round2(bvps),
			ROE: round4(0.20 + r.Float64()*0.20), ROIC: round4(0.18 + r.Float64()*0.16),
			GrossMargin: round4(0.52 + r.Float64()*0.22), OperatingMargin: round4(0.22 + r.Float64()*0.16),
			NetMargin: round4(0.18 + r.Float64()*0.12), DebtToEquity: round2(0.2 + r.Float64()*0.6),
			CurrentRatio: round2(1.3 + r.Float64()*1.2), InterestCover: round2(8 + r.Float64()*12),
			FCFPositive: true, RevenueGrowthYoY: round4(0.08 + r.Float64()*0.10),
			EPSGrowthYoY: round4(0.10 + r.Float64()*0.12), RevenueCAGR3Y: round4(0.08 + r.Float64()*0.08),
			EPSGrowth5Y: round4(0.10 + r.Float64()*0.10), EPSPositiveYears: 9 + r.Intn(8),
			DividendYears: r.Intn(15), Beta: round2(0.8 + r.Float64()*0.4),
			MarketCap: round2(2e10 * math.Pow(10, r.Float64()*2)),
		}, nil

	case archGrowth:
		// Fast revenue/EPS growth, premium multiple, GARP-friendly PEG.
		eps := price / (24 + r.Float64()*20)
		bvps := price / (6 + r.Float64()*8)
		return models.Fundamentals{
			PE: round2(price / eps), ForwardPE: round2(price / eps * 0.8),
			PB: round2(price / bvps), PS: round2(6 + r.Float64()*10),
			PEG: round2(0.7 + r.Float64()*0.3), EVEBITDA: round2(18 + r.Float64()*12),
			EarningsYield: round4(eps / price), FCFYield: round4(0.01 + r.Float64()*0.03),
			DividendYield: 0, EPSTTM: round2(eps), BVPS: round2(bvps),
			ROE: round4(0.15 + r.Float64()*0.20), ROIC: round4(0.12 + r.Float64()*0.16),
			GrossMargin: round4(0.45 + r.Float64()*0.30), OperatingMargin: round4(0.15 + r.Float64()*0.20),
			NetMargin: round4(0.10 + r.Float64()*0.15), DebtToEquity: round2(0.1 + r.Float64()*0.5),
			CurrentRatio: round2(1.5 + r.Float64()*1.5), InterestCover: round2(6 + r.Float64()*12),
			FCFPositive: r.Float64() > 0.3, RevenueGrowthYoY: round4(0.25 + r.Float64()*0.30),
			EPSGrowthYoY: round4(0.20 + r.Float64()*0.30), RevenueCAGR3Y: round4(0.22 + r.Float64()*0.20),
			EPSGrowth5Y: round4(0.25 + r.Float64()*0.25), EPSPositiveYears: 4 + r.Intn(8),
			DividendYears: 0, Beta: round2(1.1 + r.Float64()*0.7),
			MarketCap: round2(5e9 * math.Pow(10, r.Float64()*2.3)),
		}, nil

	default: // momentum / cyclical — neutral fundamentals; these trade on price action
		eps := price / (15 + r.Float64()*25)
		bvps := price / (1.5 + r.Float64()*5)
		return models.Fundamentals{
			PE: round2(price / eps), ForwardPE: round2(price / eps),
			PB: round2(price / bvps), PS: round2(1 + r.Float64()*8),
			PEG: round2(0.8 + r.Float64()*2), EVEBITDA: round2(6 + r.Float64()*14),
			EarningsYield: round4(eps / price), FCFYield: round4(0.02 + r.Float64()*0.05),
			DividendYield: round4(r.Float64() * 0.03), EPSTTM: round2(eps), BVPS: round2(bvps),
			ROE: round4(0.06 + r.Float64()*0.24), ROIC: round4(0.05 + r.Float64()*0.22),
			GrossMargin: round4(0.25 + r.Float64()*0.45), OperatingMargin: round4(0.08 + r.Float64()*0.25),
			NetMargin: round4(0.04 + r.Float64()*0.18), DebtToEquity: round2(r.Float64() * 1.8),
			CurrentRatio: round2(0.8 + r.Float64()*2.5), InterestCover: round2(2 + r.Float64()*12),
			FCFPositive: r.Float64() > 0.3, RevenueGrowthYoY: round4(-0.05 + r.Float64()*0.4),
			EPSGrowthYoY: round4(-0.05 + r.Float64()*0.4), RevenueCAGR3Y: round4(r.Float64() * 0.3),
			EPSGrowth5Y: round4(r.Float64() * 0.35), EPSPositiveYears: r.Intn(14),
			DividendYears: r.Intn(20), Beta: round2(0.9 + r.Float64()*1.2),
			MarketCap: round2(2e9 * math.Pow(10, r.Float64()*3)),
		}, nil
	}
}

func (f *FakeProvider) Candles(_ context.Context, symbol, interval string, from, to time.Time) ([]Candle, error) {
	if !to.After(from) {
		return nil, nil
	}
	a := archetypeOf(symbol)
	span := to.Sub(from).Hours() / 24
	if span < 1 {
		span = 1
	}
	var candles []Candle
	dayNum := 0.0
	prevClose := pricePath(symbol, a, 0)
	for day := from; day.Before(to); day = day.AddDate(0, 0, 1) {
		progress := dayNum / span
		closePx := pricePath(symbol, a, progress)
		open := round2(prevClose)
		high := round2(math.Max(open, closePx) * 1.008)
		low := round2(math.Min(open, closePx) * 0.992)
		candles = append(candles, Candle{
			Time: day, Open: open, High: high, Low: low, Close: closePx,
			Volume: int64(100_000 + (uint64(dayNum)*7^seed(symbol))%5_000_000), Interval: "1d",
		})
		prevClose = closePx
		dayNum++
	}
	return candles, nil
}
