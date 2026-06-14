// Package decide is the pure decision core of the strategy engine. Given a market
// snapshot, a strategy config, and the bot's current portfolio, it produces the
// trades the strategy would make — with no I/O, so every paradigm is testable in
// isolation. The expensive, price-independent screening is separable from the
// cheap, price-dependent signalling so the engine's two clocks (slow fundamentals
// lane, fast price lane) can cache and reuse work.
package decide

import "github.com/faisal-990/ProjectInvestApp/internal/platform/models"

// Action is the direction of an intent.
type Action string

const (
	Buy  Action = "buy"
	Sell Action = "sell"
)

// StockView is the engine's snapshot of one stock at decision time.
type StockView struct {
	Symbol       string
	Sector       string
	AssetClass   string // equity|bond|gold|commodity
	Price        float64
	Fundamentals models.Fundamentals
	// Closes is the historical daily close series (oldest→newest) used to derive
	// technical indicators for momentum/trend strategies.
	Closes []float64
}

// Position is a held quantity of a stock at its average cost.
type Position struct {
	Quantity float64
	AvgPrice float64
}

// Portfolio is the bot's state: available cash and current positions keyed by
// symbol.
type Portfolio struct {
	Cash      float64
	Positions map[string]Position
}

// Intent is a proposed trade the engine wants to execute. Quantity is in
// (fractional) shares; the executor turns intents into real orders via the broker.
type Intent struct {
	Action   Action  `json:"action"`
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	EstPrice float64 `json:"est_price"`
	Reason   string  `json:"reason"`
	Score    float64 `json:"score"`
}

// marketValue returns cash plus the current market value of all positions, using
// the prices in the given market snapshot.
func (p Portfolio) marketValue(byNow map[string]float64) float64 {
	total := p.Cash
	for sym, pos := range p.Positions {
		total += pos.Quantity * byNow[sym]
	}
	return total
}
