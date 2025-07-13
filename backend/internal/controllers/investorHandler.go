package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/backend/internal/service"
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

// this function is responsible to get all the investors data
// like once the investor page is loaded
func (i *InvestorHandler) HandleGetInvestor(c *gin.Context) {
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
