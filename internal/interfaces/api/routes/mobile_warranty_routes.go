package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handlers"
)

// MobileWarrantyRoutes sets up mobile warranty routes for mobile app integration
func MobileWarrantyRoutes(router *gin.RouterGroup) {
	handler := handlers.NewMobileWarrantyHandler()

	// Mobile-specific routes under /mobile prefix
	mobile := router.Group("/mobile")
	{
		// Mobile warranty scanning routes
		warranties := mobile.Group("/warranties")
		{
			// QR/Barcode scanning endpoint
			warranties.POST("/scan", handler.ScanWarranty)
			
			// Camera permissions check
			warranties.GET("/camera-permissions", handler.CheckCameraPermissions)
		}

		// Mobile claim management routes
		claims := mobile.Group("/claims")
		{
			// Photo upload for claims with compression
			claims.POST("/:claimId/photo-upload", handler.UploadClaimPhoto)
			
			// Offline data synchronization
			claims.GET("/offline-sync", handler.GetOfflineSync)
		}

		// Mobile notification routes
		notifications := mobile.Group("/notifications")
		{
			// Push notification registration
			notifications.POST("/register", handler.RegisterPushNotification)
		}

		// Mobile-optimized response format
		optimize := mobile.Group("/optimize")
		{
			// Get mobile-optimized response for any endpoint
			optimize.GET("/:endpoint", handler.GetMobileOptimizedResponse)
		}
	}
}