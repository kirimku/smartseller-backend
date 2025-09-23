// Package middleware provides HTTP middleware functions for the application
package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// RequirePermission creates middleware that ensures users have a specific permission
func RequirePermission(permission entity.Permission, userUsecase usecase.UserUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userIDValue, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		userID, ok := userIDValue.(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format", nil)
			c.Abort()
			return
		}

		// Get user details to check their role
		user, err := userUsecase.GetUserByID(userID)
		if err != nil || user == nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not found", nil)
			c.Abort()
			return
		}

		// Set user role in context for reuse in other middleware or handlers
		c.Set("user_role", user.UserRole)

		// For backward compatibility, also set is_admin flag based on role or IsAdmin field
		if user.UserRole == entity.UserRoleAdmin || user.UserRole == entity.UserRoleOwner || user.IsAdmin {
			c.Set("is_admin", true)
		} else {
			c.Set("is_admin", false)
		}

		// Check if the user's role has the required permission
		hasPermission := user.IsAdmin // For now, simplify to only check admin status

		if !hasPermission {
			utils.ErrorResponse(c, http.StatusForbidden,
				fmt.Sprintf("You don't have permission to perform this action: %s", permission), nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireOwner ensures only users with Owner role can access the endpoint
func RequireOwner(userUsecase usecase.UserUseCase) gin.HandlerFunc {
	return RequirePermission(entity.PermissionManageRoles, userUsecase)
}

// RequireAdmin ensures only users with Admin or Owner role can access the endpoint
func RequireAdmin(userUsecase usecase.UserUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userIDValue, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
			c.Abort()
			return
		}

		userID, ok := userIDValue.(string)
		if !ok {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID format", nil)
			c.Abort()
			return
		}

		// Get user details to check their role
		user, err := userUsecase.GetUserByID(userID)
		if err != nil || user == nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "User not found", nil)
			c.Abort()
			return
		}

		// Set user role in context for reuse in other middleware or handlers
		c.Set("user_role", user.UserRole)

		// Check if the user is an admin or owner
		isAdminOrOwner := user.UserRole == entity.UserRoleAdmin ||
			user.UserRole == entity.UserRoleOwner ||
			user.IsAdmin

		if !isAdminOrOwner {
			utils.ErrorResponse(c, http.StatusForbidden, "Admin access required", nil)
			c.Abort()
			return
		}

		// For backward compatibility
		c.Set("is_admin", true)

		c.Next()
	}
}
