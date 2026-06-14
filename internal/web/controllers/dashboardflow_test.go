package controllers_test

// E2E for the dashboard endpoints (public — market data is not user-specific):
// real router + DashboardService + NewsService. Only the StockRepo is faked.

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

// fakeStockRepo implements repository.StockRepo; only the read methods the
// dashboard uses are meaningful.
type fakeStockRepo struct {
	stocks  map[string]models.Stock
	candles map[string][]models.StockPrice
}

func (f *fakeStockRepo) Upsert(context.Context, *models.Stock) error { return nil }
func (f *fakeStockRepo) GetBySymbol(_ context.Context, sym string) (*models.Stock, error) {
	if s, ok := f.stocks[sym]; ok {
		return &s, nil
	}
	return nil, repository.ErrNotFound
}
func (f *fakeStockRepo) List(_ context.Context, limit, offset int) ([]models.Stock, error) {
	out := make([]models.Stock, 0, len(f.stocks))
	for _, s := range f.stocks {
		out = append(out, s)
	}
	return out, nil
}
func (f *fakeStockRepo) ListSymbols(context.Context) ([]string, error)          { return nil, nil }
func (f *fakeStockRepo) InsertCandle(context.Context, *models.StockPrice) error { return nil }
func (f *fakeStockRepo) GetCandles(_ context.Context, sym, _ string, _ int) ([]models.StockPrice, error) {
	return f.candles[sym], nil
}
func (f *fakeStockRepo) UpdatePrice(context.Context, string, float64) error { return nil }

func newDashboardRouter(repo repository.StockRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := controllers.NewDashboardHandler(service.NewDashboardService(repo), service.NewNewsService())

	r := gin.New()
	g := r.Group("/api/dashboard")
	{
		g.GET("/fundamentals", h.HandleGetStocksFundamentals)
		g.GET("/graph/:symbol", h.HandleGetStocksDetails)
		g.GET("/news", h.HandleGetStocksNews)
	}
	return r
}

func sampleRepo() *fakeStockRepo {
	return &fakeStockRepo{
		stocks: map[string]models.Stock{
			"AAPL": {Symbol: "AAPL", Name: "Apple", Sector: "Technology", AssetClass: "equity", CurrentPrice: 150,
				Fundamentals: models.Fundamentals{PE: 28, ROE: 0.4}},
		},
		candles: map[string][]models.StockPrice{
			"AAPL": {
				{Timestamp: time.Now(), Open: 149, High: 152, Low: 148, Close: 150, Volume: 1000, Interval: "1d"},
				{Timestamp: time.Now().Add(-24 * time.Hour), Open: 145, High: 150, Low: 144, Close: 149, Volume: 900, Interval: "1d"},
			},
		},
	}
}

func TestDashboard_Fundamentals(t *testing.T) {
	r := newDashboardRouter(sampleRepo())
	status, body := do(t, r, http.MethodGet, "/api/dashboard/fundamentals", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	items, ok := body["data"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected 1 stock, got %v", body["data"])
	}
	first := items[0].(map[string]any)
	if first["symbol"] != "AAPL" {
		t.Errorf("symbol = %v, want AAPL", first["symbol"])
	}
	if _, hasFund := first["fundamentals"]; !hasFund {
		t.Error("fundamentals should be included")
	}
}

func TestDashboard_StockDetailWithCandles(t *testing.T) {
	r := newDashboardRouter(sampleRepo())
	status, body := do(t, r, http.MethodGet, "/api/dashboard/graph/AAPL", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	d := dataOf(t, body)
	if d["symbol"] != "AAPL" {
		t.Errorf("symbol = %v, want AAPL", d["symbol"])
	}
	candles, ok := d["candles"].([]any)
	if !ok || len(candles) != 2 {
		t.Fatalf("expected 2 candles, got %v", d["candles"])
	}
	// Candles must be oldest-first: first close (149) precedes latest (150).
	if candles[0].(map[string]any)["close"].(float64) != 149 {
		t.Errorf("candles not oldest-first: %v", candles[0])
	}
}

func TestDashboard_UnknownSymbol404(t *testing.T) {
	r := newDashboardRouter(sampleRepo())
	if status, _ := do(t, r, http.MethodGet, "/api/dashboard/graph/ZZZZ", "", nil); status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
}

func TestDashboard_News(t *testing.T) {
	r := newDashboardRouter(sampleRepo())
	status, body := do(t, r, http.MethodGet, "/api/dashboard/news", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	if _, ok := body["data"].([]any); !ok {
		t.Errorf("expected a news data array, got %v", body["data"])
	}
}
