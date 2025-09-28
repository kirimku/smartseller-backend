package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
)

// PublicWarrantyRoutes sets up public warranty routes
func PublicWarrantyRoutes(router *gin.RouterGroup) {
	publicWarrantyHandler := handler.NewPublicWarrantyHandler()

	// Public warranty endpoints - no authentication required
	publicWarranty := router.Group("/warranty")
	{
		// POST endpoints for warranty operations
		publicWarranty.POST("/validate", publicWarrantyHandler.ValidateWarranty)
		publicWarranty.POST("/lookup", publicWarrantyHandler.LookupWarranty)
		publicWarranty.POST("/product-info", publicWarrantyHandler.GetProductInfo)
		publicWarranty.POST("/check-coverage", publicWarrantyHandler.CheckCoverage)

		// GET endpoints for direct barcode access
		publicWarranty.GET("/:barcode", publicWarrantyHandler.GetWarrantyByBarcode)
		publicWarranty.GET("/:barcode/product", publicWarrantyHandler.GetProductByBarcode)
	}
}