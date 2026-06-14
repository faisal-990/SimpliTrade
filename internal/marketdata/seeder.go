package marketdata

import (
	"context"
	"fmt"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
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
			// One bad symbol (e.g. a transient error or a delisted ticker) must
			// not abort the whole run — log and continue.
			utils.LogError("seed: skipping "+sec.Symbol, err)
			continue
		}
		count++
	}
	if count == 0 && len(universe) > 0 {
		return 0, fmt.Errorf("seeding failed for all %d symbols (check provider/key)", len(universe))
	}
	return count, nil
}

func (s *Seeder) seedOne(ctx context.Context, sec Security) error {
	// Price is the only essential field — without it the stock can't trade.
	quote, err := s.provider.Quote(ctx, sec.Symbol)
	if err != nil {
		return fmt.Errorf("quote: %w", err)
	}

	// Fundamentals enrich value strategies but are often premium-gated on free
	// tiers; treat their absence as "no data" (those gates simply skip) rather
	// than failing the symbol.
	fundamentals, err := s.provider.Fundamentals(ctx, sec.Symbol)
	if err != nil {
		utils.LogError("seed: fundamentals unavailable for "+sec.Symbol+" (continuing without)", err)
		fundamentals = models.Fundamentals{}
	}

	stock := &models.Stock{
		Symbol:       sec.Symbol,
		Name:         sec.Name,
		Exchange:     sec.Exchange,
		Sector:       sec.Sector,
		AssetClass:   AssetClassOf(sec.Symbol),
		CurrentPrice: quote.Price,
		Currency:     "USD",
		Fundamentals: fundamentals,
	}
	if err := s.store.Upsert(ctx, stock); err != nil {
		return err
	}

	// Backfill daily candles so charts + momentum indicators have data. Also
	// best-effort — a candle failure shouldn't drop an otherwise-good stock.
	if s.HistoryDays > 0 {
		to := s.now()
		from := to.AddDate(0, 0, -s.HistoryDays)
		candles, err := s.provider.Candles(ctx, sec.Symbol, "1d", from, to)
		if err != nil {
			utils.LogError("seed: candles unavailable for "+sec.Symbol+" (continuing without)", err)
			return nil
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
