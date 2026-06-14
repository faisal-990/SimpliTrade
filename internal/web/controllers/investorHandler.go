package controllers

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/faisal-990/ProjectInvestApp/internal/web/middlewares"
	"github.com/faisal-990/ProjectInvestApp/internal/web/service"
	"github.com/gin-gonic/gin"
)

type InvestorHandler struct {
	service service.InvestorService
}

func NewInvestorHandler(s service.InvestorService) *InvestorHandler {
	return &InvestorHandler{service: s}
}

// HandleGetInvestor returns the leaderboard (investors ranked by ROI).
func (i *InvestorHandler) HandleGetInvestor(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 50)
	offset := atoiDefault(c.Query("offset"), 0)
	investors, err := i.service.Leaderboard(c.Request.Context(), limit, offset)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, investors)
}

// HandleGetInvestorById returns one investor's profile + standing.
func (i *InvestorHandler) HandleGetInvestorById(c *gin.Context) {
	investor, err := i.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, investor)
}

// HandleGetInvestorTrades returns an investor's recent trades.
func (i *InvestorHandler) HandleGetInvestorTrades(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 50)
	offset := atoiDefault(c.Query("offset"), 0)
	trades, err := i.service.Trades(c.Request.Context(), c.Param("id"), limit, offset)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, trades)
}

// HandleFollowInvestor makes the caller follow an investor (visibility only).
func (i *InvestorHandler) HandleFollowInvestor(c *gin.Context) {
	if err := i.service.Follow(c.Request.Context(), middlewares.UserID(c), c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"following": true})
}

// HandleUnfollowInvestor removes a follow.
func (i *InvestorHandler) HandleUnfollowInvestor(c *gin.Context) {
	if err := i.service.Unfollow(c.Request.Context(), middlewares.UserID(c), c.Param("id")); err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, gin.H{"following": false})
}

// HandleGetFollowing returns the investors the caller follows.
func (i *InvestorHandler) HandleGetFollowing(c *gin.Context) {
	investors, err := i.service.Following(c.Request.Context(), middlewares.UserID(c))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, investors)
}

// HandleGetFeed returns the aggregated trade feed of the investors the caller
// follows.
func (i *InvestorHandler) HandleGetFeed(c *gin.Context) {
	limit := atoiDefault(c.Query("limit"), 50)
	feed, err := i.service.Feed(c.Request.Context(), middlewares.UserID(c), limit)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	httpx.OK(c, feed)
}
