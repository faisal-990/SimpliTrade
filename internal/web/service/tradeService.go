package service

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/broker"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/trading"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// TradeService validates trade requests and executes them through a Broker,
// translating domain errors into HTTP-shaped errors. It depends on the Broker
// interface, not a concrete broker, so sim vs live is decided at wiring time.
type TradeService interface {
	Buy(ctx context.Context, accountID string, req dto.TradeRequest) (*dto.TradeResponse, error)
	Sell(ctx context.Context, accountID string, req dto.TradeRequest) (*dto.TradeResponse, error)
	History(ctx context.Context, accountID string, limit, offset int) ([]dto.TradeHistoryItem, error)
	SellAll(ctx context.Context, accountID string) (int, error)
}

type tradeService struct {
	broker broker.Broker
	trades repository.TradeRepo
}

func NewTradeService(b broker.Broker, trades repository.TradeRepo) TradeService {
	return &tradeService{broker: b, trades: trades}
}

func (s *tradeService) Buy(ctx context.Context, accountID string, req dto.TradeRequest) (*dto.TradeResponse, error) {
	return s.execute(ctx, broker.Buy, accountID, req)
}

func (s *tradeService) Sell(ctx context.Context, accountID string, req dto.TradeRequest) (*dto.TradeResponse, error) {
	return s.execute(ctx, broker.Sell, accountID, req)
}

func (s *tradeService) execute(ctx context.Context, side broker.Side, accountID string, req dto.TradeRequest) (*dto.TradeResponse, error) {
	acct, err := uuid.Parse(accountID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid account identity")
	}

	order := broker.Order{
		AccountID: acct,
		Symbol:    req.Symbol,
		Side:      side,
		Quantity:  req.Quantity,
	}
	if req.IdempotencyKey != "" {
		order.IdempotencyKey = &req.IdempotencyKey
	}

	fill, err := s.broker.Execute(ctx, order)
	if err != nil {
		return nil, mapTradeError(err)
	}

	return &dto.TradeResponse{
		TradeID:    fill.TradeID.String(),
		Symbol:     fill.Symbol,
		Side:       string(fill.Side),
		Quantity:   fill.Quantity,
		Price:      fill.Price,
		TotalValue: fill.TotalValue,
		ExecutedAt: fill.ExecutedAt.Unix(),
	}, nil
}

func (s *tradeService) History(ctx context.Context, accountID string, limit, offset int) ([]dto.TradeHistoryItem, error) {
	acct, err := uuid.Parse(accountID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid account identity")
	}
	trades, err := s.trades.ListByAccount(ctx, acct, limit, offset)
	if err != nil {
		return nil, httpx.Internal("could not load trade history").WithCause(err)
	}
	items := make([]dto.TradeHistoryItem, 0, len(trades))
	for _, t := range trades {
		items = append(items, dto.TradeHistoryItem{
			TradeID:    t.ID.String(),
			Symbol:     t.Stock.Symbol,
			Side:       t.Type,
			Quantity:   t.Quantity,
			Price:      t.Price,
			TotalValue: t.TotalValue,
			Status:     t.Status,
			ExecutedAt: t.ExecutedAt.Unix(),
			Reason:     t.Reason,
		})
	}
	return items, nil
}

func (s *tradeService) SellAll(ctx context.Context, accountID string) (int, error) {
	acct, err := uuid.Parse(accountID)
	if err != nil {
		return 0, httpx.Unauthorized("invalid account identity")
	}
	n, err := s.trades.SellAll(ctx, acct)
	if err != nil {
		return 0, httpx.Internal("could not liquidate holdings").WithCause(err)
	}
	return n, nil
}

// mapTradeError translates domain/repository errors into client-facing errors.
func mapTradeError(err error) error {
	switch {
	case errors.Is(err, trading.ErrInsufficientFunds):
		return httpx.BadRequest("insufficient funds for this purchase")
	case errors.Is(err, trading.ErrInsufficientShares):
		return httpx.BadRequest("you do not hold enough shares to sell")
	case errors.Is(err, trading.ErrInvalidQuantity):
		return httpx.Validation("quantity must be positive")
	case errors.Is(err, trading.ErrInvalidPrice):
		return httpx.Internal("stock has no valid price")
	case errors.Is(err, repository.ErrNotFound):
		return httpx.NotFound("unknown symbol")
	default:
		return httpx.Internal("could not execute trade").WithCause(err)
	}
}
