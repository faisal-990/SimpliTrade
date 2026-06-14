package strategy

import "fmt"

// CustomParams is the curated, user-facing slice of a strategy that the
// "create your own investor" builder collects. BuildCustom expands it into a
// full, valid Config — filling sensible defaults for everything not exposed — so
// a hand-built investor runs through the exact same engine as the presets.
type CustomParams struct {
	ID         string
	Name       string
	Philosophy string
	Approach   string // value | quality | growth | momentum

	MaxPositions int

	// Approach-specific thresholds (0 = "don't gate on this").
	PEMax              float64 // value
	PBMax              float64 // value
	ROEMin             float64 // quality (fraction, e.g. 0.15)
	OperatingMarginMin float64 // quality (fraction)
	RevenueGrowthMin   float64 // growth (fraction)
	EPSGrowthMin       float64 // growth (fraction)
	Return6MMin        float64 // momentum (fraction)

	// Sell + risk.
	StopLossPct           float64 // fraction, e.g. 0.15
	TakeProfitVsIntrinsic float64 // multiple of intrinsic value, e.g. 1.0
	MaxPositionSize       float64 // fraction (0,1]
	CashBufferMin         float64 // fraction
	PositionSizing        string  // equal | conviction
}

// approachStyle maps a friendly approach to an engine style (which selects the
// decision paradigm via KindFor).
func approachStyle(approach string) (string, error) {
	switch approach {
	case "value":
		return "deep_value", nil
	case "quality":
		return "quality_value", nil
	case "growth":
		return "growth_compounder", nil
	case "momentum":
		return "trend_momentum", nil
	default:
		return "", fmt.Errorf("unknown approach %q (want value|quality|growth|momentum)", approach)
	}
}

func optF(v float64) *float64 {
	if v <= 0 {
		return nil
	}
	return &v
}

// BuildCustom expands curated params into a validated Config.
func BuildCustom(p CustomParams) (Config, error) {
	style, err := approachStyle(p.Approach)
	if err != nil {
		return Config{}, err
	}

	maxPos := p.MaxPositions
	if maxPos <= 0 || maxPos > 50 {
		maxPos = 20
	}
	maxPosSize := p.MaxPositionSize
	if maxPosSize <= 0 || maxPosSize > 1 {
		maxPosSize = 0.1
	}
	sizing := p.PositionSizing
	if sizing != "equal" && sizing != "conviction" {
		sizing = "equal"
	}

	var gates Gates
	switch p.Approach {
	case "value":
		gates.Valuation = &ValuationGate{PEMax: optF(p.PEMax), PBMax: optF(p.PBMax)}
		gates.FinancialSafety = &FinancialSafetyGate{FCFPositive: true}
	case "quality":
		gates.Quality = &QualityGate{ROEMin: optF(p.ROEMin), OperatingMarginMin: optF(p.OperatingMarginMin)}
	case "growth":
		gates.Growth = &GrowthGate{RevenueGrowthYoYMin: optF(p.RevenueGrowthMin), EPSGrowthYoYMin: optF(p.EPSGrowthMin)}
	case "momentum":
		gates.Momentum = &MomentumGate{Return6MMin: optF(p.Return6MMin)}
	}

	cfg := Config{
		SchemaVersion: 2,
		Identity: Identity{
			ID:         p.ID,
			Name:       p.Name,
			Style:      style,
			Philosophy: p.Philosophy,
			Enabled:    true,
		},
		Universe: Universe{
			AssetClasses: []string{"equity"},
			Geographies:  []string{"US"},
			MinPositions: 3,
			MaxPositions: maxPos,
		},
		Buy: Buy{Gates: gates},
		Sell: Sell{
			StopLossPct:           optF(p.StopLossPct),
			TakeProfitVsIntrinsic: optF(p.TakeProfitVsIntrinsic),
			MaxPositionSize:       optF(maxPosSize),
		},
		Risk: Risk{
			PositionSizing:  sizing,
			MaxPositionSize: maxPosSize,
			CashBufferMin:   p.CashBufferMin,
			Rebalance:       "dynamic",
			LeverageMax:     1.0,
		},
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
