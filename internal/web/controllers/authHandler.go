package controllers

import (
	"net/http"

	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{
		service: s,
	}
}

func (a *AuthHandler) HandleAuthLogin(c *gin.Context) {
}

func (a *AuthHandler) HandleAuthSignup(c *gin.Context) {
	var loginData dto.Login

	err := c.ShouldBindBodyWithJSON(&loginData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "the json response that you sent is not in proper form",
		})
		return
	}
	name := loginData.Name
	token, err := utils.GenerateJwt(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "server error , failed to generate jwt token",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (a *AuthHandler) HandleAuthForgotPassword(c *gin.Context) {
}

func (a *AuthHandler) HandleAuthForMe(c *gin.Context) {
}
