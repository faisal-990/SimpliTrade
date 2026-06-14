package portfolio

import (
	"math"
	"testing"
)

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestValue_GainAndAllocation(t *testing.T) {
	// 10 AAPL @ $100 cost, now $150 -> +$500 (+50%). Cash $9000.
	// Holdings value 1500, total 10500, ROI vs 10000 = +5%.
	s := Value(9000, 10000, []Position{
		{Symbol: "AAPL", Quantity: 10, AvgPrice: 100, CurrentPrice: 150},
	})

	if !approx(s.HoldingsValue, 1500) {
		t.Errorf("holdings value = %v, want 1500", s.HoldingsValue)
	}
	if !approx(s.TotalValue, 10500) {
		t.Errorf("total value = %v, want 10500", s.TotalValue)
	}
	if !approx(s.UnrealizedPL, 500) {
		t.Errorf("unrealized PL = %v, want 500", s.UnrealizedPL)
	}
	if !approx(s.UnrealizedPLPct, 0.5) {
		t.Errorf("unrealized PL%% = %v, want 0.5", s.UnrealizedPLPct)
	}
	if !approx(s.ROI, 0.05) {
		t.Errorf("ROI = %v, want 0.05", s.ROI)
	}
	if len(s.Holdings) != 1 {
		t.Fatalf("expected 1 holding, got %d", len(s.Holdings))
	}
	h := s.Holdings[0]
	if !approx(h.AllocationPct, ratio(1500, 10500)) {
		t.Errorf("allocation = %v, want %v", h.AllocationPct, ratio(1500, 10500))
	}
}

func TestValue_Loss(t *testing.T) {
	// 5 @ $200 cost, now $120 -> cost 1000, market 600, -$400 (-40%).
	s := Value(0, 1000, []Position{
		{Symbol: "X", Quantity: 5, AvgPrice: 200, CurrentPrice: 120},
	})
	if !approx(s.UnrealizedPL, -400) {
		t.Errorf("unrealized PL = %v, want -400", s.UnrealizedPL)
	}
	if !approx(s.UnrealizedPLPct, -0.4) {
		t.Errorf("unrealized PL%% = %v, want -0.4", s.UnrealizedPLPct)
	}
	// Total value 600 vs starting 1000 -> ROI -40%.
	if !approx(s.ROI, -0.4) {
		t.Errorf("ROI = %v, want -0.4", s.ROI)
	}
}

func TestValue_MultipleHoldingsAllocationsSumWithCash(t *testing.T) {
	s := Value(2000, 5000, []Position{
		{Symbol: "A", Quantity: 10, AvgPrice: 100, CurrentPrice: 100}, // 1000
		{Symbol: "B", Quantity: 20, AvgPrice: 50, CurrentPrice: 100},  // 2000
	})
	// total = 2000 cash + 3000 holdings = 5000
	if !approx(s.TotalValue, 5000) {
		t.Fatalf("total = %v, want 5000", s.TotalValue)
	}
	var allocSum float64
	for _, h := range s.Holdings {
		allocSum += h.AllocationPct
	}
	cashAlloc := ratio(2000, 5000)
	if !approx(allocSum+cashAlloc, 1.0) {
		t.Errorf("holding allocations %v + cash %v should total 1.0", allocSum, cashAlloc)
	}
}

func TestValue_EmptyPortfolioIsAllCash(t *testing.T) {
	s := Value(10000, 10000, nil)
	if !approx(s.TotalValue, 10000) || !approx(s.HoldingsValue, 0) {
		t.Errorf("all-cash: total=%v holdings=%v", s.TotalValue, s.HoldingsValue)
	}
	if !approx(s.UnrealizedPL, 0) || !approx(s.UnrealizedPLPct, 0) || !approx(s.ROI, 0) {
		t.Errorf("all-cash should have zero PL and ROI: %+v", s)
	}
}

func TestValue_ZeroStartingCapitalNoNaN(t *testing.T) {
	s := Value(0, 0, []Position{{Symbol: "A", Quantity: 1, AvgPrice: 10, CurrentPrice: 10}})
	if math.IsNaN(s.ROI) || math.IsInf(s.ROI, 0) {
		t.Errorf("ROI must be finite with zero starting capital, got %v", s.ROI)
	}
}
