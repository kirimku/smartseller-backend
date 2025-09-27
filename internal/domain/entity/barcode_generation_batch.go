package entity

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BatchGenerationStatus represents the status of a batch generation process
type BatchGenerationStatus string

const (
	BatchStatusInProgress BatchGenerationStatus = "in_progress"
	BatchStatusCompleted  BatchGenerationStatus = "completed"
	BatchStatusFailed     BatchGenerationStatus = "failed"
	BatchStatusPartial    BatchGenerationStatus = "partial"
)

// Valid validates the batch generation status
func (bgs BatchGenerationStatus) Valid() bool {
	switch bgs {
	case BatchStatusInProgress, BatchStatusCompleted, BatchStatusFailed, BatchStatusPartial:
		return true
	default:
		return false
	}
}

// String returns the string representation of BatchGenerationStatus
func (bgs BatchGenerationStatus) String() string {
	return string(bgs)
}

// Value implements the driver.Valuer interface for database storage
func (bgs BatchGenerationStatus) Value() (driver.Value, error) {
	return string(bgs), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (bgs *BatchGenerationStatus) Scan(value interface{}) error {
	if value == nil {
		*bgs = BatchStatusInProgress
		return nil
	}
	if str, ok := value.(string); ok {
		*bgs = BatchGenerationStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into BatchGenerationStatus", value)
}

// BarcodeGenerationBatch represents a batch generation record
type BarcodeGenerationBatch struct {
	// Primary identification
	ID          uuid.UUID `json:"id" db:"id"`
	BatchNumber string    `json:"batch_number" db:"batch_number"`

	// Batch associations
	ProductID    uuid.UUID `json:"product_id" db:"product_id"`
	StorefrontID uuid.UUID `json:"storefront_id" db:"storefront_id"`

	// Generation details
	RequestedQuantity int `json:"requested_quantity" db:"requested_quantity"`
	GeneratedQuantity int `json:"generated_quantity" db:"generated_quantity"`
	FailedQuantity    int `json:"failed_quantity" db:"failed_quantity"`

	// Batch metadata
	GenerationStartedAt   time.Time  `json:"generation_started_at" db:"generation_started_at"`
	GenerationCompletedAt *time.Time `json:"generation_completed_at,omitempty" db:"generation_completed_at"`
	GenerationStatus      string     `json:"generation_status" db:"generation_status"`

	// Performance tracking
	AverageGenerationTimeMs *int `json:"average_generation_time_ms,omitempty" db:"average_generation_time_ms"`
	CollisionCount          int  `json:"collision_count" db:"collision_count"`
	RetryCount              int  `json:"retry_count" db:"retry_count"`

	// Distribution details
	IntendedRecipient string  `json:"intended_recipient" db:"intended_recipient"`
	DistributionNotes *string `json:"distribution_notes,omitempty" db:"distribution_notes"`

	// Audit
	RequestedBy uuid.UUID `json:"requested_by" db:"requested_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields (not stored in database)
	SuccessRate      float64 `json:"success_rate" db:"-"`
	CollisionRate    float64 `json:"collision_rate" db:"-"`
	ProcessingTime   string  `json:"processing_time" db:"-"`
	IsCompleted      bool    `json:"is_completed" db:"-"`
	PerformanceScore string  `json:"performance_score" db:"-"`
}

// NewBarcodeGenerationBatch creates a new batch generation record
func NewBarcodeGenerationBatch(
	batchNumber string,
	productID, storefrontID, requestedBy uuid.UUID,
	requestedQuantity int,
) *BarcodeGenerationBatch {
	return &BarcodeGenerationBatch{
		ID:                  uuid.New(),
		BatchNumber:         batchNumber,
		ProductID:           productID,
		StorefrontID:        storefrontID,
		RequestedQuantity:   requestedQuantity,
		GeneratedQuantity:   0,
		FailedQuantity:      0,
		GenerationStartedAt: time.Now(),
		GenerationStatus:    string(BatchStatusInProgress),
		CollisionCount:      0,
		RetryCount:          0,
		RequestedBy:         requestedBy,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
}

// Validate validates the batch generation record
func (bgb *BarcodeGenerationBatch) Validate() error {
	// Required fields
	if bgb.BatchNumber == "" {
		return fmt.Errorf("batch_number is required")
	}
	if bgb.ProductID == uuid.Nil {
		return fmt.Errorf("product_id is required")
	}
	if bgb.StorefrontID == uuid.Nil {
		return fmt.Errorf("storefront_id is required")
	}
	if bgb.RequestedBy == uuid.Nil {
		return fmt.Errorf("requested_by is required")
	}

	// Validate quantities
	if bgb.RequestedQuantity <= 0 {
		return fmt.Errorf("requested_quantity must be positive")
	}
	if bgb.GeneratedQuantity < 0 {
		return fmt.Errorf("generated_quantity cannot be negative")
	}
	if bgb.FailedQuantity < 0 {
		return fmt.Errorf("failed_quantity cannot be negative")
	}
	if bgb.CollisionCount < 0 {
		return fmt.Errorf("collision_count cannot be negative")
	}
	if bgb.RetryCount < 0 {
		return fmt.Errorf("retry_count cannot be negative")
	}

	// Validate status
	status := BatchGenerationStatus(bgb.GenerationStatus)
	if !status.Valid() {
		return fmt.Errorf("invalid generation status: %s", bgb.GenerationStatus)
	}

	return nil
}

// UpdateProgress updates the batch progress
func (bgb *BarcodeGenerationBatch) UpdateProgress(generated, failed, collisions, retries int) {
	bgb.GeneratedQuantity = generated
	bgb.FailedQuantity = failed
	bgb.CollisionCount = collisions
	bgb.RetryCount = retries
	bgb.UpdatedAt = time.Now()
}

// Complete marks the batch as completed
func (bgb *BarcodeGenerationBatch) Complete(avgGenerationTimeMs int) {
	now := time.Now()
	bgb.GenerationCompletedAt = &now
	bgb.AverageGenerationTimeMs = &avgGenerationTimeMs

	// Determine final status
	if bgb.FailedQuantity == 0 {
		bgb.GenerationStatus = string(BatchStatusCompleted)
	} else if bgb.GeneratedQuantity > 0 {
		bgb.GenerationStatus = string(BatchStatusPartial)
	} else {
		bgb.GenerationStatus = string(BatchStatusFailed)
	}

	bgb.UpdatedAt = now
}

// MarkFailed marks the batch as failed
func (bgb *BarcodeGenerationBatch) MarkFailed(reason string) {
	bgb.GenerationStatus = string(BatchStatusFailed)
	if reason != "" {
		bgb.DistributionNotes = &reason
	}
	bgb.UpdatedAt = time.Now()
}

// ComputeFields calculates computed fields
func (bgb *BarcodeGenerationBatch) ComputeFields() {
	// Calculate success rate
	total := bgb.GeneratedQuantity + bgb.FailedQuantity
	if total > 0 {
		bgb.SuccessRate = float64(bgb.GeneratedQuantity) / float64(total) * 100
	}

	// Calculate collision rate
	totalAttempts := bgb.GeneratedQuantity + bgb.FailedQuantity + bgb.CollisionCount
	if totalAttempts > 0 {
		bgb.CollisionRate = float64(bgb.CollisionCount) / float64(totalAttempts) * 100
	}

	// Calculate processing time
	if bgb.GenerationCompletedAt != nil {
		duration := bgb.GenerationCompletedAt.Sub(bgb.GenerationStartedAt)
		bgb.ProcessingTime = formatDuration(duration)
		bgb.IsCompleted = true
	} else {
		duration := time.Since(bgb.GenerationStartedAt)
		bgb.ProcessingTime = formatDuration(duration)
		bgb.IsCompleted = false
	}

	// Calculate performance score
	bgb.PerformanceScore = bgb.calculatePerformanceScore()
}

// calculatePerformanceScore calculates a performance score based on various metrics
func (bgb *BarcodeGenerationBatch) calculatePerformanceScore() string {
	if !bgb.IsCompleted {
		return "IN_PROGRESS"
	}

	score := 100.0

	// Deduct for failed generations
	if bgb.RequestedQuantity > 0 {
		failureRate := float64(bgb.FailedQuantity) / float64(bgb.RequestedQuantity) * 100
		score -= failureRate * 2 // Each 1% failure reduces score by 2 points
	}

	// Deduct for collisions
	if bgb.CollisionRate > 1.0 {
		score -= (bgb.CollisionRate - 1.0) * 10 // Excessive collisions penalty
	}

	// Deduct for excessive time (if available)
	if bgb.AverageGenerationTimeMs != nil && *bgb.AverageGenerationTimeMs > 10 {
		excessTime := float64(*bgb.AverageGenerationTimeMs - 10)
		score -= excessTime / 10 // Each 10ms over ideal reduces score by 1 point
	}

	if score >= 95 {
		return "EXCELLENT"
	} else if score >= 85 {
		return "GOOD"
	} else if score >= 70 {
		return "FAIR"
	} else {
		return "POOR"
	}
}

// GetSummary returns a summary of the batch
func (bgb *BarcodeGenerationBatch) GetSummary() string {
	return fmt.Sprintf("Batch %s: %d/%d generated, %d failed, %.1f%% success rate",
		bgb.BatchNumber, bgb.GeneratedQuantity, bgb.RequestedQuantity,
		bgb.FailedQuantity, bgb.SuccessRate)
}

// IsInProgress checks if the batch generation is in progress
func (bgb *BarcodeGenerationBatch) IsInProgress() bool {
	return BatchGenerationStatus(bgb.GenerationStatus) == BatchStatusInProgress
}

// IsSuccessful checks if the batch was successful (completed with no failures)
func (bgb *BarcodeGenerationBatch) IsSuccessful() bool {
	return BatchGenerationStatus(bgb.GenerationStatus) == BatchStatusCompleted
}

// HasPartialSuccess checks if the batch had partial success
func (bgb *BarcodeGenerationBatch) HasPartialSuccess() bool {
	return BatchGenerationStatus(bgb.GenerationStatus) == BatchStatusPartial
}

// String returns a string representation of the batch
func (bgb *BarcodeGenerationBatch) String() string {
	return fmt.Sprintf("BarcodeGenerationBatch{ID: %s, Number: %s, Status: %s, Progress: %d/%d}",
		bgb.ID.String(), bgb.BatchNumber, bgb.GenerationStatus,
		bgb.GeneratedQuantity, bgb.RequestedQuantity)
}
