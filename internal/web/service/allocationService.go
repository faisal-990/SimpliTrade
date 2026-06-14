package service

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// AllocationService manages capped copy-trading sub-accounts: committing a slice
// of the user's cash to an investor whose bot then trades only that slice.
type AllocationService interface {
	Create(ctx context.Context, userID string, req dto.AllocateRequest) (*dto.AllocationDTO, error)
	List(ctx context.Context, userID string) ([]dto.AllocationDTO, error)
	Detail(ctx context.Context, userID, allocationID string) (*dto.AllocationDetailDTO, error)
	Stop(ctx context.Context, userID, allocationID string) error
}

type allocationService struct {
	repo repository.AllocationRepo
}

func NewAllocationService(r repository.AllocationRepo) AllocationService {
	return &allocationService{repo: r}
}

func (s *allocationService) Create(ctx context.Context, userID string, req dto.AllocateRequest) (*dto.AllocationDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	inv, err := uuid.Parse(req.InvestorID)
	if err != nil {
		return nil, httpx.BadRequest("invalid investor id")
	}
	if req.Capital <= 0 {
		return nil, httpx.Validation("capital must be positive")
	}

	if _, err := s.repo.Create(ctx, uid, inv, req.Capital); err != nil {
		if errors.Is(err, repository.ErrInsufficientCapital) {
			return nil, httpx.BadRequest("not enough cash in your account for this allocation")
		}
		return nil, httpx.Internal("could not create allocation").WithCause(err)
	}
	// Return the freshly-created allocation from the list (carries investor name).
	list, err := s.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].InvestorID == req.InvestorID && list[i].IsActive {
			return &list[i], nil
		}
	}
	return &dto.AllocationDTO{InvestorID: req.InvestorID, Capital: req.Capital, Cash: req.Capital, MarketValue: req.Capital, IsActive: true}, nil
}

func (s *allocationService) List(ctx context.Context, userID string) ([]dto.AllocationDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	views, err := s.repo.List(ctx, uid)
	if err != nil {
		return nil, httpx.Internal("could not load allocations").WithCause(err)
	}
	out := make([]dto.AllocationDTO, 0, len(views))
	for _, v := range views {
		ret := 0.0
		if v.Capital > 0 {
			ret = v.MarketValue/v.Capital - 1
		}
		out = append(out, dto.AllocationDTO{
			ID: v.ID.String(), InvestorID: v.InvestorID.String(), InvestorName: v.InvestorName,
			Strategy: v.Strategy, Capital: v.Capital, Cash: v.Cash, MarketValue: v.MarketValue,
			ReturnPct: ret, IsActive: v.IsActive,
		})
	}
	return out, nil
}

func (s *allocationService) Detail(ctx context.Context, userID, allocationID string) (*dto.AllocationDetailDTO, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	aid, err := uuid.Parse(allocationID)
	if err != nil {
		return nil, httpx.BadRequest("invalid allocation id")
	}
	act, err := s.repo.Activity(ctx, uid, aid)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httpx.NotFound("allocation not found")
		}
		return nil, httpx.Internal("could not load allocation activity").WithCause(err)
	}

	v := act.View
	ret := 0.0
	if v.Capital > 0 {
		ret = v.MarketValue/v.Capital - 1
	}
	out := &dto.AllocationDetailDTO{
		AllocationDTO: dto.AllocationDTO{
			ID: v.ID.String(), InvestorID: v.InvestorID.String(), InvestorName: v.InvestorName,
			Strategy: v.Strategy, Capital: v.Capital, Cash: v.Cash, MarketValue: v.MarketValue,
			ReturnPct: ret, IsActive: v.IsActive,
		},
		Holdings: make([]dto.AllocationHoldingDTO, 0, len(act.Holdings)),
		Trades:   make([]dto.AllocationTradeDTO, 0, len(act.Trades)),
	}
	for _, h := range act.Holdings {
		pl := (h.CurrentPrice - h.AvgPrice) * h.Quantity
		plPct := 0.0
		if h.AvgPrice > 0 {
			plPct = h.CurrentPrice/h.AvgPrice - 1 // fraction (UI multiplies by 100)
		}
		out.Holdings = append(out.Holdings, dto.AllocationHoldingDTO{
			Symbol: h.Symbol, Quantity: h.Quantity, AvgPrice: h.AvgPrice,
			CurrentPrice: h.CurrentPrice, MarketValue: h.MarketValue,
			UnrealizedPL: pl, UnrealizedPLPct: plPct,
		})
	}
	for _, t := range act.Trades {
		out.Trades = append(out.Trades, dto.AllocationTradeDTO{
			Symbol: t.Symbol, Side: t.Side, Quantity: t.Quantity,
			Price: t.Price, TotalValue: t.TotalValue, ExecutedAt: t.ExecutedAt.Unix(),
			Reason: t.Reason,
		})
	}
	return out, nil
}

func (s *allocationService) Stop(ctx context.Context, userID, allocationID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return httpx.Unauthorized("invalid user identity")
	}
	aid, err := uuid.Parse(allocationID)
	if err != nil {
		return httpx.BadRequest("invalid allocation id")
	}
	if err := s.repo.Stop(ctx, uid, aid); err != nil {
		return httpx.Internal("could not stop allocation").WithCause(err)
	}
	return nil
}
