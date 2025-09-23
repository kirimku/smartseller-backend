package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	infraRepo "github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// registerRoleRoutes registers role management routes
func (r *Router) registerRoleRoutes(api *gin.RouterGroup) {
	// Create dependencies
	userRepo := repository.NewUserRepositoryImpl(r.db)

	// Reuse wallet service components
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)
	walletDepositRepo := infraRepo.NewWalletDepositRepositoryImpl(r.db)
	transactionProcessor := service.NewTransactionProcessorImpl(r.db, walletRepo, walletTransactionRepo)

	// Create wallet service
	walletService := service.NewWalletServiceImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
		walletDepositRepo,
		transactionProcessor,
		nil, // Invoice service not needed for this context
	)

	// Create user use case with all dependencies
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService, walletService)

	// Create role handler
	roleHandler := handler.NewUserRoleHandler(userUseCase)

	// Create role routes group
	roles := api.Group("/roles")
	roles.Use(middleware.AuthMiddleware()) // Add authentication requirement
	{
		// Get available roles (authentication required)
		roles.GET("", roleHandler.GetAvailableRolesHandler)

		// Routes requiring admin permissions
		adminRoles := roles.Group("")
		adminRoles.Use(middleware.RequireAdmin(userUseCase))
		{
			// Get users by role
			adminRoles.GET("/users", roleHandler.GetUsersByRoleHandler)
		}

		// Routes requiring owner permissions
		ownerRoles := roles.Group("")
		ownerRoles.Use(middleware.AuthMiddleware())
		ownerRoles.Use(middleware.RequirePermission(entity.PermissionManageRoles, userUseCase))
		{
			// Update a user's role
			ownerRoles.PUT("/users", roleHandler.UpdateUserRoleHandler)
		}
	}
}
