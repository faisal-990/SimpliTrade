// Command seed bootstraps the stock universe into the database using the
// configured market-data provider (the FakeProvider until a real one is wired
// in at T9). It is idempotent — safe to re-run; existing symbols are updated.
package main

import (
	"context"
	"log"

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

	// Provider selection is config-driven; today only the deterministic fake
	// exists. A real provider drops in here at T9 without touching the seeder.
	var provider marketdata.Provider = marketdata.NewFakeProvider()

	seeder := marketdata.NewSeeder(provider, repository.NewStockRepo(db))
	n, err := seeder.Seed(context.Background(), marketdata.DefaultUniverse)
	if err != nil {
		log.Fatalf("❌ seed failed after %d stocks: %v", n, err)
	}
	log.Printf("✅ seeded %d stocks (provider=%s)", n, cfg.Market.Provider)
}
