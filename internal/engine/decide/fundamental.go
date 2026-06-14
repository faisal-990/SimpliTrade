package decide

import (
	"fmt"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/indicators"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/pricing"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// Assessment is the price-independent screening result for one stock under a
// strategy. The engine's slow lane can compute and cache it; the fast lane reuses
// it to signal against the live price.
type Assessment struct {
	Symbol   string
	Passes   bool    // passes all present buy gates (incl. margin of safety)
	BuyBelow float64 // intrinsic-based ceiling, 0 if the strategy has no MOS rule
	Score    float64 // YAML-weighted ranking score
	Reason   string
}

// decideFundamental implements value/quality/garp/growth/quant/contrarian/activist
// styles: sell positions that break their thesis or hit their target, then buy
// the best-scoring stocks that pass the gates.
func decideFundamental(cfg strategy.Config, market []StockView, pf Portfolio) []Intent {
	prices := priceMap(market)
	views := indexViews(market)

	var intents []Intent

	// SELL pass over current holdings.
	for sym, pos := range pf.Positions {
		v, ok := views[sym]
		if !ok {
			continue
		}
		if reason, sell := fundamentalSell(cfg, v, pos); sell {
			intents = append(intents, Intent{
				Action: Sell, Symbol: sym, Quantity: pos.Quantity,
				EstPrice: v.Price, Reason: reason,
			})
		}
	}

	// BUY pass: screen, rank, size.
	var cands []candidate
	for _, v := range market {
		if !inUniverse(cfg, v) {
			continue
		}
		a := ScreenFundamental(cfg, v)
		if !a.Passes {
			continue
		}
		cands = append(cands, candidate{Symbol: v.Symbol, Price: v.Price, Score: a.Score, Reason: a.Reason})
	}
	sortCandidates(cands)
	intents = append(intents, sizeBuys(cfg, cands, pf, prices)...)
	return intents
}

// ScreenFundamental applies the strategy's hard gates and margin-of-safety rule
// to a stock and computes its ranking score. It is price-dependent only through
// the Graham-number / margin-of-safety check; the rest is fundamentals-only.
func ScreenFundamental(cfg strategy.Config, v StockView) Assessment {
	f := v.Fundamentals
	a := Assessment{Symbol: v.Symbol}

	if g := cfg.Buy.Gates.Valuation; g != nil {
		if !passValuation(g, f, v.Price) {
			return a
		}
	}
	if g := cfg.Buy.Gates.FinancialSafety; g != nil && !passFinancialSafety(g, f) {
		return a
	}
	if g := cfg.Buy.Gates.Quality; g != nil && !passQuality(g, f) {
		return a
	}
	if g := cfg.Buy.Gates.Growth; g != nil && !passGrowth(g, f) {
		return a
	}
	if g := cfg.Buy.Gates.Stability; g != nil && !passStability(g, f) {
		return a
	}

	// Margin of safety against intrinsic value.
	if cfg.Buy.MarginOfSafetyMin != nil {
		intrinsic := pricing.IntrinsicValue(f.EPSTTM, growthEstimate(f))
		a.BuyBelow = pricing.BuyBelowPrice(intrinsic, *cfg.Buy.MarginOfSafetyMin)
		if a.BuyBelow <= 0 || v.Price > a.BuyBelow {
			return a
		}
	}

	a.Passes = true
	a.Score = scoreFundamental(cfg, v)
	a.Reason = fmt.Sprintf("passed %s gates (score %.2f)", cfg.Identity.Style, a.Score)
	return a
}

func passValuation(g *strategy.ValuationGate, f models.Fundamentals, price float64) bool {
	if g.PEMax != nil && (f.PE <= 0 || f.PE > *g.PEMax) {
		return false
	}
	if g.ForwardPEMax != nil && f.ForwardPE > *g.ForwardPEMax {
		return false
	}
	if g.PBMax != nil && f.PB > *g.PBMax {
		return false
	}
	if g.PEPBMax != nil && f.PE*f.PB > *g.PEPBMax {
		return false
	}
	if g.PSMax != nil && f.PS > *g.PSMax {
		return false
	}
	if g.PEGMax != nil && (f.PEG <= 0 || f.PEG > *g.PEGMax) {
		return false
	}
	if g.EVEBITDAMax != nil && f.EVEBITDA > *g.EVEBITDAMax {
		return false
	}
	if g.EarningsYldMin != nil && f.EarningsYield < *g.EarningsYldMin {
		return false
	}
	if g.FCFYieldMin != nil && f.FCFYield < *g.FCFYieldMin {
		return false
	}
	if g.DividendYldMin != nil && f.DividendYield < *g.DividendYldMin {
		return false
	}
	if g.UseGrahamNumber {
		gn := pricing.GrahamNumber(f.EPSTTM, f.BVPS)
		if gn <= 0 || price > gn {
			return false
		}
	}
	return true
}

func passFinancialSafety(g *strategy.FinancialSafetyGate, f models.Fundamentals) bool {
	if g.CurrentRatioMin != nil && f.CurrentRatio < *g.CurrentRatioMin {
		return false
	}
	if g.DebtToEquityMax != nil && f.DebtToEquity > *g.DebtToEquityMax {
		return false
	}
	if g.InterestCoverageMin != nil && f.InterestCover < *g.InterestCoverageMin {
		return false
	}
	if g.FCFPositive && !f.FCFPositive {
		return false
	}
	return true
}

func passQuality(g *strategy.QualityGate, f models.Fundamentals) bool {
	if g.ROEMin != nil && f.ROE < *g.ROEMin {
		return false
	}
	if g.ROICMin != nil && f.ROIC < *g.ROICMin {
		return false
	}
	if g.GrossMarginMin != nil && f.GrossMargin < *g.GrossMarginMin {
		return false
	}
	if g.OperatingMarginMin != nil && f.OperatingMargin < *g.OperatingMarginMin {
		return false
	}
	if g.NetMarginMin != nil && f.NetMargin < *g.NetMarginMin {
		return false
	}
	return true
}

func passGrowth(g *strategy.GrowthGate, f models.Fundamentals) bool {
	if g.RevenueGrowthYoYMin != nil && f.RevenueGrowthYoY < *g.RevenueGrowthYoYMin {
		return false
	}
	if g.EPSGrowthYoYMin != nil && f.EPSGrowthYoY < *g.EPSGrowthYoYMin {
		return false
	}
	if g.RevenueCAGR3YMin != nil && f.RevenueCAGR3Y < *g.RevenueCAGR3YMin {
		return false
	}
	if g.EPSGrowth5YMin != nil && f.EPSGrowth5Y < *g.EPSGrowth5YMin {
		return false
	}
	return true
}

func passStability(g *strategy.StabilityGate, f models.Fundamentals) bool {
	if g.EPSPositiveYearsMin != nil && f.EPSPositiveYears < *g.EPSPositiveYearsMin {
		return false
	}
	if g.DividendYearsMin != nil && f.DividendYears < *g.DividendYearsMin {
		return false
	}
	if g.BetaMax != nil && f.Beta > *g.BetaMax {
		return false
	}
	return true
}

// fundamentalSell evaluates the strategy's sell rules against a held position.
func fundamentalSell(cfg strategy.Config, v StockView, pos Position) (string, bool) {
	f := v.Fundamentals
	s := cfg.Sell

	if s.TakeProfitVsIntrinsic != nil {
		intrinsic := pricing.IntrinsicValue(f.EPSTTM, growthEstimate(f))
		if intrinsic > 0 && v.Price >= intrinsic**s.TakeProfitVsIntrinsic {
			return "price reached intrinsic value target", true
		}
	}
	if s.StopLossPct != nil && pos.AvgPrice > 0 && v.Price <= pos.AvgPrice*(1-*s.StopLossPct) {
		return "stop-loss triggered", true
	}

	tb := s.ThesisBreak
	switch {
	case tb.ROEBelow != nil && f.ROE < *tb.ROEBelow:
		return "thesis break: ROE fell below threshold", true
	case tb.ROICBelow != nil && f.ROIC < *tb.ROICBelow:
		return "thesis break: ROIC fell below threshold", true
	case tb.DebtToEquityAbove != nil && f.DebtToEquity > *tb.DebtToEquityAbove:
		return "thesis break: leverage exceeded threshold", true
	case tb.RevenueGrowthBelow != nil && f.RevenueGrowthYoY < *tb.RevenueGrowthBelow:
		return "thesis break: revenue growth stalled", true
	case tb.OperatingMarginBelow != nil && f.OperatingMargin < *tb.OperatingMarginBelow:
		return "thesis break: operating margin eroded", true
	case tb.InterestCoverBelow != nil && f.InterestCover < *tb.InterestCoverBelow:
		return "thesis break: interest coverage weakened", true
	case tb.FCFPositive && !f.FCFPositive:
		return "thesis break: free cash flow turned negative", true
	case tb.RSIBelow != nil:
		if rsi, ok := indicators.RSI(v.Closes, 14); ok && rsi < *tb.RSIBelow {
			return "thesis break: momentum broke down", true
		}
	}
	return "", false
}

// growthEstimate picks the best available growth rate for intrinsic value,
// preferring longer-horizon EPS growth.
func growthEstimate(f models.Fundamentals) float64 {
	switch {
	case f.EPSGrowth5Y != 0:
		return f.EPSGrowth5Y
	case f.EPSGrowthYoY != 0:
		return f.EPSGrowthYoY
	default:
		return f.RevenueGrowthYoY
	}
}

func indexViews(market []StockView) map[string]StockView {
	m := make(map[string]StockView, len(market))
	for _, v := range market {
		m[v.Symbol] = v
	}
	return m
}
