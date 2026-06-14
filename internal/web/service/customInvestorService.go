package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// CustomInvestorService lets a user author their own investor. The created
// investor is a normal bot (Investor + Account + strategy.Config), so it plugs
// into the leaderboard, follow, copy-trading, backtest, and engine unchanged.
type CustomInvestorService interface {
	Create(ctx context.Context, userID string, req dto.CreateInvestorRequest) (*dto.InvestorDTO, error)
	ListMine(ctx context.Context, userID string) ([]dto.InvestorDTO, error)
}

type customInvestorService struct {
	custom    repository.CustomStrategyRepo
	bots      repository.BotRepo
	investors repository.InvestorRepo
}

func NewCustomInvestorService(custom repository.CustomStrategyRepo, bots repository.BotRepo, investors repository.InvestorRepo) CustomInvestorService {
	return &customInvestorService{custom: custom, bots: bots, investors: investors}
}

func (s *customInvestorService) Create(ctx context.Context, userID string, req dto.CreateInvestorRequest) (*dto.InvestorDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}

	// A stable, unique strategy id keys the bot identity.
	strategyID := "custom-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
	cfg, err := strategy.BuildCustom(strategy.CustomParams{
		ID: strategyID, Name: req.Name, Philosophy: req.Philosophy, Approach: req.Approach,
		MaxPositions: req.MaxPositions,
		PEMax:        req.PEMax, PBMax: req.PBMax,
		ROEMin: req.ROEMin, OperatingMarginMin: req.OperatingMarginMin,
		RevenueGrowthMin: req.RevenueGrowthMin, EPSGrowthMin: req.EPSGrowthMin,
		Return6MMin:           req.Return6MMin,
		StopLossPct:           req.StopLossPct,
		TakeProfitVsIntrinsic: req.TakeProfitVsIntrinsic,
		MaxPositionSize:       req.MaxPositionSize,
		CashBufferMin:         req.CashBufferMin,
		PositionSizing:        req.PositionSizing,
	})
	if err != nil {
		return nil, httpx.Validation("invalid strategy: " + err.Error())
	}

	raw, err := json.Marshal(cfg)
	if err != nil {
		return nil, httpx.Internal("could not encode strategy").WithCause(err)
	}

	// Mint the bot identity (user + investor + sim account), then persist the row.
	investorID, accountID, err := s.bots.UpsertBot(ctx, strategyID, req.Name, req.Philosophy, cfg.Identity.Style)
	if err != nil {
		return nil, httpx.Internal("could not create investor identity").WithCause(err)
	}
	if err := s.custom.Create(ctx, &models.CustomStrategy{
		UserID: uid, InvestorID: investorID, AccountID: accountID,
		Name: req.Name, Style: cfg.Identity.Style, ConfigJSON: string(raw),
	}); err != nil {
		return nil, httpx.Internal("could not save investor").WithCause(err)
	}

	inv, err := s.investors.GetInvestor(ctx, investorID)
	if err != nil {
		return nil, httpx.Internal("could not load investor").WithCause(err)
	}
	d := toInvestorDTO(*inv)
	return &d, nil
}

func (s *customInvestorService) ListMine(ctx context.Context, userID string) ([]dto.InvestorDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	rows, err := s.custom.ListByUser(ctx, uid)
	if err != nil {
		return nil, httpx.Internal("could not load your investors").WithCause(err)
	}
	out := make([]dto.InvestorDTO, 0, len(rows))
	for _, r := range rows {
		inv, err := s.investors.GetInvestor(ctx, r.InvestorID)
		if err != nil {
			continue // skip any orphaned row rather than failing the list
		}
		out = append(out, toInvestorDTO(*inv))
	}
	return out, nil
}
