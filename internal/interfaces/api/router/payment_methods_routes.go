package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/logger"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// registerPaymentMethodsRoutes registers payment methods related routes
func (r *Router) registerPaymentMethodsRoutes(api *gin.RouterGroup) {
	// Initialize payment gateway using factory
	// For payment methods, we don't need user context, so pass nil
	paymentGatewayFactory := external.NewPaymentGatewayFactory(&config.AppConfig, nil)
	paymentGateway, err := paymentGatewayFactory.CreatePaymentGateway()
	if err != nil {
		logger.Error("payment_gateway_init_failed", "Failed to initialize payment gateway for payment methods", err)
		// Log error but don't panic, as this is a read-only endpoint
		logger.Warn("payment_methods_unavailable", "Payment methods endpoint will be unavailable due to payment gateway initialization failure", nil)
		return
	}

	// Initialize payment methods handler
	paymentMethodsHandler := handler.NewPaymentMethodsHandler(paymentGateway)

	// Create payment methods routes group with auth middleware (authentication required)
	paymentMethods := api.Group("/payment-methods")
	paymentMethods.Use(middleware.AuthMiddleware())
	{
		paymentMethods.GET("", paymentMethodsHandler.GetPaymentMethods)
	}
}
