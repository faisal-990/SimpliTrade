package controllers

import (
	"net/http"

	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service     service.DashboardService
	newsService service.News
}

func NewDashboardHandler(s service.DashboardService, n service.News) *DashboardHandler {
	return &DashboardHandler{
		service:     s,
		newsService: n,
	}
}

func (d *DashboardHandler) HandleGetStocksDetails(c *gin.Context) {
}

func (d *DashboardHandler) HandleGetStocksNews(c *gin.Context) {
	news := []dto.NewsDTO{}
	news, err := d.newsService.GetNews()
	if err != nil {
		utils.LogInfoF("failed to fetch news", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch news",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": news,
	})
}

func (d *DashboardHandler) HandleGetStocksFundamentals(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "success gettings stokcks from HandleGetStocksFundamentals",
	})
}
