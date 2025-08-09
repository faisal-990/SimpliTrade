package controllers

import (
	"fmt"
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
	var login dto.Login
	err := c.BindJSON(&login)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid login details send by the user",
		})
	}
	name := login.Name
	password := login.Password
	// TODO :: attach db , and check password of this user in the db
	//         if correct , send the token , else send message and fail
	c.JSON(http.StatusOK, gin.H{
		"name": name,
		"pass": password,
	})
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to login",
		})
		return
	}

	if password == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "please don't leave the password blank",
		})
		return
	}

	token, err := utils.GenerateJwt(name)
	if err != nil {
		utils.LogInfoF("failed to login user", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "could not generate token",
		})
		return
	}
	c.Writer.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
	})
}

func (a *AuthHandler) HandleAuthSignup(c *gin.Context) {
	// make sure the user doens't  exist , if the user exist
	// route the data to the login page , or send message that the user already exist
	name := c.PostForm("name")
	password := c.PostForm("password")
	if name == "" && password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "please enter the credentials",
		})
		return
	} else if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "please enter the name",
		})
		return
	} else if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "please enter the password",
		})
		return
	}
	// TODO :: Check if the user already exist from the db
	// if yes , match the password from the db , if not redirect/inform
	// that the user doens't exist

	// generate a jwt token for the user
	token, err := utils.GenerateJwt(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":  "failed to generate jwt token",
			"message1": "couldn't sign in the user",
		})
		return
	}
	c.Writer.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
	c.JSON(http.StatusOK, gin.H{
		"message": "signup successfull",
		"token":   token,
	})
}

func (a *AuthHandler) HandleAuthForgotPassword(c *gin.Context) {
	// TODO :: Implement some sort of Recovery mechanism to reset Users credentials
	// reroute them to the login page for credentials verifying and token generation
}

func (a *AuthHandler) HandleAuthForMe(c *gin.Context) {
}
