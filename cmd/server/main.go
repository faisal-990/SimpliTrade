package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/broker"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/runner"
	"github.com/faisal-990/ProjectInvestApp/internal/engine/strategy"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/config"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/mailer"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/models"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/storage"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/router"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %s", err)
	}
	utils.InitLogger(cfg.IsProd(), slog.LevelInfo)
	utils.LogInfo("configuration loaded")

	utils.LogInfo("connecting to database")
	db, err := storage.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %s", err)
	}

	// token manager (issues/validates JWTs + refresh tokens) from config
	tokenManager := auth.NewTokenManager(cfg.Auth.JWTSecret, cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	authMW := middlewares.AuthMiddleware(tokenManager)

	// loading all the layers
	// auth
	authrepo := repository.NewAuthRepo(db)
	mail := mailer.New(mailer.SMTPConfig{
		Host: cfg.Mail.Host, Port: cfg.Mail.Port,
		Username: cfg.Mail.Username, Password: cfg.Mail.Password, From: cfg.Mail.From,
	})
	authservice := service.NewAuthService(authrepo, tokenManager, mail, cfg.HTTP.AppBaseURL)
	authhandler := controllers.NewAuthHandler(authservice)

	// News
	newsservice := service.NewNewsService()

	// dashboard (read view over the stock universe the engine maintains)
	stockrepo := repository.NewStockRepo(db)
	dashboardservice := service.NewDashboardService(stockrepo)
	dashboardhandler := controllers.NewDashboardHandler(dashboardservice, newsservice)

	// investor
	investorrepo := repository.NewInvestorRepo(db)
	investorservice := service.NewInvestorService(investorrepo)
	investorhandler := controllers.NewInvestorHandler(investorservice)

	// trading (broker seam: sim today, live later via Account.Mode)
	traderepo := repository.NewTradeRepo(db)
	simBroker := broker.NewSimulatedBroker(traderepo)
	tradeservice := service.NewTradeService(broker.BrokerFor(models.ModeSim, simBroker), traderepo)

	// portfolio
	portfoliorepo := repository.NewPortfolioRepo(db)
	portfolioservice := service.NewPortfolioService(portfoliorepo)
	portfoliohandler := controllers.NewPortfolioHandler(portfolioservice, tradeservice)

	// allocations (capped copy-trading sub-accounts)
	allocationrepo := repository.NewAllocationRepo(db)
	allocationservice := service.NewAllocationService(allocationrepo)
	allocationhandler := controllers.NewAllocationHandler(allocationservice)

	// admin/dev: an in-process engine runner so "simulate market" can run a cycle
	// on demand (bots + user copy-allocations trade, leaderboard recomputes).
	engineRunner := buildEngineRunner(db, simBroker)
	adminhandler := controllers.NewAdminHandler(engineRunner, repository.NewResetRepo(db))
	adminMW := middlewares.AdminOnly(!cfg.IsProd())

	// backtest: replay an investor's strategy over historical prices.
	backtestservice, err := service.NewBacktestService(investorrepo, stockrepo, "internal/engine/strategies")
	if err != nil {
		log.Fatalf("❌ Failed to load backtest strategies: %s", err)
	}
	backtesthandler := controllers.NewBacktestHandler(backtestservice)

	utils.LogInfo("modules loaded, starting Gin engine")
	// gin.New (not Default) so we control the middleware stack explicitly.
	r := gin.New()
	r.Use(
		middlewares.Recovery(),  // panics -> clean 500
		middlewares.RequestID(), // correlation id
		middlewares.AccessLog(), // structured per-request log
		middlewares.SecurityHeaders(),
		middlewares.CORSMiddleware(cfg.HTTP.AllowedOrigins),
		middlewares.Metrics(), // RED metrics
		middlewares.RateLimit(cfg.HTTP.RateLimitRPS, cfg.HTTP.RateLimitBurst),
	)

	// Infra probes + metrics (unversioned, conventional paths).
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/readyz", readyHandler(db))
	r.GET("/metrics", middlewares.MetricsHandler())

	router.InitializeRoutes(r, authMW, authhandler, dashboardhandler, investorhandler, portfoliohandler, allocationhandler, adminhandler, backtesthandler, adminMW)
	log.Println("✅ Initialized routes")

	runServer(r, cfg.HTTP.Port)
}

// buildEngineRunner constructs an in-process engine runner for the admin
// "simulate market" control. It seeds the bot investors (idempotent), and runs
// WITHOUT a market refresher — it decides on the prices already in the DB, so
// triggering it never reaches an external API or re-prices the universe.
func buildEngineRunner(db *gorm.DB, sim *broker.SimulatedBroker) *runner.Runner {
	stockRepo := repository.NewStockRepo(db)
	portfolioRepo := repository.NewPortfolioRepo(db)
	perfRepo := repository.NewPerformanceRepo(db)
	botRepo := repository.NewBotRepo(db)
	allocRepo := repository.NewAllocationRepo(db)

	configs, err := strategy.LoadDir("internal/engine/strategies")
	if err != nil {
		log.Printf("⚠️  admin engine: load strategies: %v", err)
	}
	ctx := context.Background()
	bots := make([]runner.Bot, 0, len(configs))
	for _, c := range configs {
		if !c.Identity.Enabled {
			continue
		}
		investorID, accountID, err := botRepo.UpsertBot(ctx, c.Identity.ID, c.Identity.Name, c.Identity.Philosophy, c.Identity.Style)
		if err != nil {
			log.Printf("⚠️  admin engine: seed bot %s: %v", c.Identity.ID, err)
			continue
		}
		bots = append(bots, runner.Bot{InvestorID: investorID, AccountID: accountID, Config: c})
	}

	return runner.New(
		runner.NewDBMarketSource(stockRepo),
		runner.NewDBPortfolioSource(portfolioRepo),
		broker.BrokerFor(models.ModeSim, sim),
		runner.NewPerformanceStore(perfRepo),
		bots,
		utils.Logger(),
	).
		// Each "simulate market" cycle advances the synthetic market a step
		// (symmetric random-walk of current prices) so prices actually move both
		// ways — bots react, P&L and ROI change, and stop-loss/take-profit sells
		// can fire, instead of the universe being frozen at the seed.
		WithRefresher(runner.NewSyntheticTicker(stockRepo, 0.03, 1)).
		WithCopies(runner.NewDBCopySource(allocRepo))
}

// readyHandler reports readiness by pinging the database — load balancers use
// this to decide whether to route traffic to the instance.
func readyHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err == nil {
			err = sqlDB.PingContext(c.Request.Context())
		}
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}

// runServer starts the HTTP server and shuts it down gracefully on SIGINT/SIGTERM,
// draining in-flight requests within a timeout.
func runServer(handler http.Handler, port string) {
	srv := &http.Server{Addr: ":" + port, Handler: handler}

	go func() {
		log.Printf("🚀 Server running at :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("❌ Failed to start server: %s", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	log.Println("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %s", err)
	}
}
