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

// Refresh applies one market step to every priced stock. It moves the CURRENT
// price only — it deliberately does NOT append to the daily candle history,
// because that history is the substrate for charts and backtests and must stay a
// clean, contiguous record. Moving the current price is enough to revalue
// positions, shift ROI, and feed the bots' next decision.
func (t *SyntheticTicker) Refresh(ctx context.Context) error {
	stocks, err := t.stocks.List(ctx, 1000, 0)
	if err != nil {
		return err
	}
	for _, s := range stocks {
		prev := s.CurrentPrice
		if prev <= 0 {
			continue
		}
		// Symmetric proportional gaussian shock (no drift), so prices move both
		// ways — positions go underwater as well as up, which lets stop-loss and
		// take-profit sell rules actually trigger. Clamp to keep prices sane.
		shock := t.rng.NormFloat64() * t.vol
		next := prev * (1 + shock)
		if next < 0.01 {
			next = 0.01
		}
		if err := t.stocks.UpdatePrice(ctx, s.Symbol, next); err != nil {
			return err
		}
	}
	return nil
}
