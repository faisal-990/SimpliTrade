package middlewares

import (
	"net/http"
	"strings"

	"github.com/faisal-990/ProjectInvestApp/backend/internal/utils"
	"github.com/gin-gonic/gin"
)

func AuthMiddlewear() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "user not authorized here",
			})
			return // ❗ important: stop further processing
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid authorization format",
			})
			return
		}

		token := parts[1]
		_, err := utils.ValidateJwt(token)
		if err != nil {
			utils.LogInfoF("invalid user", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}

		// Set current user in context for downstream use
		//	c.Set("name", claims.Username)
		c.Next() // ✅ Only continue if everything succeeded
	}
}
