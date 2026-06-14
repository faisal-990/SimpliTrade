package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(s service.AuthService) *AuthHandler {
	return &AuthHandler{service: s}
}

// HandleAuthSignup creates a new user (and their default sim account) and
// returns an initial token pair.
func (a *AuthHandler) HandleAuthSignup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid signup request: "+err.Error()))
		return
	}
	resp, err := a.service.Signup(c.Request.Context(), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, resp)
}

// HandleAuthLogin authenticates by email + password and returns a token pair.
func (a *AuthHandler) HandleAuthLogin(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid login request: "+err.Error()))
		return
	}
	resp, err := a.service.Login(c.Request.Context(), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, resp)
}

// HandleAuthRefresh rotates a refresh token and returns a fresh token pair.
func (a *AuthHandler) HandleAuthRefresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid refresh request: "+err.Error()))
		return
	}
	resp, err := a.service.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, resp)
}

// HandleAuthLogout revokes the given refresh token (idempotent).
func (a *AuthHandler) HandleAuthLogout(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid logout request: "+err.Error()))
		return
	}
	if err := a.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"logged_out": true})
}

// HandleAuthForMe returns the authenticated caller's profile.
func (a *AuthHandler) HandleAuthForMe(c *gin.Context) {
	user, err := a.service.Me(c.Request.Context(), middlewares.UserID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, user)
}
