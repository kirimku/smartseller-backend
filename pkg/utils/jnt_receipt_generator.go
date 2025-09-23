package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// GenerateJNTReceiptNumber generates a JNT AWB receipt number in format JB0000000001
// This uses the transaction ID and timestamp to ensure uniqueness
func GenerateJNTReceiptNumber(ctx context.Context, transactionRepository repository.TransactionRepository, transactionID int) (string, error) {
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

	// Format as JB followed by 10 digits with zero padding
	receiptNumber := fmt.Sprintf("JB%010d", sequence)

	logger.Info("jnt_receipt_number_generated",
		"Generated JNT AWB receipt number",
		map[string]interface{}{
			"transaction_id": transactionID,
			"receipt_number": receiptNumber,
			"sequence":       sequence,
		})

	return receiptNumber, nil
}

// ValidateJNTReceiptNumber validates that a receipt number follows JNT format
func ValidateJNTReceiptNumber(receiptNumber string) bool {
	if len(receiptNumber) != 12 {
		return false
	}

	if receiptNumber[:2] != "JB" {
		return false
	}

	// Check that the rest are digits
	for i := 2; i < 12; i++ {
		if receiptNumber[i] < '0' || receiptNumber[i] > '9' {
			return false
		}
	}

	return true
}

// GenerateBasicJNTReceipt generates a basic JNT receipt number from an ID
// This is a simpler version that doesn't require database access
func GenerateBasicJNTReceipt(id int64) string {
	// Ensure the ID is within valid range
	if id < 1 {
		id = 1
	}
	if id > 9999999999 {
		id = id % 9999999999
	}

	// Format as JB followed by 10 digits with zero padding
	return fmt.Sprintf("JB%010d", id)
}
