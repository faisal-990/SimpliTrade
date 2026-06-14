package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// TwelveDataProvider implements Provider against the Twelve Data REST API. It is
// constructed only when a key is configured; otherwise the FakeProvider is used.
// Fundamental coverage depends on plan — fields the API omits are left zero, and
// a zero metric simply means the corresponding strategy gate is skipped.
type TwelveDataProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewTwelveDataProvider builds the provider. baseURL defaults to the public API
// when empty (tests pass an httptest URL).
func NewTwelveDataProvider(apiKey, baseURL string, client *http.Client) *TwelveDataProvider {
	if baseURL == "" {
		baseURL = "https://api.twelvedata.com"
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &TwelveDataProvider{apiKey: apiKey, baseURL: baseURL, client: client}
}

func (p *TwelveDataProvider) get(ctx context.Context, path string, params url.Values, out any) error {
	params.Set("apikey", p.apiKey)
	u := p.baseURL + path + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("twelvedata: %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		return ErrRateLimited
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("twelvedata: %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// --- quote ---

type tdQuote struct {
	Symbol    string `json:"symbol"`
	Close     string `json:"close"`
	Open      string `json:"open"`
	High      string `json:"high"`
	Low       string `json:"low"`
	Volume    string `json:"volume"`
	Timestamp int64  `json:"timestamp"`
}

func (q tdQuote) toQuote() Quote {
	return Quote{
		Symbol: q.Symbol,
		Price:  parseFloat(q.Close),
		Open:   parseFloat(q.Open),
		High:   parseFloat(q.High),
		Low:    parseFloat(q.Low),
		Volume: int64(parseFloat(q.Volume)),
		Time:   time.Unix(q.Timestamp, 0),
	}
}

func (p *TwelveDataProvider) Quote(ctx context.Context, symbol string) (Quote, error) {
	var q tdQuote
	if err := p.get(ctx, "/quote", url.Values{"symbol": {symbol}}, &q); err != nil {
		return Quote{}, err
	}
	if q.Symbol == "" {
		q.Symbol = symbol
	}
	return q.toQuote(), nil
}

// BatchQuotes uses Twelve Data's comma-separated symbol support: one request for
// many symbols. The API returns a single object for one symbol and a map keyed by
// symbol for many, so we handle both shapes.
func (p *TwelveDataProvider) BatchQuotes(ctx context.Context, symbols []string) (map[string]Quote, error) {
	if len(symbols) == 0 {
		return map[string]Quote{}, nil
	}
	if len(symbols) == 1 {
		q, err := p.Quote(ctx, symbols[0])
		if err != nil {
			return nil, err
		}
		return map[string]Quote{symbols[0]: q}, nil
	}

	params := url.Values{"symbol": {strings.Join(symbols, ",")}}
	var raw map[string]tdQuote
	if err := p.get(ctx, "/quote", params, &raw); err != nil {
		return nil, err
	}
	out := make(map[string]Quote, len(raw))
	for sym, q := range raw {
		if q.Symbol == "" {
			q.Symbol = sym
		}
		out[sym] = q.toQuote()
	}
	return out, nil
}

// --- candles ---

type tdTimeSeries struct {
	Values []struct {
		Datetime string `json:"datetime"`
		Open     string `json:"open"`
		High     string `json:"high"`
		Low      string `json:"low"`
		Close    string `json:"close"`
		Volume   string `json:"volume"`
	} `json:"values"`
}

func (p *TwelveDataProvider) Candles(ctx context.Context, symbol, interval string, from, to time.Time) ([]Candle, error) {
	tdInterval := "1day"
	if interval != "1d" && interval != "" {
		tdInterval = interval
	}
	params := url.Values{
		"symbol":     {symbol},
		"interval":   {tdInterval},
		"start_date": {from.Format("2006-01-02")},
		"end_date":   {to.Format("2006-01-02")},
		"order":      {"ASC"},
	}
	var ts tdTimeSeries
	if err := p.get(ctx, "/time_series", params, &ts); err != nil {
		return nil, err
	}
	candles := make([]Candle, 0, len(ts.Values))
	for _, v := range ts.Values {
		t, _ := time.Parse("2006-01-02", v.Datetime)
		candles = append(candles, Candle{
			Time: t, Open: parseFloat(v.Open), High: parseFloat(v.High),
			Low: parseFloat(v.Low), Close: parseFloat(v.Close),
			Volume: int64(parseFloat(v.Volume)), Interval: "1d",
		})
	}
	return candles, nil
}

// --- fundamentals ---

type tdStatistics struct {
	Statistics struct {
		Valuations struct {
			TrailingPE   string `json:"trailing_pe"`
			ForwardPE    string `json:"forward_pe"`
			PriceToBook  string `json:"price_to_book_mrq"`
			PriceToSales string `json:"price_to_sales_ttm"`
		} `json:"valuations_metrics"`
		Financials struct {
			GrossMargin     string `json:"gross_margin"`
			OperatingMargin string `json:"operating_margin"`
			ProfitMargin    string `json:"profit_margin"`
			ReturnOnEquity  string `json:"return_on_equity_ttm"`
		} `json:"financials"`
		StockStats struct {
			Beta string `json:"beta"`
		} `json:"stock_statistics"`
	} `json:"statistics"`
}

// Fundamentals fetches what the plan exposes (statistics endpoint). Missing
// fields stay zero; the strategy engine treats a zero metric as "gate skipped".
func (p *TwelveDataProvider) Fundamentals(ctx context.Context, symbol string) (models.Fundamentals, error) {
	var s tdStatistics
	if err := p.get(ctx, "/statistics", url.Values{"symbol": {symbol}}, &s); err != nil {
		return models.Fundamentals{}, err
	}
	v, f := s.Statistics.Valuations, s.Statistics.Financials
	return models.Fundamentals{
		PE:              parseFloat(v.TrailingPE),
		ForwardPE:       parseFloat(v.ForwardPE),
		PB:              parseFloat(v.PriceToBook),
		PS:              parseFloat(v.PriceToSales),
		GrossMargin:     parseFloat(f.GrossMargin),
		OperatingMargin: parseFloat(f.OperatingMargin),
		NetMargin:       parseFloat(f.ProfitMargin),
		ROE:             parseFloat(f.ReturnOnEquity),
		Beta:            parseFloat(s.Statistics.StockStats.Beta),
	}, nil
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}
