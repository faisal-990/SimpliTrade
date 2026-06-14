package decide

import (
	"slices"
	"sort"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
)

// Decide runs the strategy against the market and portfolio, returning the
// trades it would make. It dispatches to the evaluator for the strategy's kind;
// all evaluators share the universe filter and the sizing logic below.
func Decide(cfg strategy.Config, market []StockView, pf Portfolio) []Intent {
	switch cfg.Kind() {
	case strategy.KindMomentum:
		return decideMomentum(cfg, market, pf)
	case strategy.KindAllocation:
		return decideAllocation(cfg, market, pf)
	default:
		return decideFundamental(cfg, market, pf)
	}
}

// priceMap indexes current prices by symbol for portfolio valuation.
func priceMap(market []StockView) map[string]float64 {
	m := make(map[string]float64, len(market))
	for _, v := range market {
		m[v.Symbol] = v.Price
	}
	return m
}

// inUniverse reports whether a stock is eligible under the strategy's universe
// constraints (sector include/exclude, market-cap band).
func inUniverse(cfg strategy.Config, v StockView) bool {
	u := cfg.Universe
	if len(u.SectorsInclude) > 0 && !slices.Contains(u.SectorsInclude, v.Sector) {
		return false
	}
	if slices.Contains(u.SectorsExclude, v.Sector) {
		return false
	}
	if u.MarketCapMin != nil && v.Fundamentals.MarketCap < *u.MarketCapMin {
		return false
	}
	if u.MarketCapMax != nil && *u.MarketCapMax > 0 && v.Fundamentals.MarketCap > *u.MarketCapMax {
		return false
	}
	return true
}

// candidate is a ranked buy idea before sizing.
type candidate struct {
	Symbol string
	Price  float64
	Score  float64
	Reason string
}

// sizeBuys turns ranked candidates into fractional-share Buy intents, respecting
// the cash buffer, per-position cap, and max-positions limit. Candidates must be
// pre-sorted best-first.
func sizeBuys(cfg strategy.Config, candidates []candidate, pf Portfolio, prices map[string]float64) []Intent {
	totalValue := pf.marketValue(prices)
	if totalValue <= 0 {
		return nil
	}

	// Cash we may deploy while preserving the strategy's minimum cash buffer.
	investable := pf.Cash - cfg.Risk.CashBufferMin*totalValue
	if investable <= 0 {
		return nil
	}

	maxPosVal := cfg.Risk.MaxPositionSize * totalValue
	slots := cfg.Universe.MaxPositions - len(pf.Positions)

	// Per-name target: equal-weight splits across remaining slots; everything
	// else fills to the per-position cap in score order (conviction/pyramid).
	target := maxPosVal
	if cfg.Risk.PositionSizing == "equal" && slots > 0 {
		if eq := investable / float64(slots); eq < target {
			target = eq
		}
	}

	var intents []Intent
	for _, c := range candidates {
		if slots <= 0 || investable <= 1 || c.Price <= 0 {
			break
		}
		held := pf.Positions[c.Symbol]
		heldVal := held.Quantity * c.Price
		// "Already hold enough of it" → no room, skip (HOLD).
		room := minf(maxPosVal, target) - heldVal
		if room <= 1 {
			continue
		}
		spend := minf(room, investable)
		qty := spend / c.Price
		intents = append(intents, Intent{
			Action: Buy, Symbol: c.Symbol, Quantity: qty,
			EstPrice: c.Price, Reason: c.Reason, Score: c.Score,
		})
		investable -= spend
		if held.Quantity == 0 {
			slots-- // only a brand-new position consumes a slot
		}
	}
	return intents
}

func sortCandidates(cs []candidate) {
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].Score > cs[j].Score })
}

func minf(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
