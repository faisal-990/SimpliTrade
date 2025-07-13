package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/backend/internal/service"
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
}

func (a *AuthHandler) HandleAuthForgotPassword(c *gin.Context) {
}

func (a *AuthHandler) HandleAuthForMe(c *gin.Context) {
}
