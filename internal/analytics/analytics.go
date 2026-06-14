// Package analytics turns a user's trade history + price candles into portfolio
// performance metrics: a reconstructed equity curve, risk stats (drawdown,
// volatility, Sharpe), a market benchmark, and sector exposure. It is pure (no
// I/O): the caller loads the data and hands it in, mirroring internal/backtest.
package analytics

import (
	"math"
	"sort"
	"time"
)

// Bar is one daily close for a symbol.
type Bar struct {
	Date  time.Time
	Close float64
}

// Trade is one executed order in the account's history.
type Trade struct {
	Date     time.Time
	Symbol   string
	Side     string // "buy" | "sell"
	Quantity float64
	Price    float64
}

// Holding is a current position, used for sector exposure.
type Holding struct {
	Symbol      string
	Sector      string
	MarketValue float64
}

// Input is everything Compute needs.
type Input struct {
	StartCash        float64
	Trades           []Trade
	Histories        map[string][]Bar // daily closes per symbol (traded ∪ benchmark)
	BenchmarkSymbols []string         // equal-weighted into the "market" line
	Holdings         []Holding        // current positions, for sector breakdown
}

// Point is one value on a time series.
type Point struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// SectorSlice is one sector's share of the current portfolio.
type SectorSlice struct {
	Sector string  `json:"sector"`
	Value  float64 `json:"value"`
	Pct    float64 `json:"pct"`
}

// Result is the computed analytics.
type Result struct {
	Equity        []Point
	Benchmark     []Point
	ROI           float64
	MaxDrawdown   float64
	Volatility    float64 // annualized
	Sharpe        float64 // annualized, rf=0
	WinRate       float64
	TradeCount    int
	WinningTrades int
	BestDay       float64
	WorstDay      float64
	Sectors       []SectorSlice
}

type series struct {
	days   []int64
	closes []float64
}

func (s *series) asOf(day int64) float64 {
	i := sort.Search(len(s.days), func(i int) bool { return s.days[i] > day }) - 1
	if i < 0 {
		return 0
	}
	return s.closes[i]
}

func dayKey(t time.Time) int64 {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix()
}

// Compute reconstructs the portfolio's daily equity and derives metrics.
func Compute(in Input) Result {
	if in.StartCash <= 0 {
		in.StartCash = 100000
	}

	// Index histories and build the master timeline.
	bySym := make(map[string]*series, len(in.Histories))
	dateSet := map[int64]time.Time{}
	for sym, bars := range in.Histories {
		bb := append([]Bar(nil), bars...)
		sort.Slice(bb, func(i, j int) bool { return bb[i].Date.Before(bb[j].Date) })
		s := &series{}
		seen := map[int64]struct{}{}
		for _, b := range bb {
			dk := dayKey(b.Date)
			if _, dup := seen[dk]; dup {
				continue
			}
			seen[dk] = struct{}{}
			s.days = append(s.days, dk)
			s.closes = append(s.closes, b.Close)
			dateSet[dk] = time.Unix(dk, 0).UTC()
		}
		if len(s.days) > 0 {
			bySym[sym] = s
		}
	}
	timeline := make([]time.Time, 0, len(dateSet))
	for _, t := range dateSet {
		timeline = append(timeline, t)
	}
	sort.Slice(timeline, func(i, j int) bool { return timeline[i].Before(timeline[j]) })

	res := Result{ROI: 0}
	res.Sectors = sectorBreakdown(in.Holdings)
	if len(timeline) == 0 {
		res.Equity = []Point{{Value: in.StartCash}}
		return res
	}

	// Sort trades ascending so we can apply them as the timeline advances.
	trades := append([]Trade(nil), in.Trades...)
	sort.Slice(trades, func(i, j int) bool { return trades[i].Date.Before(trades[j].Date) })

	cash := in.StartCash
	positions := map[string]float64{}
	avgCost := map[string]float64{}
	ti := 0

	// Benchmark: equal-weight of BenchmarkSymbols, rebased to StartCash.
	base := map[string]float64{}
	for _, sym := range in.BenchmarkSymbols {
		if s := bySym[sym]; s != nil && len(s.closes) > 0 {
			base[sym] = s.closes[0]
		}
	}

	for _, d := range timeline {
		dk := dayKey(d)
		// Apply trades up to and including this day, tracking realized wins.
		for ti < len(trades) && dayKey(trades[ti].Date) <= dk {
			tr := trades[ti]
			ti++
			if tr.Side == "sell" {
				if tr.Price > avgCost[tr.Symbol] {
					res.WinningTrades++
				}
				res.TradeCount++ // count closing trades for win-rate denominator
				cash += tr.Quantity * tr.Price
				positions[tr.Symbol] -= tr.Quantity
				if positions[tr.Symbol] <= 1e-9 {
					delete(positions, tr.Symbol)
				}
			} else {
				prevQty := positions[tr.Symbol]
				total := prevQty + tr.Quantity
				if total > 0 {
					avgCost[tr.Symbol] = (avgCost[tr.Symbol]*prevQty + tr.Price*tr.Quantity) / total
				}
				positions[tr.Symbol] = total
				cash -= tr.Quantity * tr.Price
			}
		}

		equity := cash
		for sym, qty := range positions {
			if s := bySym[sym]; s != nil {
				equity += qty * s.asOf(dk)
			}
		}
		res.Equity = append(res.Equity, Point{Date: d, Value: equity})

		// Benchmark value this day.
		if len(base) > 0 {
			sum, n := 0.0, 0
			for sym, b0 := range base {
				if b0 > 0 {
					if c := bySym[sym].asOf(dk); c > 0 {
						sum += c / b0
						n++
					}
				}
			}
			if n > 0 {
				res.Benchmark = append(res.Benchmark, Point{Date: d, Value: in.StartCash * sum / float64(n)})
			}
		}
	}

	// Metrics from the equity curve.
	last := res.Equity[len(res.Equity)-1].Value
	res.ROI = last/in.StartCash - 1

	peak := res.Equity[0].Value
	var rets []float64
	for i, p := range res.Equity {
		if p.Value > peak {
			peak = p.Value
		}
		if peak > 0 {
			if dd := (peak - p.Value) / peak; dd > res.MaxDrawdown {
				res.MaxDrawdown = dd
			}
		}
		if i > 0 {
			prev := res.Equity[i-1].Value
			if prev > 0 {
				r := p.Value/prev - 1
				rets = append(rets, r)
				if r > res.BestDay {
					res.BestDay = r
				}
				if r < res.WorstDay {
					res.WorstDay = r
				}
			}
		}
	}
	mean, sd := meanStd(rets)
	res.Volatility = sd * math.Sqrt(252)
	if sd > 0 {
		res.Sharpe = mean / sd * math.Sqrt(252)
	}
	if res.TradeCount > 0 {
		res.WinRate = float64(res.WinningTrades) / float64(res.TradeCount)
	}
	return res
}

func sectorBreakdown(holdings []Holding) []SectorSlice {
	bySector := map[string]float64{}
	total := 0.0
	for _, h := range holdings {
		sec := h.Sector
		if sec == "" {
			sec = "Other"
		}
		bySector[sec] += h.MarketValue
		total += h.MarketValue
	}
	out := make([]SectorSlice, 0, len(bySector))
	for sec, v := range bySector {
		pct := 0.0
		if total > 0 {
			pct = v / total
		}
		out = append(out, SectorSlice{Sector: sec, Value: v, Pct: pct})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Value > out[j].Value })
	return out
}

func meanStd(xs []float64) (float64, float64) {
	if len(xs) == 0 {
		return 0, 0
	}
	var sum float64
	for _, x := range xs {
		sum += x
	}
	mean := sum / float64(len(xs))
	var v float64
	for _, x := range xs {
		v += (x - mean) * (x - mean)
	}
	return mean, math.Sqrt(v / float64(len(xs)))
}
