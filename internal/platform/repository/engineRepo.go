package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PerformanceRepo persists the bot leaderboard (ROI + rank per investor).
type PerformanceRepo interface {
	SavePerformance(ctx context.Context, investorID uuid.UUID, roi float64, rank int) error
}

type performanceRepo struct{ DB *gorm.DB }

func NewPerformanceRepo(db *gorm.DB) PerformanceRepo { return &performanceRepo{DB: db} }

func (r *performanceRepo) SavePerformance(ctx context.Context, investorID uuid.UUID, roi float64, rank int) error {
	perf := models.Performance{
		InvestorID: investorID,
		ROI:        roi,
		Rank:       rank,
		LastUpdate: time.Now(),
	}
	// Upsert by the unique investor_id.
	return r.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "investor_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"roi", "rank", "last_update", "updated_at"}),
	}).Create(&perf).Error
}

// BotRepo provisions and lists the engine's bot investors (a User+Account+Investor
// triple per strategy).
type BotRepo interface {
	// UpsertBot ensures a bot exists for the given strategy and returns its
	// investor and sim-account ids. Idempotent: re-running leaves balances intact.
	UpsertBot(ctx context.Context, strategyID, name, bio, style string) (investorID, accountID uuid.UUID, err error)
}

type botRepo struct{ DB *gorm.DB }

func NewBotRepo(db *gorm.DB) BotRepo { return &botRepo{DB: db} }

func (r *botRepo) UpsertBot(ctx context.Context, strategyID, name, bio, style string) (uuid.UUID, uuid.UUID, error) {
	// Synthetic, stable email keys the bot user so re-seeding is idempotent.
	email := fmt.Sprintf("%s@bots.simplitrade.local", strategyID)

	var investorID, accountID uuid.UUID
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user models.User
		err := tx.Where("email = ?", email).First(&user).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			user = models.User{
				Name: name, Email: email, Password: "x", Role: "bot",
				IsBot: true, IsActive: true, EmailVerified: true,
				Accounts: []models.Account{{Mode: models.ModeSim, Currency: "USD", Balance: models.StartingSimBalance, IsActive: true}},
			}
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			// Investor profile shares the user's id (1:1).
			if err := tx.Create(&models.Investor{ID: user.ID, Bio: bio, Strategy: style}).Error; err != nil {
				return err
			}
		case err != nil:
			return err
		}

		investorID = user.ID
		var acct models.Account
		if err := tx.Where("user_id = ? AND mode = ?", user.ID, models.ModeSim).First(&acct).Error; err != nil {
			return err
		}
		accountID = acct.ID
		return nil
	})
	return investorID, accountID, err
}
