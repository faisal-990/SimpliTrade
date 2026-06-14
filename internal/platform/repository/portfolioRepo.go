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
type PortfolioRepo interface {
	GetAccountByID(ctx context.Context, accountID uuid.UUID) (*models.Account, error)
	ListHoldings(ctx context.Context, accountID uuid.UUID) ([]models.Holding, error)
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
