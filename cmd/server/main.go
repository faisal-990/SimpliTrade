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
	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/config"
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
	authservice := service.NewAuthService(authrepo, tokenManager)
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

	router.InitializeRoutes(r, authMW, authhandler, dashboardhandler, investorhandler, portfoliohandler)
	log.Println("✅ Initialized routes")

	runServer(r, cfg.HTTP.Port)
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
