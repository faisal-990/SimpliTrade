package service

import (
	"context"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/portfolio"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// PortfolioService values an account's positions at the current market price and
// reports holdings, P&L, ROI, and allocation. Valuation math lives in the pure
// internal/portfolio package; this layer only loads data and maps DTOs.
type PortfolioService interface {
	Stats(ctx context.Context, accountID string) (*dto.PortfolioStatsDTO, error)
	Holdings(ctx context.Context, accountID string) ([]dto.HoldingDTO, error)
	// Traders ranks real users by their portfolio value (the social leaderboard).
	Traders(ctx context.Context, limit int) ([]dto.TraderDTO, error)
}

type portfolioService struct {
	repo repository.PortfolioRepo
}

func NewPortfolioService(r repository.PortfolioRepo) PortfolioService {
	return &portfolioService{repo: r}
}

func (s *portfolioService) Traders(ctx context.Context, limit int) ([]dto.TraderDTO, error) {
	rows, err := s.repo.TopTraders(ctx, limit)
	if err != nil {
		return nil, httpx.Internal("could not load traders").WithCause(err)
	}
	out := make([]dto.TraderDTO, 0, len(rows))
	for i, r := range rows {
		out = append(out, dto.TraderDTO{
			Rank:      i + 1,
			Name:      r.Name,
			AvatarURL: r.AvatarURL,
			Value:     r.Value,
			ROI:       r.Value/models.StartingSimBalance - 1,
		})
	}
	return out, nil
}

func (s *portfolioService) Stats(ctx context.Context, accountID string) (*dto.PortfolioStatsDTO, error) {
	summary, err := s.value(ctx, accountID)
	if err != nil {
		return nil, err
	}
	stats := &dto.PortfolioStatsDTO{
		Cash:            summary.Cash,
		HoldingsValue:   summary.HoldingsValue,
		TotalValue:      summary.TotalValue,
		CostBasis:       summary.CostBasis,
		UnrealizedPL:    summary.UnrealizedPL,
		UnrealizedPLPct: summary.UnrealizedPLPct,
		ROI:             summary.ROI,
		Holdings:        toHoldingDTOs(summary.Holdings),
	}
	return stats, nil
}

func (s *portfolioService) Holdings(ctx context.Context, accountID string) ([]dto.HoldingDTO, error) {
	summary, err := s.value(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return toHoldingDTOs(summary.Holdings), nil
}

// value loads the account + holdings and runs the pure valuation.
func (s *portfolioService) value(ctx context.Context, accountID string) (*portfolio.Summary, error) {
	acct, err := uuid.Parse(accountID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid account identity")
	}

	account, err := s.repo.GetAccountByID(ctx, acct)
	if err != nil {
		return nil, httpx.Internal("could not load account").WithCause(err)
	}

	holdings, err := s.repo.ListHoldings(ctx, acct)
	if err != nil {
		return nil, httpx.Internal("could not load holdings").WithCause(err)
	}

	positions := make([]portfolio.Position, 0, len(holdings))
	for _, h := range holdings {
		positions = append(positions, portfolio.Position{
			Symbol:       h.Stock.Symbol,
			Name:         h.Stock.Name,
			Quantity:     h.Quantity,
			AvgPrice:     h.AvgPrice,
			CurrentPrice: h.Stock.CurrentPrice,
		})
	}

	summary := portfolio.Value(account.Balance, models.StartingSimBalance, positions)
	return &summary, nil
}

func toHoldingDTOs(hs []portfolio.HoldingValuation) []dto.HoldingDTO {
	out := make([]dto.HoldingDTO, 0, len(hs))
	for _, h := range hs {
		out = append(out, dto.HoldingDTO{
			Symbol:          h.Symbol,
			Name:            h.Name,
			Quantity:        h.Quantity,
			AvgPrice:        h.AvgPrice,
			CurrentPrice:    h.CurrentPrice,
			CostBasis:       h.CostBasis,
			MarketValue:     h.MarketValue,
			UnrealizedPL:    h.UnrealizedPL,
			UnrealizedPLPct: h.UnrealizedPLPct,
			AllocationPct:   h.AllocationPct,
		})
	}
	return out
}
