// Package strategy parses the v2 investor strategy YAMLs into typed Config
// values. Optional gate thresholds are pointers: a nil threshold means "this
// strategy does not screen on this metric" — which is how one engine serves
// both a deep-value investor (P/E gate present) and a momentum trader (no
// valuation gates at all).
package strategy

// Config is one investor strategy parsed from a YAML file.
type Config struct {
	SchemaVersion int        `yaml:"schema_version"`
	Identity      Identity   `yaml:"identity"`
	Profile       Profile    `yaml:"profile"`
	Universe      Universe   `yaml:"universe"`
	Buy           Buy        `yaml:"buy"`
	Sell          Sell       `yaml:"sell"`
	Risk          Risk       `yaml:"risk"`
	Allocation    Allocation `yaml:"allocation"` // only for macro_* styles
}

type Identity struct {
	ID         string `yaml:"id"`
	Name       string `yaml:"name"`
	Firm       string `yaml:"firm"`
	Era        string `yaml:"era"`
	Style      string `yaml:"style"`
	Philosophy string `yaml:"philosophy"`
	Enabled    bool   `yaml:"enabled"`
}

type Profile struct {
	RiskRating         int    `yaml:"risk_rating"`
	Volatility         string `yaml:"volatility"`
	TimeHorizon        string `yaml:"time_horizon"`
	TypicalHoldingDays int    `yaml:"typical_holding_days"`
	Tagline            string `yaml:"tagline"`
}

type Universe struct {
	AssetClasses   []string `yaml:"asset_classes"`
	MarketCapMin   *float64 `yaml:"market_cap_min"`
	MarketCapMax   *float64 `yaml:"market_cap_max"`
	SectorsInclude []string `yaml:"sectors_include"`
	SectorsExclude []string `yaml:"sectors_exclude"`
	Geographies    []string `yaml:"geographies"`
	MinPositions   int      `yaml:"min_positions"`
	MaxPositions   int      `yaml:"max_positions"`
}

type Buy struct {
	Gates             Gates              `yaml:"gates"`
	MarginOfSafetyMin *float64           `yaml:"margin_of_safety_min"`
	Scoring           map[string]float64 `yaml:"scoring"`
}

// Gates are the hard pass/fail filters. Every block and field is optional; only
// the present ones are applied.
type Gates struct {
	Valuation       *ValuationGate       `yaml:"valuation"`
	FinancialSafety *FinancialSafetyGate `yaml:"financial_safety"`
	Quality         *QualityGate         `yaml:"quality"`
	Growth          *GrowthGate          `yaml:"growth"`
	Stability       *StabilityGate       `yaml:"stability"`
	Technical       *TechnicalGate       `yaml:"technical"`
	Momentum        *MomentumGate        `yaml:"momentum"`
}

type ValuationGate struct {
	PEMax           *float64 `yaml:"pe_max"`
	ForwardPEMax    *float64 `yaml:"forward_pe_max"`
	PBMax           *float64 `yaml:"pb_max"`
	PEPBMax         *float64 `yaml:"pe_pb_max"`
	PSMax           *float64 `yaml:"ps_max"`
	PEGMax          *float64 `yaml:"peg_max"`
	EVEBITDAMax     *float64 `yaml:"ev_ebitda_max"`
	EarningsYldMin  *float64 `yaml:"earnings_yield_min"`
	FCFYieldMin     *float64 `yaml:"fcf_yield_min"`
	DividendYldMin  *float64 `yaml:"dividend_yield_min"`
	UseGrahamNumber bool     `yaml:"use_graham_number"`
}

type FinancialSafetyGate struct {
	CurrentRatioMin     *float64 `yaml:"current_ratio_min"`
	DebtToEquityMax     *float64 `yaml:"debt_to_equity_max"`
	InterestCoverageMin *float64 `yaml:"interest_coverage_min"`
	NetCurrentAssetRule bool     `yaml:"net_current_asset_rule"`
	FCFPositive         bool     `yaml:"fcf_positive"`
}

type QualityGate struct {
	ROEMin                  *float64 `yaml:"roe_min"`
	ROICMin                 *float64 `yaml:"roic_min"`
	GrossMarginMin          *float64 `yaml:"gross_margin_min"`
	OperatingMarginMin      *float64 `yaml:"operating_margin_min"`
	NetMarginMin            *float64 `yaml:"net_margin_min"`
	RequireConsistentMargin bool     `yaml:"require_consistent_margins"`
}

type GrowthGate struct {
	RevenueGrowthYoYMin *float64 `yaml:"revenue_growth_yoy_min"`
	EPSGrowthYoYMin     *float64 `yaml:"eps_growth_yoy_min"`
	RevenueCAGR3YMin    *float64 `yaml:"revenue_cagr_3y_min"`
	EPSGrowth5YMin      *float64 `yaml:"eps_growth_5y_min"`
}

type StabilityGate struct {
	EPSPositiveYearsMin *int     `yaml:"eps_positive_years_min"`
	DividendYearsMin    *int     `yaml:"dividend_years_min"`
	BetaMax             *float64 `yaml:"beta_max"`
}

type TechnicalGate struct {
	PriceAboveSMADays *int     `yaml:"price_above_sma_days"`
	RSIMin            *float64 `yaml:"rsi_min"`
	RSIMax            *float64 `yaml:"rsi_max"`
	Near52WHighPct    *float64 `yaml:"near_52w_high_pct"`
	BreakoutRequired  bool     `yaml:"breakout_required"`
}

type MomentumGate struct {
	Return3MMin         *float64 `yaml:"return_3m_min"`
	Return6MMin         *float64 `yaml:"return_6m_min"`
	Return12MMin        *float64 `yaml:"return_12m_min"`
	RelativeStrengthMin *float64 `yaml:"relative_strength_min"`
}

type Sell struct {
	TakeProfitVsIntrinsic *float64    `yaml:"take_profit_vs_intrinsic"`
	StopLossPct           *float64    `yaml:"stop_loss_pct"`
	TrailingStopPct       *float64    `yaml:"trailing_stop_pct"`
	ThesisBreak           ThesisBreak `yaml:"thesis_break"`
	MaxPositionSize       *float64    `yaml:"max_position_size"`
}

type ThesisBreak struct {
	ROEBelow             *float64 `yaml:"roe_below"`
	ROICBelow            *float64 `yaml:"roic_below"`
	DebtToEquityAbove    *float64 `yaml:"debt_to_equity_above"`
	EPSDowntrendYears    *int     `yaml:"eps_downtrend_years"`
	RevenueGrowthBelow   *float64 `yaml:"revenue_growth_below"`
	OperatingMarginBelow *float64 `yaml:"operating_margin_below"`
	InterestCoverBelow   *float64 `yaml:"interest_coverage_below"`
	DividendCut          bool     `yaml:"dividend_cut"`
	FCFPositive          bool     `yaml:"fcf_positive"` // sell if FCF turns non-positive
	RSIBelow             *float64 `yaml:"rsi_below"`
}

type Risk struct {
	PositionSizing  string  `yaml:"position_sizing"` // equal|conviction|volatility_parity|pyramid
	MaxPositionSize float64 `yaml:"max_position_size"`
	CashBufferMin   float64 `yaml:"cash_buffer_min"`
	Rebalance       string  `yaml:"rebalance"`
	LeverageMax     float64 `yaml:"leverage_max"`
}

// Allocation holds asset-class target weights for macro_* styles (nil/empty for
// stock-picking strategies).
type Allocation map[string]float64
