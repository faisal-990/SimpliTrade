package decide

import (
	"testing"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

func pf64(f float64) *float64 { return &f }

// findIntent returns the first intent matching symbol+action, or nil.
func findIntent(intents []Intent, sym string, action Action) *Intent {
	for i := range intents {
		if intents[i].Symbol == sym && intents[i].Action == action {
			return &intents[i]
		}
	}
	return nil
}

// valueConfig is a minimal Graham-style fundamental strategy for tests.
func valueConfig() strategy.Config {
	return strategy.Config{
		SchemaVersion: 2,
		Identity:      strategy.Identity{ID: "test_value", Style: "deep_value", Enabled: true},
		Universe:      strategy.Universe{MaxPositions: 10},
		Buy: strategy.Buy{
			Gates: strategy.Gates{
				Valuation: &strategy.ValuationGate{PEMax: pf64(15), PBMax: pf64(1.5), UseGrahamNumber: true},
			},
			Scoring: map[string]float64{"valuation": 1, "intrinsic": 1},
		},
		Risk: strategy.Risk{MaxPositionSize: 0.2, CashBufferMin: 0},
	}
}

func TestDecide_FundamentalBuysCheapRejectsExpensive(t *testing.T) {
	cfg := valueConfig()
	// CHEAP: PE 8, PB 1.2, EPS 5, BVPS 30 -> graham ≈ 58, price 40 < graham -> pass.
	cheap := StockView{Symbol: "CHEAP", Price: 40, Fundamentals: models.Fundamentals{
		PE: 8, PB: 1.2, EPSTTM: 5, BVPS: 30,
	}}
	// RICH: PE 30 -> fails the P/E gate.
	rich := StockView{Symbol: "RICH", Price: 400, Fundamentals: models.Fundamentals{
		PE: 30, PB: 5, EPSTTM: 10, BVPS: 20,
	}}
	pf := Portfolio{Cash: 100000, Positions: map[string]Position{}}

	intents := Decide(cfg, []StockView{cheap, rich}, pf)

	if findIntent(intents, "CHEAP", Buy) == nil {
		t.Error("expected a BUY for the cheap stock that passes Graham gates")
	}
	if findIntent(intents, "RICH", Buy) != nil {
		t.Error("must NOT buy the stock that fails the P/E gate")
	}
}

func TestDecide_FundamentalTakeProfitSell(t *testing.T) {
	cfg := valueConfig()
	cfg.Sell.TakeProfitVsIntrinsic = pf64(1.0) // sell at/above intrinsic

	// EPS 5, growth ~0 -> intrinsic = 5*8.5 = 42.5; price 50 >= 42.5 -> sell.
	held := StockView{Symbol: "WIN", Price: 50, Fundamentals: models.Fundamentals{EPSTTM: 5}}
	pf := Portfolio{Cash: 0, Positions: map[string]Position{"WIN": {Quantity: 10, AvgPrice: 30}}}

	intents := Decide(cfg, []StockView{held}, pf)
	if findIntent(intents, "WIN", Sell) == nil {
		t.Error("expected a SELL when price reaches intrinsic value")
	}
}

func TestDecide_PortfolioAware_AlreadyHoldEnough_NoBuy(t *testing.T) {
	cfg := valueConfig() // max position size 20%
	cheap := StockView{Symbol: "CHEAP", Price: 40, Fundamentals: models.Fundamentals{PE: 8, PB: 1.2, EPSTTM: 5, BVPS: 30}}

	// Total value 100k, max position 20k. Already hold 500 sh * 40 = 20k -> full.
	pf := Portfolio{
		Cash:      80000,
		Positions: map[string]Position{"CHEAP": {Quantity: 500, AvgPrice: 40}},
	}
	intents := Decide(cfg, []StockView{cheap}, pf)
	if findIntent(intents, "CHEAP", Buy) != nil {
		t.Error("must not add to a position already at the max size (HOLD, not BUY)")
	}
}

func TestDecide_CashBufferRespected(t *testing.T) {
	cfg := valueConfig()
	cfg.Risk.CashBufferMin = 1.0 // must keep 100% as cash -> nothing investable
	cheap := StockView{Symbol: "CHEAP", Price: 40, Fundamentals: models.Fundamentals{PE: 8, PB: 1.2, EPSTTM: 5, BVPS: 30}}
	pf := Portfolio{Cash: 100000, Positions: map[string]Position{}}

	if intents := Decide(cfg, []StockView{cheap}, pf); len(intents) != 0 {
		t.Errorf("cash buffer of 100%% should block all buys, got %d intents", len(intents))
	}
}

func momentumConfig() strategy.Config {
	return strategy.Config{
		SchemaVersion: 2,
		Identity:      strategy.Identity{ID: "test_mom", Style: "trend_momentum", Enabled: true},
		Universe:      strategy.Universe{MaxPositions: 5},
		Buy: strategy.Buy{
			Gates: strategy.Gates{
				Technical: &strategy.TechnicalGate{PriceAboveSMADays: ptrInt(50), BreakoutRequired: true},
				Momentum:  &strategy.MomentumGate{Return3MMin: pf64(0.10)},
			},
		},
		Sell: strategy.Sell{StopLossPct: pf64(0.08)},
		Risk: strategy.Risk{MaxPositionSize: 0.3, PositionSizing: "pyramid"},
	}
}

func ptrInt(i int) *int { return &i }

// risingCloses returns n strictly increasing closes (a clean uptrend/breakout).
func risingCloses(n int, start, step float64) []float64 {
	cs := make([]float64, n)
	for i := range n {
		cs[i] = start + float64(i)*step
	}
	return cs
}

func TestDecide_MomentumBuysBreakoutStopsLoser(t *testing.T) {
	cfg := momentumConfig()
	closes := risingCloses(70, 50, 1) // uptrend, last = 119
	breakout := StockView{Symbol: "TREND", Price: closes[len(closes)-1], Closes: closes}

	// Loser held at avg 100, now 85 -> -15% > 8% stop.
	loser := StockView{Symbol: "LOSER", Price: 85, Closes: risingCloses(70, 80, 0.1)}
	pf := Portfolio{Cash: 100000, Positions: map[string]Position{"LOSER": {Quantity: 10, AvgPrice: 100}}}

	intents := Decide(cfg, []StockView{breakout, loser}, pf)
	if findIntent(intents, "TREND", Buy) == nil {
		t.Error("expected a BUY on the breakout")
	}
	if findIntent(intents, "LOSER", Sell) == nil {
		t.Error("expected a stop-loss SELL on the loser")
	}
}

func TestDecide_AllocationRebalancesToTargets(t *testing.T) {
	cfg := strategy.Config{
		SchemaVersion: 2,
		Identity:      strategy.Identity{ID: "test_alloc", Style: "macro_risk_parity", Enabled: true},
		Risk:          strategy.Risk{MaxPositionSize: 1.0},
		Allocation:    strategy.Allocation{"equity": 0.5, "bond": 0.5},
	}
	spy := StockView{Symbol: "SPY", AssetClass: "equity", Price: 100}
	bnd := StockView{Symbol: "BND", AssetClass: "bond", Price: 50}
	// All cash, total 10k -> target 5k equity + 5k bond.
	pf := Portfolio{Cash: 10000, Positions: map[string]Position{}}

	intents := Decide(cfg, []StockView{spy, bnd}, pf)
	buySPY := findIntent(intents, "SPY", Buy)
	buyBND := findIntent(intents, "BND", Buy)
	if buySPY == nil || buyBND == nil {
		t.Fatalf("expected rebalance buys for both asset classes, got %+v", intents)
	}
	if buySPY.Quantity < 49 || buySPY.Quantity > 51 {
		t.Errorf("SPY qty = %v, want ~50 ($5000/$100)", buySPY.Quantity)
	}
	if buyBND.Quantity < 99 || buyBND.Quantity > 101 {
		t.Errorf("BND qty = %v, want ~100 ($5000/$50)", buyBND.Quantity)
	}
}

func TestDecide_QualityAndSafetyGatesReject(t *testing.T) {
	cfg := strategy.Config{
		SchemaVersion: 2,
		Identity:      strategy.Identity{ID: "q", Style: "quality_value"},
		Universe:      strategy.Universe{MaxPositions: 10},
		Buy: strategy.Buy{
			Gates: strategy.Gates{
				Quality:         &strategy.QualityGate{ROICMin: pf64(0.15), GrossMarginMin: pf64(0.4)},
				FinancialSafety: &strategy.FinancialSafetyGate{DebtToEquityMax: pf64(1.0), FCFPositive: true},
			},
			Scoring: map[string]float64{"quality": 1},
		},
		Risk: strategy.Risk{MaxPositionSize: 0.2},
	}
	good := StockView{Symbol: "GOOD", Price: 50, Fundamentals: models.Fundamentals{
		ROIC: 0.25, GrossMargin: 0.6, DebtToEquity: 0.4, FCFPositive: true,
	}}
	lowQuality := StockView{Symbol: "LOWQ", Price: 50, Fundamentals: models.Fundamentals{
		ROIC: 0.05, GrossMargin: 0.6, DebtToEquity: 0.4, FCFPositive: true,
	}}
	indebted := StockView{Symbol: "DEBT", Price: 50, Fundamentals: models.Fundamentals{
		ROIC: 0.25, GrossMargin: 0.6, DebtToEquity: 3.0, FCFPositive: true,
	}}
	noFCF := StockView{Symbol: "NOFCF", Price: 50, Fundamentals: models.Fundamentals{
		ROIC: 0.25, GrossMargin: 0.6, DebtToEquity: 0.4, FCFPositive: false,
	}}
	pf := Portfolio{Cash: 100000, Positions: map[string]Position{}}

	intents := Decide(cfg, []StockView{good, lowQuality, indebted, noFCF}, pf)
	if findIntent(intents, "GOOD", Buy) == nil {
		t.Error("should buy the quality+safe name")
	}
	for _, sym := range []string{"LOWQ", "DEBT", "NOFCF"} {
		if findIntent(intents, sym, Buy) != nil {
			t.Errorf("%s should be rejected by gates", sym)
		}
	}
}

func TestDecide_FundamentalStopLossAndThesisBreak(t *testing.T) {
	cfg := valueConfig()
	cfg.Sell.StopLossPct = pf64(0.15)
	cfg.Sell.ThesisBreak.ROEBelow = pf64(0.10)

	// Held name down 20% from avg -> stop-loss fires.
	dropped := StockView{Symbol: "DROP", Price: 80, Fundamentals: models.Fundamentals{ROE: 0.2, EPSTTM: 5}}
	// Held name whose ROE collapsed -> thesis break.
	broken := StockView{Symbol: "BROKE", Price: 100, Fundamentals: models.Fundamentals{ROE: 0.05, EPSTTM: 5}}
	pf := Portfolio{Cash: 0, Positions: map[string]Position{
		"DROP":  {Quantity: 10, AvgPrice: 100},
		"BROKE": {Quantity: 10, AvgPrice: 100},
	}}

	intents := Decide(cfg, []StockView{dropped, broken}, pf)
	if findIntent(intents, "DROP", Sell) == nil {
		t.Error("expected stop-loss SELL on the dropped name")
	}
	if findIntent(intents, "BROKE", Sell) == nil {
		t.Error("expected thesis-break SELL when ROE collapsed")
	}
}

func TestDecide_AllocationTrimsOverweight(t *testing.T) {
	cfg := strategy.Config{
		SchemaVersion: 2,
		Identity:      strategy.Identity{ID: "alloc", Style: "macro_risk_parity"},
		Risk:          strategy.Risk{MaxPositionSize: 1.0},
		Allocation:    strategy.Allocation{"equity": 0.5, "bond": 0.5},
	}
	spy := StockView{Symbol: "SPY", AssetClass: "equity", Price: 100}
	bnd := StockView{Symbol: "BND", AssetClass: "bond", Price: 50}
	// Heavily overweight equity: 100 sh SPY = $10k, no bonds, no cash.
	pf := Portfolio{Cash: 0, Positions: map[string]Position{"SPY": {Quantity: 100, AvgPrice: 100}}}

	intents := Decide(cfg, []StockView{spy, bnd}, pf)
	if findIntent(intents, "SPY", Sell) == nil {
		t.Error("expected a SELL trimming the overweight equity proxy")
	}
}

func TestDecide_GrahamYAMLEndToEnd(t *testing.T) {
	cfg, err := strategy.Load("../strategies/graham.yml")
	if err != nil {
		t.Fatalf("load graham: %v", err)
	}
	// A deep-value name: low P/E & P/B, strong safety + stability, cheap vs Graham.
	cheap := StockView{
		Symbol: "VALUE", Sector: "Industrials", Price: 30,
		Fundamentals: models.Fundamentals{
			PE: 10, PB: 1.1, EPSTTM: 4, BVPS: 28, MarketCap: 5e9,
			CurrentRatio: 2.5, DebtToEquity: 0.3,
			EPSPositiveYears: 12, DividendYears: 22, EPSGrowth5Y: 0.05,
		},
	}
	pf := Portfolio{Cash: 100000, Positions: map[string]Position{}}

	intents := Decide(cfg, []StockView{cheap}, pf)
	if findIntent(intents, "VALUE", Buy) == nil {
		t.Errorf("Graham strategy should buy a cheap, safe, stable stock; got %+v", intents)
	}
}
