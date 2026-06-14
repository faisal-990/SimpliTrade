package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	service service.AnalyticsService
}

func NewAnalyticsHandler(s service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{service: s}
}

// HandleAnalytics returns the caller's portfolio performance + risk metrics.
func (h *AnalyticsHandler) HandleAnalytics(c *gin.Context) {
	res, err := h.service.Compute(c.Request.Context(), middlewares.AccountID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, res)
}
