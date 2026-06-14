package router

import (
	"net/http"

	"github.com/faisal-990/ProjectInvestApp/internal/web/controllers"
	"github.com/gin-gonic/gin"
)

func GethealthInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Server status Running",
	})
}

// InitializeRoutes wires all API routes. authMW is the auth guard applied to
// protected groups; it is built in main from the TokenManager so routing has no
// knowledge of the signing secret.
func InitializeRoutes(
	router *gin.Engine,
	authMW gin.HandlerFunc,
	authHandler *controllers.AuthHandler,
	dashboardHandler *controllers.DashboardHandler,
	investorHandler *controllers.InvestorHandler,
	portfolioHandler *controllers.PortfolioHandler,
	allocationHandler *controllers.AllocationHandler,
	adminHandler *controllers.AdminHandler,
	backtestHandler *controllers.BacktestHandler,
	oauthHandler *controllers.OAuthHandler,
	customInvestorHandler *controllers.CustomInvestorHandler,
	adminMW gin.HandlerFunc,
) {
	api := router.Group("/api")

	api.GET("/health", GethealthInfo)

	// Auth — signup/login/refresh/logout are public; /me requires a valid token.
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/signup", authHandler.HandleAuthSignup)
		authGroup.POST("/login", authHandler.HandleAuthLogin)
		authGroup.POST("/refresh", authHandler.HandleAuthRefresh)
		authGroup.POST("/logout", authHandler.HandleAuthLogout)
		authGroup.POST("/forgot-password", authHandler.HandleForgotPassword)
		authGroup.POST("/reset-password", authHandler.HandleResetPassword)
		authGroup.GET("/me", authMW, authHandler.HandleAuthForMe)
		authGroup.PUT("/me", authMW, authHandler.HandleUpdateMe)
		// OAuth sign-in (public): redirect to provider, then handle the callback.
		authGroup.GET("/oauth/:provider/login", oauthHandler.HandleLogin)
		authGroup.GET("/oauth/:provider/callback", oauthHandler.HandleCallback)
	}

	// Investor routes (protected).
	investorGroup := api.Group("/investor")
	investorGroup.Use(authMW)
	{
		investorGroup.GET("/", investorHandler.HandleGetInvestor)
		investorGroup.GET("/:id", investorHandler.HandleGetInvestorById)
		investorGroup.GET("/:id/trades", investorHandler.HandleGetInvestorTrades)
		investorGroup.GET("/:id/backtest", backtestHandler.HandleRun)
		investorGroup.POST("/:id/follow", investorHandler.HandleFollowInvestor)
		investorGroup.DELETE("/:id/follow", investorHandler.HandleUnfollowInvestor)
	}

	// Investors the caller follows + their aggregated trade feed (protected).
	api.GET("/following", authMW, investorHandler.HandleGetFollowing)
	api.GET("/feed", authMW, investorHandler.HandleGetFeed)

	// Social leaderboard of real users (protected).
	api.GET("/traders", authMW, portfolioHandler.HandleTraders)

	// User-authored ("build your own") investors (protected). Separate path so it
	// doesn't collide with /investor/:id.
	customGroup := api.Group("/custom-investors")
	customGroup.Use(authMW)
	{
		customGroup.POST("/", customInvestorHandler.HandleCreate)
		customGroup.GET("/", customInvestorHandler.HandleListMine)
		customGroup.DELETE("/:id", customInvestorHandler.HandleDelete)
	}

	// Capped copy-trading allocations (protected).
	allocGroup := api.Group("/allocations")
	allocGroup.Use(authMW)
	{
		allocGroup.POST("/", allocationHandler.HandleCreate)
		allocGroup.GET("/", allocationHandler.HandleList)
		allocGroup.GET("/:id", allocationHandler.HandleDetail)
		allocGroup.DELETE("/:id", allocationHandler.HandleStop)
	}

	// Trade routes (protected).
	tradeGroup := api.Group("/trade")
	tradeGroup.Use(authMW)
	{
		tradeGroup.POST("/buy", portfolioHandler.HandleBuyStocks)
		tradeGroup.POST("/sell", portfolioHandler.HandleSellStocks)
		tradeGroup.POST("/sell-all", portfolioHandler.HandleSellAll)
		tradeGroup.GET("/history", portfolioHandler.HandleGetUsersTradeHistory)
	}

	// Admin / dev controls (protected; AdminOnly allows any user in dev).
	adminGroup := api.Group("/admin")
	adminGroup.Use(authMW, adminMW)
	{
		adminGroup.POST("/simulate", adminHandler.HandleSimulate)
		adminGroup.POST("/reset", adminHandler.HandleResetMe)
	}

	// Portfolio routes (protected).
	portfolioGroup := api.Group("/portfolio")
	portfolioGroup.Use(authMW)
	{
		portfolioGroup.GET("/stats", portfolioHandler.HandleGetUserPortfolioStats)
		portfolioGroup.GET("/", portfolioHandler.HandleGetUsersStockHoldings)
	}

	// Dashboard routes (public — market data is not user-specific).
	dashboardGroup := api.Group("/dashboard")
	{
		dashboardGroup.GET("/fundamentals", dashboardHandler.HandleGetStocksFundamentals)
		dashboardGroup.GET("/graph/:symbol", dashboardHandler.HandleGetStocksDetails)
		dashboardGroup.GET("/news", dashboardHandler.HandleGetStocksNews)
	}
}
