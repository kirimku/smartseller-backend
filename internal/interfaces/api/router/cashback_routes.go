package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// registerCashbackRoutes registers cashback-related routes
func (r *Router) registerCashbackRoutes(api *gin.RouterGroup) {
	// Create dependencies
	userRepo := repository.NewUserRepositoryImpl(r.db)
	cashbackRepo := repository.NewCashbackRepositoryImpl(r.db)
	cashbackUseCase := usecase.NewCashbackUseCase(userRepo, cashbackRepo)
	cashbackService := service.NewCashbackService(cashbackUseCase)
	cashbackHandler := handler.NewCashbackHandler(cashbackService)

	// Secure cashback routes with auth middleware
	cashbackRoutes := api.Group("/cashback")
	cashbackRoutes.Use(middleware.AuthMiddleware())
	{
		// Get user tier info
		cashbackRoutes.GET("/tier", cashbackHandler.GetUserTierInfo)

		// Get user cashback history
		cashbackRoutes.GET("/history", cashbackHandler.GetUserCashbackHistory)

		// Calculate potential cashback for a transaction
		cashbackRoutes.POST("/calculate", cashbackHandler.CalculateCashbackPreview)
	}

	// Admin cashback routes (admin only)
	adminCashback := api.Group("/admin/cashback")
	adminCashback.Use(middleware.AuthMiddleware())
	adminCashback.Use(middleware.AdminMiddleware(r.db)) // Add admin authorization
	{
		// Get all cashback rates for all services and tiers (admin only)
		adminCashback.GET("/rates", cashbackHandler.GetAllCashbackRates)
	}
}
