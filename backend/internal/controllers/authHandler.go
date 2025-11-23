package controllers

import (
	"net/http"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/dto"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/service"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
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
	//get the email and password of the user
	var user dto.Login
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Illegal credentials format": "Please ensure that the username and password are in the correct manner",
		})
		return
	}
	ctx := c.Request.Context()
	object, err := a.service.AuthenticateUser(ctx, &user)
	if err != nil {
		switch err {
		case utils.ErrWrongPassword:
			c.JSON(http.StatusBadRequest, gin.H{
				"Invalid password": "Wrong password entered",
			})

		case utils.ErrInvalidEmail:
			c.JSON(http.StatusBadRequest, gin.H{
				"Invalid email": "please enter a valid email",
			})

		case utils.ErrInvalidPassword:
			c.JSON(http.StatusBadRequest, gin.H{
				"Invalid password type": "please enter a valid password type",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"Internal server error": err,
			})
		}
		return
	}

	token, err := utils.GenerateJwt(object.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Token generation failed": err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Token": token,
	})

	return
}

func (a *AuthHandler) HandleAuthSignup(c *gin.Context) {
}

func (a *AuthHandler) HandleAuthForgotPassword(c *gin.Context) {
}

func (a *AuthHandler) HandleAuthForMe(c *gin.Context) {
}
