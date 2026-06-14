package service

import (
	"context"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/analytics"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// analyticsLookback is how many daily bars to reconstruct the equity curve over.
const analyticsLookback = 180

// benchmarkSize is how many universe names form the equal-weight market line.
const benchmarkSize = 25

// AnalyticsService reconstructs a user's portfolio performance + risk metrics
// from their trade history and price candles (reusing the candle data; no new
// state). It is read-only.
type AnalyticsService interface {
	Compute(ctx context.Context, accountID string) (*dto.AnalyticsDTO, error)
}

type analyticsService struct {
	trades     repository.TradeRepo
	portfolios repository.PortfolioRepo
	stocks     repository.StockRepo
}

func NewAnalyticsService(trades repository.TradeRepo, portfolios repository.PortfolioRepo, stocks repository.StockRepo) AnalyticsService {
	return &analyticsService{trades: trades, portfolios: portfolios, stocks: stocks}
}

func (s *analyticsService) Compute(ctx context.Context, accountID string) (*dto.AnalyticsDTO, error) {
	acct, err := uuid.Parse(accountID)
	if err != nil {
		return nil, httpx.Unauthorized("invalid account identity")
	}

	trades, err := s.trades.ListByAccount(ctx, acct, 1000, 0)
	if err != nil {
		return nil, httpx.Internal("could not load trades").WithCause(err)
	}
	holdings, err := s.portfolios.ListHoldings(ctx, acct)
	if err != nil {
		return nil, httpx.Internal("could not load holdings").WithCause(err)
	}

	// Symbols we need candles for: everything traded, plus a market sample.
	need := map[string]struct{}{}
	in := analytics.Input{StartCash: models.StartingSimBalance}
	for _, t := range trades {
		need[t.Stock.Symbol] = struct{}{}
		in.Trades = append(in.Trades, analytics.Trade{
			Date: t.ExecutedAt, Symbol: t.Stock.Symbol, Side: t.Type,
			Quantity: t.Quantity, Price: t.Price,
		})
	}
	// Live prices, so the curve's last point matches portfolio stats and reflects
	// market moves that haven't been written to the daily candle history yet.
	livePrice := map[string]float64{}
	for _, h := range holdings {
		in.Holdings = append(in.Holdings, analytics.Holding{
			Symbol: h.Stock.Symbol, Sector: h.Stock.Sector,
			MarketValue: h.Quantity * h.Stock.CurrentPrice,
		})
		livePrice[h.Stock.Symbol] = h.Stock.CurrentPrice
	}

	// Market benchmark = an equal-weight sample of the universe.
	universe, err := s.stocks.List(ctx, benchmarkSize, 0)
	if err != nil {
		return nil, httpx.Internal("could not load market sample").WithCause(err)
	}
	for _, st := range universe {
		need[st.Symbol] = struct{}{}
		in.BenchmarkSymbols = append(in.BenchmarkSymbols, st.Symbol)
		livePrice[st.Symbol] = st.CurrentPrice
	}

	now := time.Now().UTC()
	in.Histories = make(map[string][]analytics.Bar, len(need))
	for sym := range need {
		candles, err := s.stocks.GetCandles(ctx, sym, "1d", analyticsLookback)
		if err != nil || len(candles) == 0 {
			continue
		}
		bars := make([]analytics.Bar, 0, len(candles)+1)
		for _, c := range candles {
			bars = append(bars, analytics.Bar{Date: c.Timestamp, Close: c.Close})
		}
		// Append a live "now" bar so the curve ends at the current price.
		if lp := livePrice[sym]; lp > 0 {
			bars = append(bars, analytics.Bar{Date: now, Close: lp})
		}
		in.Histories[sym] = bars
	}

	r := analytics.Compute(in)
	out := &dto.AnalyticsDTO{
		StartValue: in.StartCash, ROI: r.ROI, MaxDrawdown: r.MaxDrawdown,
		Volatility: r.Volatility, Sharpe: r.Sharpe, WinRate: r.WinRate,
		TradeCount: r.TradeCount, BestDay: r.BestDay, WorstDay: r.WorstDay,
		BenchmarkName: "Market (equal-weight)",
		Equity:        make([]dto.AnalyticsPoint, 0, len(r.Equity)),
		Benchmark:     make([]dto.AnalyticsPoint, 0, len(r.Benchmark)),
		Sectors:       make([]dto.SectorSliceDTO, 0, len(r.Sectors)),
	}
	if n := len(r.Equity); n > 0 {
		out.CurrentValue = r.Equity[n-1].Value
	}
	for _, p := range r.Equity {
		out.Equity = append(out.Equity, dto.AnalyticsPoint{Date: p.Date.Unix(), Value: p.Value})
	}
	for _, p := range r.Benchmark {
		out.Benchmark = append(out.Benchmark, dto.AnalyticsPoint{Date: p.Date.Unix(), Value: p.Value})
	}
	for _, sec := range r.Sectors {
		out.Sectors = append(out.Sectors, dto.SectorSliceDTO{Sector: sec.Sector, Value: sec.Value, Pct: sec.Pct})
	}
	return out, nil
}
