package router

import (
	"github.com/faisal-990/ProjectInvestApp/backend/internal/controllers/auth"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/controllers/dashboard"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/controllers/investor"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/controllers/portfolio"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
)


func InitializeRoutes(router *gin.Engine) {

	router.Use(middlewares.LoggerMiddleware())
	api := router.Group("/api")

	// Auth Routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", auth.HandleAuthLogin)
		authGroup.POST("/signup", auth.HandleAuthSignup)
		authGroup.POST("/forgot-password", auth.HandleAuthForgotPassword)

		// ✅ Protect only /me
		authGroup.POST("/me", middlewares.AuthMiddlewear(), auth.HandleAuthForMe)
	}

	// Investor Routes
	investorGroup := api.Group("/investor")
	{
		investorGroup.GET("/", investor.HandleGetInvestor)
		investorGroup.GET("/:id", investor.HandleGetInvestorById)
		investorGroup.GET("/:id/trades", investor.HandleGetInvestorTrades)

		// ✅ Protect only this one
		investorGroup.DELETE("/:id/follow", middlewares.AuthMiddlewear(), investor.HandleUnfollowInvestor)
		investorGroup.POST("/:id/follow", investor.HandleFollowInvestor)
	}

	// Trade Routes — ✅ protect the whole group
	tradeGroup := api.Group("/trade")
	tradeGroup.Use(middlewares.AuthMiddlewear())
	{
		tradeGroup.POST("/buy", portfolio.HandleBuyStocks)
		tradeGroup.POST("/sell", portfolio.HandleSellStocks)
		tradeGroup.GET("/history", portfolio.HandleGetUsersTradeHistory)
	}

	// Portfolio Routes — ✅ protect the whole group
	portfolioGroup := api.Group("/portfolio")
	portfolioGroup.Use(middlewares.AuthMiddlewear())
	{
		portfolioGroup.GET("/stats", portfolio.HandleGetUserPortfolioStats)
		portfolioGroup.GET("/", portfolio.HandleGetUsersStockHoldings)
	}

	// Dashboard Routes — ❌ no auth middleware
	dashboardGroup := api.Group("/dashboard")
	{
		dashboardGroup.GET("/fundamentals", dashboard.HandleGetStocksFundamentals)
		dashboardGroup.GET("/graph/:symbol", dashboard.HandleGetStocksDetails)
		dashboardGroup.GET("/news", dashboard.HandleGetStocksNews)
	}
}

