package marketdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// tdServer returns an httptest server emulating the Twelve Data endpoints we use,
// so the adapter's request-building and response-parsing are tested without a key
// or network.
func tdServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {
		sym := r.URL.Query().Get("symbol")
		if strings.Contains(sym, ",") {
			_, _ = w.Write([]byte(`{
				"AAPL":{"symbol":"AAPL","close":"150.00","open":"148.0","high":"151.0","low":"147.0","volume":"1000","timestamp":1700000000},
				"MSFT":{"symbol":"MSFT","close":"300.00","open":"298.0","high":"302.0","low":"297.0","volume":"2000","timestamp":1700000000}
			}`))
			return
		}
		_, _ = w.Write([]byte(`{"symbol":"AAPL","close":"150.00","open":"148.0","high":"151.0","low":"147.0","volume":"1000","timestamp":1700000000}`))
	})
	mux.HandleFunc("/time_series", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"values":[
			{"datetime":"2026-01-02","open":"149","high":"152","low":"148","close":"150","volume":"1000"},
			{"datetime":"2026-01-03","open":"150","high":"155","low":"149","close":"154","volume":"1200"}
		]}`))
	})
	mux.HandleFunc("/statistics", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"statistics":{
			"valuations_metrics":{"trailing_pe":"28.5","forward_pe":"25.0","price_to_book_mrq":"45.2","price_to_sales_ttm":"7.1"},
			"financials":{"gross_margin":"0.43","operating_margin":"0.30","profit_margin":"0.25","return_on_equity_ttm":"1.5"},
			"stock_statistics":{"beta":"1.25"}
		}}`))
	})
	return httptest.NewServer(mux)
}

func TestTwelveData_Quote(t *testing.T) {
	srv := tdServer(t)
	defer srv.Close()
	p := NewTwelveDataProvider("test-key", srv.URL, srv.Client())

	q, err := p.Quote(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("Quote: %v", err)
	}
	if q.Symbol != "AAPL" || q.Price != 150 || q.High != 151 {
		t.Errorf("parsed quote = %+v", q)
	}
}

func TestTwelveData_BatchQuotes(t *testing.T) {
	srv := tdServer(t)
	defer srv.Close()
	p := NewTwelveDataProvider("test-key", srv.URL, srv.Client())

	quotes, err := p.BatchQuotes(context.Background(), []string{"AAPL", "MSFT"})
	if err != nil {
		t.Fatalf("BatchQuotes: %v", err)
	}
	if len(quotes) != 2 || quotes["MSFT"].Price != 300 {
		t.Errorf("parsed batch = %+v", quotes)
	}
}

func TestTwelveData_Candles(t *testing.T) {
	srv := tdServer(t)
	defer srv.Close()
	p := NewTwelveDataProvider("test-key", srv.URL, srv.Client())

	candles, err := p.Candles(context.Background(), "AAPL", "1d", time.Now().AddDate(0, 0, -5), time.Now())
	if err != nil {
		t.Fatalf("Candles: %v", err)
	}
	if len(candles) != 2 || candles[1].Close != 154 {
		t.Errorf("parsed candles = %+v", candles)
	}
}

func TestTwelveData_Fundamentals(t *testing.T) {
	srv := tdServer(t)
	defer srv.Close()
	p := NewTwelveDataProvider("test-key", srv.URL, srv.Client())

	f, err := p.Fundamentals(context.Background(), "AAPL")
	if err != nil {
		t.Fatalf("Fundamentals: %v", err)
	}
	if f.PE != 28.5 || f.GrossMargin != 0.43 || f.Beta != 1.25 {
		t.Errorf("parsed fundamentals = %+v", f)
	}
}

func TestTwelveData_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	p := NewTwelveDataProvider("k", srv.URL, srv.Client())
	if _, err := p.Quote(context.Background(), "AAPL"); err != ErrRateLimited {
		t.Fatalf("err = %v, want ErrRateLimited", err)
	}
}
