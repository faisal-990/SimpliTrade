// Package trading holds the pure settlement math for executing trades against an
// account: how a buy or sell changes cash, position size, and average cost. It
// has no dependencies (no DB, no HTTP), so every rule and edge is unit-testable
// in isolation. The repository layer applies these results inside a row-locked
// transaction; the broker layer chooses sim vs live execution.
package trading

import "errors"

var (
	// ErrInsufficientFunds means the account lacks cash to cover a buy.
	ErrInsufficientFunds = errors.New("trading: insufficient funds")
	// ErrInsufficientShares means the account holds fewer shares than a sell asks.
	ErrInsufficientShares = errors.New("trading: insufficient shares")
	// ErrInvalidQuantity means quantity was not strictly positive.
	ErrInvalidQuantity = errors.New("trading: quantity must be positive")
	// ErrInvalidPrice means price was not strictly positive.
	ErrInvalidPrice = errors.New("trading: price must be positive")
)

// Position is an account's holding of a single stock.
type Position struct {
	Quantity float64
	AvgPrice float64
}

// BuyResult is the computed outcome of a buy: the new cash balance, the new
// position (with recomputed average cost), and the cash spent.
type BuyResult struct {
	NewBalance  float64
	NewPosition Position
	Cost        float64
}

// SellResult is the computed outcome of a sell: the new cash balance, the
// remaining position, and the proceeds received.
type SellResult struct {
	NewBalance  float64
	NewPosition Position
	Proceeds    float64
}

// ComputeBuy returns the result of buying qty shares at price against an account
// with the given balance and existing position. The average price is recomputed
// as a cost-weighted blend of the old and new shares. It does not mutate inputs.
func ComputeBuy(balance float64, current Position, price, qty float64) (BuyResult, error) {
	if qty <= 0 {
		return BuyResult{}, ErrInvalidQuantity
	}
	if price <= 0 {
		return BuyResult{}, ErrInvalidPrice
	}
	cost := price * qty
	if cost > balance {
		return BuyResult{}, ErrInsufficientFunds
	}

	newQty := current.Quantity + qty
	// Weighted-average cost basis across existing and newly bought shares.
	newAvg := (current.AvgPrice*current.Quantity + price*qty) / newQty

	return BuyResult{
		NewBalance:  balance - cost,
		NewPosition: Position{Quantity: newQty, AvgPrice: newAvg},
		Cost:        cost,
	}, nil
}

// ComputeSell returns the result of selling qty shares at price. Average cost is
// unchanged by a sale; when the position is fully closed the remaining average
// resets to zero. It does not mutate inputs.
func ComputeSell(balance float64, current Position, price, qty float64) (SellResult, error) {
	if qty <= 0 {
		return SellResult{}, ErrInvalidQuantity
	}
	if price <= 0 {
		return SellResult{}, ErrInvalidPrice
	}
	if qty > current.Quantity {
		return SellResult{}, ErrInsufficientShares
	}

	proceeds := price * qty
	remaining := current.Quantity - qty
	newPos := Position{Quantity: remaining, AvgPrice: current.AvgPrice}
	if remaining == 0 {
		newPos.AvgPrice = 0
	}

	return SellResult{
		NewBalance:  balance + proceeds,
		NewPosition: newPos,
		Proceeds:    proceeds,
	}, nil
}
