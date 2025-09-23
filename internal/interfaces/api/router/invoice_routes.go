package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/config"
	domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external/durianpay"

	"github.com/kirimku/smartseller-backend/internal/infrastructure/external/sicepat"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	infrastructureservice "github.com/kirimku/smartseller-backend/internal/infrastructure/service"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	restHandler "github.com/kirimku/smartseller-backend/internal/interfaces/rest/handler"
	"github.com/kirimku/smartseller-backend/pkg/logger"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
)

// registerInvoiceRoutes registers invoice-related routes
func (r *Router) registerInvoiceRoutes(api *gin.RouterGroup) {
	// Initialize repositories
	invoiceRepo := repository.NewInvoiceRepositoryImpl(r.db)
	transactionRepo := repository.NewTransactionRepositoryImpl(r.db)
	walletRepo := repository.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := repository.NewWalletTransactionRepositoryImpl(r.db)
	debtRepo := repository.NewDebtRepository(r.db)
	userRepo := repository.NewUserRepositoryImpl(r.db)
	walletDepositRepo := repository.NewWalletDepositRepositoryImpl(r.db)

	// Create transaction processor (needed for wallet service)
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

	// Initialize payment gateway using factory pattern
	var paymentGateway domainservice.PaymentGateway
	var err error

	paymentGatewayFactory := external.NewPaymentGatewayFactory(&config.AppConfig, userUseCase)

	// Validate gateway configuration first
	if err := paymentGatewayFactory.ValidateGatewayConfig(config.AppConfig.Payment.Gateway); err != nil {
		logger.Error("payment_gateway_config_invalid", "Payment gateway configuration validation failed", err)
		panic("Critical error: Payment gateway configuration is invalid")
	}

	// Create payment gateway instance
	paymentGateway, err = paymentGatewayFactory.CreatePaymentGateway()
	if err != nil {
		logger.Error("payment_gateway_init_failed", "Failed to initialize payment gateway", err)
		// Instead of using a non-functional payment gateway, panic to trigger app restart
		panic("Critical error: Failed to initialize payment gateway")
	}

	// Get existing transaction service
	transactionService, err := r.getTransactionService()
	if err != nil {
		logger.Error("transaction_service_init_failed", "Failed to initialize transaction service", err)
		panic("Critical error: Failed to initialize transaction service")
	}

	// Initialize courier clients (using existing config structure)
	jntClient, err := external.NewJNTClient(&config.AppConfig)
	if err != nil {
		logger.Info("jnt_client_init_failed", "Failed to initialize JNT client", err)
		jntClient = nil // Set to nil if initialization fails
	}

	jneClient, err := external.NewJNEClient(&config.AppConfig)
	if err != nil {
		logger.Info("jne_client_init_failed", "Failed to initialize JNE client", err)
		jneClient = nil // Set to nil if initialization fails
	}

	sicepatClient, err := external.NewSiCepatClient(&config.AppConfig)
	if err != nil {
		logger.Info("sicepat_client_init_failed", "Failed to initialize Sicepat client", err)
		sicepatClient = nil // Set to nil if initialization fails
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
	})
	if err != nil {
		logger.Error("Failed to initialize SAPX client: %v", err)
		panic(err)
	}

	// Initialize PackageCategoryService
	packageCategoryService := domainservice.NewPackageCategoryService()

	// Initialize ResiGenerator for SiCepat
	resiGenerator := infrastructureservice.NewSiCepatResiGenerator(transactionRepo, &config.AppConfig)

	// Initialize LogisticBookingService
	logisticBookingService := service.NewLogisticBookingService(
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

	// Initialize debt management service
	debtManagementService := service.NewDebtManagementService(
		debtRepo,
		walletRepo,
		walletTransactionRepo,
		&config.AppConfig,
	)

	// Initialize invoice service
	invoiceService := service.NewInvoiceService(
		invoiceRepo,
		transactionRepo,
		paymentGateway,
		&config.AppConfig,
		logisticBookingService, // Use our initialized logistic service
		nil,                    // TODO: Initialize BarcodeService
		r.shippingService,      // Use the shipping service from router
		debtManagementService,  // Add debt management service
	)

	// Initialize invoice handler
	invoiceHandler := handler.NewInvoiceHandler(
		invoiceService,
		transactionService,
	)

	// Initialize DurianPay WebhookClient for signature verification
	durianPayWebhookClient, err := durianpay.NewWebhookClient(&config.AppConfig.Payment.DurianPay)
	if err != nil {
		logger.Error("durianpay_webhook_client_init_failed", "Failed to initialize DurianPay WebhookClient", err)
		durianPayWebhookClient = nil // Set to nil if initialization fails, handler will log warnings
	}

	// Initialize dedicated DurianPay webhook handler with WebhookClient
	durianPayWebhookHandler := restHandler.NewDurianPayWebhookHandler(invoiceService, durianPayWebhookClient)

	// Create invoice routes group with auth middleware
	invoices := api.Group("/invoices")
	invoices.Use(middleware.AuthMiddleware())
	{
		invoices.GET("/:id", invoiceHandler.GetInvoice)
		invoices.DELETE("/:id", invoiceHandler.CancelInvoice)
	}

	// Add routes under transactions
	transactions := api.Group("/transactions")
	transactions.Use(middleware.AuthMiddleware())
	{
		transactions.POST("/:id/invoices", invoiceHandler.CreateInvoice)
		transactions.GET("/:id/invoice", invoiceHandler.GetInvoiceByTransaction)
	}

	// Webhook endpoints (no auth required)
	webhooks := api.Group("/webhooks/payments")
	{
		// Use generic handler for Xendit (keep existing functionality)
		webhooks.POST("/xendit", invoiceHandler.HandlePaymentCallback)
		// Use dedicated handler for DurianPay (new implementation)
		webhooks.POST("/durianpay", durianPayWebhookHandler.HandleDurianPayWebhook)
	}
}

// getTransactionService retrieves (or creates) the transaction service
func (r *Router) getTransactionService() (usecase.TransactionService, error) {
	// Initialize repositories
	transactionRepo := repository.NewTransactionRepositoryImpl(r.db)
	userRepo := repository.NewUserRepositoryImpl(r.db)
	cashbackRepo := repository.NewCashbackRepositoryImpl(r.db)
	costComponentRepo := repository.NewCostComponentRepositoryImpl(r.db)

	// Get courier service
	courierService := r.getCourierService()

	// Create shipping fee service using the courier service
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

	// Initialize wallet repositories
	walletRepo := repository.NewWalletRepositoryImpl(r.db)
	walletTransactionRepo := repository.NewWalletTransactionRepositoryImpl(r.db)

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

	// Create transaction service
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

	return transactionService, nil
}
