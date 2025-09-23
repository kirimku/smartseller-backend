package router

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/config"
	domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/cron"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external/sicepat"
	infraRepo "github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	restHandler "github.com/kirimku/smartseller-backend/internal/interfaces/rest/handler"
	"github.com/kirimku/smartseller-backend/pkg/cache"
	"github.com/kirimku/smartseller-backend/pkg/email"
	"github.com/kirimku/smartseller-backend/pkg/logger"
	"github.com/kirimku/smartseller-backend/pkg/metrics"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router handles HTTP routing using Gin framework
type Router struct {
	db                     *sqlx.DB
	router                 *gin.Engine
	memCache               cache.Cache             // Store the cache for reuse
	courierService         *service.CourierService // Store the courier service for reuse
	emailService           email.EmailSender       // Changed type to interface instead of concrete type
	invoiceService         domainservice.InvoiceService
	shippingService        domainservice.ShippingService
	logisticBookingService domainservice.LogisticBookingService
	trackingCron           *cron.TrackingCron // Add tracking cron
}

// NewRouter creates a new instance of Router
func NewRouter(
	db *sqlx.DB,
	emailService email.EmailSender,
	invoiceService domainservice.InvoiceService,
	shippingService domainservice.ShippingService,
	logisticBookingService domainservice.LogisticBookingService,
	trackingCron *cron.TrackingCron,
) *Router {
	router := gin.Default()

	// Add common middleware here
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Add Prometheus metrics middleware
	metricsCollector := metrics.GetGlobalMetricsCollector()
	router.Use(metrics.PrometheusMiddleware(metricsCollector))

	// Setup session middleware with configuration from config
	store := cookie.NewStore([]byte(config.AppConfig.SessionConfig.Key))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   config.AppConfig.SessionConfig.MaxAge,
		HttpOnly: true,
		Secure:   config.AppConfig.SessionConfig.Secure,
		SameSite: config.AppConfig.SessionConfig.SameSite,
		Domain:   config.AppConfig.SessionConfig.Domain,
	})
	router.Use(sessions.Sessions("kirimku_session", store))

	// Add security middlewares
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())

	// Initialize shared cache
	memCache := cache.NewInMemoryCache(5*time.Minute, 10*time.Minute)

	return &Router{
		db:                     db,
		router:                 router,
		memCache:               memCache,
		emailService:           emailService,
		invoiceService:         invoiceService,
		shippingService:        shippingService,
		logisticBookingService: logisticBookingService,
		trackingCron:           trackingCron,
	}
}

// Setup configures all routes and returns the engine
func (r *Router) Setup() *gin.Engine {
	// Health check endpoint for monitoring (without prefix)
	r.router.GET("/health", func(c *gin.Context) {
		// Simple health check response
		c.JSON(200, gin.H{
			"status":      "healthy",
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"service":     "kirimku-backend",
			"version":     config.AppConfig.Version,
			"environment": config.AppConfig.Environment,
		})
	})

	// Secure Prometheus metrics endpoint for monitoring
	// Add authentication middleware based on configuration
	metricsGroup := r.router.Group("/metrics")
	r.setupMetricsSecurity(metricsGroup)
	metricsGroup.GET("", gin.WrapH(promhttp.Handler()))

	// Health check endpoint for monitoring (with API prefix)
	r.router.GET("/api/v1/health", func(c *gin.Context) {
		// Simple health check response
		c.JSON(200, gin.H{
			"status":      "healthy",
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"service":     "kirimku-backend",
			"version":     config.AppConfig.Version,
			"environment": config.AppConfig.Environment,
		})
	})

	// Test endpoint for debugging tracking performance (NO AUTH)
	r.router.GET("/api/v1/test-tracking/:tracking_number", func(c *gin.Context) {
		trackingNumber := c.Param("tracking_number")
		logger.Info("test_tracking_start", fmt.Sprintf("Starting test tracking for %s", trackingNumber), nil)

		startTime := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		status, err := r.shippingService.GetTrackingStatus(ctx, trackingNumber)
		duration := time.Since(startTime)

		logger.Info("test_tracking_complete", fmt.Sprintf("Test tracking completed in %v", duration), map[string]interface{}{
			"tracking_number": trackingNumber,
			"duration_ms":     duration.Milliseconds(),
			"success":         err == nil,
		})

		if err != nil {
			c.JSON(500, gin.H{
				"error":           err.Error(),
				"tracking_number": trackingNumber,
				"duration_ms":     duration.Milliseconds(),
			})
			return
		}

		c.JSON(200, gin.H{
			"tracking_number": trackingNumber,
			"duration_ms":     duration.Milliseconds(),
			"status":          status,
		})
	})

	// API routes with versioning
	apiV1 := r.router.Group("/api/v1")

	// Register routes
	r.registerAuthRoutes(apiV1)
	r.registerUserRoutes(apiV1)
	r.registerAddressRoutes(apiV1)
	r.registerCourierRoutes(apiV1)
	r.registerPackageRoutes(apiV1)
	r.registerCashbackRoutes(apiV1)
	r.registerWalletRoutes(apiV1)
	r.registerNotificationRoutes(apiV1)
	r.registerAdminRoutes(apiV1)
	r.registerTransactionRoutes(apiV1)
	r.registerInvoiceRoutes(apiV1)
	r.registerPaymentMethodsRoutes(apiV1)     // Add payment methods routes
	r.registerRoleRoutes(apiV1)               // Add role management routes
	r.registerTrackingRoutes(apiV1)           // Add tracking routes
	r.registerDebtRoutes(apiV1)               // Add debt management routes
	r.registerRefundRequestRoutes(apiV1)      // Add refund request routes
	r.registerAdminRefundRequestRoutes(apiV1) // Add admin refund request routes

	return r.router
}

// Run starts the HTTP server
func (r *Router) Run(addr string) error {
	return r.router.Run(addr)
}

// registerAuthRoutes registers authentication routes
func (r *Router) registerAuthRoutes(api *gin.RouterGroup) {
	// Create dependencies
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)

	// Create wallet repositories
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)
	walletDepositRepo := infraRepo.NewWalletDepositRepositoryImpl(r.db)

	// Create transaction processor (needed for wallet service)
	transactionProcessor := service.NewTransactionProcessorImpl(r.db, walletRepo, walletTransactionRepo)

	// Create minimal wallet service just for wallet creation
	walletService := service.NewWalletServiceImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
		walletDepositRepo,
		transactionProcessor,
		nil, // Invoice service not needed for wallet creation
	)

	// Pass email service and wallet service to user use case - using proper dependency injection
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService, walletService)

	authHandler := handler.NewAuthHandler(userUseCase)

	// Auth routes
	auth := api.Group("/auth")
	auth.Use(middleware.APISecurityMiddleware()) // Add CSP-friendly headers for auth endpoints
	{
		// Google OAuth routes
		google := auth.Group("/google")
		{
			google.GET("/login", authHandler.LoginHandler)
			google.POST("/callback", authHandler.GoogleCallback)
		}

		// Form-based auth routes
		auth.POST("/register", authHandler.RegisterHandler)
		auth.POST("/login", authHandler.LoginWithCredentialsHandler)
		auth.POST("/forgot-password", authHandler.ForgotPasswordHandler)
		auth.POST("/reset-password", authHandler.ResetPasswordHandler)

		// Other auth routes
		auth.POST("/refresh", authHandler.RefreshTokenHandler)
		auth.POST("/logout", authHandler.LogoutHandler)
	}
}

// registerUserRoutes registers user-related routes
func (r *Router) registerUserRoutes(api *gin.RouterGroup) {
	// Create dependencies
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)

	// Create wallet repositories and services (needed for user profile with wallet balance)
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)
	walletDepositRepo := infraRepo.NewWalletDepositRepositoryImpl(r.db)
	transactionProcessor := service.NewTransactionProcessorImpl(r.db, walletRepo, walletTransactionRepo)

	walletService := service.NewWalletServiceImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
		walletDepositRepo,
		transactionProcessor,
		nil, // Invoice service not needed for user profile
	)

	// Create user use case with wallet service dependency
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService, walletService)

	// Create user handler
	userHandler := handler.NewUserHandler(userUseCase)

	// Setup user routes
	SetupUserRoutes(api, userHandler)
}

// registerAddressRoutes registers address routes
func (r *Router) registerAddressRoutes(api *gin.RouterGroup) {
	// Create dependencies
	addressRepo := infraRepo.NewAddressRepositoryImpl(r.db)
	addressService := service.NewAddressService(addressRepo)
	addressHandler := handler.NewAddressHandler(addressService)

	// Create address routes group
	addresses := api.Group("/addresses")
	addresses.Use(middleware.AuthMiddleware()) // Auth middleware
	{
		addresses.POST("", addressHandler.CreateAddress)
		addresses.GET("", addressHandler.ListAddresses)
		addresses.GET("/:id", addressHandler.GetAddress)
		addresses.PUT("/:id", addressHandler.UpdateAddress)
		addresses.DELETE("/:id", addressHandler.DeleteAddress)
		addresses.POST("/:id/main", addressHandler.SetMainAddress)
	}

	// Location data endpoints with auth middleware
	locationData := api.Group("")
	locationData.Use(middleware.AuthMiddleware()) // Add auth middleware
	{
		locationData.GET("/districts", addressHandler.GetDistricts)
		locationData.GET("/provinces", addressHandler.GetProvinces)
		locationData.GET("/cities", addressHandler.GetCities)
		locationData.GET("/postcodes", addressHandler.GetPostcodes)
		locationData.POST("/search-location", addressHandler.SearchLocation)
	}
}

// registerCourierRoutes registers courier-related routes
func (r *Router) registerCourierRoutes(api *gin.RouterGroup) {
	// Create external service clients with proper error handling
	jneClient, err := external.NewJNEClient(&config.AppConfig)
	if err != nil {
		logger.Error("Failed to initialize JNE client", "error", err)
		// Instead of using a non-functional client, panic to trigger app restart

		// TODO: bring back this when jne integration finish
		// panic(fmt.Sprintf("Critical error: Failed to initialize JNE client: %v", err))
	}

	sicepatClient, err := external.NewSiCepatClient(&config.AppConfig)
	if err != nil {
		logger.Error("Failed to initialize SiCepat client", "error", err)
		// Instead of using a non-functional client, panic to trigger app restart
		// TODO: bring back this when jne integration finish
		// panic(fmt.Sprintf("Critical error: Failed to initialize SiCepat client: %v", err))
	}

	jntClient, err := external.NewJNTClient(&config.AppConfig)
	if err != nil {
		logger.Error("Failed to initialize JNT client", "error", err)
		// Instead of using a non-functional client, panic to trigger app restart
		panic(fmt.Sprintf("Critical error: Failed to initialize JNT client: %v", err))
	}

	sapxClient, err := external.NewSAPXClient(&external.SAPXClientConfig{
		APIURL:             config.AppConfig.SAPXConfig.APIURL,
		APITrackerURL:      config.AppConfig.SAPXConfig.APITrackerURL,
		APIKeyPickup:       config.AppConfig.SAPXConfig.APIKeyPickup,
		APIKeyDropoff:      config.AppConfig.SAPXConfig.APIKeyDropoff,
		CustomerCodeNonCOD: config.AppConfig.SAPXConfig.CustomerCodeNonCOD,
		CustomerCodeCOD:    config.AppConfig.SAPXConfig.CustomerCodeCOD,
		MappingCode:        config.AppConfig.SAPXMapping,
	})
	if err != nil {
		logger.Error("Failed to initialize SAPX client", "error", err)
		// SAPX is optional, so we don't panic but log the error
		sapxClient = nil
	}

	// Create courier cache using the correct package path
	memCache := r.memCache

	// Create courier repository for admin settings integration
	courierRepo := infraRepo.NewCourierRepositoryImpl(r.db)

	// Create insurance service for courier insurance calculations
	insuranceService := domainservice.NewInsuranceCalculationService()

	// Create cashback dependencies for courier service
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)
	cashbackRepo := infraRepo.NewCashbackRepositoryImpl(r.db)
	cashbackUseCase := usecase.NewCashbackUseCase(userRepo, cashbackRepo)
	cashbackService := service.NewCashbackService(cashbackUseCase)
	cashbackServiceMapper := domainservice.NewCashbackServiceMapper()

	// Create the courier service with the correct interface types
	// The compiler error suggests the service expects interfaces, not concrete types
	courierService := service.NewCourierService(
		jneClient,     // This is already the correct type (*external.JNEClient)
		sicepatClient, // This is already the correct type (*external.SiCepatClient)
		jntClient,     // This is already the correct type (*external.JNTClient)
		sapxClient,    // SAPX client (can be nil if disabled)
		courierRepo,   // Courier repository for admin settings
		memCache,
		insuranceService,      // Insurance service for calculating delivery insurance
		cashbackService,       // Cashback service for calculating cashback info
		cashbackServiceMapper, // Service mapper for courier+service to cashback service mapping
	)

	// Create the courier handler with the real service implementation
	courierHandler := handler.NewCourierHandler(courierService)

	// Create courier routes group with auth middleware
	couriers := api.Group("")
	couriers.Use(middleware.AuthMiddleware()) // Auth middleware
	{
		// Register courier endpoints that require authentication
		couriers.GET("/couriers", courierHandler.GetCouriers)
	}
}

// registerPackageRoutes registers package-related routes
func (r *Router) registerPackageRoutes(api *gin.RouterGroup) {
	// Create dependencies
	packageCategoryService := domainservice.NewPackageCategoryService()
	packageHandler := handler.NewPackageHandler(packageCategoryService)

	// Create package routes group with auth middleware
	packages := api.Group("")
	packages.Use(middleware.AuthMiddleware()) // Auth middleware
	{
		// Register package endpoints that require authentication
		packages.GET("/package-categories", packageHandler.GetPackageCategories)
	}
}

// registerWalletRoutes registers wallet-related routes
func (r *Router) registerWalletRoutes(api *gin.RouterGroup) {
	// Create repositories
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)
	walletDepositRepo := infraRepo.NewWalletDepositRepositoryImpl(r.db)

	// Create domain services
	transactionProcessor := service.NewTransactionProcessorImpl(r.db, walletRepo, walletTransactionRepo)
	// Use the passed invoice service instead of creating a new one
	invoiceService := r.invoiceService

	// Create wallet service using the domain services
	walletService := service.NewWalletServiceImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
		walletDepositRepo,
		transactionProcessor,
		invoiceService,
	)

	// Create wallet use case
	walletUseCase := usecase.NewWalletUseCase(walletService)

	// Create handlers
	walletHandler := handler.NewWalletHandler(walletUseCase)

	// Setup routes
	SetupWalletRoutes(api, walletHandler)
}

// Helper methods for dependency reuse

// registerTrackingRoutes registers tracking-related routes
func (r *Router) registerTrackingRoutes(api *gin.RouterGroup) {
	// Initialize tracking repositories - assuming they are already in main.go
	trackingRepo := infraRepo.NewTrackingRepository(r.db)
	discrepancyRepo := infraRepo.NewDiscrepancyRepository(r.db)
	debtRepo := infraRepo.NewDebtRepository(r.db)
	transactionRepo := infraRepo.NewTransactionRepositoryImpl(r.db)
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)
	notificationRepo := infraRepo.NewNotificationRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)
	walletDepositRepo := infraRepo.NewWalletDepositRepositoryImpl(r.db)

	// Initialize additional repositories needed for transaction service
	cashbackRepo := infraRepo.NewCashbackRepositoryImpl(r.db)
	costComponentRepo := infraRepo.NewCostComponentRepositoryImpl(r.db)

	// Initialize wallet service dependencies
	transactionProcessor := service.NewTransactionProcessorImpl(r.db, walletRepo, walletTransactionRepo)

	// Initialize debt management service
	debtManagementService := service.NewDebtManagementService(
		debtRepo,
		walletRepo,
		walletTransactionRepo,
		&config.AppConfig,
	)

	// Initialize minimal wallet service for discrepancy service
	walletService := service.NewWalletServiceImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
		walletDepositRepo,
		transactionProcessor,
		nil, // invoiceService not needed for this context
	)

	// Initialize courier clients
	jntClient, err := external.NewJNTClient(&config.AppConfig)
	if err != nil {
		logger.Error("Failed to initialize JNT client", "client", "jnt", "error", err)
		jntClient = nil
	}

	jneClient, err := external.NewJNEClient(&config.AppConfig)
	if err != nil {
		logger.Error("Failed to initialize JNE client", "client", "jne", "error", err)
		jneClient = nil
	}

	sicepatClient, err := external.NewSiCepatClient(&config.AppConfig)
	if err != nil {
		logger.Error("Failed to initialize SiCepat client", "client", "sicepat", "error", err)
		sicepatClient = nil
	}

	// Initialize SiCepat tracking service from unified client
	var sicepatTrackingService *sicepat.TrackingServiceImpl
	if sicepatClient != nil {
		sicepatTrackingService = sicepatClient.GetTrackingService()
	}

	sapxClient, err := external.NewSAPXClient(&external.SAPXClientConfig{
		APIURL:             config.AppConfig.SAPXConfig.APIURL,
		APITrackerURL:      config.AppConfig.SAPXConfig.APITrackerURL,
		APIKeyPickup:       config.AppConfig.SAPXConfig.APIKeyPickup,
		APIKeyDropoff:      config.AppConfig.SAPXConfig.APIKeyDropoff,
		CustomerCodeNonCOD: config.AppConfig.SAPXConfig.CustomerCodeNonCOD,
		CustomerCodeCOD:    config.AppConfig.SAPXConfig.CustomerCodeCOD,
		MappingCode:        config.AppConfig.SAPXMapping,
	})
	if err != nil {
		logger.Error("Failed to initialize SAPX client", "client", "sapx", "error", err)
		sapxClient = nil
	}

	// Create courier service for shipping fee service
	courierRepo2 := infraRepo.NewCourierRepositoryImpl(r.db) // Create another instance for shipping fee service

	// Create insurance service for courier service
	insuranceService2 := domainservice.NewInsuranceCalculationService()

	// Create cashback dependencies for courier service
	userRepo2 := infraRepo.NewUserRepositoryImpl(r.db)
	cashbackRepo2 := infraRepo.NewCashbackRepositoryImpl(r.db)
	cashbackUseCase2 := usecase.NewCashbackUseCase(userRepo2, cashbackRepo2)
	cashbackService2 := service.NewCashbackService(cashbackUseCase2)
	cashbackServiceMapper2 := domainservice.NewCashbackServiceMapper()

	courierService := service.NewCourierService(
		jneClient,
		sicepatClient,
		jntClient,
		sapxClient,
		courierRepo2,
		r.memCache,
		insuranceService2,
		cashbackService2,       // Cashback service for calculating cashback info
		cashbackServiceMapper2, // Service mapper for courier+service to cashback service mapping
	)

	// Create shipping fee service
	shippingFeeService := service.NewShippingFeeService(
		courierService,
		r.memCache,
	)

	// Create insurance calculation service
	insuranceService := domainservice.NewInsuranceCalculationService()

	// Create location validation service
	locationService := domainservice.NewLocationValidationService()

	// Create cashback usecase
	cashbackUsecase := usecase.NewCashbackUseCase(
		userRepo,
		cashbackRepo,
	)

	// Create transaction service with all required dependencies
	transactionService := usecase.NewTransactionService(
		transactionRepo,
		userRepo,
		cashbackRepo,
		cashbackUsecase,
		shippingFeeService,
		insuranceService,
		costComponentRepo,
		locationService,
		transactionProcessor,
		debtManagementService,
	)

	// Initialize discrepancy service (dependency for shipping service)
	discrepancyService := service.NewDiscrepancyService(
		transactionRepo,
		trackingRepo,
		discrepancyRepo,
		debtRepo,
		walletRepo,
		userRepo,
		walletService,
		&config.AppConfig,
	)

	// Initialize shipping service
	shippingService := service.NewShippingService(
		trackingRepo,
		discrepancyRepo,
		debtRepo,
		transactionRepo,
		notificationRepo,
		discrepancyService,
		transactionService, // Pass proper transaction service
		&config.AppConfig,
		jntClient,
		jneClient,
		sicepatClient,
		sicepatTrackingService, // Pass the SiCepat tracking service
		sapxClient,
	)

	// Create webhook handler (in rest package)
	webhookHandler := restHandler.NewTrackingWebhookHandler(shippingService)

	// Create API handler (in rest package)
	apiHandler := restHandler.NewTrackingAPIHandler(shippingService, discrepancyService)

	// Register webhook routes (no auth required)
	webhooks := api.Group("/webhooks/tracking")
	{
		webhooks.POST("/jne", webhookHandler.HandleJNEWebhook)
		webhooks.POST("/sicepat", webhookHandler.HandleSiCepatWebhook)
		webhooks.POST("/ninjavan", webhookHandler.HandleNinjaVanWebhook)
	}

	// Register authenticated API routes
	tracking := api.Group("/tracking")
	tracking.Use(middleware.AuthMiddleware())
	{
		tracking.GET("/:tracking_number", apiHandler.GetTrackingStatus)
		tracking.GET("/:tracking_number/history", apiHandler.GetTrackingHistory)
		tracking.GET("/:tracking_number/url", apiHandler.GetTrackingURL)
		tracking.POST("/:tracking_number/refresh", apiHandler.RefreshTracking)
		tracking.POST("/bulk-refresh", apiHandler.BulkRefreshTracking)
	}

	discrepancies := api.Group("/discrepancies")
	discrepancies.Use(middleware.AuthMiddleware())
	{
		discrepancies.GET("/user/:user_id", apiHandler.GetUserDiscrepancies)
		discrepancies.GET("/transaction/:transaction_id", apiHandler.DetectDiscrepancies)
		discrepancies.POST("/:discrepancy_id/process", apiHandler.ProcessDiscrepancy)
	}
}

// registerDebtRoutes registers debt management routes
func (r *Router) registerDebtRoutes(api *gin.RouterGroup) {
	// Initialize repositories
	debtRepo := infraRepo.NewDebtRepository(r.db)
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)

	// Initialize debt management service
	debtManagementService := service.NewDebtManagementService(
		debtRepo,
		walletRepo,
		walletTransactionRepo,
		&config.AppConfig,
	)

	// Create debt handler
	debtHandler := restHandler.NewDebtHandler(debtManagementService)

	// Register debt routes
	debt := api.Group("/debt")
	debt.Use(middleware.AuthMiddleware())
	{
		// User debt information
		debt.GET("/user/:user_id", debtHandler.GetUserDebt)
		debt.GET("/user/:user_id/summary", debtHandler.GetDebtSummary)
		debt.GET("/user/:user_id/history", debtHandler.GetDebtHistory)
		debt.GET("/user/:user_id/check", debtHandler.CheckDebtBeforeTransaction)

		// Debt settlement
		debt.POST("/user/:user_id/settle", debtHandler.SettleDebt)
		debt.POST("/user/:user_id/auto-settle", debtHandler.AutoSettleFromWallet)

		// Debt entry management
		debt.PUT("/entry/:debt_entry_id/status", debtHandler.UpdateDebtStatus)
		debt.POST("/entry/:debt_entry_id/dispute", debtHandler.DisputeDebt)
	}
}

// registerRefundRequestRoutes registers refund request routes
func (r *Router) registerRefundRequestRoutes(api *gin.RouterGroup) {
	// Create repositories
	refundRequestRepo := infraRepo.NewRefundRequestRepositoryImpl(r.db)
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)
	transactionRepo := infraRepo.NewTransactionRepositoryImpl(r.db)
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)

	// Create transaction processor for wallet operations
	transactionProcessor := service.NewTransactionProcessorImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
	)

	// Create user use case for getting user emails
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService, nil)

	// Create service
	refundRequestService := service.NewRefundRequestService(
		r.db,
		refundRequestRepo,
		walletRepo,
		transactionRepo,
		userRepo,
		transactionProcessor,
		walletTransactionRepo,
	)

	// Create handlers
	refundRequestHandler := handler.NewRefundRequestHandler(refundRequestService, userUseCase)

	// User refund request routes
	refundRequests := api.Group("/refund-requests")
	refundRequests.Use(middleware.AuthMiddleware()) // Require authentication
	{
		refundRequests.POST("", refundRequestHandler.CreateRefundRequest)
		refundRequests.GET("", refundRequestHandler.GetUserRefundRequests)
		refundRequests.GET("/:id", refundRequestHandler.GetRefundRequestByID)
		refundRequests.PUT("/:id/cancel", refundRequestHandler.CancelRefundRequest)
		refundRequests.GET("/stats", refundRequestHandler.GetUserRefundStats)
	}
}

// registerAdminRefundRequestRoutes registers admin refund request routes
func (r *Router) registerAdminRefundRequestRoutes(api *gin.RouterGroup) {
	// Create service dependencies
	transactionRepo := infraRepo.NewTransactionRepositoryImpl(r.db)
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)
	refundRequestRepo := infraRepo.NewRefundRequestRepositoryImpl(r.db)
	walletRepo := infraRepo.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := infraRepo.NewWalletTransactionRepositoryImpl(r.db)

	// Create transaction processor for wallet operations
	transactionProcessor := service.NewTransactionProcessorImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
	)

	// Create user use case for getting user emails
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService, nil)

	refundRequestService := service.NewRefundRequestService(
		r.db,
		refundRequestRepo,
		walletRepo,
		transactionRepo,
		userRepo,
		transactionProcessor,
		walletTransactionRepo,
	)

	// Create handlers
	adminRefundRequestHandler := handler.NewAdminRefundRequestHandler(refundRequestService, userUseCase)

	// Admin refund request routes
	adminRefundRequests := api.Group("/admin/refund-requests")
	adminRefundRequests.Use(middleware.AuthMiddleware())      // Require authentication
	adminRefundRequests.Use(middleware.AdminMiddleware(r.db)) // Require admin access
	{
		adminRefundRequests.GET("", adminRefundRequestHandler.GetAllRefundRequests)
		adminRefundRequests.GET("/:id", adminRefundRequestHandler.GetRefundRequestByID)
		adminRefundRequests.PUT("/:id/approve", adminRefundRequestHandler.ApproveRefundRequest)
		adminRefundRequests.PUT("/:id/reject", adminRefundRequestHandler.RejectRefundRequest)
		adminRefundRequests.PUT("/:id/complete", adminRefundRequestHandler.CompleteRefundRequest)
		adminRefundRequests.PUT("/:id/retry-deduct", adminRefundRequestHandler.RetryDeductWallet)
		adminRefundRequests.GET("/stats", adminRefundRequestHandler.GetRefundStats)
		adminRefundRequests.GET("/export", adminRefundRequestHandler.ExportRefundRequests)
	}
}

// setupMetricsSecurity configures security for the metrics endpoint
func (r *Router) setupMetricsSecurity(metricsGroup *gin.RouterGroup) {
	// Option 1: IP Allowlisting (recommended for Grafana Cloud)
	allowedIPs := getEnvStringSlice("METRICS_ALLOWED_IPS", []string{})
	if len(allowedIPs) > 0 {
		logger.Info("Configuring IP allowlist for metrics endpoint", "allowed_ips", allowedIPs)
		metricsGroup.Use(middleware.MetricsAuthMiddleware(allowedIPs))
		return
	}

	// Option 2: Basic Authentication
	username := getEnvWithDefault("METRICS_USERNAME", "")
	password := getEnvWithDefault("METRICS_PASSWORD", "")
	if username != "" && password != "" {
		logger.Info("Configuring basic auth for metrics endpoint")
		metricsGroup.Use(middleware.BasicAuthMiddleware(username, password))
		return
	}

	// Option 3: Disable metrics endpoint if no security configured
	if !getEnvAsBool("METRICS_ALLOW_PUBLIC", false) {
		logger.Warn("Metrics endpoint disabled - no security configuration found")
		metricsGroup.Use(func(c *gin.Context) {
			c.JSON(403, gin.H{"error": "Metrics endpoint disabled for security"})
			c.Abort()
		})
		return
	}

	logger.Warn("Metrics endpoint is publicly accessible - configure METRICS_ALLOWED_IPS or METRICS_USERNAME/METRICS_PASSWORD for security")
}

// Helper function to get environment variable as string slice
func getEnvStringSlice(key string, defaultVal []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return strings.Split(value, ",")
}

// Helper function to get environment variable with default
func getEnvWithDefault(key, defaultVal string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value
}

// Helper function to get environment variable as bool
func getEnvAsBool(key string, defaultVal bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultVal
	}
	return value == "true" || value == "1" || value == "yes"
}
