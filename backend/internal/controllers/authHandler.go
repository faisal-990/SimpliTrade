package controllers

import (
	"net/http"
	"strings"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/dto"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/service"
	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		case gorm.ErrRecordNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"Error": "Please signup , you dont have any existing account",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"Internal server error": err,
			})
		}
		return
	}

	token, err := utils.GenerateJwt(object.ID)
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
	//get the name,email,password for signup
	var object dto.Signup

	if err := c.BindJSON(&object); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Invalid data fields sent",
		})
		return
	}

	//check if the email already exists in our databse , then redirect them to endpoint
	ctx := c.Request.Context()
	object.Email = strings.ToLower(strings.TrimSpace(object.Email))
	object.Name = strings.TrimSpace(object.Name)

	id, err := a.service.CreateUser(ctx, &object)
	if err != nil {

		switch err {
		case utils.ErrUserAlreadyExist:
			c.JSON(http.StatusConflict, gin.H{
				"Message": "User you are trying to signup already exist , please head to login",
			})
			break
		case utils.ErrInvalidEmail:
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Please enter a valid email address",
			})
			break
		case utils.ErrInvalidPassword:

			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Please enter a valid password",
			})
			break
		case utils.ErrNoName:

			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Please enter a name",
			})
			break
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
		}

		return
	}
	token, err := utils.GenerateJwt(id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Failed to generate token after signing up the user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Message": "Welcome to the app , enjoy your investing",
		"Token":   token,
	})
	return
}

func (a *AuthHandler) HandleAuthForgotPassword(c *gin.Context) {
}

func (a *AuthHandler) HandleAuthForMe(c *gin.Context) {
}
