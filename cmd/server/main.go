package main

import (
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/config"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/storage"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/router"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
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

	// dashboard
	dashboardrepo := repository.NewDashboardRepo(db)
	dashboardservice := service.NewDashboardService(dashboardrepo)
	dashboardhandler := controllers.NewDashboardHandler(dashboardservice, newsservice)

	// investor
	investorrepo := repository.NewInvestorRepo(db)
	investorservice := service.NewInvestorService(investorrepo)
	investorhandler := controllers.NewInvestorHandler(investorservice)

	// portfolio
	portfoliorepo := repository.NewPortfolioRepo(db)
	portfolioservice := service.NewPortfolioService(portfoliorepo)
	portfoliohandler := controllers.NewPortfolioHandler(portfolioservice)

	utils.LogInfo("modules loaded, starting Gin engine")
	r := gin.Default()

	r.Use(middlewares.CORSMiddleware())

	router.InitializeRoutes(r, authMW, authhandler, dashboardhandler, investorhandler, portfoliohandler)

	log.Println("✅ Initialized routes")

	port := cfg.HTTP.Port
	fmt.Println("*****************************")
	fmt.Printf("STARTING SERVER AT:%s\n", time.Now())
	fmt.Println("*****************************")
	log.Printf("🚀 Server running at :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %s", err)
	}
}
