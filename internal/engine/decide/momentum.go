package decide

import (
	"github.com/faisal-990/ProjectInvestApp/internal/engine/indicators"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
)

const (
	rsiPeriod      = 14
	yearDays       = 252
	breakoutWindow = 20 // recent-high lookback for a breakout
)

// decideMomentum implements trend/macro-momentum styles: exit on stops or
// momentum breakdown, enter on technical + momentum breakouts.
func decideMomentum(cfg strategy.Config, market []StockView, pf Portfolio) []Intent {
	prices := priceMap(market)
	views := indexViews(market)

	var intents []Intent
	for sym, pos := range pf.Positions {
		v, ok := views[sym]
		if !ok {
			continue
		}
		if reason, sell := momentumSell(cfg, v, pos); sell {
			intents = append(intents, Intent{
				Action: Sell, Symbol: sym, Quantity: pos.Quantity,
				EstPrice: v.Price, Reason: reason,
			})
		}
	}

	var cands []candidate
	for _, v := range market {
		if !inUniverse(cfg, v) {
			continue
		}
		if score, ok := momentumScreen(cfg, v); ok {
			cands = append(cands, candidate{Symbol: v.Symbol, Price: v.Price, Score: score, Reason: "momentum/trend breakout"})
		}
	}
	sortCandidates(cands)
	intents = append(intents, sizeBuys(cfg, cands, pf, prices)...)
	return intents
}

// momentumScreen applies technical + momentum gates and returns a momentum score.
func momentumScreen(cfg strategy.Config, v StockView) (float64, bool) {
	if t := cfg.Buy.Gates.Technical; t != nil {
		if t.PriceAboveSMADays != nil {
			sma, ok := indicators.SMA(v.Closes, *t.PriceAboveSMADays)
			if !ok || v.Price < sma {
				return 0, false
			}
		}
		if t.RSIMin != nil || t.RSIMax != nil {
			rsi, ok := indicators.RSI(v.Closes, rsiPeriod)
			if !ok {
				return 0, false
			}
			if t.RSIMin != nil && rsi < *t.RSIMin {
				return 0, false
			}
			if t.RSIMax != nil && rsi > *t.RSIMax {
				return 0, false
			}
		}
		if t.Near52WHighPct != nil {
			high, ok := indicators.High(v.Closes, yearDays)
			if !ok || high <= 0 || v.Price < high*(1-*t.Near52WHighPct) {
				return 0, false
			}
		}
		if t.BreakoutRequired {
			high, ok := indicators.High(v.Closes, breakoutWindow)
			if !ok || v.Price < high {
				return 0, false
			}
		}
	}

	if m := cfg.Buy.Gates.Momentum; m != nil {
		if m.Return3MMin != nil {
			r, ok := indicators.ReturnOver(v.Closes, 63)
			if !ok || r < *m.Return3MMin {
				return 0, false
			}
		}
		if m.Return6MMin != nil {
			r, ok := indicators.ReturnOver(v.Closes, sixMonthsDays)
			if !ok || r < *m.Return6MMin {
				return 0, false
			}
		}
		if m.Return12MMin != nil {
			r, ok := indicators.ReturnOver(v.Closes, yearDays)
			if !ok || r < *m.Return12MMin {
				return 0, false
			}
		}
		if m.RelativeStrengthMin != nil {
			// We have no cross-sectional index here; RSI(14) is used as a
			// per-stock relative-strength proxy (0–100, same scale).
			rs, ok := indicators.RSI(v.Closes, rsiPeriod)
			if !ok || rs < *m.RelativeStrengthMin {
				return 0, false
			}
		}
	}

	// Score by trailing 6-month return (fallback to 3-month).
	if r, ok := indicators.ReturnOver(v.Closes, sixMonthsDays); ok {
		return clamp01(r / 0.5), true
	}
	if r, ok := indicators.ReturnOver(v.Closes, 63); ok {
		return clamp01(r / 0.3), true
	}
	return 0.5, true
}

// momentumSell applies hard stop-loss, trailing stop, and momentum-breakdown exits.
func momentumSell(cfg strategy.Config, v StockView, pos Position) (string, bool) {
	s := cfg.Sell
	if s.StopLossPct != nil && pos.AvgPrice > 0 && v.Price <= pos.AvgPrice*(1-*s.StopLossPct) {
		return "stop-loss triggered", true
	}
	if s.TrailingStopPct != nil {
		// Trail from the recent high as a proxy for the high since entry.
		if high, ok := indicators.High(v.Closes, breakoutWindow); ok && high > 0 && v.Price <= high*(1-*s.TrailingStopPct) {
			return "trailing stop triggered", true
		}
	}
	if s.ThesisBreak.RSIBelow != nil {
		if rsi, ok := indicators.RSI(v.Closes, rsiPeriod); ok && rsi < *s.ThesisBreak.RSIBelow {
			return "momentum broke down", true
		}
	}
	return "", false
}
