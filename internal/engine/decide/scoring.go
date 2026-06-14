package decide

import (
	"github.com/faisal-990/ProjectInvestApp/internal/engine/indicators"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/pricing"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
)

const sixMonthsDays = 126 // ~trading days in 6 months

// scoreFundamental ranks a stock by the strategy's YAML scoring weights. Each
// factor is normalized to ~[0,1] and combined by its weight, so a strategy that
// weights "quality" highly prefers high-ROIC names, one that weights "intrinsic"
// prefers the cheapest-vs-value names, etc. — faithful to each investor's emphasis.
func scoreFundamental(cfg strategy.Config, v StockView) float64 {
	var score float64
	for key, weight := range cfg.Buy.Scoring {
		score += weight * factorScore(key, v)
	}
	return score
}

func factorScore(key string, v StockView) float64 {
	f := v.Fundamentals
	switch key {
	case "valuation":
		return clamp01((40 - f.PE) / 40) // cheaper P/E scores higher
	case "quality":
		return clamp01((f.ROIC + f.ROE) / 0.6)
	case "growth":
		return clamp01((f.RevenueGrowthYoY + f.EPSGrowthYoY) / 1.0)
	case "intrinsic":
		iv := pricing.IntrinsicValue(f.EPSTTM, growthEstimate(f))
		if iv <= 0 || v.Price <= 0 {
			return 0
		}
		return clamp01((iv - v.Price) / iv) // bigger discount scores higher
	case "earnings_yield":
		return clamp01(f.EarningsYield / 0.15)
	case "stability":
		return clamp01(float64(f.EPSPositiveYears) / 15)
	case "momentum":
		if r, ok := indicators.ReturnOver(v.Closes, sixMonthsDays); ok {
			return clamp01(r / 0.3)
		}
		return 0
	default:
		return 0
	}
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
