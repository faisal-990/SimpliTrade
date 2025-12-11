package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Pair struct {
	first, second int
}

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

	// Other fields omitted for brevity (you had them correctly)
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

	// Sell rules + portfolio omitted but included in your original code…
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
}

func main() {

	timestart := time.Now()

	// Load stock data
	dir, _ := os.Getwd()
	rawStocks, _ := os.ReadFile(dir + "/internal/seed/mock_stocks_list.json")

	var data []Stock
	json.Unmarshal(rawStocks, &data)

	// Load YAML strategy
	rawRules, _ := os.ReadFile(dir + "/internal/strategies/benjamin.yml")

	var config StrategyConfig
	yaml.Unmarshal(rawRules, &config)

	// Output file
	f, err := os.OpenFile(dir+"/output.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// -------------------------------
	//   CONCURRENCY BEGINS HERE
	// -------------------------------

	n := 100 // desired workers
	total := len(data)

	if n > total {
		n = total
	}

	// Channel carrying chosen stocks
	result := make(chan Stock, n)

	var wg sync.WaitGroup
	wg.Add(n)

	// Compute chunk size (ceiling division)
	chunkSize := (total + n - 1) / n

	// Spawn workers (FAN OUT)
	for i := 0; i < n; i++ {

		start := i * chunkSize
		end := start + chunkSize

		if start >= total {
			wg.Done()
			continue
		}
		if end > total {
			end = total
		}

		go func(s, e int) {
			defer wg.Done()

			for j := s; j < e; j++ {
				if satisfy(&config, &data[j]) {
					result <- data[j]
				}
			}
		}(start, end)
	}

	// CLOUSER goroutine closes result channel when all workers finish.
	go func() {
		wg.Wait()
		close(result)
	}()

	// FAN-IN: collect stocks from all workers + write to file
	for stock := range result {

		b, _ := json.Marshal(stock)
		f.Write(b)
		f.Write([]byte("\n"))
	}

	fmt.Println("Time taken:", time.Since(timestart))
}

func satisfy(cfg *StrategyConfig, s *Stock) bool {
	time.Sleep(time.Millisecond * 1)
	// Valuation rules
	if s.Metrics.PERatio > cfg.Filters.Valuation.PEMax {
		return false
	}
	if s.Metrics.PBRatio > cfg.Filters.Valuation.PBMax {
		return false
	}
	if (s.Metrics.PERatio * s.Metrics.PBRatio) > cfg.Filters.Valuation.PePbMax {
		return false
	}

	// Graham number
	if cfg.Filters.Valuation.UseGrahamNumber {
		graham := math.Sqrt(22.5 * s.Metrics.EPSTTM * s.Metrics.BookValuePerShare)
		if s.Price > graham {
			return false
		}
	}

	// Financial Safety
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

	// Stability
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
