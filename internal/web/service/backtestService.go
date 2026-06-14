package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/backtest"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/google/uuid"
)

// fundamentalsCaveat is surfaced to users: the replay uses today's fundamentals,
// since the platform stores only the latest snapshot, not a historical series.
const fundamentalsCaveat = "Prices and technical signals are fully historical; fundamental screens use the latest available fundamentals (no point-in-time fundamentals are stored). Treat results as indicative, not a precise historical record."

// maxBacktestBars caps how many daily bars per symbol a run replays.
const maxBacktestBars = 400

// BacktestService replays an investor's strategy over historical prices.
type BacktestService interface {
	Run(ctx context.Context, investorID string, days int, startCash float64) (*dto.BacktestResultDTO, error)
}

type backtestService struct {
	investors repository.InvestorRepo
	stocks    repository.StockRepo
	custom    repository.CustomStrategyRepo
	byName    map[string]strategy.Config // preset configs keyed by investor name
}

// NewBacktestService loads the preset strategy configs once and indexes them by
// the investor name they identify. Custom (user-authored) investors are resolved
// at run time from the custom-strategy repo by investor id.
func NewBacktestService(investors repository.InvestorRepo, stocks repository.StockRepo, custom repository.CustomStrategyRepo, strategiesDir string) (BacktestService, error) {
	configs, err := strategy.LoadDir(strategiesDir)
	if err != nil {
		return nil, err
	}
	byName := make(map[string]strategy.Config, len(configs))
	for _, c := range configs {
		byName[strings.ToLower(c.Identity.Name)] = c
	}
	return &backtestService{investors: investors, stocks: stocks, custom: custom, byName: byName}, nil
}

// customConfig returns the stored config for a user-authored investor, if any.
func (s *backtestService) customConfig(ctx context.Context, investorID uuid.UUID) (strategy.Config, bool) {
	if s.custom == nil {
		return strategy.Config{}, false
	}
	row, err := s.custom.GetByInvestorID(ctx, investorID)
	if err != nil {
		return strategy.Config{}, false
	}
	var cfg strategy.Config
	if err := json.Unmarshal([]byte(row.ConfigJSON), &cfg); err != nil {
		return strategy.Config{}, false
	}
	return cfg, true
}

func (s *backtestService) Run(ctx context.Context, investorID string, days int, startCash float64) (*dto.BacktestResultDTO, error) {
	id, err := uuid.Parse(investorID)
	if err != nil {
		return nil, httpx.BadRequest("invalid investor id")
	}
	inv, err := s.investors.GetInvestor(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, httpx.NotFound("investor not found")
		}
		return nil, httpx.Internal("could not load investor").WithCause(err)
	}
	// Custom investors carry their config in the DB (resolved by id); presets are
	// matched by name. Check custom first so a custom investor never collides with
	// a preset of the same name.
	cfg, ok := s.customConfig(ctx, id)
	if !ok {
		cfg, ok = s.byName[strings.ToLower(inv.Name)]
	}
	if !ok {
		return nil, httpx.BadRequest("no strategy definition found for this investor")
	}

	if days <= 0 || days > maxBacktestBars {
		days = 180
	}
	if startCash <= 0 {
		startCash = 100000
	}

	stocks, err := s.stocks.List(ctx, 500, 0)
	if err != nil {
		return nil, httpx.Internal("could not load market universe").WithCause(err)
	}
	hist := make([]backtest.SymbolHistory, 0, len(stocks))
	for _, st := range stocks {
		candles, err := s.stocks.GetCandles(ctx, st.Symbol, "1d", days)
		if err != nil || len(candles) == 0 {
			continue
		}
		bars := make([]backtest.Bar, 0, len(candles))
		for _, c := range candles { // GetCandles is newest-first; Run re-sorts
			bars = append(bars, backtest.Bar{Date: c.Timestamp, Close: c.Close})
		}
		hist = append(hist, backtest.SymbolHistory{
			Symbol: st.Symbol, Sector: st.Sector, AssetClass: st.AssetClass,
			Fundamentals: st.Fundamentals, Bars: bars,
		})
	}
	if len(hist) == 0 {
		return nil, httpx.BadRequest("no historical price data available to backtest")
	}

	r := backtest.Run(cfg, hist, backtest.Params{StartCash: startCash})

	out := &dto.BacktestResultDTO{
		InvestorID: inv.ID.String(), InvestorName: inv.Name, Strategy: inv.Strategy,
		StartCash: r.StartCash, FinalValue: r.FinalValue, EndCash: r.EndCash, ROI: r.ROI,
		MaxDrawdown: r.MaxDrawdown, WinRate: r.WinRate, TradeCount: r.TradeCount,
		BuyCount: r.BuyCount, SellCount: r.SellCount, HeldCount: len(r.Holdings),
		StartDate: r.StartDate.Unix(), EndDate: r.EndDate.Unix(),
		Equity:   make([]dto.BacktestPoint, 0, len(r.Equity)),
		Trades:   make([]dto.BacktestTrade, 0, len(r.Trades)),
		Holdings: make([]dto.BacktestHolding, 0, len(r.Holdings)),
		Note:     fundamentalsCaveat,
	}
	for _, p := range r.Equity {
		out.Equity = append(out.Equity, dto.BacktestPoint{Date: p.Date.Unix(), Value: p.Value})
	}
	for _, h := range r.Holdings {
		out.Holdings = append(out.Holdings, dto.BacktestHolding{
			Symbol: h.Symbol, Quantity: h.Quantity, AvgPrice: h.AvgPrice, LastPrice: h.LastPrice,
			MarketValue: h.MarketValue, UnrealizedPL: h.UnrealizedPL,
		})
	}
	// Most recent trades first for display.
	for i := len(r.Trades) - 1; i >= 0; i-- {
		t := r.Trades[i]
		out.Trades = append(out.Trades, dto.BacktestTrade{
			Date: t.Date.Unix(), Side: t.Side, Symbol: t.Symbol,
			Quantity: t.Quantity, Price: t.Price, TotalValue: t.TotalValue,
		})
	}
	return out, nil
}
