package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/database"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/router"
	"github.com/kirimku/smartseller-backend/pkg/email"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

func main() {
	// Initialize logger
	logger.Init("INFO")
	logger.Info("application_start", "SmartSeller backend starting up", nil)

	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.Connect(config.AppConfig.Database.URL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("database_close_error", "Failed to close database connection", err, nil)
		}
	}()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	logger.Info("database_connected", "Successfully connected to database", nil)

	// Initialize services
	emailService := email.NewEmailService()

	// Initialize router with minimal services
	r := router.NewRouter(
		db,
		emailService,
	)

	// Setup routes
	engine := r.SetupRoutes()

	// Configure server
	port := config.AppConfig.Port
	if port == "" {
		port = "8090"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("server_starting", "SmartSeller backend server starting", map[string]interface{}{
			"port": port,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	logger.Info("server_started", "SmartSeller backend server started successfully", map[string]interface{}{
		"port": port,
	})

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("server_shutdown_start", "SmartSeller backend server shutting down", nil)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server_shutdown_error", "Server forced to shutdown", err, nil)
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("server_shutdown_complete", "SmartSeller backend server shutdown complete", nil)
}
