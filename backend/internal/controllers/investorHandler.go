package controllers

import (
	"net/http"

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
	value, exist := c.Get("name")
	if !exist {
		c.JSON(http.StatusOK, gin.H{
			"message": "logged in but user name not found",
		})
	}
	name := value.(string)

	c.JSON(http.StatusOK, gin.H{
		"welcome ": name,
		"message":  "so finally you have implemented jwt , awesome , way to go bro , you are the man!!!!!",
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
