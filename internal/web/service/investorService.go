package service

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// InvestorService powers the social layer: the leaderboard, investor profiles,
// their trade feeds, and the follow graph (visibility-only — following surfaces
// an investor's moves, it does not copy them).
type InvestorService interface {
	Leaderboard(ctx context.Context, limit, offset int) ([]dto.InvestorDTO, error)
	Get(ctx context.Context, investorID string) (*dto.InvestorDTO, error)
	Trades(ctx context.Context, investorID string, limit, offset int) ([]dto.TradeHistoryItem, error)
	Follow(ctx context.Context, followerID, investorID string) error
	Unfollow(ctx context.Context, followerID, investorID string) error
	Following(ctx context.Context, followerID string) ([]dto.InvestorDTO, error)
	Feed(ctx context.Context, followerID string, limit int) ([]dto.FeedItem, error)
}

type investorService struct {
	repo repository.InvestorRepo
}

func NewInvestorService(r repository.InvestorRepo) InvestorService {
	return &investorService{repo: r}
}

func (s *investorService) Leaderboard(ctx context.Context, limit, offset int) ([]dto.InvestorDTO, error) {
	items, err := s.repo.ListInvestors(ctx, limit, offset)
	if err != nil {
		return nil, httpx.Internal("could not load leaderboard").WithCause(err)
	}
	out := make([]dto.InvestorDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toInvestorDTO(it))
	}
	return out, nil
}

func (s *investorService) Get(ctx context.Context, investorID string) (*dto.InvestorDTO, error) {
	id, err := uuid.Parse(investorID)
	if err != nil {
		return nil, httpx.BadRequest("invalid investor id")
	}
	item, err := s.repo.GetInvestor(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httpx.NotFound("investor not found")
		}
		return nil, httpx.Internal("could not load investor").WithCause(err)
	}
	d := toInvestorDTO(*item)
	return &d, nil
}

func (s *investorService) Trades(ctx context.Context, investorID string, limit, offset int) ([]dto.TradeHistoryItem, error) {
	id, err := uuid.Parse(investorID)
	if err != nil {
		return nil, httpx.BadRequest("invalid investor id")
	}
	trades, err := s.repo.ListInvestorTrades(ctx, id, limit, offset)
	if err != nil {
		return nil, httpx.Internal("could not load investor trades").WithCause(err)
	}
	out := make([]dto.TradeHistoryItem, 0, len(trades))
	for _, t := range trades {
		out = append(out, dto.TradeHistoryItem{
			TradeID: t.ID.String(), Symbol: t.Stock.Symbol, Side: t.Type,
			Quantity: t.Quantity, Price: t.Price, TotalValue: t.TotalValue,
			Status: t.Status, ExecutedAt: t.ExecutedAt.Unix(),
		})
	}
	return out, nil
}

func (s *investorService) Follow(ctx context.Context, followerID, investorID string) error {
	follower, investor, err := parsePair(followerID, investorID)
	if err != nil {
		return err
	}
	if follower == investor {
		return httpx.BadRequest("you cannot follow yourself")
	}
	if err := s.repo.Follow(ctx, follower, investor); err != nil {
		return httpx.Internal("could not follow investor").WithCause(err)
	}
	return nil
}

func (s *investorService) Unfollow(ctx context.Context, followerID, investorID string) error {
	follower, investor, err := parsePair(followerID, investorID)
	if err != nil {
		return err
	}
	if err := s.repo.Unfollow(ctx, follower, investor); err != nil {
		return httpx.Internal("could not unfollow investor").WithCause(err)
	}
	return nil
}

func (s *investorService) Following(ctx context.Context, followerID string) ([]dto.InvestorDTO, error) {
	id, err := uuid.Parse(followerID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	items, err := s.repo.ListFollowedInvestors(ctx, id)
	if err != nil {
		return nil, httpx.Internal("could not load followed investors").WithCause(err)
	}
	out := make([]dto.InvestorDTO, 0, len(items))
	for _, it := range items {
		out = append(out, toInvestorDTO(it))
	}
	return out, nil
}

func (s *investorService) Feed(ctx context.Context, followerID string, limit int) ([]dto.FeedItem, error) {
	follower, err := uuid.Parse(followerID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid user identity")
	}
	trades, err := s.repo.FeedTrades(ctx, follower, limit)
	if err != nil {
		return nil, httpx.Internal("could not load feed").WithCause(err)
	}
	out := make([]dto.FeedItem, 0, len(trades))
	for _, t := range trades {
		out = append(out, dto.FeedItem{
			InvestorID:   t.Account.UserID.String(),
			InvestorName: t.Account.User.Name,
			Symbol:       t.Stock.Symbol,
			Side:         t.Type,
			Quantity:     t.Quantity,
			Price:        t.Price,
			ExecutedAt:   t.ExecutedAt.Unix(),
		})
	}
	return out, nil
}

func parsePair(followerID, investorID string) (uuid.UUID, uuid.UUID, error) {
	follower, err := uuid.Parse(followerID)
	if err != nil {
		return uuid.Nil, uuid.Nil, httpx.Unauthorized("invalid user identity")
	}
	investor, err := uuid.Parse(investorID)
	if err != nil {
		return uuid.Nil, uuid.Nil, httpx.BadRequest("invalid investor id")
	}
	return follower, investor, nil
}

func toInvestorDTO(it repository.InvestorSummary) dto.InvestorDTO {
	return dto.InvestorDTO{
		ID: it.ID.String(), Name: it.Name, Bio: it.Bio,
		Strategy: it.Strategy, ROI: it.ROI, Rank: it.Rank,
	}
}
