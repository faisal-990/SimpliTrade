package main

import (
	"log"
	"os"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/controllers"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/db"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/middlewares"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/repository"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/router"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/service"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	utils.LogInfo("###LOADING env files")
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env files")
	}
	log.Println("✅ Loaded .env files")

	utils.LogInfo("###CONNECTING to db......")
	db, err := db.Connect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %s", err)
	}
	log.Println("✅ Connected to DB")

	// loading all the layers
	// auth
	authrepo := repository.NewAuthRepo(db)
	authservice := service.NewAuthService(authrepo)
	authhandler := controllers.NewAuthHandler(authservice)

	// dashboard
	dashboardrepo := repository.NewDashboardRepo(db)
	dashboardservice := service.NewDashboardService(dashboardrepo)
	dashboardhandler := controllers.NewDashboardHandler(dashboardservice)

	// investor
	investorrepo := repository.NewInvestorRepo(db)
	investorservice := service.NewInvestorService(investorrepo)
	investorhandler := controllers.NewInvestorHandler(investorservice)

	// portfolio
	portfoliorepo := repository.NewPortfolioRepo(db)
	portfolioservice := service.NewPortfolioService(portfoliorepo)
	portfoliohandler := controllers.NewPortfolioHandler(portfolioservice)

	utils.LogInfo("Loaded all the modules , Now starting gin Engine")
	r := gin.Default()
	log.Println("✅ Created Gin engine")

	r.Use(middlewares.CORSMiddleware())

	router.InitializeRoutes(r, authhandler, dashboardhandler, investorhandler, portfoliohandler)

	log.Println("✅ Initialized routes")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Server running at :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %s", err)
	}
}
