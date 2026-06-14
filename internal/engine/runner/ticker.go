package runner

import (
	"context"
	"math/rand"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
)

// SyntheticTicker advances the simulated market by one step: it nudges every
// stock's price by a small random walk and appends a fresh candle.
//
// The admin "simulate market" control uses it so each cycle actually MOVES the
// market. Without it, the engine decides against unchanged prices: bots buy once
// and then freeze — no new trades (their positions already match the target) and
// no mark-to-market P&L (value == cost), so every investor's ROI sticks at 0.
// A synthetic tick keeps the sandbox alive — prices drift, positions gain/lose,
// strategies react — without calling a real, rate-limited feed (which also can't
// move prices while the real market is closed).
type SyntheticTicker struct {
	stocks repository.StockRepo
	vol    float64 // per-tick volatility: std-dev of the proportional price shock
	rng    *rand.Rand
}

// NewSyntheticTicker builds a ticker. vol is the per-cycle volatility (e.g. 0.02
// ≈ 2% typical move); seed makes a session reproducible.
func NewSyntheticTicker(stocks repository.StockRepo, vol float64, seed int64) *SyntheticTicker {
	if vol <= 0 {
		vol = 0.02
	}
	return &SyntheticTicker{stocks: stocks, vol: vol, rng: rand.New(rand.NewSource(seed))}
}

// reversion is how strongly each step pulls the price back toward its anchor
// (the seeded daily close). Without this, a plain multiplicative random walk
// compounds and a few names drift to absurd prices over many cycles.
const reversion = 0.15

// Refresh applies one market step to every priced stock. It moves the CURRENT
// price only — it deliberately does NOT append to the daily candle history,
// because that history is the substrate for charts and backtests and must stay a
// clean, contiguous record.
//
// The step is a mean-reverting (Ornstein–Uhlenbeck-style) walk: a symmetric
// shock plus a pull back toward the stock's anchor price, then clamped to a band
// around the anchor. This keeps prices oscillating realistically instead of
// running away, so ROI and the leaderboard stay sane.
func (t *SyntheticTicker) Refresh(ctx context.Context) error {
	stocks, err := t.stocks.List(ctx, 1000, 0)
	if err != nil {
		return err
	}
	anchors, err := t.stocks.LatestCloses(ctx)
	if err != nil {
		return err
	}
	for _, s := range stocks {
		prev := s.CurrentPrice
		if prev <= 0 {
			continue
		}
		anchor := anchors[s.Symbol]
		if anchor <= 0 {
			anchor = prev
		}
		shock := t.rng.NormFloat64() * t.vol
		next := prev + reversion*(anchor-prev) + prev*shock
		// Keep within a sane band around the anchor.
		if lo := anchor * 0.4; next < lo {
			next = lo
		}
		if hi := anchor * 2.5; next > hi {
			next = hi
		}
		if next < 0.01 {
			next = 0.01
		}
		if err := t.stocks.UpdatePrice(ctx, s.Symbol, next); err != nil {
			return err
		}
	}
	return nil
}
