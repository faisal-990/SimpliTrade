package marketdata

import (
	"context"
	"testing"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/google/uuid"
)

// memSeedStore is an in-memory SeedStore — the only substitution; the Seeder and
// FakeProvider under test are the real production code.
type memSeedStore struct {
	stocks  map[string]models.Stock
	candles []models.StockPrice
}

func newMemSeedStore() *memSeedStore {
	return &memSeedStore{stocks: map[string]models.Stock{}}
}

func (m *memSeedStore) Upsert(_ context.Context, s *models.Stock) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New() // mimic GORM populating the PK so candles can reference it
	}
	m.stocks[s.Symbol] = *s
	return nil
}

func (m *memSeedStore) InsertCandle(_ context.Context, p *models.StockPrice) error {
	m.candles = append(m.candles, *p)
	return nil
}

func TestSeeder_PopulatesStocksWithPriceAndFundamentals(t *testing.T) {
	store := newMemSeedStore()
	seeder := NewSeeder(NewFakeProvider(), store)

	universe := DefaultUniverse[:5]
	n, err := seeder.Seed(context.Background(), universe)
	if err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if n != len(universe) {
		t.Fatalf("seeded %d, want %d", n, len(universe))
	}
	if len(store.stocks) != len(universe) {
		t.Fatalf("store has %d stocks, want %d", len(store.stocks), len(universe))
	}

	for _, sec := range universe {
		stock, ok := store.stocks[sec.Symbol]
		if !ok {
			t.Fatalf("missing seeded stock %s", sec.Symbol)
		}
		if stock.Name != sec.Name || stock.Sector != sec.Sector {
			t.Errorf("%s metadata not carried through: %+v", sec.Symbol, stock)
		}
		if stock.CurrentPrice <= 0 {
			t.Errorf("%s seeded with non-positive price %v", sec.Symbol, stock.CurrentPrice)
		}
		if stock.Fundamentals.PE <= 0 {
			t.Errorf("%s seeded without fundamentals (PE=%v)", sec.Symbol, stock.Fundamentals.PE)
		}
	}
}

func TestSeeder_BackfillsCandleHistory(t *testing.T) {
	store := newMemSeedStore()
	seeder := NewSeeder(NewFakeProvider(), store)
	seeder.HistoryDays = 30

	if _, err := seeder.Seed(context.Background(), DefaultUniverse[:1]); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if len(store.candles) == 0 {
		t.Fatal("expected candle history to be backfilled")
	}
	for _, c := range store.candles {
		if c.StockID == uuid.Nil {
			t.Error("candle not linked to a stock id")
		}
		if c.Interval != "1d" {
			t.Errorf("candle interval = %q, want 1d", c.Interval)
		}
	}
}

func TestSeeder_NoHistoryWhenDisabled(t *testing.T) {
	store := newMemSeedStore()
	seeder := NewSeeder(NewFakeProvider(), store)
	seeder.HistoryDays = 0

	if _, err := seeder.Seed(context.Background(), DefaultUniverse[:3]); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if len(store.candles) != 0 {
		t.Errorf("expected no candles when HistoryDays=0, got %d", len(store.candles))
	}
}
