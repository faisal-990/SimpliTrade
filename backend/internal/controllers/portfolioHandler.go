package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type PortfolioHandler struct {
	service service.PortfolioService
}

func NewPortfolioHandler(s service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{
		service: s,
	}
}

func (p *PortfolioHandler) HandleBuyStocks(c *gin.Context) {
}

func (p *PortfolioHandler) HandleSellStocks(c *gin.Context) {
}

func (p *PortfolioHandler) HandleGetUsersTradeHistory(c *gin.Context) {
}

func (p *PortfolioHandler) HandleGetUsersStockHoldings(c *gin.Context) {
}

func (p *PortfolioHandler) HandleGetUserPortfolioStats(c *gin.Context) {
}
