package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handlers"
)

// CustomerClaimRoutes sets up customer warranty claim routes
func CustomerClaimRoutes(router *gin.RouterGroup) {
	handler := handlers.NewCustomerClaimHandler()

	// Customer claim routes group
	claimsGroup := router.Group("/claims")
	{
		// Claim submission and management
		claimsGroup.POST("/submit", handler.SubmitClaim)
		claimsGroup.GET("", handler.ListClaims)
		claimsGroup.GET("/:id", handler.GetClaimDetails)
		claimsGroup.PUT("/:id", handler.UpdateClaim)
		
		// Claim attachments
		claimsGroup.POST("/attachments/upload", handler.UploadAttachment)
		claimsGroup.GET("/:id/attachments", handler.GetClaimAttachments)
		
		// Claim feedback
		claimsGroup.POST("/feedback", handler.SubmitFeedback)
		
		// Claim timeline
		claimsGroup.GET("/:id/timeline", handler.GetClaimTimeline)
	}
}