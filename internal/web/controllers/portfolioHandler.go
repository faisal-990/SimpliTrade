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

// HandleSellAll liquidates all of the caller's holdings at current prices.
func (p *PortfolioHandler) HandleSellAll(c *gin.Context) {
	n, err := p.trades.SellAll(c.Request.Context(), middlewares.AccountID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"sold": n})
}

// HandleGetUsersStockHoldings returns the caller's holdings valued at the
// current market price.
func (p *PortfolioHandler) HandleGetUsersStockHoldings(c *gin.Context) {
	holdings, err := p.service.Holdings(c.Request.Context(), middlewares.AccountID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, holdings)
}

// HandleGetUserPortfolioStats returns the account valuation: total value, P&L,
// ROI, and allocation.
func (p *PortfolioHandler) HandleGetUserPortfolioStats(c *gin.Context) {
	stats, err := p.service.Stats(c.Request.Context(), middlewares.AccountID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, stats)
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
