package utils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

var (
	// Global instance of the development receipt tracker
	devReceiptTracker *DevSiCepatReceiptTracker
)

// isDevelopmentMode checks if we're running in development mode
func isDevelopmentMode() bool {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "development" // default
	}
	return appEnv == "development" || appEnv == "local"
}

// GenerateSiCepatReceiptNumber generates a SiCepat AWB receipt number
// In development mode, uses the predefined range 888889340571-888889341570
// In production mode, uses the original format SC0000000001
func GenerateSiCepatReceiptNumber(ctx context.Context, transactionRepository repository.TransactionRepository, transactionID int) (string, error) {
	// Check if we're in development mode
	if isDevelopmentMode() {
		return generateDevelopmentReceipt(transactionID)
	}

	// Production mode - use original logic
	return generateProductionReceipt(transactionID)
}

// generateDevelopmentReceipt generates a receipt number from the development range
func generateDevelopmentReceipt(transactionID int) (string, error) {
	// Initialize tracker if not already done
	if devReceiptTracker == nil {
		devReceiptTracker = NewDevSiCepatReceiptTracker()
	}

	receiptNumber, err := devReceiptTracker.GetNextAvailableReceipt(transactionID)
	if err != nil {
		// Log usage stats for debugging
		stats := devReceiptTracker.GetUsageStats()
		logger.Error("dev_sicepat_receipt.generation_failed",
			"Failed to generate development receipt number",
			map[string]interface{}{
				"error":          err.Error(),
				"transaction_id": transactionID,
				"usage_stats":    stats,
			})
		return "", fmt.Errorf("failed to generate development SiCepat receipt number: %w", err)
	}

	return receiptNumber, nil
}

// generateProductionReceipt generates a receipt number using the original production logic
func generateProductionReceipt(transactionID int) (string, error) {
	// Use a combination of transaction ID and timestamp for uniqueness
	// This approach avoids the need for a separate sequence table

	// Get current timestamp in microseconds for uniqueness
	timestamp := time.Now().UnixMicro()

	// Create a unique sequence by combining transaction ID and timestamp
	// Use modulo to keep it within a reasonable range
	sequence := (int64(transactionID)*1000000 + timestamp) % 9999999999

	// Ensure minimum sequence number is 1
	if sequence < 1 {
		sequence = 1
	}

	// Format as SC followed by 10 digits with zero padding
	receiptNumber := fmt.Sprintf("SC%010d", sequence)

	logger.Info("sicepat_receipt_number_generated",
		"Generated SiCepat AWB receipt number (production)",
		map[string]interface{}{
			"transaction_id": transactionID,
			"receipt_number": receiptNumber,
			"sequence":       sequence,
			"mode":           "production",
		})

	return receiptNumber, nil
}

// GetDevelopmentReceiptStats returns statistics about development receipt usage
func GetDevelopmentReceiptStats() map[string]interface{} {
	if !isDevelopmentMode() {
		return map[string]interface{}{
			"mode":  "production",
			"stats": "Development receipt tracking not available in production mode",
		}
	}

	// Initialize tracker if not already done
	if devReceiptTracker == nil {
		devReceiptTracker = NewDevSiCepatReceiptTracker()
	}

	stats := devReceiptTracker.GetUsageStats()
	stats["mode"] = "development"

	return stats
}

// ResetDevelopmentReceipts clears all used development receipt numbers (for testing)
func ResetDevelopmentReceipts() error {
	if !isDevelopmentMode() {
		return fmt.Errorf("reset only available in development mode")
	}

	// Initialize tracker if not already done
	if devReceiptTracker == nil {
		devReceiptTracker = NewDevSiCepatReceiptTracker()
	}

	return devReceiptTracker.ResetUsedNumbers()
}

// ValidateSiCepatReceiptNumber validates that a receipt number follows SiCepat format
func ValidateSiCepatReceiptNumber(receiptNumber string) bool {
	if len(receiptNumber) != 12 {
		return false
	}

	if receiptNumber[:2] != "SC" {
		return false
	}

	// Check that the remaining 10 characters are digits
	for i := 2; i < 12; i++ {
		if receiptNumber[i] < '0' || receiptNumber[i] > '9' {
			return false
		}
	}

	return true
}
