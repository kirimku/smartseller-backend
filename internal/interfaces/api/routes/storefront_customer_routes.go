package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/middleware"
	"github.com/kirimku/smartseller-backend/internal/interfaces/http/handlers"
)

// SetupStorefrontCustomerRoutes sets up routes for storefront-specific customer operations
func SetupStorefrontCustomerRoutes(
	router *gin.Engine,
	tenantMiddleware *middleware.TenantMiddleware,
	customerAuthMiddleware *middleware.CustomerAuthMiddleware,
	customerAuthHandler *handlers.CustomerAuthHandler,
	addressHandler *handler.AddressHandler,
) {
	// Storefront-specific customer routes with tenant resolution
	api := router.Group("/api/v1")
	{
		storefront := api.Group("/storefront/:slug")
		storefront.Use(tenantMiddleware.ResolveTenant())
		{
			// Public authentication endpoints (no auth required)
			auth := storefront.Group("/auth")
			{
				auth.POST("/register", customerAuthHandler.RegisterCustomer)
				auth.POST("/login", customerAuthHandler.LoginCustomer)
				auth.POST("/forgot-password", customerAuthHandler.RequestPasswordReset)
				auth.POST("/reset-password", customerAuthHandler.ConfirmPasswordReset)
				auth.POST("/verify-email", customerAuthHandler.VerifyEmail)
				auth.POST("/resend-verification", customerAuthHandler.ResendVerificationEmail)
				auth.POST("/validate-reset-token", customerAuthHandler.ValidateResetToken)
			}
			
			// Public endpoints with optional authentication
			public := storefront.Group("")
			public.Use(customerAuthMiddleware.OptionalCustomerAuth())
			{
				public.POST("/auth/refresh", customerAuthHandler.RefreshToken)
				public.POST("/auth/logout", customerAuthHandler.LogoutCustomer)
			}
		}
		
		// Protected customer endpoints (require authentication)
		protected := storefront.Group("")
		protected.Use(customerAuthMiddleware.CustomerAuthRequired())
		{
			// Profile management - using customer auth handler for basic profile operations
			profile := protected.Group("/profile")
			{
				profile.POST("/change-password", customerAuthHandler.ChangePassword)
				// TODO: Implement profile management endpoints
				// profile.GET("", customerHandler.GetProfile)
				// profile.PUT("", customerHandler.UpdateProfile)
				// profile.POST("/upload-avatar", customerHandler.UploadAvatar)
			}
			
			// Address management - using dedicated address handler
			addresses := protected.Group("/addresses")
			{
				addresses.GET("/:id", addressHandler.GetAddress)
				addresses.PUT("/:id", addressHandler.UpdateAddress)
				addresses.DELETE("/:id", addressHandler.DeleteAddress)
				// TODO: Implement customer-specific address endpoints
				// addresses.GET("", customerHandler.GetCustomerAddresses)
				// addresses.POST("", customerHandler.CreateCustomerAddress)
				// addresses.POST("/:id/default", customerHandler.SetDefaultAddress)
			}
			
			// Address validation and utilities
			addressUtils := protected.Group("/address-utils")
			{
				addressUtils.POST("/validate", addressHandler.ValidateAddress)
				addressUtils.POST("/geocode", addressHandler.GeocodeAddress)
				addressUtils.GET("/nearby", addressHandler.GetNearbyAddresses)
			}
		}
		
		// Optional authentication endpoints (for guest users)
		optional := storefront.Group("")
		optional.Use(customerAuthMiddleware.OptionalCustomerAuth())
		{
			// TODO: Implement product catalog endpoints
			// products := optional.Group("/products")
			// {
			//     products.GET("", productHandler.GetStorefrontProducts)
			//     products.GET("/:id", productHandler.GetStorefrontProduct)
			// }
			
			// TODO: Implement category endpoints
			// categories := optional.Group("/categories")
			// {
			//     categories.GET("", categoryHandler.GetStorefrontCategories)
			//     categories.GET("/:id", categoryHandler.GetStorefrontCategory)
			// }
			
			// TODO: Implement search endpoints
			// search := optional.Group("/search")
			// {
			//     search.GET("/products", searchHandler.SearchStorefrontProducts)
			// }
		}
		
		// TODO: Implement shopping cart endpoints
		// cart := protected.Group("/cart")
		// {
		//     cart.GET("", cartHandler.GetCart)
		//     cart.POST("/items", cartHandler.AddItem)
		//     cart.PUT("/items/:id", cartHandler.UpdateItem)
		//     cart.DELETE("/items/:id", cartHandler.RemoveItem)
		// }
		
		// TODO: Implement order endpoints
		// orders := protected.Group("/orders")
		// {
		//     orders.GET("", orderHandler.GetCustomerOrders)
		//     orders.POST("", orderHandler.CreateOrder)
		//     orders.GET("/:id", orderHandler.GetOrder)
		//     orders.GET("/:id/status", orderHandler.GetOrderStatus)
		// }
		
		// TODO: Implement checkout endpoints
		// checkout := protected.Group("/checkout")
		// {
		//     checkout.POST("", checkoutHandler.CreateOrder)
		//     checkout.POST("/payment", checkoutHandler.ProcessPayment)
		//     checkout.GET("/confirmation/:id", checkoutHandler.GetOrderConfirmation)
		// }
	}
}