package repository

import (
	"context"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ResetRepo restores a user to a clean slate — used by the dev "reset my
// account" control (handy after the simulated market data is regenerated).
type ResetRepo interface {
	ResetUser(ctx context.Context, userID string) error
}

type resetRepo struct{ DB *gorm.DB }

func NewResetRepo(db *gorm.DB) ResetRepo { return &resetRepo{DB: db} }

// ResetUser deletes the user's holdings, trades, and copy sub-accounts, and
// restores their primary cash to the starting balance — all in one transaction.
func (r *resetRepo) ResetUser(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var accountIDs []uuid.UUID
		if err := tx.Model(&models.Account{}).Where("user_id = ?", uid).Pluck("id", &accountIDs).Error; err != nil {
			return err
		}
		if len(accountIDs) > 0 {
			if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.Holding{}).Error; err != nil {
				return err
			}
			if err := tx.Unscoped().Where("account_id IN ?", accountIDs).Delete(&models.Trade{}).Error; err != nil {
				return err
			}
		}
		// Drop copy sub-accounts entirely.
		if err := tx.Where("user_id = ? AND kind = ?", uid, models.KindCopy).Delete(&models.Account{}).Error; err != nil {
			return err
		}
		// Restore the primary account to the starting balance.
		return tx.Model(&models.Account{}).
			Where("user_id = ? AND kind = ?", uid, models.KindPrimary).
			Update("balance", models.StartingSimBalance).Error
	})
}
