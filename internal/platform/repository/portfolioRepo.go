package repository

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PortfolioRepo reads the data needed to value an account: its cash balance and
// its holdings (each joined with its stock for symbol + current price).
// TraderRow is one real user on the traders leaderboard: their identity plus
// their primary-account portfolio value (cash + holdings at current price).
type TraderRow struct {
	UserID    uuid.UUID `gorm:"column:user_id"`
	Name      string    `gorm:"column:name"`
	AvatarURL string    `gorm:"column:avatar_url"`
	Value     float64   `gorm:"column:value"`
}

type PortfolioRepo interface {
	GetAccountByID(ctx context.Context, accountID uuid.UUID) (*models.Account, error)
	ListHoldings(ctx context.Context, accountID uuid.UUID) ([]models.Holding, error)
	// TopTraders ranks real (non-bot) users by their primary sim portfolio value.
	TopTraders(ctx context.Context, limit int) ([]TraderRow, error)
}

type portfolioRepo struct {
	DB *gorm.DB
}

func NewPortfolioRepo(db *gorm.DB) PortfolioRepo {
	return &portfolioRepo{DB: db}
}

func (r *portfolioRepo) GetAccountByID(ctx context.Context, accountID uuid.UUID) (*models.Account, error) {
	var acct models.Account
	err := r.DB.WithContext(ctx).First(&acct, "id = ?", accountID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		utils.LogError("repo: get account by id", err)
		return nil, err
	}
	return &acct, nil
}

func (r *portfolioRepo) TopTraders(ctx context.Context, limit int) ([]TraderRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var rows []TraderRow
	err := r.DB.WithContext(ctx).
		Table("accounts a").
		Select(`u.id AS user_id, u.name AS name, COALESCE(u.avatar_url, '') AS avatar_url,
			a.balance + COALESCE(SUM(h.quantity * s.current_price), 0) AS value`).
		Joins("JOIN users u ON u.id = a.user_id").
		Joins("LEFT JOIN holdings h ON h.account_id = a.id").
		Joins("LEFT JOIN stocks s ON s.id = h.stock_id").
		Where("a.kind = ? AND a.mode = ? AND u.is_bot = ? AND u.is_active = ?",
			models.KindPrimary, models.ModeSim, false, true).
		Group("u.id, u.name, u.avatar_url, a.balance").
		Order("value DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		utils.LogError("repo: top traders", err)
		return nil, err
	}
	return rows, nil
}

func (r *portfolioRepo) ListHoldings(ctx context.Context, accountID uuid.UUID) ([]models.Holding, error) {
	var holdings []models.Holding
	err := r.DB.WithContext(ctx).
		Preload("Stock").
		Where("account_id = ?", accountID).
		Find(&holdings).Error
	if err != nil {
		utils.LogError("repo: list holdings", err)
		return nil, err
	}
	return holdings, nil
}
