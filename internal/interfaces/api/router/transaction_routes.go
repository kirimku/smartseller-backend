package router

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/config"
	domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external/sicepat"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	infrastructureservice "github.com/kirimku/smartseller-backend/internal/infrastructure/service"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/pkg/logger"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// registerTransactionRoutes registers transaction-related routes
func (r *Router) registerTransactionRoutes(api *gin.RouterGroup) {
	// Initialize repositories
	transactionRepo := repository.NewTransactionRepositoryImpl(r.db)
	userRepo := repository.NewUserRepositoryImpl(r.db)
	cashbackRepo := repository.NewCashbackRepositoryImpl(r.db)
	costComponentRepo := repository.NewCostComponentRepositoryImpl(r.db)
	invoiceRepo := repository.NewInvoiceRepositoryImpl(r.db)

	// Initialize wallet repositories (need these early for user usecase)
	walletRepo := repository.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := repository.NewWalletTransactionRepositoryImpl(r.db)
	walletDepositRepo := repository.NewWalletDepositRepositoryImpl(r.db)

	// Initialize debt repository and debt management service
	debtRepo := repository.NewDebtRepository(r.db)
	debtManagementService := service.NewDebtManagementService(
		debtRepo,
		walletRepo,
		walletTransactionRepo,
		&config.AppConfig,
	)

	// Initialize transaction processor
	transactionProcessor := service.NewTransactionProcessorImpl(r.db, walletRepo, walletTransactionRepo)

	// Initialize wallet service for user usecase
	walletService := service.NewWalletServiceImpl(
		r.db,
		walletRepo,
		walletTransactionRepo,
		walletDepositRepo,
		transactionProcessor,
		nil, // Invoice service not needed for this context
	)

	// Initialize user usecase (needed for payment gateway)
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService, walletService)

	// Initialize payment gateway using factory
	paymentGatewayFactory := external.NewPaymentGatewayFactory(&config.AppConfig, userUseCase)
	paymentGateway, err := paymentGatewayFactory.CreatePaymentGateway()
	if err != nil {
		logger.Error("payment_gateway_init_failed", "Failed to initialize payment gateway", err)
		// Instead of using a non-functional payment gateway, panic to trigger app restart
		panic("Critical error: Failed to initialize payment gateway")
	}

	// Initialize courier clients for logistic service
	jneClient, err := external.NewJNEClient(&config.AppConfig)
	if err != nil {
		logger.Error("jne_client_init_failed", "Failed to initialize JNE client", err)
		// TODO: remove this when integration compeleted
		// panic(fmt.Sprintf("Critical error: Failed to initialize JNE client: %v", err))
	}

	sicepatClient, err := external.NewSiCepatClient(&config.AppConfig)
	if err != nil {
		logger.Error("sicepat_client_init_failed", "Failed to initialize SiCepat client", err)
		// TODO: remove this when integration compeleted
		// panic(fmt.Sprintf("Critical error: Failed to initialize SiCepat client: %v", err))
	}

	jntClient, err := external.NewJNTClient(&config.AppConfig)
	if err != nil {
		logger.Error("jnt_client_init_failed", "Failed to initialize JNT client", err)
		panic(fmt.Sprintf("Critical error: Failed to initialize JNT client: %v", err))
	}

	// Initialize SiCepat booking service following DI pattern
	var sicepatBookingService sicepat.BookingService
	if config.AppConfig.SiCepatConfig.APIKey != "" {
		sicepatBookingService = sicepat.NewBookingServiceWithMapping(
			config.AppConfig.SiCepatConfig.PickupURL,
			config.AppConfig.SiCepatConfig.APIKey,
			true,                            // useReceipt
			30*time.Second,                  // timeout
			config.AppConfig.SiCepatMappingOrigin,      // origin mapping
			config.AppConfig.SiCepatMappingDestination, // destination mapping
		)
	}

	// Initialize SAPX client
	sapxClient, err := external.NewSAPXClient(&external.SAPXClientConfig{
		APIURL:             config.AppConfig.SAPXConfig.APIURL,
		APITrackerURL:      config.AppConfig.SAPXConfig.APITrackerURL,
		APIKeyPickup:       config.AppConfig.SAPXConfig.APIKeyPickup,
		APIKeyDropoff:      config.AppConfig.SAPXConfig.APIKeyDropoff,
		CustomerCodeCOD:    config.AppConfig.SAPXConfig.CustomerCodeCOD,
		CustomerCodeNonCOD: config.AppConfig.SAPXConfig.CustomerCodeNonCOD,
		Timeout:            config.AppConfig.SAPXConfig.Timeout,
		MaxRetries:         config.AppConfig.SAPXConfig.MaxRetries,
		MappingCode:        config.AppConfig.SAPXMapping,
	})
	if err != nil {
		logger.Error("Failed to initialize SAPX client", "error", err)
		sapxClient = nil
	}

	// Initialize logistic booking service
	// Initialize PackageCategoryService
	packageCategoryService := domainservice.NewPackageCategoryService()

	// Initialize ResiGenerator for SiCepat
	resiGenerator := infrastructureservice.NewSiCepatResiGenerator(transactionRepo, &config.AppConfig)

	logisticService := service.NewLogisticBookingService(
		transactionRepo,
		userRepo,
		&config.AppConfig,
		packageCategoryService,
		jntClient,
		jneClient,
		sicepatClient,
		sicepatBookingService,
		sapxClient,
		resiGenerator,
	)

	// Initialize invoice service
	invoiceService := service.NewInvoiceService(
		invoiceRepo,
		transactionRepo,
		paymentGateway,
		&config.AppConfig,
		logisticService,       // Initialize LogisticBookingService
		nil,                   // TODO: Initialize BarcodeService
		r.shippingService,     // Use the shipping service from router
		debtManagementService, // Add debt management service
	)

	// Reuse the existing courier service from courier routes
	// This is already initialized in registerCourierRoutes
	courierService := r.getCourierService()

	// Create shipping fee service using the courier service
	shippingFeeService := service.NewShippingFeeService(
		courierService,
		r.memCache, // Assuming memCache is accessible as a field in Router
	)

	// Create insurance calculation service
	insuranceService := domainservice.NewInsuranceCalculationService()

	// Create location validation service
	locationService := domainservice.NewLocationValidationService()

	// Create cashback usecase
	cashbackUsecase := usecase.NewCashbackUseCase(
		userRepo,     // First parameter should be UserRepository
		cashbackRepo, // Second parameter should be CashbackRepository
	)

	// Create transaction usecase with required dependencies
	transactionUsecase := usecase.NewTransactionService(
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

	// Initialize wallet payment processor
	walletTransactionRefRepo := repository.NewWalletTransactionReferenceRepositoryImpl(r.db)
	stdLogger := log.New(os.Stdout, "wallet: ", log.LstdFlags)

	walletDeductionService := service.NewWalletDeductionService(
		r.db,
		walletRepo,
		walletTransactionRepo,
		walletTransactionRefRepo,
		transactionRepo,
		transactionProcessor,
	)
	walletRefundService := service.NewWalletRefundService(
		walletRepo,
		walletTransactionRepo,
		walletTransactionRefRepo,
		transactionRepo,
		r.db,
		stdLogger,
	)
	walletPaymentProcessor := service.NewWalletPaymentProcessor(
		walletRepo,
		walletTransactionRepo,
		walletTransactionRefRepo,
		transactionRepo,
		walletDeductionService,
		walletRefundService,
	)

	courierRoutingCodeMapper := domainservice.NewCourierRoutingCodeMapper(&config.AppConfig)

	// Initialize transaction handler
	transactionHandler := handler.NewTransactionHandler(transactionUsecase, invoiceService, walletPaymentProcessor, courierRoutingCodeMapper)

	// Initialize analytics use case and handler
	analyticsUseCase := usecase.NewTransactionAnalyticsUseCase(transactionRepo)
	analyticsHandler := handler.NewTransactionAnalyticsHandler(analyticsUseCase)

	// Create transaction routes group with auth middleware
	transactions := api.Group("/transactions")
	transactions.Use(middleware.AuthMiddleware())
	{
		transactions.POST("", transactionHandler.CreateTransaction)
		transactions.GET("", transactionHandler.GetTransactions)
		transactions.GET("/:id", transactionHandler.GetTransaction)
		// transactions.PUT("/:id/state", transactionHandler.UpdateTransactionState)
		// transactions.DELETE("/:id", transactionHandler.CancelTransaction)

		// Analytics and reporting endpoints
		transactions.GET("/summary", analyticsHandler.GetTransactionSummary)
		transactions.GET("/analytics", analyticsHandler.GetTransactionAnalytics)

		// Export endpoints
		export := transactions.Group("/export")
		{
			export.GET("/csv", analyticsHandler.ExportTransactionsCSV)
			export.GET("/xlsx", analyticsHandler.ExportTransactionsXLSX)
		}
	}
}

// getCourierService returns the existing courier service or creates it if needed
// This avoids duplicating courier service initialization code
func (r *Router) getCourierService() *service.CourierService {
	// If we have a cached courier service, return it
	if r.courierService != nil {
		return r.courierService
	}

	// Otherwise, initialize it the same way as in registerCourierRoutes
	// Create external service clients with proper error handling
	jneClient, err := external.NewJNEClient(&config.AppConfig)
	if err != nil {
		logger.Error("jne_client_init_failed", "Failed to initialize JNE client", err)
		// TODO: remove this when integration compeleted
		// panic(fmt.Sprintf("Critical error: Failed to initialize JNE client: %v", err))
	}

	sicepatClient, err := external.NewSiCepatClient(&config.AppConfig)
	if err != nil {
		logger.Error("sicepat_client_init_failed", "Failed to initialize SiCepat client", err)
		// TODO: remove this when integration compeleted
		// panic(fmt.Sprintf("Critical error: Failed to initialize SiCepat client: %v", err))
	}

	jntClient, err := external.NewJNTClient(&config.AppConfig)
	if err != nil {
		logger.Error("jnt_client_init_failed", "Failed to initialize JNT client", err)
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
		logger.Error("sapx_client_init_failed", "Failed to initialize SAPX client", err)
		// SAPX is optional, so we don't panic but log the error
		sapxClient = nil
	}

	// Create courier service
	courierRepo := repository.NewCourierRepositoryImpl(r.db)

	// Create insurance service for courier service
	insuranceService := domainservice.NewInsuranceCalculationService()

	// Create cashback dependencies for courier service
	userRepo := repository.NewUserRepositoryImpl(r.db)
	cashbackRepo := repository.NewCashbackRepositoryImpl(r.db)
	cashbackUseCase := usecase.NewCashbackUseCase(userRepo, cashbackRepo)
	cashbackService := service.NewCashbackService(cashbackUseCase)
	cashbackServiceMapper := domainservice.NewCashbackServiceMapper()

	courierService := service.NewCourierService(
		jneClient,
		sicepatClient,
		jntClient,
		sapxClient,
		courierRepo,
		r.memCache,
		insuranceService,
		cashbackService,       // Cashback service for calculating cashback info
		cashbackServiceMapper, // Service mapper for courier+service to cashback service mapping
	)

	// Store for future use
	r.courierService = courierService

	return courierService
}
