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
		// Step 1: Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			helpers.Error(c, http.StatusUnauthorized, "authorization header required")
			c.Abort() // stop the request — don't call the next handler
			return
		}

		// Step 2: Header must be in format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			helpers.Error(c, http.StatusUnauthorized, "authorization format must be: Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// Step 3: Parse and validate the token
		claims, err := config.ParseAccessToken(tokenStr)
		if err != nil {
			// Covers: expired token, tampered signature, wrong format
			helpers.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// Step 4: Verify the user still exists and is active in the DB
		// This catches cases where an account was deactivated after login
		var count int64
		config.DB.Table("users").
			Where("id = ? AND is_active = true AND deleted_at IS NULL", claims.UserID).
			Count(&count)

		if count == 0 {
			helpers.Error(c, http.StatusUnauthorized, "account not found or deactivated")
			c.Abort()
			return
		}

		// Step 5: Inject claims into context — available in all downstream handlers
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)

		// Step 6: Continue to the next middleware or handler
		c.Next()
	}
}
