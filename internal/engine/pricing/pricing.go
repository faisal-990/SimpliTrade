// Package pricing holds the pure valuation formulas the strategy engine uses to
// decide what a stock is worth — independent of its current market price. These
// are the "decisions already made" (Benjamin Graham's formulas) turned into
// numbers; the engine compares the live price against them.
package pricing

import "math"

// Graham's intrinsic-value constants (from The Intelligent Investor). g is the
// expected annual growth rate; growth is capped so optimistic inputs can't
// inflate value without bound.
const (
	grahamBasePE        = 8.5
	grahamGrowthFactor  = 2.0
	grahamMaxGrowthRate = 0.15 // 15% cap
	grahamMultiple      = 22.5 // the classic P/E × P/B ceiling
)

// GrahamNumber returns √(22.5 · EPS · BVPS), the maximum price Graham's defensive
// investor would pay. It returns 0 when EPS or BVPS is non-positive (the stock
// is not a defensive candidate), so callers treat 0 as "no valid floor".
func GrahamNumber(eps, bvps float64) float64 {
	if eps <= 0 || bvps <= 0 {
		return 0
	}
	return math.Sqrt(grahamMultiple * eps * bvps)
}

// IntrinsicValue returns Graham's revised intrinsic value V = EPS·(8.5 + 2g),
// where g is the expected growth rate as a percentage and is capped at 15%.
// growth is a fraction (0.10 = 10%). Returns 0 for non-positive EPS.
func IntrinsicValue(eps, growth float64) float64 {
	if eps <= 0 {
		return 0
	}
	g := growth
	if g > grahamMaxGrowthRate {
		g = grahamMaxGrowthRate
	}
	if g < 0 {
		g = 0
	}
	return eps * (grahamBasePE + grahamGrowthFactor*(g*100))
}

// BuyBelowPrice applies a margin of safety to an intrinsic value: the highest
// price at which a buy is allowed. A mosMin of 0.30 means "buy only at or below
// 70% of intrinsic value". Returns 0 when intrinsic is 0.
func BuyBelowPrice(intrinsic, mosMin float64) float64 {
	if intrinsic <= 0 {
		return 0
	}
	if mosMin < 0 {
		mosMin = 0
	}
	return intrinsic * (1 - mosMin)
}
