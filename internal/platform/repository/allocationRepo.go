package repository

import (
	"context"
	"errors"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrInsufficientCapital means the primary account can't fund the requested cap.
var ErrInsufficientCapital = errors.New("repository: insufficient capital in primary account")

// AllocationView is a copy sub-account projected with the investor it mirrors and
// its current valuation.
type AllocationView struct {
	ID           uuid.UUID
	InvestorID   uuid.UUID
	InvestorName string
	Strategy     string
	Capital      float64
	Cash         float64
	MarketValue  float64 // cash + holdings at current price
	IsActive     bool
	CreatedAt    time.Time
}

// CopyAccount identifies an active copy sub-account for the engine to trade.
type CopyAccount struct {
	AccountID  uuid.UUID
	InvestorID uuid.UUID
}

// AllocationHolding is one position the bot opened inside a copy sub-account.
type AllocationHolding struct {
	Symbol       string
	Quantity     float64
	AvgPrice     float64
	CurrentPrice float64
	MarketValue  float64
}

// AllocationTrade is one order the bot executed inside a copy sub-account.
type AllocationTrade struct {
	Symbol     string
	Side       string
	Quantity   float64
	Price      float64
	TotalValue float64
	ExecutedAt time.Time
	Reason     string
}

// AllocationActivity is what the mirrored investor's bot has done with a user's
// allocated capital: its current positions and recent orders.
type AllocationActivity struct {
	View     AllocationView
	Holdings []AllocationHolding
	Trades   []AllocationTrade
}

// AllocationRepo manages capped copy sub-accounts: funding them from the primary
// account, listing them, and stopping (liquidating) them.
type AllocationRepo interface {
	Create(ctx context.Context, userID, investorID uuid.UUID, capital float64) (*models.Account, error)
	List(ctx context.Context, userID uuid.UUID) ([]AllocationView, error)
	Activity(ctx context.Context, userID, allocationID uuid.UUID) (*AllocationActivity, error)
	Stop(ctx context.Context, userID, allocationID uuid.UUID) error
	ListActive(ctx context.Context) ([]CopyAccount, error)
}

type allocationRepo struct{ DB *gorm.DB }

func NewAllocationRepo(db *gorm.DB) AllocationRepo { return &allocationRepo{DB: db} }

// Create transfers `capital` from the user's primary account into a new capped
// copy account that mirrors the given investor — atomically, with the primary
// row locked so the balance can't go negative.
func (r *allocationRepo) Create(ctx context.Context, userID, investorID uuid.UUID, capital float64) (*models.Account, error) {
	if capital <= 0 {
		return nil, errors.New("repository: capital must be positive")
	}
	var copyAcct *models.Account
	err := r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var primary models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND kind = ? AND mode = ?", userID, models.KindPrimary, models.ModeSim).
			First(&primary).Error; err != nil {
			return err
		}
		if capital > primary.Balance {
			return ErrInsufficientCapital
		}
		if err := tx.Model(&primary).Update("balance", primary.Balance-capital).Error; err != nil {
			return err
		}
		acct := &models.Account{
			UserID:     userID,
			Kind:       models.KindCopy,
			Mode:       models.ModeSim,
			Currency:   "USD",
			Balance:    capital, // the copy account trades only this slice
			Capital:    capital,
			InvestorID: &investorID,
			IsActive:   true,
		}
		if err := tx.Create(acct).Error; err != nil {
			return err
		}
		copyAcct = acct
		return nil
	})
	if err != nil {
		return nil, err
	}
	return copyAcct, nil
}

func (r *allocationRepo) List(ctx context.Context, userID uuid.UUID) ([]AllocationView, error) {
	var accounts []models.Account
	if err := r.DB.WithContext(ctx).
		Where("user_id = ? AND kind = ?", userID, models.KindCopy).
		Order("created_at DESC").Find(&accounts).Error; err != nil {
		utils.LogError("repo: list allocations", err)
		return nil, err
	}

	out := make([]AllocationView, 0, len(accounts))
	for _, a := range accounts {
		v := AllocationView{
			ID: a.ID, Capital: a.Capital, Cash: a.Balance,
			MarketValue: a.Balance, IsActive: a.IsActive, CreatedAt: a.CreatedAt,
		}
		if a.InvestorID != nil {
			v.InvestorID = *a.InvestorID
			// investor name + strategy
			var inv struct {
				Name     string
				Strategy string
			}
			r.DB.WithContext(ctx).Table("investors").
				Select("users.name AS name, investors.strategy AS strategy").
				Joins("JOIN users ON users.id = investors.id").
				Where("investors.id = ?", *a.InvestorID).Scan(&inv)
			v.InvestorName = inv.Name
			v.Strategy = inv.Strategy
		}
		// add holdings market value at current price
		var holdings []models.Holding
		r.DB.WithContext(ctx).Preload("Stock").Where("account_id = ?", a.ID).Find(&holdings)
		for _, h := range holdings {
			v.MarketValue += h.Quantity * h.Stock.CurrentPrice
		}
		out = append(out, v)
	}
	return out, nil
}

// Activity returns what the mirrored investor's bot has done with a user's
// allocated capital: the copy sub-account's current positions and recent orders.
// Ownership is enforced (the copy account must belong to the user).
func (r *allocationRepo) Activity(ctx context.Context, userID, allocationID uuid.UUID) (*AllocationActivity, error) {
	var acct models.Account
	if err := r.DB.WithContext(ctx).
		Where("id = ? AND user_id = ? AND kind = ?", allocationID, userID, models.KindCopy).
		First(&acct).Error; err != nil {
		return nil, err
	}

	view := AllocationView{
		ID: acct.ID, Capital: acct.Capital, Cash: acct.Balance,
		MarketValue: acct.Balance, IsActive: acct.IsActive, CreatedAt: acct.CreatedAt,
	}
	if acct.InvestorID != nil {
		view.InvestorID = *acct.InvestorID
		var inv struct {
			Name     string
			Strategy string
		}
		r.DB.WithContext(ctx).Table("investors").
			Select("users.name AS name, investors.strategy AS strategy").
			Joins("JOIN users ON users.id = investors.id").
			Where("investors.id = ?", *acct.InvestorID).Scan(&inv)
		view.InvestorName = inv.Name
		view.Strategy = inv.Strategy
	}

	var holdings []models.Holding
	if err := r.DB.WithContext(ctx).Preload("Stock").
		Where("account_id = ?", acct.ID).Find(&holdings).Error; err != nil {
		return nil, err
	}
	outHoldings := make([]AllocationHolding, 0, len(holdings))
	for _, h := range holdings {
		mv := h.Quantity * h.Stock.CurrentPrice
		view.MarketValue += mv
		outHoldings = append(outHoldings, AllocationHolding{
			Symbol: h.Stock.Symbol, Quantity: h.Quantity, AvgPrice: h.AvgPrice,
			CurrentPrice: h.Stock.CurrentPrice, MarketValue: mv,
		})
	}

	var trades []models.Trade
	if err := r.DB.WithContext(ctx).Preload("Stock").
		Where("account_id = ?", acct.ID).
		Order("executed_at DESC").Limit(50).Find(&trades).Error; err != nil {
		return nil, err
	}
	outTrades := make([]AllocationTrade, 0, len(trades))
	for _, t := range trades {
		outTrades = append(outTrades, AllocationTrade{
			Reason: t.Reason,
			Symbol: t.Stock.Symbol, Side: t.Type, Quantity: t.Quantity,
			Price: t.Price, TotalValue: t.TotalValue, ExecutedAt: t.ExecutedAt,
		})
	}

	return &AllocationActivity{View: view, Holdings: outHoldings, Trades: outTrades}, nil
}

// Stop liquidates the copy account's holdings at current price, returns all cash
// to the primary account, and marks the copy inactive — atomically.
func (r *allocationRepo) Stop(ctx context.Context, userID, allocationID uuid.UUID) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var copyAcct models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ? AND kind = ?", allocationID, userID, models.KindCopy).
			First(&copyAcct).Error; err != nil {
			return err
		}

		var holdings []models.Holding
		if err := tx.Preload("Stock").Where("account_id = ?", copyAcct.ID).Find(&holdings).Error; err != nil {
			return err
		}
		proceeds := copyAcct.Balance
		for _, h := range holdings {
			proceeds += h.Quantity * h.Stock.CurrentPrice
		}
		if err := tx.Where("account_id = ?", copyAcct.ID).Delete(&models.Holding{}).Error; err != nil {
			return err
		}

		var primary models.Account
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ? AND kind = ? AND mode = ?", userID, models.KindPrimary, models.ModeSim).
			First(&primary).Error; err != nil {
			return err
		}
		if err := tx.Model(&primary).Update("balance", primary.Balance+proceeds).Error; err != nil {
			return err
		}
		return tx.Model(&copyAcct).Updates(map[string]any{"balance": 0, "is_active": false}).Error
	})
}

func (r *allocationRepo) ListActive(ctx context.Context) ([]CopyAccount, error) {
	var accounts []models.Account
	if err := r.DB.WithContext(ctx).
		Where("kind = ? AND is_active = ? AND investor_id IS NOT NULL", models.KindCopy, true).
		Find(&accounts).Error; err != nil {
		return nil, err
	}
	out := make([]CopyAccount, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, CopyAccount{AccountID: a.ID, InvestorID: *a.InvestorID})
	}
	return out, nil
}
