package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// SetupUserRoutes sets up user-related routes
func SetupUserRoutes(router *gin.RouterGroup, userHandler *handler.UserHandler) {
	userRouter := router.Group("/users")
	{
		// Get current user profile - requires authentication
		userRouter.GET("/me", middleware.AuthRequired(), userHandler.GetUserProfile)
	}
}
