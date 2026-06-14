// Command engine is Tower 2: the strategy daemon. It seeds the bot investors,
// then on each tick refreshes market prices, runs every bot's strategy, executes
// the resulting trades through the simulated broker, and recomputes the
// leaderboard. All state lives in Postgres, so the engine resumes cleanly on
// restart.
package main

import (
	"context"
	"log"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/faisal-990/ProjectInvestApp/internal/broker"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/runner"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/marketdata"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/config"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/storage"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ config: %v", err)
	}
	utils.InitLogger(cfg.IsProd(), slog.LevelInfo)
	logger := utils.Logger()

	db, err := storage.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("❌ db: %v", err)
	}

	// Repositories + broker.
	stockRepo := repository.NewStockRepo(db)
	portfolioRepo := repository.NewPortfolioRepo(db)
	tradeRepo := repository.NewTradeRepo(db)
	perfRepo := repository.NewPerformanceRepo(db)
	botRepo := repository.NewBotRepo(db)
	sim := broker.NewSimulatedBroker(tradeRepo)

	// Load strategies and provision a bot per enabled strategy.
	configs, err := strategy.LoadDir(cfg.Engine.StrategiesDir)
	if err != nil {
		log.Fatalf("❌ load strategies: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	bots := make([]runner.Bot, 0, len(configs))
	for _, c := range configs {
		if !c.Identity.Enabled {
			continue
		}
		investorID, accountID, err := botRepo.UpsertBot(ctx, c.Identity.ID, c.Identity.Name, c.Identity.Philosophy, c.Identity.Style)
		if err != nil {
			log.Fatalf("❌ seed bot %s: %v", c.Identity.ID, err)
		}
		bots = append(bots, runner.Bot{InvestorID: investorID, AccountID: accountID, Config: c})
	}
	logger.Info("engine: bots provisioned", "count", len(bots))

	// Provider + runner wiring. Config selects fake (default) or a real provider
	// (e.g. twelvedata) when MARKET_API_KEY is set — see marketdata.NewProvider.
	provider := marketdata.NewProvider(cfg.Market.Provider, cfg.Market.APIKey, cfg.Market.RatePerMin)
	r := runner.New(
		runner.NewDBMarketSource(stockRepo),
		runner.NewDBPortfolioSource(portfolioRepo),
		broker.BrokerFor(models.ModeSim, sim),
		runner.NewPerformanceStore(perfRepo),
		bots,
		logger,
	).WithRefresher(runner.NewDBRefresher(provider, stockRepo, marketdata.Symbols()))

	r.Run(ctx, cfg.Engine.TickInterval)
}
