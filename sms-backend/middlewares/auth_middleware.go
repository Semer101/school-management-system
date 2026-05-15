package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sms-backend/config"
	"sms-backend/helpers"
)

// AuthMiddleware validates the JWT on every protected route.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helpers.Error(c, http.StatusUnauthorized, "authorization header required")
			c.Abort()
			return
		}

		// Header must be in format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			helpers.Error(c, http.StatusUnauthorized, "authorization format must be: Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]

		claims, err := config.ParseAccessToken(tokenStr)
		if err != nil {
			helpers.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// Verify the user still exists and is active in the DB.
		var count int64
		config.DB.Table("users").
			Where("id = ? AND is_active = true AND deleted_at IS NULL", claims.UserID).
			Count(&count)

		if count == 0 {
			helpers.Error(c, http.StatusUnauthorized, "account not found or deactivated")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// SSEAuthMiddleware validates the short-lived SSE token passed as ?sse_token=<token>.
// EventSource cannot set headers. Instead of passing the full access JWT in the URL
// (which gets written to every proxy/access log), clients call POST /api/notifications/sse-token
// to get a 60-second token, then pass that as ?sse_token=. This keeps long-lived secrets out of logs.
func SSEAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sseToken := c.Query("sse_token")
		if sseToken == "" {
			helpers.Error(c, http.StatusUnauthorized, "sse_token query parameter required — call POST /api/notifications/sse-token first")
			c.Abort()
			return
		}

		claims, err := config.ParseSSEToken(sseToken)
		if err != nil {
			helpers.Error(c, http.StatusUnauthorized, "invalid or expired SSE token")
			c.Abort()
			return
		}

		// Confirm user is still active.
		var count int64
		config.DB.Table("users").
			Where("id = ? AND is_active = true AND deleted_at IS NULL", claims.UserID).
			Count(&count)
		if count == 0 {
			helpers.Error(c, http.StatusUnauthorized, "account not found or deactivated")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)

		c.Next()
	}
}
