package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// AdminMiddleware creates a middleware for admin route protection
func AdminMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
			c.Abort()
			return
		}

		// Create user repository
		userRepo := repository.NewUserRepositoryImpl(db)

		// Get the user
		user, err := userRepo.GetUserByID(userID.(string))
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Error retrieving user", nil)
			c.Abort()
			return
		}

		// Check if user is admin
		if user.IsAdmin != true {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied. Admin privileges required.", nil)
			c.Abort()
			return
		}

		// User is admin, proceed
		c.Next()
	}
}
