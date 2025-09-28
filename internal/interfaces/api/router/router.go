package router

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	infraRepo "github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
	"github.com/kirimku/smartseller-backend/internal/interfaces/http/handlers"
	customerMiddleware "github.com/kirimku/smartseller-backend/internal/interfaces/api/middleware"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/routes"

	"github.com/kirimku/smartseller-backend/pkg/email"
	"github.com/kirimku/smartseller-backend/pkg/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Router represents the main router structure
type Router struct {
	db           *sqlx.DB
	emailService email.EmailSender
}

// NewRouter creates a new router instance
func NewRouter(db *sqlx.DB, emailService email.EmailSender) *Router {
	return &Router{
		db:           db,
		emailService: emailService,
	}
}

// SetupRoutes configures all the routes for the application
func (r *Router) SetupRoutes() *gin.Engine {
	router := gin.New()

	// Add core middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Add CORS and Security middleware
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())

	// Session middleware
	store := cookie.NewStore([]byte(config.AppConfig.SessionConfig.Key))
	store.Options(sessions.Options{
		MaxAge:   config.AppConfig.SessionConfig.MaxAge,
		Secure:   config.AppConfig.SessionConfig.Secure,
		HttpOnly: true,
		Domain:   config.AppConfig.SessionConfig.Domain,
		SameSite: config.AppConfig.SessionConfig.SameSite,
	})
	router.Use(sessions.Sessions("smartseller-session", store))

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "smartseller-backend"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Setup API routes
	r.setupAPIRoutes(router)

	return router
}

// setupAPIRoutes configures the API routes
func (r *Router) setupAPIRoutes(router *gin.Engine) {
	// Create a default structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create existing repositories
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)
	productRepo := infraRepo.NewPostgreSQLProductRepository(r.db)
	productCategoryRepo := infraRepo.NewPostgreSQLProductCategoryRepository(r.db)
	productVariantRepo := infraRepo.NewPostgreSQLProductVariantRepository(r.db)
	productVariantOptionRepo := infraRepo.NewPostgreSQLProductVariantOptionRepository(r.db)
	productImageRepo := infraRepo.NewPostgreSQLProductImageRepository(r.db)

	// Initialize tenant infrastructure first
	tenantConfig := tenant.DefaultTenantConfig()
	tenantCache := tenant.NewInMemoryTenantCache(1000, 5*time.Minute)
	
	// Initialize repositories with proper parameters
	customerRepo := repository.NewPostgreSQLCustomerRepository(r.db, nil, &repository.NoOpMetricsCollector{})
	customerAddressRepo := repository.NewPostgreSQLCustomerAddressRepository(r.db, nil, &repository.NoOpMetricsCollector{})
	storefrontRepo := repository.NewPostgreSQLStorefrontRepository(r.db, nil, &repository.NoOpMetricsCollector{})
	
	// Initialize tenant resolver with all dependencies
	tenantResolver := tenant.NewTenantResolver(r.db.DB, tenantConfig, tenantCache, storefrontRepo)
	
	// Update repositories with tenant resolver
	customerRepo = repository.NewPostgreSQLCustomerRepository(r.db, tenantResolver, &repository.NoOpMetricsCollector{})
	customerAddressRepo = repository.NewPostgreSQLCustomerAddressRepository(r.db, tenantResolver, &repository.NoOpMetricsCollector{})
	storefrontRepo = repository.NewPostgreSQLStorefrontRepository(r.db, tenantResolver, &repository.NoOpMetricsCollector{})

	// Initialize services
	validationService := service.NewValidationServiceSimple(customerRepo, storefrontRepo, customerAddressRepo)
	customerService := service.NewCustomerServiceSimple(customerRepo, tenantResolver)
	
	// Type assert email service to MailgunService for customer services
	mailgunService, ok := r.emailService.(*email.MailgunService)
	if !ok {
		panic("Email service must be MailgunService for customer authentication features")
	}
	
	customerPasswordResetService := service.NewCustomerPasswordResetService(customerRepo, mailgunService)
	customerEmailVerificationService := service.NewCustomerEmailVerificationService(customerRepo, mailgunService)
	
	// Initialize customer auth handler
	customerAuthHandler := handlers.NewCustomerAuthHandler(
		customerService,
		customerPasswordResetService,
		customerEmailVerificationService,
		validationService,
	)

	// Initialize address service and handler
	addressService := service.NewCustomerAddressServiceSimple(customerAddressRepo, customerRepo, tenantResolver)
	addressHandler := handler.NewAddressHandler(addressService, validationService)

	// Initialize tenant and customer auth middleware
	tenantMiddleware := customerMiddleware.NewTenantMiddleware(tenantResolver, "localhost")
	customerAuthMiddleware := customerMiddleware.NewCustomerAuthMiddleware()

	// Create use cases (existing)
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService)
	productUseCase := usecase.NewProductUseCase(
		productRepo,
		productCategoryRepo,
		productVariantRepo,
		productVariantOptionRepo,
		productImageRepo,
		logger,
	)

	// Create handlers (existing)
	authHandler := handler.NewAuthHandler(userUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	productHandler := handler.NewProductHandler(productUseCase, logger)

	// TODO: Temporary warranty barcode handler (will be replaced with full implementation)
	warrantyBarcodeHandler := handler.NewWarrantyBarcodeHandler(logger)
	
	// Warranty claim handler
	warrantyClaimHandler := handler.NewWarrantyClaimHandler(logger)
	
	// Claim attachment and timeline handlers
	claimAttachmentHandler := handler.NewClaimAttachmentHandler()
	claimTimelineHandler := handler.NewClaimTimelineHandler()
	
	// Repair ticket handler
	repairTicketHandler := handler.NewRepairTicketHandler()
	
	// Batch generation handler
	batchGenerationHandler := handler.NewBatchGenerationHandler()

	// Setup storefront customer routes
	routes.SetupStorefrontCustomerRoutes(router, tenantMiddleware, customerAuthMiddleware, customerAuthHandler, addressHandler)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes
		auth := v1.Group("/auth")
		{
			auth.GET("/login", authHandler.LoginHandler)
			auth.GET("/google/callback", authHandler.GoogleCallback)
			auth.POST("/register", authHandler.RegisterHandler)
			auth.POST("/login", authHandler.LoginWithCredentialsHandler)
			auth.POST("/refresh", authHandler.RefreshTokenHandler)
			auth.POST("/forgot-password", authHandler.ForgotPasswordHandler)
			auth.POST("/reset-password", authHandler.ResetPasswordHandler)
			auth.POST("/logout", middleware.AuthMiddleware(), authHandler.LogoutHandler)
			
			// Secure token endpoints
			auth.POST("/set-secure-tokens", authHandler.SetSecureTokensHandler)
			auth.POST("/clear-secure-tokens", authHandler.ClearSecureTokensHandler)
			auth.GET("/secure-check", authHandler.SecureCheckHandler)
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			users.GET("/profile", userHandler.GetUserProfile)
			users.GET("/me", userHandler.GetUserProfile) // Alias for /profile to match OpenAPI docs
		}

		// TODO: Phase 4 Implementation - Customer Routes
		// Uncomment and complete customer routes once repositories are implemented
		/*
			customers := v1.Group("/customers")
			{
				customers.POST("/register", customerHandler.RegisterCustomer)

				protected := customers.Group("")
				protected.Use(middleware.AuthMiddleware())
				{
					protected.GET("/:id", customerHandler.GetCustomer)
					protected.GET("/by-email", customerHandler.GetCustomerByEmail)
					protected.PUT("/:id", customerHandler.UpdateCustomer)
					protected.POST("/:id/deactivate", customerHandler.DeactivateCustomer)
					protected.POST("/:id/reactivate", customerHandler.ReactivateCustomer)
					protected.GET("/search", customerHandler.SearchCustomers)
					protected.GET("/stats", customerHandler.GetCustomerStats)
					protected.POST("/:id/addresses", customerHandler.CreateCustomerAddress)
					protected.GET("/:id/addresses", customerHandler.GetCustomerAddresses)
					protected.POST("/:customer_id/addresses/:address_id/default", customerHandler.SetDefaultAddress)
					protected.GET("/:id/addresses/default", customerHandler.GetDefaultAddress)
				}
			}
		*/

		// TODO: Phase 4 Implementation - Storefront Routes
		// Uncomment and complete storefront routes once repositories are implemented
		/*
			storefronts := v1.Group("/storefronts")
			storefronts.Use(middleware.AuthMiddleware())
			{
				storefronts.POST("/", storefrontHandler.CreateStorefront)
				storefronts.GET("/:id", storefrontHandler.GetStorefront)
				storefronts.GET("/by-slug", storefrontHandler.GetStorefrontBySlug)
				storefronts.PUT("/:id", storefrontHandler.UpdateStorefront)
				storefronts.DELETE("/:id", storefrontHandler.DeleteStorefront)
				storefronts.POST("/:id/activate", storefrontHandler.ActivateStorefront)
				storefronts.POST("/:id/deactivate", storefrontHandler.DeactivateStorefront)
				storefronts.POST("/:id/suspend", storefrontHandler.SuspendStorefront)
				storefronts.PUT("/:id/settings", storefrontHandler.UpdateStorefrontSettings)
				storefronts.GET("/search", storefrontHandler.SearchStorefronts)
				storefronts.GET("/:id/stats", storefrontHandler.GetStorefrontStats)
				storefronts.POST("/validate-domain", storefrontHandler.ValidateCustomDomain)
			}
		*/

		// TODO: Phase 4 Implementation - Address Routes
		// Uncomment and complete address routes once repositories are implemented
		/*
			addresses := v1.Group("/addresses")
			addresses.Use(middleware.AuthMiddleware())
			{
				addresses.GET("/:id", addressHandler.GetAddress)
				addresses.PUT("/:id", addressHandler.UpdateAddress)
				addresses.DELETE("/:id", addressHandler.DeleteAddress)
				addresses.POST("/validate", addressHandler.ValidateAddress)
				addresses.POST("/geocode", addressHandler.GeocodeAddress)
				addresses.POST("/nearby", addressHandler.GetNearbyAddresses)
				addresses.POST("/bulk", addressHandler.BulkCreateAddresses)
				addresses.PUT("/bulk", addressHandler.BulkUpdateAddresses)
				addresses.DELETE("/bulk", addressHandler.BulkDeleteAddresses)
				addresses.POST("/stats", addressHandler.GetAddressStats)
				addresses.POST("/distribution", addressHandler.GetAddressDistribution)
			}
		*/

		// Product routes (protected)
		products := v1.Group("/products")
		products.Use(middleware.AuthMiddleware())
		{
			// CRUD operations
			products.POST("/", productHandler.CreateProduct)
			products.GET("/", productHandler.ListProducts)
			products.GET("/:id", productHandler.GetProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}

		// Admin Warranty routes (protected) - Phase 7 Implementation
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		{
			warranty := admin.Group("/warranty")
			{
				// Barcode management routes
				barcodes := warranty.Group("/barcodes")
				{
					// Barcode generation and management
					barcodes.POST("/generate", warrantyBarcodeHandler.GenerateBarcodes)
					barcodes.GET("/", warrantyBarcodeHandler.ListBarcodes)
					barcodes.GET("/:id", warrantyBarcodeHandler.GetBarcode)
					barcodes.POST("/:id/activate", warrantyBarcodeHandler.ActivateBarcode)
					barcodes.POST("/bulk-activate", warrantyBarcodeHandler.BulkActivateBarcodes)

					// Statistics and validation
					barcodes.GET("/stats", warrantyBarcodeHandler.GetBarcodeStats)
					barcodes.GET("/validate/:barcode_value", warrantyBarcodeHandler.ValidateBarcode)
				}

				// Batch generation routes
				batches := warranty.Group("/claims/:id/batches")
				{
					// Batch CRUD operations
					batches.POST("/", batchGenerationHandler.CreateBatch)
					batches.GET("/", batchGenerationHandler.ListBatches)
					batches.GET("/:batchId", batchGenerationHandler.GetBatch)
					batches.DELETE("/:batchId", batchGenerationHandler.DeleteBatch)

					// Batch management
					batches.GET("/:batchId/progress", batchGenerationHandler.GetBatchProgress)
					batches.POST("/:batchId/cancel", batchGenerationHandler.CancelBatch)

					// Batch analysis
					batches.GET("/:batchId/collisions", batchGenerationHandler.GetBatchCollisions)
					batches.GET("/:batchId/statistics", batchGenerationHandler.GetBatchStatistics)
				}

				// Claim management routes
				claims := warranty.Group("/claims")
				{
					// Claim listing and retrieval
					claims.GET("/", warrantyClaimHandler.ListClaims)
					claims.GET("/:id", warrantyClaimHandler.GetClaim)

					// Claim validation and processing
					claims.POST("/:id/validate", warrantyClaimHandler.ValidateClaim)
					claims.POST("/:id/reject", warrantyClaimHandler.RejectClaim)
					claims.POST("/:id/assign", warrantyClaimHandler.AssignTechnician)
					claims.POST("/:id/complete", warrantyClaimHandler.CompleteClaim)

					// Claim management
					claims.POST("/:id/notes", warrantyClaimHandler.AddClaimNotes)
					claims.POST("/bulk-status", warrantyClaimHandler.BulkUpdateClaimStatus)

					// Statistics
					claims.GET("/stats", warrantyClaimHandler.GetClaimStatistics)

					// Attachment management routes
					attachments := claims.Group("/:id/attachments")
					{
						attachments.GET("/", claimAttachmentHandler.ListAttachments)
						attachments.POST("/upload", claimAttachmentHandler.UploadAttachment)
						attachments.GET("/:attachment_id/download", claimAttachmentHandler.DownloadAttachment)
						attachments.DELETE("/:attachment_id", claimAttachmentHandler.DeleteAttachment)
						attachments.POST("/:attachment_id/approve", claimAttachmentHandler.ApproveAttachment)
					}

					// Timeline management routes
					timeline := claims.Group("/:id/timeline")
					{
						timeline.GET("/", claimTimelineHandler.GetClaimTimeline)
						timeline.POST("/", claimTimelineHandler.CreateTimelineEntry)
						timeline.GET("/:entry_id", claimTimelineHandler.GetTimelineEntry)
						timeline.PUT("/:entry_id", claimTimelineHandler.UpdateTimelineEntry)
						timeline.DELETE("/:entry_id", claimTimelineHandler.DeleteTimelineEntry)
					}

					// Repair ticket management routes
					repairTickets := claims.Group("/:id/repair-tickets")
					{
						repairTickets.POST("/", repairTicketHandler.CreateRepairTicket)
						repairTickets.GET("/", repairTicketHandler.ListRepairTickets)
						repairTickets.GET("/:ticketId", repairTicketHandler.GetRepairTicket)
						repairTickets.PUT("/:ticketId", repairTicketHandler.UpdateRepairTicket)
						repairTickets.PUT("/:ticketId/assign", repairTicketHandler.AssignTechnician)
						repairTickets.PUT("/:ticketId/complete", repairTicketHandler.CompleteRepair)
						repairTickets.PUT("/:ticketId/quality-check", repairTicketHandler.QualityCheck)
						repairTickets.GET("/statistics", repairTicketHandler.GetRepairStatistics)
					}
				}
			}
		}

		// Initialize customer authentication middleware for public and customer routes
		customerAuth := customerMiddleware.NewCustomerAuthMiddleware()

		// Public API routes (no authentication required) - Phase 8 Implementation
		public := v1.Group("/public")
		public.Use(customerAuth.RateLimitMiddleware(50)) // 50 requests per minute for public endpoints
		public.Use(customerAuth.SecurityHeadersMiddleware())
		public.Use(customerAuth.CORSMiddleware())
		{
			// Public warranty validation endpoints
			routes.PublicWarrantyRoutes(public)
		}

		// Customer API routes (authentication required for customers) - Phase 8 Implementation
		customer := v1.Group("/customer")
		
		// Apply rate limiting and security middleware to all customer endpoints
		customer.Use(customerAuth.RateLimitMiddleware(100)) // 100 requests per minute for customer endpoints
		customer.Use(customerAuth.SecurityHeadersMiddleware())
		customer.Use(customerAuth.CORSMiddleware())
		
		// Public customer endpoints (no authentication required)
		customerPublic := customer.Group("/public")
		{
			// Customer warranty registration and management endpoints (public access)
			routes.CustomerWarrantyRoutes(customerPublic)
		}
		
		// Protected customer endpoints (authentication required)
		customerProtected := customer.Group("/protected")
		customerProtected.Use(customerAuth.CustomerAuthRequired())
		{
			// Customer claim submission and management endpoints
			routes.CustomerClaimRoutes(customerProtected)
			
			// Customer claim tracking endpoints
			routes.SetupCustomerTrackingRoutes(customerProtected)
			
			// Mobile warranty endpoints for mobile app integration
			routes.MobileWarrantyRoutes(customerProtected)
		}
	}
}
