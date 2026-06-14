package repository

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CustomStrategyRepo persists user-authored investors. Listing/loading their
// configs lets the engine and backtester treat them exactly like preset bots.
type CustomStrategyRepo interface {
	Create(ctx context.Context, cs *models.CustomStrategy) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.CustomStrategy, error)
	ListAll(ctx context.Context) ([]models.CustomStrategy, error)
	GetByInvestorID(ctx context.Context, investorID uuid.UUID) (*models.CustomStrategy, error)
}

type customStrategyRepo struct{ DB *gorm.DB }

func NewCustomStrategyRepo(db *gorm.DB) CustomStrategyRepo { return &customStrategyRepo{DB: db} }

func (r *customStrategyRepo) Create(ctx context.Context, cs *models.CustomStrategy) error {
	return r.DB.WithContext(ctx).Create(cs).Error
}

func (r *customStrategyRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.CustomStrategy, error) {
	var out []models.CustomStrategy
	err := r.DB.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&out).Error
	if err != nil {
		utils.LogError("repo: list custom strategies", err)
		return nil, err
	}
	return out, nil
}

func (r *customStrategyRepo) ListAll(ctx context.Context) ([]models.CustomStrategy, error) {
	var out []models.CustomStrategy
	if err := r.DB.WithContext(ctx).Find(&out).Error; err != nil {
		utils.LogError("repo: list all custom strategies", err)
		return nil, err
	}
	return out, nil
}

func (r *customStrategyRepo) GetByInvestorID(ctx context.Context, investorID uuid.UUID) (*models.CustomStrategy, error) {
	var cs models.CustomStrategy
	err := r.DB.WithContext(ctx).Where("investor_id = ?", investorID).First(&cs).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		utils.LogError("repo: get custom strategy", err)
		return nil, err
	}
	return &cs, nil
}
