package marketdata

import (
	"context"
	"fmt"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
)

// SeedStore is the narrow persistence surface the Seeder needs. repository.StockRepo
// satisfies it; tests provide an in-memory implementation.
type SeedStore interface {
	Upsert(ctx context.Context, stock *models.Stock) error
	InsertCandle(ctx context.Context, price *models.StockPrice) error
}

// Seeder populates the stock universe (identity + current price + fundamentals)
// and a short price history from a Provider. Run once to bootstrap an empty DB
// (cmd/seed); the engine's poller keeps prices fresh thereafter.
type Seeder struct {
	provider Provider
	store    SeedStore
	now      func() time.Time
	// HistoryDays controls how many daily candles to backfill per symbol.
	HistoryDays int
}

// NewSeeder builds a Seeder with sensible defaults (30 days of history).
func NewSeeder(p Provider, store SeedStore) *Seeder {
	return &Seeder{provider: p, store: store, now: time.Now, HistoryDays: 30}
}

// Seed upserts every security in the universe and backfills its price history.
// It returns the number of stocks successfully seeded. A per-symbol failure is
// wrapped with the symbol for context and aborts the run, so a misconfigured
// provider surfaces immediately rather than silently seeding a partial universe.
func (s *Seeder) Seed(ctx context.Context, universe []Security) (int, error) {
	count := 0
	for _, sec := range universe {
		if err := s.seedOne(ctx, sec); err != nil {
			return count, fmt.Errorf("seeding %s: %w", sec.Symbol, err)
		}
		count++
	}
	return count, nil
}

func (s *Seeder) seedOne(ctx context.Context, sec Security) error {
	fundamentals, err := s.provider.Fundamentals(ctx, sec.Symbol)
	if err != nil {
		return err
	}
	quote, err := s.provider.Quote(ctx, sec.Symbol)
	if err != nil {
		return err
	}

	stock := &models.Stock{
		Symbol:       sec.Symbol,
		Name:         sec.Name,
		Exchange:     sec.Exchange,
		Sector:       sec.Sector,
		CurrentPrice: quote.Price,
		Currency:     "USD",
		Fundamentals: fundamentals,
	}
	if err := s.store.Upsert(ctx, stock); err != nil {
		return err
	}

	// Backfill daily candles so the dashboard graph has data on day one.
	if s.HistoryDays > 0 {
		to := s.now()
		from := to.AddDate(0, 0, -s.HistoryDays)
		candles, err := s.provider.Candles(ctx, sec.Symbol, "1d", from, to)
		if err != nil {
			return err
		}
		for _, c := range candles {
			price := &models.StockPrice{
				StockID:   stock.ID,
				Timestamp: c.Time,
				Open:      c.Open,
				High:      c.High,
				Low:       c.Low,
				Close:     c.Close,
				Volume:    c.Volume,
				Interval:  c.Interval,
			}
			if err := s.store.InsertCandle(ctx, price); err != nil {
				return err
			}
		}
	}
	return nil
}
