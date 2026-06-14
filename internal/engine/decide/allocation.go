package decide

import (
	"fmt"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
)

// rebalanceBand is the drift (as a fraction of total value) an asset class may
// stray from its target before the allocator trades to correct it — avoids
// churning on small deviations.
const rebalanceBand = 0.05

// decideAllocation implements macro risk-parity / all-weather styles (Dalio): it
// holds asset-class ETF proxies at target weights and rebalances toward them when
// drift exceeds the band. It picks one proxy symbol per asset class from the
// market snapshot.
func decideAllocation(cfg strategy.Config, market []StockView, pf Portfolio) []Intent {
	if len(cfg.Allocation) == 0 {
		return nil
	}
	prices := priceMap(market)
	totalValue := pf.marketValue(prices)
	if totalValue <= 0 {
		return nil
	}

	// One representative proxy per asset class, and current value held per class.
	proxy := map[string]StockView{}
	for _, v := range market {
		if _, ok := proxy[v.AssetClass]; !ok {
			proxy[v.AssetClass] = v
		}
	}
	current := map[string]float64{}
	for sym, pos := range pf.Positions {
		current[viewAssetClass(market, sym)] += pos.Quantity * prices[sym]
	}

	band := rebalanceBand * totalValue
	var intents []Intent
	for class, weight := range cfg.Allocation {
		if class == "cash" {
			continue // cash is the residual, not traded
		}
		p, ok := proxy[class]
		if !ok || p.Price <= 0 {
			continue
		}
		diff := weight*totalValue - current[class]
		switch {
		case diff > band:
			intents = append(intents, Intent{
				Action: Buy, Symbol: p.Symbol, Quantity: diff / p.Price,
				EstPrice: p.Price, Reason: fmt.Sprintf("rebalance: raise %s to %.0f%%", class, weight*100),
			})
		case diff < -band:
			held := pf.Positions[p.Symbol]
			qty := minf(-diff/p.Price, held.Quantity)
			if qty > 0 {
				intents = append(intents, Intent{
					Action: Sell, Symbol: p.Symbol, Quantity: qty,
					EstPrice: p.Price, Reason: fmt.Sprintf("rebalance: trim %s to %.0f%%", class, weight*100),
				})
			}
		}
	}
	return intents
}

func viewAssetClass(market []StockView, symbol string) string {
	for _, v := range market {
		if v.Symbol == symbol {
			return v.AssetClass
		}
	}
	return "equity"
}
