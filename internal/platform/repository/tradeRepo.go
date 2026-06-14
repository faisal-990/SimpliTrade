package repository

import (
	"context"
	"errors"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/trading"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TradeRepo executes trades atomically and reads trade history. Execution runs
// in a single transaction with the account row locked FOR UPDATE, so concurrent
// trades on the same account serialize and balances never go negative.
type TradeRepo interface {
	ExecuteBuy(ctx context.Context, accountID uuid.UUID, symbol string, qty float64, idemKey *string) (*models.Trade, error)
	ExecuteSell(ctx context.Context, accountID uuid.UUID, symbol string, qty float64, idemKey *string) (*models.Trade, error)
	ListByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]models.Trade, error)
}

type tradeRepo struct {
	DB *gorm.DB
}

func NewTradeRepo(db *gorm.DB) TradeRepo {
	return &tradeRepo{DB: db}
}

// tradeSide is the kind of execution; it selects which pure settlement function
// runs and the recorded trade type.
type tradeSide string

const (
	sideBuy  tradeSide = "buy"
	sideSell tradeSide = "sell"
)

func (r *tradeRepo) ExecuteBuy(ctx context.Context, accountID uuid.UUID, symbol string, qty float64, idemKey *string) (*models.Trade, error) {
	return r.execute(ctx, sideBuy, accountID, symbol, qty, idemKey)
}

func (r *tradeRepo) ExecuteSell(ctx context.Context, accountID uuid.UUID, symbol string, qty float64, idemKey *string) (*models.Trade, error) {
	return r.execute(ctx, sideSell, accountID, symbol, qty, idemKey)
}

func (r *tradeRepo) execute(ctx context.Context, side tradeSide, accountID uuid.UUID, symbol string, qty float64, idemKey *string) (*models.Trade, error) {
	var result *models.Trade
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Idempotency: a repeated key returns the original trade, never re-executes.
		if idemKey != nil {
			var existing models.Trade
			err := tx.Where("idempotency_key = ?", *idemKey).First(&existing).Error
			if err == nil {
				result = &existing
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		var stock models.Stock
		if err := tx.Where("symbol = ?", symbol).First(&stock).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		// Lock the account row so concurrent trades serialize.
		var account models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&account, "id = ?", accountID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		// Load the current position (may not exist yet).
		var holding models.Holding
		hErr := tx.Where("account_id = ? AND stock_id = ?", accountID, stock.ID).First(&holding).Error
		hasHolding := hErr == nil
		if hErr != nil && !errors.Is(hErr, gorm.ErrRecordNotFound) {
			return hErr
		}
		current := trading.Position{Quantity: holding.Quantity, AvgPrice: holding.AvgPrice}
		price := stock.CurrentPrice

		var newBalance, total float64
		var newPos trading.Position
		switch side {
		case sideBuy:
			res, err := trading.ComputeBuy(account.Balance, current, price, qty)
			if err != nil {
				return err
			}
			newBalance, newPos, total = res.NewBalance, res.NewPosition, res.Cost
		case sideSell:
			res, err := trading.ComputeSell(account.Balance, current, price, qty)
			if err != nil {
				return err
			}
			newBalance, newPos, total = res.NewBalance, res.NewPosition, res.Proceeds
		}

		// Persist the new balance.
		if err := tx.Model(&account).Update("balance", newBalance).Error; err != nil {
			return err
		}

		// Persist the new position: delete when fully closed, else upsert.
		if newPos.Quantity == 0 {
			if hasHolding {
				if err := tx.Delete(&holding).Error; err != nil {
					return err
				}
			}
		} else if hasHolding {
			if err := tx.Model(&holding).Updates(map[string]any{
				"quantity":  newPos.Quantity,
				"avg_price": newPos.AvgPrice,
			}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Create(&models.Holding{
				AccountID: accountID, StockID: stock.ID,
				Quantity: newPos.Quantity, AvgPrice: newPos.AvgPrice,
			}).Error; err != nil {
				return err
			}
		}

		trade := &models.Trade{
			AccountID:      accountID,
			StockID:        stock.ID,
			Type:           string(side),
			Quantity:       qty,
			Price:          price,
			TotalValue:     total,
			ExecutedAt:     time.Now(),
			Status:         "executed",
			IdempotencyKey: idemKey,
		}
		if err := tx.Create(trade).Error; err != nil {
			return err
		}
		result = trade
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *tradeRepo) ListByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]models.Trade, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var trades []models.Trade
	err := r.DB.WithContext(ctx).
		Preload("Stock").
		Where("account_id = ?", accountID).
		Order("executed_at DESC").
		Limit(limit).Offset(offset).
		Find(&trades).Error
	if err != nil {
		utils.LogError("repo: list trades by account", err)
		return nil, err
	}
	return trades, nil
}
