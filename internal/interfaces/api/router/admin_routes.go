package router

import (
	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external/jnt"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// registerAdminRoutes registers admin dashboard routes
func (r *Router) registerAdminRoutes(api *gin.RouterGroup) { // Create dependencies
	userRepo := repository.NewUserRepositoryImpl(r.db)
	cashbackRepo := repository.NewCashbackRepositoryImpl(r.db)
	courierRepo := repository.NewCourierRepositoryImpl(r.db)
	transactionRepo := repository.NewTransactionRepositoryImpl(r.db)
	invoiceRepo := repository.NewInvoiceRepositoryImpl(r.db)
	trackingRepo := repository.NewTrackingRepository(r.db)
	walletRepo := repository.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := repository.NewWalletTransactionRepositoryImpl(r.db)

	// Create transaction processor service
	transactionProcessor := service.NewTransactionProcessorImpl(r.db, walletRepo, walletTransactionRepo)

	// Create PostPaymentProcessor for admin retry booking functionality
	// Reuse existing services from router (logisticBookingService, barcodeService, shippingService)
	postPaymentProcessor := service.NewPostPaymentProcessor(
		transactionRepo,
		r.logisticBookingService, // Use logistic booking service from router
		nil,                      // barcodeService - can be nil for retry functionality
		r.shippingService,        // Use shipping service from router
	)

	// Create usecases
	cashbackUseCase := usecase.NewCashbackUseCase(userRepo, cashbackRepo)
	adminUserUseCase := usecase.NewAdminUserUseCase(userRepo)
	adminCourierUseCase := usecase.NewAdminCourierUseCase(courierRepo)
	adminTransactionUseCase := usecase.NewAdminTransactionUseCase(transactionRepo, userRepo, invoiceRepo, trackingRepo, cashbackRepo, transactionProcessor, postPaymentProcessor)

	// Create services
	userService := service.NewUserService(userRepo)
	cashbackService := service.NewCashbackService(cashbackUseCase)

	// Create JNT client for logistics cancellation (only JNT is supported initially)
	jntClient, err := external.NewJNTClient(&config.AppConfig)
	if err != nil {
		// Log error but continue - cancellation service will handle missing client gracefully
		// logger.Error("Failed to initialize JNT client for cancellation", err)
		jntClient = nil
	}

	// Create logistics cancellation service
	logisticsCancellationService := service.NewLogisticsCancellationService(jntClient)

	// Create transaction cancellation service
	transactionCancellationService := service.NewTransactionCancellationService(
		transactionRepo,
		trackingRepo,
		logisticsCancellationService,
		r.db,
	)

	// Create JNT tracking service dependencies
	var jntTrackingServiceClient *jnt.TrackingServiceImpl
	jntClient2, err := external.NewJNTClient(&config.AppConfig)
	if err != nil {
		// Log error but continue - JNT tracking service will handle missing client gracefully
		jntTrackingServiceClient = nil
	} else {
		jntTrackingServiceClient = jntClient2.GetTrackingService()
	}

	jntTrackingService := service.NewJNTTrackingCronService(
		trackingRepo,
		jntTrackingServiceClient,
		r.shippingService,
	)

	// Create handlers
	adminDashboardHandler := handler.NewAdminDashboardHandler(userService, cashbackService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserUseCase)
	adminCourierHandler := handler.NewAdminCourierHandler(adminCourierUseCase)
	adminTransactionHandler := handler.NewAdminTransactionHandler(adminTransactionUseCase, transactionCancellationService)
	adminJNTTrackingHandler := handler.NewAdminJNTTrackingHandler(jntTrackingService)
	adminUnifiedTrackingHandler := handler.NewAdminUnifiedTrackingHandler(r.trackingCron)
	adminTrackingHandler := handler.NewAdminTrackingHandler(r.shippingService)

	// Admin routes (protected by auth and admin middleware)
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware(r.db)) // Pass database connection
	{
		// Dashboard routes
		dashboard := admin.Group("/dashboard")
		{
			// Tier distribution
			dashboard.GET("/tier-distribution", adminDashboardHandler.GetTierDistribution)

			// Cashback statistics
			dashboard.GET("/cashback-statistics", adminDashboardHandler.GetCashbackStatistics)

			// Top cashback users
			dashboard.GET("/top-cashback-users", adminDashboardHandler.GetTopCashbackUsers)
		}

		// User management routes
		admin.GET("/users", adminUserHandler.GetUsers)

		// Courier management routes
		admin.GET("/couriers", adminCourierHandler.GetCouriers)
		admin.GET("/couriers/enabled", adminCourierHandler.GetEnabledCourierCodes)
		admin.GET("/couriers/summary", adminCourierHandler.GetCourierStatusSummary)
		admin.GET("/couriers/:courier_code", adminCourierHandler.GetCourierByCode)
		admin.PUT("/couriers/:courier_code/status", adminCourierHandler.UpdateCourierStatus)

		// Transaction management routes
		admin.GET("/transactions", adminTransactionHandler.GetTransactions)
		admin.GET("/transactions/download/csv", adminTransactionHandler.DownloadTransactionsCSV)
		admin.GET("/transactions/:id", adminTransactionHandler.GetTransactionDetail)
		admin.PUT("/transactions/:id/state", adminTransactionHandler.UpdateTransactionState)
		admin.POST("/transactions/:id/notes", adminTransactionHandler.AddTransactionNote)
		admin.POST("/transactions/:id/refund", adminTransactionHandler.ProcessRefund)
		admin.POST("/transactions/:id/cancel", adminTransactionHandler.CancelTransaction)
		admin.POST("/transactions/:id/retry-cashback", adminTransactionHandler.RetryCashback)
		admin.POST("/transactions/:id/retry-booking", adminTransactionHandler.RetryBooking)

		// JNT tracking admin routes
		jntTracking := admin.Group("/jnt-tracking")
		{
			jntTracking.GET("/status", adminJNTTrackingHandler.GetCronStatus)
			jntTracking.POST("/trigger", adminJNTTrackingHandler.TriggerCron)
			jntTracking.GET("/runs", adminJNTTrackingHandler.GetExecutionRuns)
			jntTracking.GET("/metrics", adminJNTTrackingHandler.GetMetrics)
			jntTracking.GET("/config", adminJNTTrackingHandler.GetConfiguration)
			jntTracking.GET("/circuit-breaker", adminJNTTrackingHandler.GetCircuitBreakerStatus)
			jntTracking.POST("/circuit-breaker/reset", adminJNTTrackingHandler.ResetCircuitBreakers)
			jntTracking.GET("/error-analytics", adminJNTTrackingHandler.GetErrorAnalytics)
		}

		// Unified tracking admin routes
		trackingCron := admin.Group("/tracking-cron")
		{
			trackingCron.GET("/status", adminUnifiedTrackingHandler.GetCronStatus)
			trackingCron.GET("/health", adminUnifiedTrackingHandler.GetHealthCheck)
			trackingCron.GET("/metrics", adminUnifiedTrackingHandler.GetMetrics)
			trackingCron.GET("/couriers", adminUnifiedTrackingHandler.GetRegisteredCouriers)
			trackingCron.POST("/couriers/refresh", adminUnifiedTrackingHandler.RefreshCourierList)

			// Manual trigger endpoints
			trackingCron.POST("/trigger", adminUnifiedTrackingHandler.TriggerAllCouriers)
			trackingCron.POST("/trigger/jnt", adminUnifiedTrackingHandler.TriggerJNTOnly)
			trackingCron.POST("/trigger/sicepat", adminUnifiedTrackingHandler.TriggerSicepatOnly)
			trackingCron.POST("/trigger/custom", adminUnifiedTrackingHandler.TriggerWithTimeout)
		}

		// Admin tracking management routes
		tracking := admin.Group("/tracking")
		{
			tracking.GET("/:booking_code", adminTrackingHandler.GetTrackingStatus)
			tracking.GET("/:booking_code/history", adminTrackingHandler.GetTrackingHistory)
			tracking.GET("/:booking_code/url", adminTrackingHandler.GetTrackingURL)
			tracking.POST("/:booking_code/refresh", adminTrackingHandler.RefreshTracking)
			tracking.POST("/bulk-refresh", adminTrackingHandler.BulkRefreshTracking)
			tracking.GET("/transaction/:transaction_id", adminTrackingHandler.GetTransactionTracking)
		}
	}
}
