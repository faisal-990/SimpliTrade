package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// FundamentalsSource supplies company fundamentals only — the slice of the
// market-data surface that Twelve Data's free tier doesn't provide. A composite
// provider pairs it with a price/candle source (see CompositeProvider).
type FundamentalsSource interface {
	Fundamentals(ctx context.Context, symbol string) (models.Fundamentals, error)
}

// FinnhubProvider fetches fundamentals from Finnhub's /stock/metric endpoint.
// Finnhub's free tier (~60 req/min) covers valuation, quality, and growth
// metrics — enough to drive the value strategies' screens and thesis-break exits
// that the price-only feed leaves dark.
//
// Unit conventions (Finnhub → our model): margins, ROE/ROIC, growth and yields
// arrive as PERCENT numbers (e.g. 43.3 = 43.3%) and are divided by 100 to match
// the strategy thresholds, which are fractions (e.g. roe_below: 0.10). Market cap
// arrives in MILLIONS and is scaled to absolute dollars.
type FinnhubProvider struct {
	apiKey string
	base   string
	http   *http.Client
}

func NewFinnhubProvider(apiKey string) *FinnhubProvider {
	return &FinnhubProvider{
		apiKey: apiKey,
		base:   "https://finnhub.io/api/v1",
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

// finnhubMetric mirrors the fields we use from the "metric" object. Pointers
// distinguish "absent" from a real zero so we never fabricate a 0% margin.
type finnhubMetric struct {
	PETTM            *float64 `json:"peTTM"`
	PENormAnnual     *float64 `json:"peNormalizedAnnual"`
	PBQuarterly      *float64 `json:"pbQuarterly"`
	PBAnnual         *float64 `json:"pbAnnual"`
	PSTTM            *float64 `json:"psTTM"`
	MarketCap        *float64 `json:"marketCapitalization"` // in millions
	ROETTM           *float64 `json:"roeTTM"`               // percent
	ROITTM           *float64 `json:"roiTTM"`               // percent (used as ROIC proxy)
	GrossMarginTTM   *float64 `json:"grossMarginTTM"`       // percent
	OperatingMargin  *float64 `json:"operatingMarginTTM"`   // percent
	NetMarginTTM     *float64 `json:"netProfitMarginTTM"`   // percent
	DebtToEquityQ    *float64 `json:"totalDebt/totalEquityQuarterly"`
	CurrentRatioQ    *float64 `json:"currentRatioQuarterly"`
	InterestCover    *float64 `json:"netInterestCoverageTTM"`
	RevenueGrowthYoY *float64 `json:"revenueGrowthTTMYoy"` // percent
	EPSGrowthYoY     *float64 `json:"epsGrowthTTMYoy"`     // percent
	RevenueGrowth5Y  *float64 `json:"revenueGrowth5Y"`     // percent
	EPSGrowth5Y      *float64 `json:"epsGrowth5Y"`         // percent
	EPSTTM           *float64 `json:"epsTTM"`
	EPSNormAnnual    *float64 `json:"epsNormalizedAnnual"`
	BVPSQuarterly    *float64 `json:"bookValuePerShareQuarterly"`
	Beta             *float64 `json:"beta"`
	DividendYield    *float64 `json:"currentDividendYieldTTM"` // percent
	FCFPerShareTTM   *float64 `json:"freeCashFlowPerShareTTM"`
}

func (p *FinnhubProvider) Fundamentals(ctx context.Context, symbol string) (models.Fundamentals, error) {
	q := url.Values{"symbol": {symbol}, "metric": {"all"}, "token": {p.apiKey}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.base+"/stock/metric?"+q.Encode(), nil)
	if err != nil {
		return models.Fundamentals{}, err
	}
	resp, err := p.http.Do(req)
	if err != nil {
		return models.Fundamentals{}, fmt.Errorf("finnhub: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return models.Fundamentals{}, ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return models.Fundamentals{}, fmt.Errorf("finnhub: status %d", resp.StatusCode)
	}

	var body struct {
		Metric finnhubMetric `json:"metric"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return models.Fundamentals{}, fmt.Errorf("finnhub: decode: %w", err)
	}
	return mapFinnhub(body.Metric), nil
}

// mapFinnhub converts Finnhub's percent/millions units into the model's
// fraction/absolute conventions. A missing metric stays zero, which the engine
// treats as "gate skipped" rather than a hard fail.
func mapFinnhub(m finnhubMetric) models.Fundamentals {
	pct := func(p *float64) float64 { // percent → fraction
		if p == nil {
			return 0
		}
		return *p / 100
	}
	f := models.Fundamentals{
		PE:               firstNonNil(m.PETTM, m.PENormAnnual),
		PB:               firstNonNil(m.PBQuarterly, m.PBAnnual),
		PS:               deref(m.PSTTM),
		MarketCap:        deref(m.MarketCap) * 1_000_000, // millions → dollars
		ROE:              pct(m.ROETTM),
		ROIC:             pct(m.ROITTM),
		GrossMargin:      pct(m.GrossMarginTTM),
		OperatingMargin:  pct(m.OperatingMargin),
		NetMargin:        pct(m.NetMarginTTM),
		DebtToEquity:     deref(m.DebtToEquityQ),
		CurrentRatio:     deref(m.CurrentRatioQ),
		InterestCover:    deref(m.InterestCover),
		RevenueGrowthYoY: pct(m.RevenueGrowthYoY),
		EPSGrowthYoY:     pct(m.EPSGrowthYoY),
		RevenueCAGR3Y:    pct(m.RevenueGrowth5Y), // 5Y stands in for the 3Y CAGR gate
		EPSGrowth5Y:      pct(m.EPSGrowth5Y),
		EPSTTM:           firstNonNil(m.EPSTTM, m.EPSNormAnnual),
		BVPS:             deref(m.BVPSQuarterly),
		Beta:             deref(m.Beta),
		DividendYield:    pct(m.DividendYield),
		FCFPositive:      deref(m.FCFPerShareTTM) > 0,
	}
	return f
}

func deref(p *float64) float64 {
	if p == nil {
		return 0
	}
	return *p
}

func firstNonNil(ps ...*float64) float64 {
	for _, p := range ps {
		if p != nil {
			return *p
		}
	}
	return 0
}
