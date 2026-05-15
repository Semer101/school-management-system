package middlewares

import (
	"fmt"
	"net/http"

	"sms-backend/config"
	"sms-backend/helpers"
	"sms-backend/models"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware checks that the authenticated user has one of the allowed roles.
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role injected by AuthMiddleware — always run AuthMiddleware first
		userRole := c.GetString("role")

		if userRole == "" {
			helpers.Error(c, http.StatusUnauthorized, "no role found — ensure AuthMiddleware runs first")
			c.Abort()
			return
		}

		// Check if user's role is in the allowed list
		for _, role := range allowedRoles {
			if userRole == role {
				c.Next() // role matches — allow through
				return
			}
		}

		// None of the allowed roles matched
		helpers.Error(c, http.StatusForbidden, "access denied: insufficient permissions")
		c.Abort()
	}
}

// OwnerOrAdmin checks that the user is either an Admin or the owner of the resource.
func OwnerOrAdmin(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		userRole := c.GetString("role")

		// Admins can access any resource
		if userRole == models.RoleAdmin {
			c.Next()
			return
		}

		// Parse the resource ID from the URL param (e.g. /students/:id)
		var resourceID uint
		if _, err := fmt.Sscanf(c.Param(paramName), "%d", &resourceID); err != nil {
			helpers.Error(c, http.StatusBadRequest, "invalid resource id")
			c.Abort()
			return
		}

		// Check ownership — the requesting user must match the resource ID
		if userID != resourceID {
			helpers.Error(c, http.StatusForbidden, "access denied: you can only access your own data")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ParentOwnsStudent checks that the student being requested belongs to this parent.
func ParentOwnsStudent() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		userRole := c.GetString("role")

		// Admins and Teachers bypass this check
		if userRole == models.RoleAdmin || userRole == models.RoleTeacher {
			c.Next()
			return
		}

		// Only apply to parents
		if userRole != models.RoleParent {
			c.Next()
			return
		}

		studentID := c.Param("studentID")

		// Convert to uint-safe comparison
		var count int64
		config.DB.Model(&models.Student{}).
			Where("id = ? AND parent_id = ?", studentID, userID).
			Count(&count)

		if count == 0 {
			helpers.Error(c, http.StatusForbidden,
				"access denied: this student is not linked to your account")
			c.Abort()
			return
		}

		c.Next()
	}
}
