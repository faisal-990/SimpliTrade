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

// CopyAllocation is a user's capped sub-account that mirrors an investor.
type CopyAllocation struct {
	AccountID  uuid.UUID
	InvestorID uuid.UUID
}

// CopySource lists the active copy allocations to trade each cycle. Optional.
type CopySource interface {
	ListActiveCopies(ctx context.Context) ([]CopyAllocation, error)
}

// Runner ties the pieces together for a set of bots.
type Runner struct {
	market  MarketSource
	pf      PortfolioSource
	broker  broker.Broker
	perf    PerformanceStore
	bots    []Bot
	log     *slog.Logger
	refresh Refresher   // optional
	copies  CopySource  // optional
	clock   MarketClock // optional: when set, cycles run only during market hours
	wasOpen bool        // tracks open→closed transitions for the end-of-day log
}

// WithRefresher sets an optional market refresher run at the start of each cycle.
func (r *Runner) WithRefresher(ref Refresher) *Runner {
	r.refresh = ref
	return r
}

// WithCopies sets an optional source of user copy-allocations to trade each cycle.
func (r *Runner) WithCopies(src CopySource) *Runner {
	r.copies = src
	return r
}

// WithClock gates Run to the market session: cycles execute only when the clock
// reports the market open. Run starts immediately at the open and idles
// otherwise — production scheduling with no manual trigger.
func (r *Runner) WithClock(c MarketClock) *Runner {
	r.clock = c
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
		name       string
		roi        float64
	}
	scores := make([]score, 0, len(r.bots))
	botOrders := 0

	for _, bot := range r.bots {
		if !bot.Config.Identity.Enabled {
			continue
		}
		botOrders += r.tradeBot(ctx, bot, snapshot)

		// Revalue after trading to score the bot for the leaderboard.
		pf, err := r.pf.Load(ctx, bot.AccountID)
		if err != nil {
			r.log.Error("engine: reload portfolio", "investor", bot.Config.Identity.ID, "err", err)
			continue
		}
		scores = append(scores, score{bot.InvestorID, bot.Config.Identity.Name, roiOf(pf, prices)})
	}

	// Rank by ROI (highest first), tie-break by investor id for stable order.
	sort.SliceStable(scores, func(i, j int) bool {
		if scores[i].roi != scores[j].roi {
			return scores[i].roi > scores[j].roi
		}
		return scores[i].investorID.String() < scores[j].investorID.String()
	})
	for rank, s := range scores {
		if err := r.perf.SavePerformance(ctx, s.investorID, s.roi, rank+1); err != nil {
			r.log.Error("engine: save performance", "investor", s.investorID, "err", err)
		}
	}

	// Trade users' capped copy allocations using the mirrored investor's strategy.
	// These are not ranked — they belong to users, not the leaderboard.
	copyOrders := r.tradeCopies(ctx, snapshot)

	// Per-cycle summary: a compact digest of what the engine just did.
	if len(scores) > 0 {
		leader := scores[0]
		r.log.Info("engine: cycle complete",
			"bot_orders", botOrders,
			"copy_orders", copyOrders,
			"bots_ranked", len(scores),
			"leader", leader.name,
			"leader_roi", leader.roi,
		)
	}
	return nil
}

// tradeCopies runs each active copy allocation through its mirrored investor's
// strategy, trading the capped sub-account. Returns the number of executed orders.
func (r *Runner) tradeCopies(ctx context.Context, snapshot []decide.StockView) int {
	if r.copies == nil {
		return 0
	}
	allocs, err := r.copies.ListActiveCopies(ctx)
	if err != nil {
		r.log.Error("engine: list copy allocations", "err", err)
		return 0
	}
	if len(allocs) == 0 {
		return 0
	}
	// Index strategy configs by the investor they belong to.
	cfgByInvestor := make(map[uuid.UUID]strategy.Config, len(r.bots))
	for _, b := range r.bots {
		cfgByInvestor[b.InvestorID] = b.Config
	}
	executed := 0
	for _, a := range allocs {
		cfg, ok := cfgByInvestor[a.InvestorID]
		if !ok {
			continue
		}
		executed += r.tradeBot(ctx, Bot{InvestorID: a.InvestorID, AccountID: a.AccountID, Config: cfg}, snapshot)
	}
	return executed
}

// tradeBot loads the bot's portfolio, decides, and executes the resulting
// intents. Returns the number of orders that filled.
func (r *Runner) tradeBot(ctx context.Context, bot Bot, snapshot []decide.StockView) int {
	pf, err := r.pf.Load(ctx, bot.AccountID)
	if err != nil {
		r.log.Error("engine: load portfolio", "investor", bot.Config.Identity.ID, "err", err)
		return 0
	}
	intents := dedupe(decide.Decide(bot.Config, snapshot, pf))
	executed := 0
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
			continue
		}
		executed++
	}
	return executed
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
