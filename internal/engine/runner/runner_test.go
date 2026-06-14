package runner

import (
	"context"
	"io"
	"log/slog"
	"maps"
	"sync"
	"testing"

	"github.com/faisal-990/ProjectInvestApp/internal/broker"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/decide"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/google/uuid"
)

func pf64(f float64) *float64 { return &f }

func quietLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

// --- fakes (the only substitutions; Runner + Decide under test are real) ---

type fakeMarket struct{ views []decide.StockView }

func (f *fakeMarket) Snapshot(context.Context) ([]decide.StockView, error) { return f.views, nil }

// fakePortfolios stores a portfolio per account and applies executed orders so
// re-loads reflect trades (needed for ROI scoring).
type fakePortfolios struct {
	mu sync.Mutex
	m  map[uuid.UUID]*decide.Portfolio
}

func (f *fakePortfolios) Load(_ context.Context, id uuid.UUID) (decide.Portfolio, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p := f.m[id]
	// return a copy
	cp := decide.Portfolio{Cash: p.Cash, Positions: map[string]decide.Position{}}
	maps.Copy(cp.Positions, p.Positions)
	return cp, nil
}

// fakeBroker settles orders against the fakePortfolios at the snapshot price.
type fakeBroker struct {
	pf     *fakePortfolios
	prices map[string]float64
	execs  int
}

func (b *fakeBroker) Execute(_ context.Context, o broker.Order) (broker.Fill, error) {
	b.pf.mu.Lock()
	defer b.pf.mu.Unlock()
	price := b.prices[o.Symbol]
	p := b.pf.m[o.AccountID]
	cost := price * o.Quantity
	if o.Side == broker.Buy {
		if cost > p.Cash {
			return broker.Fill{}, broker.ErrInvalidSide // stand-in failure
		}
		pos := p.Positions[o.Symbol]
		newQty := pos.Quantity + o.Quantity
		p.Positions[o.Symbol] = decide.Position{Quantity: newQty, AvgPrice: price}
		p.Cash -= cost
	} else {
		pos := p.Positions[o.Symbol]
		p.Positions[o.Symbol] = decide.Position{Quantity: pos.Quantity - o.Quantity, AvgPrice: pos.AvgPrice}
		p.Cash += cost
	}
	b.execs++
	return broker.Fill{Symbol: o.Symbol, Quantity: o.Quantity, Price: price}, nil
}

type fakePerf struct {
	mu    sync.Mutex
	ranks map[uuid.UUID]int
	rois  map[uuid.UUID]float64
}

func (f *fakePerf) SavePerformance(_ context.Context, investorID uuid.UUID, roi float64, rank int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ranks[investorID] = rank
	f.rois[investorID] = roi
	return nil
}

func valueBot(id string, style string) strategy.Config {
	return strategy.Config{
		SchemaVersion: 2,
		Identity:      strategy.Identity{ID: id, Style: style, Enabled: true},
		Universe:      strategy.Universe{MaxPositions: 10},
		Buy: strategy.Buy{
			Gates:   strategy.Gates{Valuation: &strategy.ValuationGate{PEMax: pf64(15)}},
			Scoring: map[string]float64{"valuation": 1},
		},
		Risk: strategy.Risk{MaxPositionSize: 0.5, CashBufferMin: 0},
	}
}

func TestRunner_BotBuysAndPerformanceRanked(t *testing.T) {
	cheap := decide.StockView{Symbol: "CHEAP", Price: 40, Fundamentals: models.Fundamentals{PE: 8, PB: 1, EPSTTM: 5, BVPS: 30}}
	market := &fakeMarket{views: []decide.StockView{cheap}}
	prices := map[string]float64{"CHEAP": 40}

	// Two bots: one active value bot with cash, one disabled.
	activeAcct, idleAcct := uuid.New(), uuid.New()
	activeInv, idleInv := uuid.New(), uuid.New()
	pf := &fakePortfolios{m: map[uuid.UUID]*decide.Portfolio{
		activeAcct: {Cash: 100000, Positions: map[string]decide.Position{}},
		idleAcct:   {Cash: 100000, Positions: map[string]decide.Position{}},
	}}
	b := &fakeBroker{pf: pf, prices: prices}
	perf := &fakePerf{ranks: map[uuid.UUID]int{}, rois: map[uuid.UUID]float64{}}

	idle := valueBot("idle", "deep_value")
	idle.Identity.Enabled = false

	bots := []Bot{
		{InvestorID: activeInv, AccountID: activeAcct, Config: valueBot("active", "deep_value")},
		{InvestorID: idleInv, AccountID: idleAcct, Config: idle},
	}
	r := New(market, pf, b, perf, bots, quietLog())

	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	// The active bot should have traded; the disabled bot should not be scored.
	if b.execs == 0 {
		t.Error("expected the active bot to execute at least one buy")
	}
	if pf.m[activeAcct].Cash >= 100000 {
		t.Error("active bot's cash should have decreased after buying")
	}
	if _, scored := perf.ranks[idleInv]; scored {
		t.Error("disabled bot must not appear on the leaderboard")
	}
	if perf.ranks[activeInv] != 1 {
		t.Errorf("active bot rank = %d, want 1 (only scored bot)", perf.ranks[activeInv])
	}
}

func TestRunner_RanksByROIDescending(t *testing.T) {
	// Two bots already holding the same stock at different cost bases; the one
	// with the lower cost basis has the higher ROI and should rank #1.
	market := &fakeMarket{views: []decide.StockView{
		{Symbol: "X", Price: 150, Fundamentals: models.Fundamentals{PE: 30}}, // fails gate -> no new trades
	}}
	winnerAcct, loserAcct := uuid.New(), uuid.New()
	winnerInv, loserInv := uuid.New(), uuid.New()
	// ROI is total value vs starting capital (100k). Winner holds more market
	// value (1000·150 = 150k → +50%); loser holds less (500·150 = 75k → -25%).
	pf := &fakePortfolios{m: map[uuid.UUID]*decide.Portfolio{
		winnerAcct: {Cash: 0, Positions: map[string]decide.Position{"X": {Quantity: 1000, AvgPrice: 50}}},
		loserAcct:  {Cash: 0, Positions: map[string]decide.Position{"X": {Quantity: 500, AvgPrice: 140}}},
	}}
	b := &fakeBroker{pf: pf, prices: map[string]float64{"X": 150}}
	perf := &fakePerf{ranks: map[uuid.UUID]int{}, rois: map[uuid.UUID]float64{}}

	bots := []Bot{
		{InvestorID: loserInv, AccountID: loserAcct, Config: valueBot("loser", "deep_value")},
		{InvestorID: winnerInv, AccountID: winnerAcct, Config: valueBot("winner", "deep_value")},
	}
	r := New(market, pf, b, perf, bots, quietLog())
	if err := r.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if perf.ranks[winnerInv] != 1 {
		t.Errorf("higher-ROI bot rank = %d, want 1", perf.ranks[winnerInv])
	}
	if perf.ranks[loserInv] != 2 {
		t.Errorf("lower-ROI bot rank = %d, want 2", perf.ranks[loserInv])
	}
	if perf.rois[winnerInv] <= perf.rois[loserInv] {
		t.Errorf("winner ROI %.3f should exceed loser ROI %.3f", perf.rois[winnerInv], perf.rois[loserInv])
	}
}

func TestDedupe_SellWinsOverBuy(t *testing.T) {
	out := dedupe([]decide.Intent{
		{Action: decide.Buy, Symbol: "A", Quantity: 1},
		{Action: decide.Sell, Symbol: "A", Quantity: 2},
		{Action: decide.Buy, Symbol: "B", Quantity: 3},
	})
	bySym := map[string]decide.Intent{}
	for _, in := range out {
		bySym[in.Symbol] = in
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 deduped intents, got %d", len(out))
	}
	if bySym["A"].Action != decide.Sell {
		t.Errorf("symbol A should resolve to SELL, got %s", bySym["A"].Action)
	}
}
