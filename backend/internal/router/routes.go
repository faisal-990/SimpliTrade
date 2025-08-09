package router

import (
	"github.com/faisal-990/ProjectInvestApp/backend/internal/controllers"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func InitializeRoutes(
	router *gin.Engine,
	authHandler *controllers.AuthHandler,
	dashboardHandler *controllers.DashboardHandler,
	investorHandler *controllers.InvestorHandler,
	portfolioHandler *controllers.PortfolioHandler,
) {
	router.Use(middlewares.LoggerMiddleware())
	api := router.Group("/api")

	// Auth Routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", authHandler.HandleAuthLogin)
		authGroup.POST("/signup", authHandler.HandleAuthSignup)
		authGroup.POST("/forgot-password", middlewares.AuthMiddlewear())
		authGroup.POST("/me", middlewares.AuthMiddlewear(), authHandler.HandleAuthForMe)
	}

	// Investor Routes
	investorGroup := api.Group("/investor")
	investorGroup.Use(middlewares.AuthMiddlewear())
	{
		investorGroup.GET("/", investorHandler.HandleGetInvestor)
		investorGroup.GET("/:id", investorHandler.HandleGetInvestorById)
		investorGroup.GET("/:id/trades", investorHandler.HandleGetInvestorTrades)
		investorGroup.DELETE("/:id/follow", middlewares.AuthMiddlewear(), investorHandler.HandleUnfollowInvestor)
		investorGroup.POST("/:id/follow", investorHandler.HandleFollowInvestor)
	}

	// Trade Routes
	tradeGroup := api.Group("/trade")
	tradeGroup.Use(middlewares.AuthMiddlewear())
	{
		tradeGroup.POST("/buy", portfolioHandler.HandleBuyStocks)
		tradeGroup.POST("/sell", portfolioHandler.HandleSellStocks)
		tradeGroup.GET("/history", portfolioHandler.HandleGetUsersTradeHistory)
	}

	// Portfolio Routes
	portfolioGroup := api.Group("/portfolio")
	portfolioGroup.Use(middlewares.AuthMiddlewear())
	{
		portfolioGroup.GET("/stats", portfolioHandler.HandleGetUserPortfolioStats)
		portfolioGroup.GET("/", portfolioHandler.HandleGetUsersStockHoldings)
	}

	// Dashboard Routes
	dashboardGroup := api.Group("/dashboard")
	{
		dashboardGroup.GET("/fundamentals", dashboardHandler.HandleGetStocksFundamentals)
		dashboardGroup.GET("/graph/:symbol", dashboardHandler.HandleGetStocksDetails)
		dashboardGroup.GET("/news", dashboardHandler.HandleGetStocksNews)
	}
}
