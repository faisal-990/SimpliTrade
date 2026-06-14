package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/engine/runner"
	"github.com/faisal-990/ProjectInvestApp/internal/platform/repository"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/gin-gonic/gin"
)

// AdminHandler exposes dev/admin controls: trigger a market cycle on demand and
// reset the caller's own portfolio.
type AdminHandler struct {
	engine *runner.Runner
	reset  repository.ResetRepo
}

func NewAdminHandler(engine *runner.Runner, reset repository.ResetRepo) *AdminHandler {
	return &AdminHandler{engine: engine, reset: reset}
}

// HandleSimulate runs one engine cycle: every bot decides + trades, every user
// copy-allocation trades, and the leaderboard is recomputed — so you can watch
// strategies act immediately instead of waiting for the scheduled tick.
func (h *AdminHandler) HandleSimulate(c *gin.Context) {
	if err := h.engine.RunOnce(c.Request.Context()); err != nil {
		httpx.Fail(c, httpx.Internal("market simulation failed").WithCause(err))
		return
	}
	httpx.OK(c, gin.H{"simulated": true})
}

// HandleResetMe wipes the caller's holdings, trades, and copy allocations and
// restores their primary cash to the starting balance — a clean slate for the
// current user only.
func (h *AdminHandler) HandleResetMe(c *gin.Context) {
	if err := h.reset.ResetUser(c.Request.Context(), middlewares.UserID(c)); err != nil {
		httpx.Fail(c, httpx.Internal("could not reset account").WithCause(err))
		return
	}
	httpx.OK(c, gin.H{"reset": true})
}
