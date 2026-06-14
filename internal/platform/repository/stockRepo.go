package repository

import (
	"context"
	"errors"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StockRepo is the persistence boundary for the tradable universe and its price
// history. Reads here back the user-facing dashboard (Tower 1); writes come from
// the engine's market poller (Tower 2).
type StockRepo interface {
	// Upsert inserts or updates a stock keyed by its unique symbol, refreshing
	// price, fundamentals, and metadata.
	Upsert(ctx context.Context, stock *models.Stock) error
	GetBySymbol(ctx context.Context, symbol string) (*models.Stock, error)
	List(ctx context.Context, limit, offset int) ([]models.Stock, error)
	ListSymbols(ctx context.Context) ([]string, error)

	InsertCandle(ctx context.Context, price *models.StockPrice) error
	GetCandles(ctx context.Context, symbol, interval string, limit int) ([]models.StockPrice, error)
}

type stockRepo struct {
	DB *gorm.DB
}

func NewStockRepo(db *gorm.DB) StockRepo {
	return &stockRepo{DB: db}
}

func (r *stockRepo) Upsert(ctx context.Context, stock *models.Stock) error {
	// On symbol conflict, refresh the volatile fields the poller maintains.
	return r.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "exchange", "sector", "current_price", "currency",
			"fundamentals", "updated_at",
		}),
	}).Create(stock).Error
}

func (r *stockRepo) GetBySymbol(ctx context.Context, symbol string) (*models.Stock, error) {
	var stock models.Stock
	err := r.DB.WithContext(ctx).Where("symbol = ?", symbol).First(&stock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		utils.LogError("repo: get stock by symbol", err)
		return nil, err
	}
	return &stock, nil
}

func (r *stockRepo) List(ctx context.Context, limit, offset int) ([]models.Stock, error) {
	if limit <= 0 || limit > 500 {
		limit = 100 // sane default + hard cap so listing never scans unbounded
	}
	var stocks []models.Stock
	err := r.DB.WithContext(ctx).Order("symbol ASC").Limit(limit).Offset(offset).Find(&stocks).Error
	if err != nil {
		utils.LogError("repo: list stocks", err)
		return nil, err
	}
	return stocks, nil
}

func (r *stockRepo) ListSymbols(ctx context.Context) ([]string, error) {
	var symbols []string
	err := r.DB.WithContext(ctx).Model(&models.Stock{}).Order("symbol ASC").Pluck("symbol", &symbols).Error
	if err != nil {
		utils.LogError("repo: list symbols", err)
		return nil, err
	}
	return symbols, nil
}

func (r *stockRepo) InsertCandle(ctx context.Context, price *models.StockPrice) error {
	return r.DB.WithContext(ctx).Create(price).Error
}

func (r *stockRepo) GetCandles(ctx context.Context, symbol, interval string, limit int) ([]models.StockPrice, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	var prices []models.StockPrice
	err := r.DB.WithContext(ctx).
		Joins("JOIN stocks ON stocks.id = stock_prices.stock_id").
		Where("stocks.symbol = ? AND stock_prices.interval = ?", symbol, interval).
		Order("stock_prices.timestamp DESC").
		Limit(limit).
		Find(&prices).Error
	if err != nil {
		utils.LogError("repo: get candles", err)
		return nil, err
	}
	return prices, nil
}
