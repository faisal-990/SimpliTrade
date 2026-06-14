package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/dto"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type AllocationHandler struct {
	service service.AllocationService
}

func NewAllocationHandler(s service.AllocationService) *AllocationHandler {
	return &AllocationHandler{service: s}
}

// HandleCreate opens a capped copy allocation for the caller.
func (h *AllocationHandler) HandleCreate(c *gin.Context) {
	var req dto.AllocateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Fail(c, httpx.Validation("invalid allocation request: "+err.Error()))
		return
	}
	alloc, err := h.service.Create(c.Request.Context(), middlewares.UserID(c), req)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.Created(c, alloc)
}

// HandleList returns the caller's copy allocations.
func (h *AllocationHandler) HandleList(c *gin.Context) {
	allocs, err := h.service.List(c.Request.Context(), middlewares.UserID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, allocs)
}

// HandleDetail returns one allocation plus what the bot did with the capital:
// its current holdings and recent trades.
func (h *AllocationHandler) HandleDetail(c *gin.Context) {
	detail, err := h.service.Detail(c.Request.Context(), middlewares.UserID(c), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, detail)
}

// HandleStop liquidates a copy allocation and returns the cash to the primary account.
func (h *AllocationHandler) HandleStop(c *gin.Context) {
	if err := h.service.Stop(c.Request.Context(), middlewares.UserID(c), c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"stopped": true})
}
