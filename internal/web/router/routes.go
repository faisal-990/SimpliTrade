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
		authGroup.GET("/me", authMW, authHandler.HandleAuthForMe)
	}

	// Investor routes (protected).
	investorGroup := api.Group("/investor")
	investorGroup.Use(authMW)
	{
		investorGroup.GET("/", investorHandler.HandleGetInvestor)
		investorGroup.GET("/:id", investorHandler.HandleGetInvestorById)
		investorGroup.GET("/:id/trades", investorHandler.HandleGetInvestorTrades)
		investorGroup.POST("/:id/follow", investorHandler.HandleFollowInvestor)
		investorGroup.DELETE("/:id/follow", investorHandler.HandleUnfollowInvestor)
	}

	// Aggregated feed of the investors the caller follows (protected).
	api.GET("/feed", authMW, investorHandler.HandleGetFeed)

	// Trade routes (protected).
	tradeGroup := api.Group("/trade")
	tradeGroup.Use(authMW)
	{
		tradeGroup.POST("/buy", portfolioHandler.HandleBuyStocks)
		tradeGroup.POST("/sell", portfolioHandler.HandleSellStocks)
		tradeGroup.GET("/history", portfolioHandler.HandleGetUsersTradeHistory)
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
