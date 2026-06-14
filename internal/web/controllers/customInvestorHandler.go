package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type CustomInvestorHandler struct {
	service service.CustomInvestorService
}

func NewCustomInvestorHandler(s service.CustomInvestorService) *CustomInvestorHandler {
	return &CustomInvestorHandler{service: s}
}

// HandleCreate builds a new user-authored investor.
func (h *CustomInvestorHandler) HandleCreate(c *gin.Context) {
	var req dto.CreateInvestorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid investor: "+err.Error()))
		return
	}
	inv, err := h.service.Create(c.Request.Context(), middlewares.UserID(c), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, inv)
}

// HandleListMine returns the investors the caller created.
func (h *CustomInvestorHandler) HandleListMine(c *gin.Context) {
	items, err := h.service.ListMine(c.Request.Context(), middlewares.UserID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, items)
}

// HandleDelete removes a custom investor the caller created.
func (h *CustomInvestorHandler) HandleDelete(c *gin.Context) {
	if err := h.service.Delete(c.Request.Context(), middlewares.UserID(c), c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"deleted": true})
}
