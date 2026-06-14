package middlewares

import "github.com/gin-gonic/gin"

// Context keys under which the auth middleware stores the caller's identity.
// They are unexported; downstream handlers read them through the typed accessors
// below so the key strings live in exactly one place.
const (
	ctxUserID    = "userID"
	ctxAccountID = "accountID"
	ctxRole      = "role"
)

// UserID returns the authenticated user's ID, or "" if unauthenticated.
func UserID(c *gin.Context) string { return getString(c, ctxUserID) }

// AccountID returns the caller's active account ID, or "" if unauthenticated.
func AccountID(c *gin.Context) string { return getString(c, ctxAccountID) }

// Role returns the caller's role, or "" if unauthenticated.
func Role(c *gin.Context) string { return getString(c, ctxRole) }

func getString(c *gin.Context, key string) string {
	if v, ok := c.Get(key); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
