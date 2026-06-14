package middlewares

import (
	"strings"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/auth"
	"github.com/faisal-990/ProjectInvestApp/internal/web/httpx"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the Bearer access token and populates the request
// context with the caller's identity (user, account, role) for downstream
// handlers. It is constructed with a TokenManager rather than reading a global,
// so the signing secret flows from config.
func AuthMiddleware(tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			httpx.Fail(c, httpx.Unauthorized("authorization header required"))
			return
		}

		// Expected format: "Bearer <token>".
		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || token == "" {
			httpx.Fail(c, httpx.Unauthorized("authorization header must be 'Bearer <token>'"))
			return
		}

		claims, err := tm.ParseAccessToken(token)
		if err != nil {
			httpx.Fail(c, httpx.Unauthorized("invalid or expired token"))
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxAccountID, claims.AccountID)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}
