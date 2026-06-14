package controllers_test

// E2E for the trade endpoints: real router + auth middleware + PortfolioHandler +
// TradeService. The Broker and trade repo are faked (the data layer), so we test
// the real HTTP/validation/error-mapping path without a DB. (Reuses the do/
// dataOf/errCodeOf helpers from authflow_test.go — same test package.)

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/broker"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/trading"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// fakeBroker records the last order and returns a canned fill or error.
type fakeBroker struct {
	last broker.Order
	err  error
}

func (f *fakeBroker) Execute(_ context.Context, o broker.Order) (broker.Fill, error) {
	f.last = o
	if f.err != nil {
		return broker.Fill{}, f.err
	}
	return broker.Fill{
		TradeID: uuid.New(), Symbol: o.Symbol, Side: o.Side,
		Quantity: o.Quantity, Price: 100, TotalValue: 100 * o.Quantity,
		ExecutedAt: time.Now(),
	}, nil
}

// fakeHistoryRepo serves trade history; execution methods are unused (the broker
// is faked) so they no-op.
type fakeHistoryRepo struct{ trades []models.Trade }

func (f *fakeHistoryRepo) ExecuteBuy(context.Context, uuid.UUID, string, float64, *string, string) (*models.Trade, error) {
	return nil, nil
}
func (f *fakeHistoryRepo) ExecuteSell(context.Context, uuid.UUID, string, float64, *string, string) (*models.Trade, error) {
	return nil, nil
}
func (f *fakeHistoryRepo) ListByAccount(context.Context, uuid.UUID, int, int) ([]models.Trade, error) {
	return f.trades, nil
}
func (f *fakeHistoryRepo) SellAll(context.Context, uuid.UUID) (int, error) { return 0, nil }

func newTradeRouter(b broker.Broker, repo repository.TradeRepo) (*gin.Engine, string) {
	gin.SetMode(gin.TestMode)
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	svc := service.NewTradeService(b, repo)
	h := controllers.NewPortfolioHandler(nil, svc)
	mw := middlewares.AuthMiddleware(tm)

	r := gin.New()
	g := r.Group("/api/trade")
	g.Use(mw)
	{
		g.POST("/buy", h.HandleBuyStocks)
		g.POST("/sell", h.HandleSellStocks)
		g.GET("/history", h.HandleGetUsersTradeHistory)
	}
	// A valid access token carrying a real account id.
	token, _, _ := tm.GenerateAccessToken(uuid.New().String(), uuid.New().String(), "user")
	return r, token
}

func TestTrade_BuySuccess(t *testing.T) {
	fb := &fakeBroker{}
	r, token := newTradeRouter(fb, &fakeHistoryRepo{})

	status, body := do(t, r, http.MethodPost, "/api/trade/buy", token,
		map[string]any{"symbol": "AAPL", "quantity": 10})
	if status != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (%v)", status, body)
	}
	d := dataOf(t, body)
	if d["side"] != "buy" || d["symbol"] != "AAPL" {
		t.Errorf("response = %v", d)
	}
	if fb.last.Side != broker.Buy || fb.last.Quantity != 10 {
		t.Errorf("broker received order %+v", fb.last)
	}
}

func TestTrade_SellSuccess(t *testing.T) {
	fb := &fakeBroker{}
	r, token := newTradeRouter(fb, &fakeHistoryRepo{})
	status, _ := do(t, r, http.MethodPost, "/api/trade/sell", token,
		map[string]any{"symbol": "MSFT", "quantity": 3})
	if status != http.StatusCreated {
		t.Fatalf("status = %d, want 201", status)
	}
	if fb.last.Side != broker.Sell {
		t.Errorf("expected sell order, got %+v", fb.last)
	}
}

func TestTrade_InsufficientFundsIs400(t *testing.T) {
	r, token := newTradeRouter(&fakeBroker{err: trading.ErrInsufficientFunds}, &fakeHistoryRepo{})
	status, body := do(t, r, http.MethodPost, "/api/trade/buy", token,
		map[string]any{"symbol": "AAPL", "quantity": 999999})
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (%v)", status, body)
	}
}

func TestTrade_UnknownSymbolIs404(t *testing.T) {
	r, token := newTradeRouter(&fakeBroker{err: repository.ErrNotFound}, &fakeHistoryRepo{})
	status, _ := do(t, r, http.MethodPost, "/api/trade/buy", token,
		map[string]any{"symbol": "ZZZZ", "quantity": 1})
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
}

func TestTrade_RequiresAuth(t *testing.T) {
	r, _ := newTradeRouter(&fakeBroker{}, &fakeHistoryRepo{})
	status, _ := do(t, r, http.MethodPost, "/api/trade/buy", "",
		map[string]any{"symbol": "AAPL", "quantity": 1})
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
}

func TestTrade_InvalidQuantityIs400(t *testing.T) {
	r, token := newTradeRouter(&fakeBroker{}, &fakeHistoryRepo{})
	status, body := do(t, r, http.MethodPost, "/api/trade/buy", token,
		map[string]any{"symbol": "AAPL", "quantity": 0})
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	if errCodeOf(body) != "validation_error" {
		t.Errorf("error code = %q, want validation_error", errCodeOf(body))
	}
}

func TestTrade_History(t *testing.T) {
	repo := &fakeHistoryRepo{trades: []models.Trade{
		{ID: uuid.New(), Type: "buy", Quantity: 10, Price: 100, TotalValue: 1000, Status: "executed", Stock: models.Stock{Symbol: "AAPL"}},
	}}
	r, token := newTradeRouter(&fakeBroker{}, repo)
	status, body := do(t, r, http.MethodGet, "/api/trade/history", token, nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%v)", status, body)
	}
	items, ok := body["data"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected 1 history item, got %v", body["data"])
	}
}
