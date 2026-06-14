package controllers_test

// E2E for the portfolio endpoints: real router + auth middleware + PortfolioHandler
// + PortfolioService + pure valuation. Only the read repository is faked (the
// data layer). Reuses do/dataOf from authflow_test.go (same test package).

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakePortfolioRepo struct {
	account  *models.Account
	holdings []models.Holding
}

func (f *fakePortfolioRepo) GetAccountByID(context.Context, uuid.UUID) (*models.Account, error) {
	return f.account, nil
}
func (f *fakePortfolioRepo) ListHoldings(context.Context, uuid.UUID) ([]models.Holding, error) {
	return f.holdings, nil
}
func (f *fakePortfolioRepo) TopTraders(context.Context, int) ([]repository.TraderRow, error) {
	return nil, nil
}

func newPortfolioRouter(repo *fakePortfolioRepo) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	svc := service.NewPortfolioService(repo)
	h := controllers.NewPortfolioHandler(svc, nil)
	mw := middlewares.AuthMiddleware(tm)

	r := gin.New()
	g := r.Group("/api/portfolio")
	g.Use(mw)
	{
		g.GET("/stats", h.HandleGetUserPortfolioStats)
		g.GET("/", h.HandleGetUsersStockHoldings)
	}
	token, _, _ := tm.GenerateAccessToken(uuid.New().String(), repo.account.ID.String(), "user")
	return r, token
}

func TestPortfolio_StatsComputesValuation(t *testing.T) {
	acctID := uuid.New()
	repo := &fakePortfolioRepo{
		// Spent 10k of the 100k start: 90k cash + 100 sh @ avg $100 cost basis.
		account: &models.Account{ID: acctID, Balance: 90000},
		holdings: []models.Holding{
			{Quantity: 100, AvgPrice: 100, Stock: models.Stock{Symbol: "AAPL", Name: "Apple", CurrentPrice: 120}},
		},
	}
	r, token := newPortfolioRouter(repo)

	status, body := do(t, r, http.MethodGet, "/api/portfolio/stats", token, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	d := dataOf(t, body)

	// market = 100*120 = 12000; total = 90000 + 12000 = 102000.
	if got := d["total_value"].(float64); got != 102000 {
		t.Errorf("total_value = %v, want 102000", got)
	}
	// unrealized = 12000 - 10000 = 2000.
	if got := d["unrealized_pl"].(float64); got != 2000 {
		t.Errorf("unrealized_pl = %v, want 2000", got)
	}
	// ROI = 102000/100000 - 1 = 0.02.
	if got := d["roi"].(float64); got != 0.02 {
		t.Errorf("roi = %v, want 0.02", got)
	}
}

func TestPortfolio_HoldingsList(t *testing.T) {
	acctID := uuid.New()
	repo := &fakePortfolioRepo{
		account: &models.Account{ID: acctID, Balance: 50000},
		holdings: []models.Holding{
			{Quantity: 10, AvgPrice: 50, Stock: models.Stock{Symbol: "MSFT", Name: "Microsoft", CurrentPrice: 60}},
		},
	}
	r, token := newPortfolioRouter(repo)

	status, body := do(t, r, http.MethodGet, "/api/portfolio/", token, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	items, ok := body["data"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected 1 holding, got %v", body["data"])
	}
	h := items[0].(map[string]any)
	if h["symbol"] != "MSFT" {
		t.Errorf("symbol = %v, want MSFT", h["symbol"])
	}
	if h["market_value"].(float64) != 600 {
		t.Errorf("market_value = %v, want 600", h["market_value"])
	}
}

func TestPortfolio_RequiresAuth(t *testing.T) {
	repo := &fakePortfolioRepo{account: &models.Account{ID: uuid.New(), Balance: 100000}}
	r, _ := newPortfolioRouter(repo)
	if status, _ := do(t, r, http.MethodGet, "/api/portfolio/stats", "", nil); status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
}
