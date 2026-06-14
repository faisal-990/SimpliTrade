package strategy

import (
	"testing"
)

const strategiesDir = "../strategies"

func TestLoadDir_AllTwentyInvestorsParse(t *testing.T) {
	configs, err := LoadDir(strategiesDir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(configs) != 20 {
		t.Fatalf("loaded %d v2 strategies, want 20 (legacy v1 files should be skipped)", len(configs))
	}

	kinds := map[Kind]int{}
	seen := map[string]bool{}
	for _, c := range configs {
		if c.Identity.ID == "" || c.Identity.Style == "" {
			t.Errorf("strategy missing id/style: %+v", c.Identity)
		}
		if seen[c.Identity.ID] {
			t.Errorf("duplicate strategy id %q", c.Identity.ID)
		}
		seen[c.Identity.ID] = true
		if c.Profile.RiskRating < 1 || c.Profile.RiskRating > 10 {
			t.Errorf("%s: risk_rating %d out of 1..10", c.Identity.ID, c.Profile.RiskRating)
		}
		if c.Risk.MaxPositionSize <= 0 || c.Risk.MaxPositionSize > 1 {
			t.Errorf("%s: bad max_position_size %v", c.Identity.ID, c.Risk.MaxPositionSize)
		}
		kinds[c.Kind()]++
	}

	if kinds[KindAllocation] != 1 {
		t.Errorf("allocation strategies = %d, want 1 (Dalio)", kinds[KindAllocation])
	}
	if kinds[KindMomentum] != 2 {
		t.Errorf("momentum strategies = %d, want 2 (Livermore, Druckenmiller)", kinds[KindMomentum])
	}
	if kinds[KindFundamental] != 17 {
		t.Errorf("fundamental strategies = %d, want 17", kinds[KindFundamental])
	}
}

func TestLoad_GrahamGatesPresent(t *testing.T) {
	c, err := Load(strategiesDir + "/graham.yml")
	if err != nil {
		t.Fatalf("Load graham: %v", err)
	}
	if c.Buy.Gates.Valuation == nil || c.Buy.Gates.Valuation.PEMax == nil {
		t.Fatal("graham should have a valuation P/E gate")
	}
	if *c.Buy.Gates.Valuation.PEMax != 15 {
		t.Errorf("graham pe_max = %v, want 15", *c.Buy.Gates.Valuation.PEMax)
	}
	if !c.Buy.Gates.Valuation.UseGrahamNumber {
		t.Error("graham should use the Graham number")
	}
	if c.Buy.MarginOfSafetyMin == nil || *c.Buy.MarginOfSafetyMin != 0.33 {
		t.Errorf("graham margin_of_safety_min = %v, want 0.33", c.Buy.MarginOfSafetyMin)
	}
}

func TestLoad_WoodHasNoValuationGate(t *testing.T) {
	// A disruptive-growth investor screens on growth, not valuation — the
	// present-only gate design must leave the valuation block nil.
	c, err := Load(strategiesDir + "/wood.yml")
	if err != nil {
		t.Fatalf("Load wood: %v", err)
	}
	if c.Buy.Gates.Valuation != nil {
		t.Errorf("wood should have NO valuation gate, got %+v", c.Buy.Gates.Valuation)
	}
	if c.Buy.Gates.Growth == nil || c.Buy.Gates.Growth.RevenueGrowthYoYMin == nil {
		t.Fatal("wood should have a revenue-growth gate")
	}
}

func TestLoad_DalioAllocation(t *testing.T) {
	c, err := Load(strategiesDir + "/dalio.yml")
	if err != nil {
		t.Fatalf("Load dalio: %v", err)
	}
	if c.Kind() != KindAllocation {
		t.Fatalf("dalio kind = %s, want allocation", c.Kind())
	}
	if len(c.Allocation) == 0 {
		t.Fatal("dalio should declare asset-class allocation targets")
	}
	var sum float64
	for _, w := range c.Allocation {
		sum += w
	}
	if sum < 0.99 || sum > 1.01 {
		t.Errorf("dalio allocation weights sum to %v, want ~1.0", sum)
	}
}

func TestLoad_MomentumStops(t *testing.T) {
	c, err := Load(strategiesDir + "/livermore.yml")
	if err != nil {
		t.Fatalf("Load livermore: %v", err)
	}
	if c.Kind() != KindMomentum {
		t.Fatalf("livermore kind = %s, want momentum", c.Kind())
	}
	if c.Sell.StopLossPct == nil {
		t.Error("a trend trader must have a hard stop-loss")
	}
	if c.Buy.Gates.Technical == nil {
		t.Error("livermore should screen on technicals")
	}
}
