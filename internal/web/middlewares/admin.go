package middlewares

import (
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/gin-gonic/gin"
)

// AdminOnly guards admin/dev endpoints. In dev it allows any authenticated user
// (so the "simulate market" control is usable locally); in prod it requires the
// admin role. Must run after AuthMiddleware (which sets the role in context).
func AdminOnly(devMode bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if devMode || Role(c) == "admin" {
			c.Next()
			return
		}
		httpx.Fail(c, httpx.Forbidden("admin access required"))
	}
}
