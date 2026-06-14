package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type InvestorHandler struct {
	service service.InvestorService
}

func NewInvestorHandler(s service.InvestorService) *InvestorHandler {
	return &InvestorHandler{
		service: s,
	}
}

// HandleGetInvestor lists investors. Full implementation lands in T7; for now it
// confirms the auth pipeline by echoing the authenticated caller's id.
func (i *InvestorHandler) HandleGetInvestor(c *gin.Context) {
	httpx.OK(c, gin.H{
		"caller_user_id": middlewares.UserID(c),
		"message":        "investor listing not implemented yet (T7)",
	})
}

// this function is responsible to get a sinlge investor data based on the investor id
func (i *InvestorHandler) HandleGetInvestorById(c *gin.Context) {
}

func (i *InvestorHandler) HandleFollowInvestor(c *gin.Context) {
}

func (i *InvestorHandler) HandleUnfollowInvestor(c *gin.Context) {
}

func (i *InvestorHandler) HandleGetInvestorTrades(c *gin.Context) {
}
