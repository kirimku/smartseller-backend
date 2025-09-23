package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// registerNotificationRoutes registers notification-related routes
func (r *Router) registerNotificationRoutes(api *gin.RouterGroup) {
	// Create dependencies
	notificationRepo := repository.NewNotificationRepositoryImpl(r.db)
	notificationService := service.NewNotificationService(notificationRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// Notifications routes (protected by auth middleware)
	notifications := api.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware())
	{
		// Get user's notifications
		notifications.GET("", notificationHandler.GetUserNotifications)

		// Mark a notification as read
		notifications.PUT("/:id/read", notificationHandler.MarkNotificationAsRead)

		// Mark all notifications as read
		notifications.PUT("/read-all", notificationHandler.MarkAllNotificationsAsRead)
	}
}
