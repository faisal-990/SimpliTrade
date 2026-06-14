package runner

import (
	"context"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/decide"
	"github.com/faisal-990/ProjectInvestApp/internal/marketdata"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/google/uuid"
)

// candleHistory is how many recent daily closes to load per stock for the
// momentum/trend indicators (~1 trading year).
const candleHistory = 260

// DBMarketSource builds the engine's market snapshot from the database: current
// price + fundamentals per stock, plus recent closes for technical indicators.
type DBMarketSource struct {
	stocks repository.StockRepo
}

func NewDBMarketSource(stocks repository.StockRepo) *DBMarketSource {
	return &DBMarketSource{stocks: stocks}
}

func (s *DBMarketSource) Snapshot(ctx context.Context) ([]decide.StockView, error) {
	stocks, err := s.stocks.List(ctx, 500, 0)
	if err != nil {
		return nil, err
	}
	views := make([]decide.StockView, 0, len(stocks))
	for _, st := range stocks {
		closes, err := s.closesFor(ctx, st.Symbol)
		if err != nil {
			return nil, err
		}
		views = append(views, decide.StockView{
			Symbol:       st.Symbol,
			Sector:       st.Sector,
			AssetClass:   st.AssetClass,
			Price:        st.CurrentPrice,
			Fundamentals: st.Fundamentals,
			Closes:       closes,
		})
	}
	return views, nil
}

// closesFor returns recent daily closes oldest→newest (GetCandles returns
// newest-first, so we reverse).
func (s *DBMarketSource) closesFor(ctx context.Context, symbol string) ([]float64, error) {
	candles, err := s.stocks.GetCandles(ctx, symbol, "1d", candleHistory)
	if err != nil {
		return nil, err
	}
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[len(candles)-1-i] = c.Close
	}
	return closes, nil
}

// DBPortfolioSource loads an account's cash + positions from the database.
type DBPortfolioSource struct {
	portfolios repository.PortfolioRepo
}

func NewDBPortfolioSource(p repository.PortfolioRepo) *DBPortfolioSource {
	return &DBPortfolioSource{portfolios: p}
}

func (s *DBPortfolioSource) Load(ctx context.Context, accountID uuid.UUID) (decide.Portfolio, error) {
	acct, err := s.portfolios.GetAccountByID(ctx, accountID)
	if err != nil {
		return decide.Portfolio{}, err
	}
	holdings, err := s.portfolios.ListHoldings(ctx, accountID)
	if err != nil {
		return decide.Portfolio{}, err
	}
	positions := make(map[string]decide.Position, len(holdings))
	for _, h := range holdings {
		positions[h.Stock.Symbol] = decide.Position{Quantity: h.Quantity, AvgPrice: h.AvgPrice}
	}
	return decide.Portfolio{Cash: acct.Balance, Positions: positions}, nil
}

// DBRefresher polls the provider for fresh quotes and updates DB prices + appends
// a candle. This is the engine's market poller (slow lane).
type DBRefresher struct {
	provider marketdata.Provider
	stocks   repository.StockRepo
	symbols  []string
}

func NewDBRefresher(p marketdata.Provider, stocks repository.StockRepo, symbols []string) *DBRefresher {
	return &DBRefresher{provider: p, stocks: stocks, symbols: symbols}
}

func (r *DBRefresher) Refresh(ctx context.Context) error {
	quotes, err := r.provider.BatchQuotes(ctx, r.symbols)
	if err != nil {
		return err
	}
	for sym, q := range quotes {
		if err := r.stocks.UpdatePrice(ctx, sym, q.Price); err != nil {
			return err
		}
	}
	return nil
}

// PerformanceStoreAdapter adapts the PerformanceRepo to the runner's interface.
type PerformanceStoreAdapter struct {
	repo repository.PerformanceRepo
}

func NewPerformanceStore(repo repository.PerformanceRepo) *PerformanceStoreAdapter {
	return &PerformanceStoreAdapter{repo: repo}
}

func (a *PerformanceStoreAdapter) SavePerformance(ctx context.Context, investorID uuid.UUID, roi float64, rank int) error {
	return a.repo.SavePerformance(ctx, investorID, roi, rank)
}
