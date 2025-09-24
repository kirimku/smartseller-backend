package router

import (
	"log/slog"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/config"
	infraRepo "github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
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

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

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

	// Create repositories
	userRepo := infraRepo.NewUserRepositoryImpl(r.db)
	productRepo := infraRepo.NewPostgreSQLProductRepository(r.db)
	productCategoryRepo := infraRepo.NewPostgreSQLProductCategoryRepository(r.db)
	productVariantRepo := infraRepo.NewPostgreSQLProductVariantRepository(r.db)
	productVariantOptionRepo := infraRepo.NewPostgreSQLProductVariantOptionRepository(r.db)
	productImageRepo := infraRepo.NewPostgreSQLProductImageRepository(r.db)

	// Create use cases
	userUseCase := usecase.NewUserUseCase(userRepo, r.emailService)
	productUseCase := usecase.NewProductUseCase(
		productRepo,
		productCategoryRepo,
		productVariantRepo,
		productVariantOptionRepo,
		productImageRepo,
		logger,
	)

	// Create handlers
	authHandler := handler.NewAuthHandler(userUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	productHandler := handler.NewProductHandler(productUseCase, logger)

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
		}

		// User routes (protected)
		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			users.GET("/profile", userHandler.GetUserProfile)
		}

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
	}
}
