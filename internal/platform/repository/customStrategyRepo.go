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
	// Delete removes a custom investor the user owns. Any active copy-allocations
	// to it are liquidated first (cash returned to each owner), then the bot
	// identity is removed (follows, performance, account, holdings, trades cascade).
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
		// Liquidate any active copy-allocations to this investor first, returning
		// each owner's cash to their primary account — so deleting is one click.
		var copies []models.Account
		if err := tx.Where("kind = ? AND investor_id = ? AND is_active = ?", models.KindCopy, investorID, true).
			Find(&copies).Error; err != nil {
			return err
		}
		for _, c := range copies {
			var holdings []models.Holding
			if err := tx.Preload("Stock").Where("account_id = ?", c.ID).Find(&holdings).Error; err != nil {
				return err
			}
			proceeds := c.Balance
			for _, h := range holdings {
				proceeds += h.Quantity * h.Stock.CurrentPrice
			}
			if err := tx.Where("account_id = ?", c.ID).Delete(&models.Holding{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Account{}).
				Where("user_id = ? AND kind = ? AND mode = ?", c.UserID, models.KindPrimary, models.ModeSim).
				Update("balance", gorm.Expr("balance + ?", proceeds)).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Account{}).Where("id = ?", c.ID).
				Updates(map[string]any{"balance": 0, "is_active": false}).Error; err != nil {
				return err
			}
		}
		// Remove rows referencing the investor that aren't covered by a cascading
		// FK (performances has no ON DELETE CASCADE), then the investor + user.
		if err := tx.Where("investor_id = ?", investorID).Delete(&models.Performance{}).Error; err != nil {
			return err
		}
		if err := tx.Where("investor_id = ?", investorID).Delete(&models.Follow{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", investorID).Delete(&models.Investor{}).Error; err != nil {
			return err
		}
		// Deleting the user cascades to its account → holdings + trades.
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
