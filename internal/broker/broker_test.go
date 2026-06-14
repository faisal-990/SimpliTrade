package broker

import (
	"context"
	"errors"
	"testing"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/trading"
	"github.com/google/uuid"
)

// fakeTradeRepo records calls and returns canned results, so SimulatedBroker's
// mapping logic is tested without a database.
type fakeTradeRepo struct {
	lastSide   string
	lastSymbol string
	lastQty    float64
	trade      *models.Trade
	err        error
}

func (f *fakeTradeRepo) ExecuteBuy(_ context.Context, acct uuid.UUID, symbol string, qty float64, _ *string) (*models.Trade, error) {
	f.lastSide, f.lastSymbol, f.lastQty = "buy", symbol, qty
	return f.trade, f.err
}

func (f *fakeTradeRepo) ExecuteSell(_ context.Context, acct uuid.UUID, symbol string, qty float64, _ *string) (*models.Trade, error) {
	f.lastSide, f.lastSymbol, f.lastQty = "sell", symbol, qty
	return f.trade, f.err
}

func (f *fakeTradeRepo) ListByAccount(_ context.Context, _ uuid.UUID, _, _ int) ([]models.Trade, error) {
	return nil, nil
}
func (f *fakeTradeRepo) SellAll(_ context.Context, _ uuid.UUID) (int, error) { return 0, nil }

func TestSimulatedBroker_BuyMapsToFill(t *testing.T) {
	id := uuid.New()
	repo := &fakeTradeRepo{trade: &models.Trade{ID: id, Quantity: 10, Price: 100, TotalValue: 1000}}
	b := NewSimulatedBroker(repo)

	fill, err := b.Execute(context.Background(), Order{AccountID: uuid.New(), Symbol: "AAPL", Side: Buy, Quantity: 10})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.lastSide != "buy" || repo.lastSymbol != "AAPL" || repo.lastQty != 10 {
		t.Errorf("repo got side=%s symbol=%s qty=%v", repo.lastSide, repo.lastSymbol, repo.lastQty)
	}
	if fill.TradeID != id || fill.TotalValue != 1000 || fill.Side != Buy {
		t.Errorf("fill = %+v", fill)
	}
}

func TestSimulatedBroker_SellRoutesToSell(t *testing.T) {
	repo := &fakeTradeRepo{trade: &models.Trade{ID: uuid.New()}}
	b := NewSimulatedBroker(repo)
	if _, err := b.Execute(context.Background(), Order{Symbol: "MSFT", Side: Sell, Quantity: 5}); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if repo.lastSide != "sell" {
		t.Errorf("expected sell routing, got %s", repo.lastSide)
	}
}

func TestSimulatedBroker_PropagatesError(t *testing.T) {
	repo := &fakeTradeRepo{err: trading.ErrInsufficientFunds}
	b := NewSimulatedBroker(repo)
	if _, err := b.Execute(context.Background(), Order{Symbol: "AAPL", Side: Buy, Quantity: 10}); !errors.Is(err, trading.ErrInsufficientFunds) {
		t.Fatalf("err = %v, want ErrInsufficientFunds", err)
	}
}

func TestSimulatedBroker_InvalidSide(t *testing.T) {
	b := NewSimulatedBroker(&fakeTradeRepo{})
	if _, err := b.Execute(context.Background(), Order{Symbol: "AAPL", Side: "hold", Quantity: 1}); !errors.Is(err, ErrInvalidSide) {
		t.Fatalf("err = %v, want ErrInvalidSide", err)
	}
}

func TestBrokerFor_DefaultsToSimulated(t *testing.T) {
	sim := NewSimulatedBroker(&fakeTradeRepo{})
	if got := BrokerFor(models.ModeSim, sim); got != sim {
		t.Error("sim mode should map to the simulated broker")
	}
	if got := BrokerFor(models.ModeLive, sim); got != sim {
		t.Error("live mode currently falls back to the simulated broker")
	}
}
