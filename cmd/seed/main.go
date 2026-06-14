// Command seed bootstraps the stock universe into the database using the
// configured market-data provider (FakeProvider by default; a real provider when
// MARKET_API_KEY is set). It is idempotent — safe to re-run; existing symbols are
// updated.
//
// SEED_LIMIT (env) caps how many symbols are seeded — useful on a free API tier
// where seeding the whole universe would exceed the rate/daily limits. Unset or
// 0 means the full universe.
package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/faisal-990/ProjectInvestApp/internal/marketdata"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/config"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ config: %v", err)
	}

	db, err := storage.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("❌ db: %v", err)
	}

	// Provider selection is config-driven (fake by default; twelvedata when a
	// key is set). The seeder is unchanged regardless of source.
	provider := marketdata.NewProvider(cfg.Market.Provider, cfg.Market.APIKey, cfg.Market.RatePerMin)
	provider = marketdata.WithFundamentals(provider, cfg.Market.FundamentalsProvider, cfg.Market.FundamentalsAPIKey)

	universe := marketdata.DefaultUniverse
	if limit := seedLimit(); limit > 0 && limit < len(universe) {
		universe = universe[:limit]
		log.Printf("ℹ️  SEED_LIMIT=%d — seeding a subset (free-tier friendly)", limit)
	}

	seeder := marketdata.NewSeeder(provider, repository.NewStockRepo(db))
	if d := seedHistoryDays(); d > 0 {
		seeder.HistoryDays = d // SEED_HISTORY_DAYS overrides the default backfill window
	}
	n, err := seeder.Seed(context.Background(), universe)
	if err != nil {
		log.Fatalf("❌ seed failed after %d stocks: %v", n, err)
	}
	log.Printf("✅ seeded %d stocks (provider=%s)", n, cfg.Market.Provider)
}

func seedLimit() int {
	n, _ := strconv.Atoi(os.Getenv("SEED_LIMIT"))
	return n
}

func seedHistoryDays() int {
	n, _ := strconv.Atoi(os.Getenv("SEED_HISTORY_DAYS"))
	return n
}
