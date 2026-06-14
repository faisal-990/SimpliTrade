package middlewares

import (
	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/gin-gonic/gin"
)

// Recovery converts a panic in any handler into a clean 500 error envelope
// instead of crashing the server, logging the panic with the request id.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				utils.Logger().Error("panic recovered",
					"panic", r,
					"path", c.Request.URL.Path,
					"request_id", RequestIDOf(c),
				)
				httpx.Fail(c, httpx.Internal("an unexpected error occurred"))
			}
		}()
		c.Next()
	}
}
