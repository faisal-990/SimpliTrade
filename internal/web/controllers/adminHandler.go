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
	// buildEngine rebuilds the runner per call so investors created at runtime
	// (e.g. a just-authored custom investor) are always part of the cycle.
	buildEngine func() *runner.Runner
	reset       repository.ResetRepo
}

func NewAdminHandler(buildEngine func() *runner.Runner, reset repository.ResetRepo) *AdminHandler {
	return &AdminHandler{buildEngine: buildEngine, reset: reset}
}

// HandleSimulate runs one engine cycle: every bot decides + trades, every user
// copy-allocation trades, and the leaderboard is recomputed — so you can watch
// strategies act immediately instead of waiting for the scheduled tick.
func (h *AdminHandler) HandleSimulate(c *gin.Context) {
	if err := h.buildEngine().RunOnce(c.Request.Context()); err != nil {
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
