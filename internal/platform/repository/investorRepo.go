package repository

import (
	"context"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InvestorSummary is a bot investor projected with its display name and current
// leaderboard standing — what the listing and profile endpoints return.
type InvestorSummary struct {
	ID       uuid.UUID `gorm:"column:id"`
	Name     string    `gorm:"column:name"`
	Bio      string    `gorm:"column:bio"`
	Strategy string    `gorm:"column:strategy"`
	ROI      float64   `gorm:"column:roi"`
	Rank     int       `gorm:"column:rank"`
}

// InvestorRepo reads the bot investors, their trades, and the follow graph.
type InvestorRepo interface {
	ListInvestors(ctx context.Context, limit, offset int) ([]InvestorSummary, error)
	GetInvestor(ctx context.Context, id uuid.UUID) (*InvestorSummary, error)
	ListInvestorTrades(ctx context.Context, investorID uuid.UUID, limit, offset int) ([]models.Trade, error)

	Follow(ctx context.Context, followerID, investorID uuid.UUID) error
	Unfollow(ctx context.Context, followerID, investorID uuid.UUID) error
	ListFollowing(ctx context.Context, followerID uuid.UUID) ([]uuid.UUID, error)
	FeedTrades(ctx context.Context, followerID uuid.UUID, limit int) ([]models.Trade, error)
}

type investorRepo struct {
	DB *gorm.DB
}

func NewInvestorRepo(db *gorm.DB) InvestorRepo {
	return &investorRepo{DB: db}
}

// investorSelect is the projection joining investor identity (users), profile
// (investors), and standing (performances). Missing performance rows default to
// ROI 0 / a large rank so unranked bots sort last.
const investorSelect = `investors.id AS id, users.name AS name, investors.bio AS bio,
	investors.strategy AS strategy,
	COALESCE(performances.roi, 0) AS roi,
	COALESCE(performances.rank, 1000000) AS rank`

func (r *investorRepo) baseQuery(ctx context.Context) *gorm.DB {
	return r.DB.WithContext(ctx).
		Table("investors").
		Select(investorSelect).
		Joins("JOIN users ON users.id = investors.id").
		Joins("LEFT JOIN performances ON performances.investor_id = investors.id").
		Where("users.is_bot = ?", true)
}

func (r *investorRepo) ListInvestors(ctx context.Context, limit, offset int) ([]InvestorSummary, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var out []InvestorSummary
	err := r.baseQuery(ctx).Order("rank ASC").Limit(limit).Offset(offset).Scan(&out).Error
	if err != nil {
		utils.LogError("repo: list investors", err)
		return nil, err
	}
	return out, nil
}

func (r *investorRepo) GetInvestor(ctx context.Context, id uuid.UUID) (*InvestorSummary, error) {
	var out InvestorSummary
	err := r.baseQuery(ctx).Where("investors.id = ?", id).Scan(&out).Error
	if err != nil {
		utils.LogError("repo: get investor", err)
		return nil, err
	}
	if out.ID == uuid.Nil {
		return nil, ErrNotFound
	}
	return &out, nil
}

func (r *investorRepo) ListInvestorTrades(ctx context.Context, investorID uuid.UUID, limit, offset int) ([]models.Trade, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var trades []models.Trade
	err := r.DB.WithContext(ctx).
		Joins("JOIN accounts ON accounts.id = trades.account_id").
		Where("accounts.user_id = ? AND accounts.mode = ?", investorID, models.ModeSim).
		Preload("Stock").
		Order("trades.executed_at DESC").
		Limit(limit).Offset(offset).
		Find(&trades).Error
	if err != nil {
		utils.LogError("repo: list investor trades", err)
		return nil, err
	}
	return trades, nil
}

func (r *investorRepo) Follow(ctx context.Context, followerID, investorID uuid.UUID) error {
	// Idempotent via the unique (investor, follower) index.
	return r.DB.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.Follow{InvestorID: investorID, FollowerID: followerID}).Error
}

func (r *investorRepo) Unfollow(ctx context.Context, followerID, investorID uuid.UUID) error {
	return r.DB.WithContext(ctx).
		Where("investor_id = ? AND follower_id = ?", investorID, followerID).
		Delete(&models.Follow{}).Error
}

func (r *investorRepo) ListFollowing(ctx context.Context, followerID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.DB.WithContext(ctx).
		Model(&models.Follow{}).
		Where("follower_id = ?", followerID).
		Pluck("investor_id", &ids).Error
	if err != nil {
		utils.LogError("repo: list following", err)
		return nil, err
	}
	return ids, nil
}

func (r *investorRepo) FeedTrades(ctx context.Context, followerID uuid.UUID, limit int) ([]models.Trade, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var trades []models.Trade
	err := r.DB.WithContext(ctx).
		Joins("JOIN accounts ON accounts.id = trades.account_id").
		Joins("JOIN follows ON follows.investor_id = accounts.user_id").
		Where("follows.follower_id = ? AND accounts.mode = ?", followerID, models.ModeSim).
		Preload("Stock").
		Preload("Account.User").
		Order("trades.executed_at DESC").
		Limit(limit).
		Find(&trades).Error
	if err != nil {
		utils.LogError("repo: feed trades", err)
		return nil, err
	}
	return trades, nil
}
