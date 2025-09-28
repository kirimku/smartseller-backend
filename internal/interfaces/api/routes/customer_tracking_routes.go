package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handlers"
)

// SetupCustomerTrackingRoutes sets up routes for customer claim tracking
func SetupCustomerTrackingRoutes(router *gin.RouterGroup) {
	wsHandler := handlers.NewWebSocketHandler()
	trackingHandler := handlers.NewCustomerTrackingHandler(wsHandler)

	// Customer claim tracking routes
	tracking := router.Group("/claims/:id")
	{
		// Get current claim status and progress
		tracking.GET("/status", trackingHandler.GetClaimStatus)
		
		// Get claim updates and notifications
		tracking.GET("/updates", trackingHandler.GetClaimUpdates)
		
		// Send communication/message about claim
		tracking.POST("/communication", trackingHandler.SendCommunication)
		
		// Mark updates as read
		tracking.POST("/updates/mark-read", trackingHandler.MarkUpdatesAsRead)
	}
	
	// WebSocket endpoint for real-time updates
	router.GET("/ws", wsHandler.HandleWebSocket)
}