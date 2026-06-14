// Package broker is the seam between trading logic and how an order is actually
// filled and settled. SimulatedBroker settles against our own database (virtual
// money); a future LiveBroker will call a real brokerage API. The trade service
// depends only on the Broker interface and selects an implementation from the
// account's mode (BrokerFor), so switching a user to real money is a data change,
// not a code change.
package broker

import (
	"context"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/google/uuid"
)

// Side is the direction of an order.
type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

// Order is a request to trade a quantity of a symbol for an account.
type Order struct {
	AccountID      uuid.UUID
	Symbol         string
	Side           Side
	Quantity       float64
	IdempotencyKey *string
}

// Fill is the result of executing an order.
type Fill struct {
	TradeID    uuid.UUID
	Symbol     string
	Side       Side
	Quantity   float64
	Price      float64
	TotalValue float64
	ExecutedAt time.Time
}

// Broker fills and settles orders. Implementations differ in where money moves;
// the contract (and its errors, e.g. trading.ErrInsufficientFunds) is identical.
type Broker interface {
	Execute(ctx context.Context, order Order) (Fill, error)
}

// SimulatedBroker settles orders against the application database via TradeRepo
// (virtual money). It is the only broker today; LiveBroker arrives at T9/T10.
type SimulatedBroker struct {
	trades repository.TradeRepo
}

// NewSimulatedBroker builds a SimulatedBroker over the given trade repository.
func NewSimulatedBroker(trades repository.TradeRepo) *SimulatedBroker {
	return &SimulatedBroker{trades: trades}
}

func (b *SimulatedBroker) Execute(ctx context.Context, order Order) (Fill, error) {
	var (
		trade *models.Trade
		err   error
	)
	switch order.Side {
	case Buy:
		trade, err = b.trades.ExecuteBuy(ctx, order.AccountID, order.Symbol, order.Quantity, order.IdempotencyKey)
	case Sell:
		trade, err = b.trades.ExecuteSell(ctx, order.AccountID, order.Symbol, order.Quantity, order.IdempotencyKey)
	default:
		return Fill{}, ErrInvalidSide
	}
	if err != nil {
		return Fill{}, err
	}
	return Fill{
		TradeID:    trade.ID,
		Symbol:     order.Symbol,
		Side:       order.Side,
		Quantity:   trade.Quantity,
		Price:      trade.Price,
		TotalValue: trade.TotalValue,
		ExecutedAt: trade.ExecutedAt,
	}, nil
}

// BrokerFor selects a broker implementation for an account mode. Today every
// mode maps to the SimulatedBroker; when LiveBroker lands, only this function
// changes — callers are untouched.
func BrokerFor(mode models.AccountMode, sim *SimulatedBroker) Broker {
	switch mode {
	case models.ModeLive:
		// TODO(T9/T10): return the real LiveBroker once KYC + provider are wired.
		return sim
	default:
		return sim
	}
}
