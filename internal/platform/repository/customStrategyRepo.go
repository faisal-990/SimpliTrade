package repository

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrInvestorInUse means a custom investor can't be deleted because money is
// still allocated to it.
var ErrInvestorInUse = errors.New("repository: investor has active allocations")

// CustomStrategyRepo persists user-authored investors. Listing/loading their
// configs lets the engine and backtester treat them exactly like preset bots.
type CustomStrategyRepo interface {
	Create(ctx context.Context, cs *models.CustomStrategy) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]models.CustomStrategy, error)
	ListAll(ctx context.Context) ([]models.CustomStrategy, error)
	GetByInvestorID(ctx context.Context, investorID uuid.UUID) (*models.CustomStrategy, error)
	// Delete removes a custom investor the user owns (identity cascades). It
	// refuses while active copy-allocations still reference it.
	Delete(ctx context.Context, userID, investorID uuid.UUID) error
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

func (r *customStrategyRepo) Delete(ctx context.Context, userID, investorID uuid.UUID) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cs models.CustomStrategy
		if err := tx.Where("investor_id = ? AND user_id = ?", investorID, userID).First(&cs).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound // not found, or not owned by this user
			}
			return err
		}
		// Refuse while anyone still has money allocated to this investor.
		var active int64
		if err := tx.Model(&models.Account{}).
			Where("kind = ? AND investor_id = ? AND is_active = ?", models.KindCopy, investorID, true).
			Count(&active).Error; err != nil {
			return err
		}
		if active > 0 {
			return ErrInvestorInUse
		}
		// Delete the identity — FK cascades remove follows, performance, the bot's
		// account, holdings and trades.
		if err := tx.Where("id = ?", investorID).Delete(&models.Investor{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", investorID).Delete(&models.User{}).Error; err != nil {
			return err
		}
		return tx.Delete(&cs).Error
	})
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
