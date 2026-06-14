package controllers

import (
	"strconv"

	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type BacktestHandler struct {
	service service.BacktestService
}

func NewBacktestHandler(s service.BacktestService) *BacktestHandler {
	return &BacktestHandler{service: s}
}

// HandleRun replays the investor's strategy over historical prices.
// GET /investor/:id/backtest?days=180&cash=100000
func (h *BacktestHandler) HandleRun(c *gin.Context) {
	days := atoiDefault(c.Query("days"), 180)
	cash := 100000.0
	if v := c.Query("cash"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			cash = f
		}
	}
	result, err := h.service.Run(c.Request.Context(), c.Param("id"), days, cash)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, result)
}
