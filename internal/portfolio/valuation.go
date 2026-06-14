// Package portfolio holds the pure valuation math for an account: valuing
// positions at the current market price and deriving P&L, ROI, and allocation.
// It has no dependencies (no DB, no HTTP), so every number is unit-testable
// against hand-computed fixtures.
package portfolio

import "math"

// Position is one holding to be valued: how many shares, at what average cost,
// and the current market price.
type Position struct {
	Symbol       string
	Name         string
	Quantity     float64
	AvgPrice     float64
	CurrentPrice float64
}

// HoldingValuation is a single position valued at the current price.
type HoldingValuation struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	Quantity        float64 `json:"quantity"`
	AvgPrice        float64 `json:"avg_price"`
	CurrentPrice    float64 `json:"current_price"`
	CostBasis       float64 `json:"cost_basis"`        // quantity * avg price
	MarketValue     float64 `json:"market_value"`      // quantity * current price
	UnrealizedPL    float64 `json:"unrealized_pl"`     // market value - cost basis
	UnrealizedPLPct float64 `json:"unrealized_pl_pct"` // fraction of cost basis
	AllocationPct   float64 `json:"allocation_pct"`    // fraction of total portfolio value
}

// Summary is the whole-account valuation.
type Summary struct {
	Cash            float64            `json:"cash"`
	HoldingsValue   float64            `json:"holdings_value"`
	TotalValue      float64            `json:"total_value"` // cash + holdings
	CostBasis       float64            `json:"cost_basis"`
	UnrealizedPL    float64            `json:"unrealized_pl"`
	UnrealizedPLPct float64            `json:"unrealized_pl_pct"`
	ROI             float64            `json:"roi"` // total value vs starting capital
	Holdings        []HoldingValuation `json:"holdings"`
}

// Value computes the full valuation for an account holding `positions` with
// `cash` on hand, where `startingCapital` is the capital the account began with
// (used for ROI). It is pure and does not mutate its inputs. Money values are
// rounded to cents and ratios to 4 decimals for stable, presentable output.
func Value(cash, startingCapital float64, positions []Position) Summary {
	holdings := make([]HoldingValuation, 0, len(positions))
	var holdingsValue, costBasis float64

	for _, p := range positions {
		cost := p.Quantity * p.AvgPrice
		market := p.Quantity * p.CurrentPrice
		holdingsValue += market
		costBasis += cost

		hv := HoldingValuation{
			Symbol:       p.Symbol,
			Name:         p.Name,
			Quantity:     p.Quantity,
			AvgPrice:     money(p.AvgPrice),
			CurrentPrice: money(p.CurrentPrice),
			CostBasis:    money(cost),
			MarketValue:  money(market),
			UnrealizedPL: money(market - cost),
		}
		hv.UnrealizedPLPct = ratio(market-cost, cost)
		holdings = append(holdings, hv)
	}

	totalValue := cash + holdingsValue

	// Allocation is each holding's share of total portfolio value (incl. cash),
	// computed after totalValue is known.
	for i := range holdings {
		holdings[i].AllocationPct = ratio(holdings[i].MarketValue, totalValue)
	}

	return Summary{
		Cash:            money(cash),
		HoldingsValue:   money(holdingsValue),
		TotalValue:      money(totalValue),
		CostBasis:       money(costBasis),
		UnrealizedPL:    money(holdingsValue - costBasis),
		UnrealizedPLPct: ratio(holdingsValue-costBasis, costBasis),
		ROI:             ratio(totalValue-startingCapital, startingCapital),
		Holdings:        holdings,
	}
}

func money(v float64) float64 { return math.Round(v*100) / 100 }

// ratio returns num/den rounded to 4 decimals, or 0 when den is 0 (avoids NaN
// for empty/all-cash portfolios).
func ratio(num, den float64) float64 {
	if den == 0 {
		return 0
	}
	return math.Round((num/den)*10000) / 10000
}
