package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handlers"
)

// CustomerWarrantyRoutes sets up customer warranty routes
func CustomerWarrantyRoutes(router *gin.RouterGroup) {
	handler := handlers.NewCustomerWarrantyHandler()

	// Customer warranty registration and management routes
	warranties := router.Group("/warranties")
	{
		// Register a new warranty
		warranties.POST("/register", handler.RegisterWarranty)
		
		// List customer warranties with filtering and pagination
		warranties.GET("", handler.GetWarranties)
		
		// Get detailed warranty information
		warranties.GET("/:id", handler.GetWarrantyDetails)
		
		// Update warranty information (customer details, preferences)
		warranties.PUT("/:id", handler.UpdateWarranty)
	}
}