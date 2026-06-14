// Package runner orchestrates the strategy engine: on each tick it runs every
// bot investor's Decide() against the current market, executes the resulting
// intents through the broker, and recomputes the performance leaderboard. It
// depends only on narrow interfaces, so the orchestration is testable with fakes
// while the real implementations are thin DB/provider adapters.
package runner

import (
	"context"
	"log/slog"
	"sort"

	"github.com/faisal-990/ProjectInvestApp/internal/broker"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/decide"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/portfolio"
	"github.com/google/uuid"
)

// Bot is an engine-driven investor persona: its strategy and the account it
// trades.
type Bot struct {
	InvestorID uuid.UUID
	AccountID  uuid.UUID
	Config     strategy.Config
}

// MarketSource provides the current market snapshot the engine decides against.
type MarketSource interface {
	Snapshot(ctx context.Context) ([]decide.StockView, error)
}

// PortfolioSource loads an account's cash + positions for decision-making and
// valuation.
type PortfolioSource interface {
	Load(ctx context.Context, accountID uuid.UUID) (decide.Portfolio, error)
}

// PerformanceStore persists each bot's ROI and leaderboard rank.
type PerformanceStore interface {
	SavePerformance(ctx context.Context, investorID uuid.UUID, roi float64, rank int) error
}

// Refresher pulls fresh market data (prices) before a decision cycle. It is
// optional; the slow lane uses it to keep DB prices current.
type Refresher interface {
	Refresh(ctx context.Context) error
}

// Runner ties the pieces together for a set of bots.
type Runner struct {
	market  MarketSource
	pf      PortfolioSource
	broker  broker.Broker
	perf    PerformanceStore
	bots    []Bot
	log     *slog.Logger
	refresh Refresher // optional
}

// WithRefresher sets an optional market refresher run at the start of each cycle.
func (r *Runner) WithRefresher(ref Refresher) *Runner {
	r.refresh = ref
	return r
}

// New builds a Runner.
func New(market MarketSource, pf PortfolioSource, b broker.Broker, perf PerformanceStore, bots []Bot, log *slog.Logger) *Runner {
	return &Runner{market: market, pf: pf, broker: b, perf: perf, bots: bots, log: log}
}

// RunOnce executes a single decision cycle: every enabled bot decides and trades,
// then the leaderboard is recomputed. A snapshot failure aborts the cycle; any
// per-bot or per-order failure (e.g. insufficient funds) is logged and skipped —
// one bot's bad tick never blocks the others.
func (r *Runner) RunOnce(ctx context.Context) error {
	// Slow lane: refresh prices first (best-effort — stale data beats no cycle).
	if r.refresh != nil {
		if err := r.refresh.Refresh(ctx); err != nil {
			r.log.Error("engine: market refresh", "err", err)
		}
	}

	snapshot, err := r.market.Snapshot(ctx)
	if err != nil {
		return err
	}
	prices := make(map[string]float64, len(snapshot))
	for _, v := range snapshot {
		prices[v.Symbol] = v.Price
	}

	type score struct {
		investorID uuid.UUID
		roi        float64
	}
	scores := make([]score, 0, len(r.bots))

	for _, bot := range r.bots {
		if !bot.Config.Identity.Enabled {
			continue
		}
		r.tradeBot(ctx, bot, snapshot)

		// Revalue after trading to score the bot for the leaderboard.
		pf, err := r.pf.Load(ctx, bot.AccountID)
		if err != nil {
			r.log.Error("engine: reload portfolio", "investor", bot.Config.Identity.ID, "err", err)
			continue
		}
		scores = append(scores, score{bot.InvestorID, roiOf(pf, prices)})
	}

	// Rank by ROI (highest first) and persist.
	sort.SliceStable(scores, func(i, j int) bool { return scores[i].roi > scores[j].roi })
	for rank, s := range scores {
		if err := r.perf.SavePerformance(ctx, s.investorID, s.roi, rank+1); err != nil {
			r.log.Error("engine: save performance", "investor", s.investorID, "err", err)
		}
	}
	return nil
}

// tradeBot loads the bot's portfolio, decides, and executes the resulting intents.
func (r *Runner) tradeBot(ctx context.Context, bot Bot, snapshot []decide.StockView) {
	pf, err := r.pf.Load(ctx, bot.AccountID)
	if err != nil {
		r.log.Error("engine: load portfolio", "investor", bot.Config.Identity.ID, "err", err)
		return
	}
	intents := dedupe(decide.Decide(bot.Config, snapshot, pf))
	for _, in := range intents {
		order := broker.Order{
			AccountID: bot.AccountID,
			Symbol:    in.Symbol,
			Side:      sideOf(in.Action),
			Quantity:  in.Quantity,
		}
		if _, err := r.broker.Execute(ctx, order); err != nil {
			// Insufficient funds / shares are normal outcomes, not failures.
			r.log.Debug("engine: order skipped", "investor", bot.Config.Identity.ID,
				"symbol", in.Symbol, "action", in.Action, "err", err)
		}
	}
}

// dedupe ensures at most one intent per symbol; a SELL wins over a BUY for the
// same symbol in the same cycle (exit beats entry).
func dedupe(intents []decide.Intent) []decide.Intent {
	bySymbol := make(map[string]decide.Intent, len(intents))
	for _, in := range intents {
		if existing, ok := bySymbol[in.Symbol]; ok {
			if existing.Action == decide.Sell || in.Action != decide.Sell {
				continue // keep the existing sell, or ignore a second buy
			}
		}
		bySymbol[in.Symbol] = in
	}
	out := make([]decide.Intent, 0, len(bySymbol))
	for _, in := range bySymbol {
		out = append(out, in)
	}
	return out
}

// roiOf values a portfolio at current prices and returns ROI vs starting capital.
func roiOf(pf decide.Portfolio, prices map[string]float64) float64 {
	positions := make([]portfolio.Position, 0, len(pf.Positions))
	for sym, pos := range pf.Positions {
		positions = append(positions, portfolio.Position{
			Symbol: sym, Quantity: pos.Quantity, AvgPrice: pos.AvgPrice, CurrentPrice: prices[sym],
		})
	}
	return portfolio.Value(pf.Cash, models.StartingSimBalance, positions).ROI
}

func sideOf(a decide.Action) broker.Side {
	if a == decide.Sell {
		return broker.Sell
	}
	return broker.Buy
}
