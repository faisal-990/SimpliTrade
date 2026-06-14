package controllers

import (
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/engine/runner"
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/gin-gonic/gin"
)

// MarketHandler reports US market open/closed status for the UI banner.
type MarketHandler struct {
	clock *runner.USEquityClock
}

func NewMarketHandler(clock *runner.USEquityClock) *MarketHandler {
	return &MarketHandler{clock: clock}
}

// HandleStatus returns whether the market is open now and the next transition.
func (h *MarketHandler) HandleStatus(c *gin.Context) {
	open, next := h.clock.Status(time.Now())
	out := dto.MarketStatusDTO{Open: open}
	if !next.IsZero() {
		out.NextChange = next.Unix()
	}
	httpx.OK(c, out)
}
