package middlewares

import (
	"time"

	"github.com/faisal-990/ProjectInvestApp/internal/platform/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// requestIDHeader is the header carrying the per-request correlation id.
const requestIDHeader = "X-Request-ID"

const ctxRequestID = "requestID"

// RequestID assigns each request a correlation id (honoring an inbound
// X-Request-ID if present), exposes it on the response, and stores it in context
// so logs can be tied together.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set(ctxRequestID, id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

// RequestIDOf returns the request's correlation id.
func RequestIDOf(c *gin.Context) string { return getString(c, ctxRequestID) }

// AccessLog emits one structured log line per request with method, path, status,
// latency, client IP, and request id.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		utils.Logger().Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"request_id", RequestIDOf(c),
		)
	}
}
