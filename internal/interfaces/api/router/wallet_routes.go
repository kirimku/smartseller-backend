package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// SetupWalletRoutes sets up wallet-related routes
func SetupWalletRoutes(router *gin.RouterGroup, walletHandler *handler.WalletHandler) {
	walletRouter := router.Group("/wallets")
	{
		// Public endpoint - protected at handler level
		walletRouter.POST("", middleware.AuthRequired(), walletHandler.CreateWallet)

		// Get wallet for logged-in user
		walletRouter.GET("/me", middleware.AuthRequired(), walletHandler.GetUserWallet)

		// Get wallet by ID - protected at handler level
		walletRouter.GET("/:id", middleware.AuthRequired(), walletHandler.GetWallet)

		// Get wallet balance - protected at handler level
		walletRouter.GET("/:id/balance", middleware.AuthRequired(), walletHandler.GetWalletBalance)

		// Get transaction history - protected at handler level
		walletRouter.GET("/:id/transactions", middleware.AuthRequired(), walletHandler.GetTransactionHistory)

		// Create deposit request - protected at handler level
		walletRouter.POST("/deposits", middleware.AuthRequired(), walletHandler.CreateDeposit)

		// Process payment from wallet - protected at handler level
		walletRouter.POST("/payments", middleware.AuthRequired(), walletHandler.ProcessPayment)

		// Update wallet status (admin only) - protected at handler level
		walletRouter.PUT("/:id/status", middleware.AuthRequired(), walletHandler.UpdateWalletStatus)
	}
}
