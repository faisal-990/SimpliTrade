package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"math"
	"os"
)

type Stock struct {
	Ticker      string  `json:"ticker"`
	CompanyName string  `json:"company_name"`
	Sector      string  `json:"sector"`
	Price       float64 `json:"price"`

	Metrics struct {
		PERatio           float64 `json:"pe_ratio"`
		PBRatio           float64 `json:"pb_ratio"`
		CurrentRatio      float64 `json:"current_ratio"`
		DebtToEquity      float64 `json:"debt_to_equity"`
		NetCurrentAssets  float64 `json:"net_current_assets"`
		LongTermDebt      float64 `json:"long_term_debt"`
		EPSTTM            float64 `json:"eps_ttm"`
		BookValuePerShare float64 `json:"book_value_per_share"`
	} `json:"metrics"`

	History struct {
		EPSGrowth10Y            float64 `json:"eps_growth_10y"`
		YearsPositiveEPS        int     `json:"years_positive_eps"`
		YearsContinuousDividend int     `json:"years_continuous_dividend"`
		DividendYield           float64 `json:"dividend_yield"`
	} `json:"history"`
}

type StrategyConfig struct {
	StrategyName string `yaml:"strategy_name"`
	StrategyID   string `yaml:"strategy_id"`
	Enabled      bool   `yaml:"enabled"`

	Filters struct {
		Valuation struct {
			PEMax           float64 `yaml:"pe_max"`
			PBMax           float64 `yaml:"pb_max"`
			PePbMax         float64 `yaml:"pe_pb_max"`
			MOSMin          float64 `yaml:"margin_of_safety_min"`
			UseGrahamNumber bool    `yaml:"use_graham_number"`
		} `yaml:"valuation"`

		FinancialSafety struct {
			CurrentRatioMin     float64 `yaml:"current_ratio_min"`
			DebtToEquityMax     float64 `yaml:"debt_to_equity_max"`
			NetCurrentAssetRule bool    `yaml:"net_current_asset_rule"`
		} `yaml:"financial_safety"`

		Stability struct {
			EPSPositiveYearsMin int     `yaml:"eps_positive_years_min"`
			DividendYearsMin    int     `yaml:"dividend_years_min"`
			EPSGrowth10YrMin    float64 `yaml:"eps_growth_10yr_min"`
		} `yaml:"stability"`
	} `yaml:"filters"`

	Scoring struct {
		IntrinsicValue struct {
			BasePE           float64 `yaml:"base_pe"`
			GrowthMultiplier float64 `yaml:"growth_multiplier"`
			MaxGrowthRate    float64 `yaml:"max_growth_rate"`
			Weight           float64 `yaml:"weight"`
		} `yaml:"intrinsic_value"`

		Stability struct {
			EPSVarianceWindow int     `yaml:"eps_variance_window"`
			Weight            float64 `yaml:"weight"`
		} `yaml:"stability"`

		Valuation struct {
			Weight float64 `yaml:"weight"`
		} `yaml:"valuation"`

		EarningsYield struct {
			RequiredMultipleOfBondYield float64 `yaml:"required_multiple_of_bond_yield"`
			Weight                      float64 `yaml:"weight"`
		} `yaml:"earnings_yield"`
	} `yaml:"scoring"`

	SellRules struct {
		MOSLostThreshold    float64 `yaml:"mos_lost_threshold"`
		OvervaluationBuffer float64 `yaml:"overvaluation_buffer"`

		ViolationFilters struct {
			PeLimit           float64 `yaml:"pe_limit"`
			PbLimit           float64 `yaml:"pb_limit"`
			PePbLimit         float64 `yaml:"pe_pb_limit"`
			DebtToEquityLimit float64 `yaml:"debt_to_equity_limit"`
			CurrentRatioLimit float64 `yaml:"current_ratio_limit"`
			EPSNegative       bool    `yaml:"eps_negative"`
			DividendCut       bool    `yaml:"dividend_cut"`
		} `yaml:"violation_filters"`

		EPSTrend struct {
			CheckTrend     bool `yaml:"check_trend"`
			DowntrendYears int  `yaml:"downtrend_years"`
		} `yaml:"eps_trend"`

		DebtIncreaseCheck bool    `yaml:"debt_increase_check"`
		MaxPositionSize   float64 `yaml:"max_position_size"`
	} `yaml:"sell_rules"`

	Portfolio struct {
		EquityTarget   float64 `yaml:"equity_target"`
		BondTarget     float64 `yaml:"bond_target"`
		RebalanceUpper float64 `yaml:"rebalance_upper"`
		RebalanceLower float64 `yaml:"rebalance_lower"`
		MinPositions   int     `yaml:"min_positions"`
		MaxPositions   int     `yaml:"max_positions"`
	} `yaml:"portfolio"`

	ModernAdjustments struct {
		Enable bool `yaml:"enable"`

		SectorExceptions map[string]struct {
			CurrentRatioMin float64 `yaml:"current_ratio_min,omitempty"`
			DebtToEquityMax float64 `yaml:"debt_to_equity_max,omitempty"`
		} `yaml:"sector_exceptions"`

		QualityChecks struct {
			ROEMin float64 `yaml:"roe_min"`
		} `yaml:"quality_checks"`
	} `yaml:"modern_adjustments"`
}

func main() {
	//open the mock stocks file
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	stocks, err := os.ReadFile(dir + "/internal/seed/mock_stocks_list.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var data []Stock
	err = json.Unmarshal(stocks, &data)
	if err != nil {
		fmt.Println(err)
		return
	}
	rules, err := os.ReadFile(dir + "/internal/strategies/benjamin.yml")
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(string(rules))
	var config StrategyConfig

	err = yaml.Unmarshal(rules, &config)
	if err != nil {
		fmt.Println(err)
		return
	}

	//opening the file to write the output to
	f, err := os.OpenFile(dir+"/output.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	//apply the grahams rule to pick stocks
	for i := 0; i < len(data); i++ {
		if satisfy(&config, &data[i]) == true {
			_, err = f.WriteString(data[i].CompanyName + ",")
			if err != nil {
				fmt.Println(err)

			}
		}
	}
	//create a new stockuniverse file to put stocks that satisfies grahams philosophy
}
func satisfy(cfg *StrategyConfig, s *Stock) bool {

	// --- Valuation ---
	if s.Metrics.PERatio > cfg.Filters.Valuation.PEMax {
		return false
	}
	if s.Metrics.PBRatio > cfg.Filters.Valuation.PBMax {
		return false
	}
	if (s.Metrics.PERatio * s.Metrics.PBRatio) > cfg.Filters.Valuation.PePbMax {
		return false
	}

	// Graham Number check
	if cfg.Filters.Valuation.UseGrahamNumber {
		grahamNumber := math.Sqrt(22.5 * s.Metrics.EPSTTM * s.Metrics.BookValuePerShare)
		if s.Price > grahamNumber {
			return false
		}
	}

	// // Margin of Safety check
	// intrinsicValue := intrinsicValueCalc(cfg, s.Metrics.EPSTTM) // define helper
	// if s.Price > intrinsicValue*(1.0-cfg.Filters.Valuation.MOSMin) {
	// 	return false
	// }
	//
	// // --- Financial Safety ---
	if s.Metrics.CurrentRatio < cfg.Filters.FinancialSafety.CurrentRatioMin {
		return false
	}
	if s.Metrics.DebtToEquity > cfg.Filters.FinancialSafety.DebtToEquityMax {
		return false
	}
	if cfg.Filters.FinancialSafety.NetCurrentAssetRule {
		if s.Metrics.LongTermDebt > s.Metrics.NetCurrentAssets {
			return false
		}
	}

	// --- Stability ---
	if s.History.YearsPositiveEPS < cfg.Filters.Stability.EPSPositiveYearsMin {
		return false
	}
	if s.History.YearsContinuousDividend < cfg.Filters.Stability.DividendYearsMin {
		return false
	}
	if s.History.EPSGrowth10Y < cfg.Filters.Stability.EPSGrowth10YrMin {
		return false
	}

	return true
}
func intrinsicValueCalc(cfg *StrategyConfig, eps float64) float64 {
	g := cfg.Scoring.IntrinsicValue.MaxGrowthRate
	return eps * (cfg.Scoring.IntrinsicValue.BasePE + cfg.Scoring.IntrinsicValue.GrowthMultiplier*g)
}
