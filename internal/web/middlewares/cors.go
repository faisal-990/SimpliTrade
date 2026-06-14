package middlewares

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware builds CORS from a configured allowlist. Passing ["*"] allows
// any origin (dev); in prod, pass the real frontend origins. Credentials are
// only allowed when the origin is explicitly listed (the CORS spec forbids
// credentials with a "*" origin).
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 1 && allowedOrigins[0] == "*"
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: !allowAll,
		MaxAge:           12 * time.Hour,
	})
}
