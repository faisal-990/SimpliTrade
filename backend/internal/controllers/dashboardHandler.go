package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(s service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: s,
	}
}

func (d *DashboardHandler) HandleGetStocksDetails(c *gin.Context) {
}

func (d *DashboardHandler) HandleGetStocksNews(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "market is hot right now ",
	})
}

func (d *DashboardHandler) HandleGetStocksFundamentals(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "success gettings stokcks from HandleGetStocksFundamentals",
	})
}
