package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/storage"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/router"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
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
	db, err := storage.Connect()
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %s", err)
	}
	log.Println("✅ Connected to DB")

	// loading all the layers
	// auth
	authrepo := repository.NewAuthRepo(db)
	authservice := service.NewAuthService(authrepo)
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

	// gemerating token to test jwt aith
	//token, err := utils.GenerateJwt("sawez")
	//if err != nil {
	//log.Fatal("failed to generate token")
	//}
	//fmt.Printf("TOKEN: %s\n", token)
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
	fmt.Println("*****************************")
	fmt.Printf("STARTING SERVER AT:%s\n", time.Now())
	fmt.Println("*****************************")
	log.Printf("🚀 Server running at :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %s", err)
	}
}
