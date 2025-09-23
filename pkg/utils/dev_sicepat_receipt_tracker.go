package utils

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// DevSiCepatReceiptTracker manages development receipt numbers for SiCepat
type DevSiCepatReceiptTracker struct {
	filePath    string
	usedNumbers map[string]bool
	mu          sync.RWMutex
	rangeStart  int64
	rangeEnd    int64
}

// DevReceiptData represents the structure of the tracking file
type DevReceiptData struct {
	UsedNumbers []string `json:"used_numbers"`
	LastUpdated string   `json:"last_updated"`
	RangeStart  int64    `json:"range_start"`
	RangeEnd    int64    `json:"range_end"`
}

// NewDevSiCepatReceiptTracker creates a new development receipt tracker
func NewDevSiCepatReceiptTracker() *DevSiCepatReceiptTracker {
	tracker := &DevSiCepatReceiptTracker{
		filePath:    "./dev_sicepat_receipts.json",
		usedNumbers: make(map[string]bool),
		rangeStart:  888889340571, // Start of development range
		rangeEnd:    888889341570, // End of development range
	}

	// Load existing data
	tracker.loadFromFile()

	return tracker
}

// GetNextAvailableReceipt returns the next available receipt number in the development range
func (t *DevSiCepatReceiptTracker) GetNextAvailableReceipt(transactionID int) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Try to find an unused number in the range
	maxAttempts := 100 // Prevent infinite loops
	attempts := 0

	for attempts < maxAttempts {
		// Generate a random number in the range
		receiptNum := t.rangeStart + rand.Int63n(t.rangeEnd-t.rangeStart+1)
		receiptStr := fmt.Sprintf("%d", receiptNum)

		// Check if this number is already used
		if !t.usedNumbers[receiptStr] {
			// Mark as used and save
			t.usedNumbers[receiptStr] = true
			err := t.saveToFile()
			if err != nil {
				logger.Error("dev_sicepat_receipt.save_error",
					"Failed to save receipt tracking file",
					map[string]interface{}{
						"error":          err.Error(),
						"transaction_id": transactionID,
						"receipt_number": receiptStr,
					})
				// Continue anyway, just log the error
			}

			logger.Info("dev_sicepat_receipt.assigned",
				"Assigned development SiCepat receipt number",
				map[string]interface{}{
					"transaction_id": transactionID,
					"receipt_number": receiptStr,
					"range_start":    t.rangeStart,
					"range_end":      t.rangeEnd,
					"used_count":     len(t.usedNumbers),
					"total_range":    t.rangeEnd - t.rangeStart + 1,
				})

			return receiptStr, nil
		}

		attempts++
	}

	// If we've exhausted attempts, check if we're running out of numbers
	usedCount := len(t.usedNumbers)
	totalRange := t.rangeEnd - t.rangeStart + 1

	if int64(usedCount) >= int64(float64(totalRange)*0.9) { // 90% used
		logger.Error("dev_sicepat_receipt.range_exhausted",
			"Development receipt range nearly exhausted",
			map[string]interface{}{
				"used_count":     usedCount,
				"total_range":    totalRange,
				"range_start":    t.rangeStart,
				"range_end":      t.rangeEnd,
				"transaction_id": transactionID,
			})
	}

	return "", fmt.Errorf("failed to find available receipt number after %d attempts (used: %d/%d)",
		maxAttempts, usedCount, totalRange)
}

// IsReceiptUsed checks if a receipt number is already used
func (t *DevSiCepatReceiptTracker) IsReceiptUsed(receiptNumber string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.usedNumbers[receiptNumber]
}

// GetUsageStats returns statistics about receipt usage
func (t *DevSiCepatReceiptTracker) GetUsageStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	totalRange := t.rangeEnd - t.rangeStart + 1
	usedCount := len(t.usedNumbers)

	return map[string]interface{}{
		"range_start":      t.rangeStart,
		"range_end":        t.rangeEnd,
		"total_range":      totalRange,
		"used_count":       usedCount,
		"available_count":  totalRange - int64(usedCount),
		"usage_percentage": float64(usedCount) / float64(totalRange) * 100,
	}
}

// ResetUsedNumbers clears all used numbers (for testing purposes)
func (t *DevSiCepatReceiptTracker) ResetUsedNumbers() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.usedNumbers = make(map[string]bool)
	return t.saveToFile()
}

// loadFromFile loads the tracking data from the JSON file
func (t *DevSiCepatReceiptTracker) loadFromFile() {
	data, err := os.ReadFile(t.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start fresh
			logger.Info("dev_sicepat_receipt.new_file",
				"Creating new development receipt tracking file",
				map[string]interface{}{
					"file_path":   t.filePath,
					"range_start": t.rangeStart,
					"range_end":   t.rangeEnd,
				})
			return
		}
		logger.Error("dev_sicepat_receipt.load_error",
			"Failed to load receipt tracking file",
			map[string]interface{}{
				"error":     err.Error(),
				"file_path": t.filePath,
			})
		return
	}

	var receiptData DevReceiptData
	if err := json.Unmarshal(data, &receiptData); err != nil {
		logger.Error("dev_sicepat_receipt.parse_error",
			"Failed to parse receipt tracking file",
			map[string]interface{}{
				"error":     err.Error(),
				"file_path": t.filePath,
			})
		return
	}

	// Load used numbers into map
	for _, num := range receiptData.UsedNumbers {
		t.usedNumbers[num] = true
	}

	// Update range if specified in file
	if receiptData.RangeStart > 0 && receiptData.RangeEnd > 0 {
		t.rangeStart = receiptData.RangeStart
		t.rangeEnd = receiptData.RangeEnd
	}

	logger.Info("dev_sicepat_receipt.loaded",
		"Loaded development receipt tracking data",
		map[string]interface{}{
			"file_path":    t.filePath,
			"used_count":   len(t.usedNumbers),
			"range_start":  t.rangeStart,
			"range_end":    t.rangeEnd,
			"last_updated": receiptData.LastUpdated,
		})
}

// saveToFile saves the tracking data to the JSON file
func (t *DevSiCepatReceiptTracker) saveToFile() error {
	// Convert map to slice
	usedNumbers := make([]string, 0, len(t.usedNumbers))
	for num := range t.usedNumbers {
		usedNumbers = append(usedNumbers, num)
	}

	receiptData := DevReceiptData{
		UsedNumbers: usedNumbers,
		LastUpdated: time.Now().Format(time.RFC3339),
		RangeStart:  t.rangeStart,
		RangeEnd:    t.rangeEnd,
	}

	data, err := json.MarshalIndent(receiptData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal receipt data: %w", err)
	}

	return os.WriteFile(t.filePath, data, 0644)
}

// Initialize random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}
