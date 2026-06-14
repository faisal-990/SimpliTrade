package controllers

import (
	"strconv"

	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type PortfolioHandler struct {
	service service.PortfolioService
	trades  service.TradeService
}

func NewPortfolioHandler(s service.PortfolioService, t service.TradeService) *PortfolioHandler {
	return &PortfolioHandler{service: s, trades: t}
}

// HandleBuyStocks executes a buy for the caller's account.
func (p *PortfolioHandler) HandleBuyStocks(c *gin.Context) {
	var req dto.TradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid trade request: "+err.Error()))
		return
	}
	resp, err := p.trades.Buy(c.Request.Context(), middlewares.AccountID(c), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, resp)
}

// HandleSellStocks executes a sell for the caller's account.
func (p *PortfolioHandler) HandleSellStocks(c *gin.Context) {
	var req dto.TradeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid trade request: "+err.Error()))
		return
	}
	resp, err := p.trades.Sell(c.Request.Context(), middlewares.AccountID(c), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, resp)
}

// HandleGetUsersTradeHistory returns the caller's trade history (paginated).
func (p *PortfolioHandler) HandleGetUsersTradeHistory(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 50)
	offset := atoiDefault(c.Query("offset"), 0)
	items, err := p.trades.History(c.Request.Context(), middlewares.AccountID(c), limit, offset)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, items)
}

// HandleGetUsersStockHoldings returns the caller's holdings (T4).
func (p *PortfolioHandler) HandleGetUsersStockHoldings(c *gin.Context) {
	httpx.OK(c, gin.H{"message": "holdings not implemented yet (T4)"})
}

// HandleGetUserPortfolioStats returns portfolio P&L/ROI (T4).
func (p *PortfolioHandler) HandleGetUserPortfolioStats(c *gin.Context) {
	httpx.OK(c, gin.H{"message": "portfolio stats not implemented yet (T4)"})
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil && n >= 0 {
		return n
	}
	return def
}
