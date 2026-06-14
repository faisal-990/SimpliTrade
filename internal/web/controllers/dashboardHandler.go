package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service     service.DashboardService
	newsService service.News
}

func NewDashboardHandler(s service.DashboardService, n service.News) *DashboardHandler {
	return &DashboardHandler{service: s, newsService: n}
}

// HandleGetStocksFundamentals lists the stock universe with fundamentals.
func (d *DashboardHandler) HandleGetStocksFundamentals(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 100)
	offset := atoiDefault(c.Query("offset"), 0)
	stocks, err := d.service.Fundamentals(c.Request.Context(), limit, offset)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, stocks)
}

// HandleGetStocksDetails returns one stock's detail + recent candles for its chart.
func (d *DashboardHandler) HandleGetStocksDetails(c *gin.Context) {
	detail, err := d.service.StockDetail(c.Request.Context(), c.Param("symbol"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, detail)
}

// HandleGetStocksNews returns recent market headlines.
func (d *DashboardHandler) HandleGetStocksNews(c *gin.Context) {
	news, err := d.newsService.GetNews()
	if err != nil {
		httpx.Fail(c, httpx.Internal("could not load news").WithCause(err))
		return
	}
	httpx.OK(c, news)
}
