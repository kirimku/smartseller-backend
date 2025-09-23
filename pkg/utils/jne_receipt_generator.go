package utils

import (
	"context"
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// JNE AWB Receipt prefixes based on service type
var AWBPrefix = map[string]string{
	"JNE REG":      "BLJC",
	"JNE YES":      "BLJC",
	"JNE Trucking": "BLJT",
}

// buildJNEServiceName constructs the JNE service name from courier and service type
func buildJNEServiceName(courier, serviceType string) string {
	// Only build service name if it's a JNE courier
	if courier != "jne" {
		return ""
	}

	// Map service types to JNE service names
	switch serviceType {
	case "REG", "CTC":
		return "JNE REG"
	case "YES", "CTCYES":
		return "JNE YES"
	case "JTR", "Trucking":
		return "JNE Trucking"
	default:
		return "JNE REG" // Default to REG service
	}
}

// GenerateJNEReceiptNumber generates a JNE AWB receipt number based on service type and transaction
// Similar to Ruby's JneService::GenerateAwbReceipt
func GenerateJNEReceiptNumber(ctx context.Context, transactionRepository repository.TransactionRepository, transaction *entity.Transaction) (string, error) {
	// Build service name from courier and courier service type
	serviceName := buildJNEServiceName(transaction.Courier, transaction.CourierServiceType)
	if serviceName == "" {
		serviceName = "JNE REG" // Default service
	}

	// Get AWB prefix based on service name
	prefix, exists := AWBPrefix[serviceName]
	if !exists {
		prefix = "BLJC" // Default prefix
	}

	var suffix string
	var err error

	// Generate suffix based on transaction type (similar to Ruby logic)
	if isBukasendTransaction(transaction) {
		if isBukasendOOMTransaction(transaction) {
			suffix, err = generateBukasendOOMSuffixAWBNumber(transaction)
		} else {
			suffix, err = generateBukasendSuffixAWBNumber(transaction)
		}
	} else {
		suffix = generateRegularSuffixAWBNumber(transaction)
	}

	if err != nil {
		logger.Error("generate_jne_receipt_suffix_failed",
			"Failed to generate JNE AWB receipt suffix",
			map[string]interface{}{
				"transaction_id": transaction.ID,
				"service_name":   serviceName,
				"error":          err.Error(),
			})
		return "", fmt.Errorf("failed to generate JNE receipt suffix: %w", err)
	}

	receiptNumber := prefix + suffix

	logger.Info("jne_receipt_number_generated",
		"Generated JNE AWB receipt number",
		map[string]interface{}{
			"transaction_id": transaction.ID,
			"service_name":   serviceName,
			"prefix":         prefix,
			"suffix":         suffix,
			"receipt_number": receiptNumber,
		})

	return receiptNumber, nil
}

// isBukasendTransaction checks if transaction is bukasend type
func isBukasendTransaction(transaction *entity.Transaction) bool {
	// This logic should match Ruby's @transaction.bukasend? method
	// For now, we'll check if transaction ID contains "BPT" pattern
	return strings.Contains(transaction.BookingCode, "BPT") ||
		strings.Contains(fmt.Sprintf("%d", transaction.ID), "BPT")
}

// isBukasendOOMTransaction checks if transaction is bukasend OOM type
func isBukasendOOMTransaction(transaction *entity.Transaction) bool {
	// This logic should match Ruby's @transaction.bukasend_oom? method
	// For now, return false as we don't have this field in our entity
	return false
}

// generateRegularSuffixAWBNumber generates suffix for regular transactions
// Ruby: @transaction.transaction_id.to_s.rjust(12, '0')
func generateRegularSuffixAWBNumber(transaction *entity.Transaction) string {
	transactionIDStr := fmt.Sprintf("%d", transaction.ID)
	// Right justify with zeros to 12 characters
	for len(transactionIDStr) < 12 {
		transactionIDStr = "0" + transactionIDStr
	}
	return transactionIDStr
}

// generateBukasendSuffixAWBNumber generates suffix for bukasend transactions
// Ruby: trx_id = @transaction.transaction_id.tr('BPT', ”)
//
//	suffix_awb_number = trx_id.to_i.to_s(36).upcase + 'BPT'
//	suffix_awb_number.rjust(11, '0')
func generateBukasendSuffixAWBNumber(transaction *entity.Transaction) (string, error) {
	transactionIDStr := fmt.Sprintf("%d", transaction.ID)

	// Remove 'BPT' characters (Ruby's tr('BPT', ''))
	cleanID := strings.ReplaceAll(transactionIDStr, "B", "")
	cleanID = strings.ReplaceAll(cleanID, "P", "")
	cleanID = strings.ReplaceAll(cleanID, "T", "")

	// Convert to integer then to base36
	transactionID, err := strconv.Atoi(cleanID)
	if err != nil {
		return "", fmt.Errorf("failed to convert transaction ID to integer: %w", err)
	}

	base36 := strings.ToUpper(strconv.FormatInt(int64(transactionID), 36))
	suffixAWBNumber := base36 + "BPT"

	// Right justify with zeros to 11 characters
	for len(suffixAWBNumber) < 11 {
		suffixAWBNumber = "0" + suffixAWBNumber
	}

	return suffixAWBNumber, nil
}

// generateBukasendOOMSuffixAWBNumber generates suffix for bukasend OOM transactions
// Ruby: trx_id = @transaction.transaction_id.tr('BPT', ”)
//
//	hash = Digest::MD5.hexdigest(trx_id)[0...7]
//	suffix_awb_number = hash.upcase + 'BPT'
//	suffix_awb_number.rjust(11, '0')
func generateBukasendOOMSuffixAWBNumber(transaction *entity.Transaction) (string, error) {
	transactionIDStr := fmt.Sprintf("%d", transaction.ID)

	// Remove 'BPT' characters (Ruby's tr('BPT', ''))
	cleanID := strings.ReplaceAll(transactionIDStr, "B", "")
	cleanID = strings.ReplaceAll(cleanID, "P", "")
	cleanID = strings.ReplaceAll(cleanID, "T", "")

	// Generate MD5 hash and take first 7 characters
	hash := md5.Sum([]byte(cleanID))
	hashStr := fmt.Sprintf("%x", hash)[:7]

	suffixAWBNumber := strings.ToUpper(hashStr) + "BPT"

	// Right justify with zeros to 11 characters
	for len(suffixAWBNumber) < 11 {
		suffixAWBNumber = "0" + suffixAWBNumber
	}

	return suffixAWBNumber, nil
}

// ValidateJNEReceiptNumber validates that a receipt number follows JNE format
func ValidateJNEReceiptNumber(receiptNumber string) bool {
	if len(receiptNumber) < 4 {
		return false
	}

	// Check if it starts with a valid prefix
	prefix := receiptNumber[:4]
	for _, validPrefix := range AWBPrefix {
		if prefix == validPrefix {
			return true
		}
	}

	return false
}
