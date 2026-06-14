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

// DBRefresher polls the provider for fresh quotes and updates DB prices. This is
// the engine's market poller. It polls only the symbols actually in the database
// (what was seeded), not a static universe — so seeding a subset (SEED_LIMIT)
// keeps API usage proportionally small, which matters under a free-tier quota.
type DBRefresher struct {
	provider marketdata.Provider
	stocks   repository.StockRepo
	limit    int // 0 = all symbols; >0 caps symbols per refresh (free-tier safety)
}

func NewDBRefresher(p marketdata.Provider, stocks repository.StockRepo) *DBRefresher {
	return &DBRefresher{provider: p, stocks: stocks}
}

// WithLimit caps how many symbols are refreshed per cycle. On a tiny free-tier
// quota, refreshing the whole universe every tick blows the rate limit; capping
// keeps each refresh within budget. 0 means refresh everything.
func (r *DBRefresher) WithLimit(n int) *DBRefresher {
	r.limit = n
	return r
}

func (r *DBRefresher) Refresh(ctx context.Context) error {
	symbols, err := r.stocks.ListSymbols(ctx)
	if err != nil {
		return err
	}
	if len(symbols) == 0 {
		return nil
	}
	if r.limit > 0 && len(symbols) > r.limit {
		symbols = symbols[:r.limit]
	}
	quotes, err := r.provider.BatchQuotes(ctx, symbols)
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

// DBCopySource lists active copy allocations from the database for the engine.
type DBCopySource struct {
	allocs repository.AllocationRepo
}

func NewDBCopySource(allocs repository.AllocationRepo) *DBCopySource {
	return &DBCopySource{allocs: allocs}
}

func (s *DBCopySource) ListActiveCopies(ctx context.Context) ([]CopyAllocation, error) {
	rows, err := s.allocs.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]CopyAllocation, 0, len(rows))
	for _, r := range rows {
		out = append(out, CopyAllocation{AccountID: r.AccountID, InvestorID: r.InvestorID})
	}
	return out, nil
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
