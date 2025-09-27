package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// BarcodeGenerationBatchRepository defines the interface for batch generation tracking
type BarcodeGenerationBatchRepository interface {
	// CreateBatch creates a new batch generation record
	CreateBatch(ctx context.Context, batch *entity.BarcodeGenerationBatch) error

	// GetBatch retrieves a batch by its ID
	GetBatch(ctx context.Context, batchID uuid.UUID) (*entity.BarcodeGenerationBatch, error)

	// GetBatchByNumber retrieves a batch by its batch number
	GetBatchByNumber(ctx context.Context, batchNumber string) (*entity.BarcodeGenerationBatch, error)

	// UpdateBatch updates an existing batch record
	UpdateBatch(ctx context.Context, batch *entity.BarcodeGenerationBatch) error

	// GetBatchesByStorefront retrieves batches for a specific storefront
	GetBatchesByStorefront(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.BarcodeGenerationBatch, error)

	// GetBatchesByProduct retrieves batches for a specific product
	GetBatchesByProduct(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*entity.BarcodeGenerationBatch, error)

	// GetBatchesByStatus retrieves batches by generation status
	GetBatchesByStatus(ctx context.Context, storefrontID uuid.UUID, status string, limit, offset int) ([]*entity.BarcodeGenerationBatch, error)

	// GetBatchesByDateRange retrieves batches within a date range
	GetBatchesByDateRange(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*entity.BarcodeGenerationBatch, error)

	// GetWithFilters retrieves batches with filters and pagination
	GetWithFilters(ctx context.Context, filters *BatchFilters) ([]*entity.BarcodeGenerationBatch, error)

	// Count counts batches with optional filters
	Count(ctx context.Context, filters *BatchFilters) (int, error)

	// Delete soft deletes a batch record
	Delete(ctx context.Context, batchID uuid.UUID) error

	// GetBatchStatistics retrieves batch generation statistics
	GetBatchStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time) (*BatchStatistics, error)

	// GenerateBatchNumber generates a unique batch number
	GenerateBatchNumber(ctx context.Context, storefrontID uuid.UUID) (string, error)
}

// BarcodeCollisionRepository defines the interface for collision tracking
type BarcodeCollisionRepository interface {
	// LogCollision logs a barcode collision for monitoring
	LogCollision(ctx context.Context, attemptedBarcode string, attempt int, batchID *uuid.UUID) error

	// GetCollisionStats retrieves collision statistics
	GetCollisionStats(ctx context.Context, startDate, endDate *time.Time) (*CollisionStats, error)

	// GetCollisionsByBatch retrieves collisions for a specific batch
	GetCollisionsByBatch(ctx context.Context, batchID uuid.UUID) ([]*BarcodeCollision, error)

	// GetCollisionsByDateRange retrieves collisions within a date range
	GetCollisionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*BarcodeCollision, error)

	// CleanupOldCollisions removes collision logs older than the specified number of days
	CleanupOldCollisions(ctx context.Context, olderThanDays int) error

	// GetCollisionTrends retrieves collision trends over time
	GetCollisionTrends(ctx context.Context, period string, startDate, endDate time.Time) ([]*CollisionTrend, error)
}

// BatchFilters represents filters for batch queries
type BatchFilters struct {
	StorefrontID        *uuid.UUID `json:"storefront_id,omitempty"`
	ProductID           *uuid.UUID `json:"product_id,omitempty"`
	RequestedBy         *uuid.UUID `json:"requested_by,omitempty"`
	GenerationStatus    *string    `json:"generation_status,omitempty"`
	CreatedAfter        *time.Time `json:"created_after,omitempty"`
	CreatedBefore       *time.Time `json:"created_before,omitempty"`
	CompletedAfter      *time.Time `json:"completed_after,omitempty"`
	CompletedBefore     *time.Time `json:"completed_before,omitempty"`
	MinQuantity         *int       `json:"min_quantity,omitempty"`
	MaxQuantity         *int       `json:"max_quantity,omitempty"`
	IntendedRecipient   *string    `json:"intended_recipient,omitempty"`
	BatchNumber         *string    `json:"batch_number,omitempty"`
	Search              *string    `json:"search,omitempty"`
	Page                int        `json:"page"`
	PageSize            int        `json:"page_size"`
	SortBy              string     `json:"sort_by"`
	SortDirection       string     `json:"sort_direction"`
	IncludeBarcodeCount bool       `json:"include_barcode_count"`
	IncludePerformance  bool       `json:"include_performance"`
}

// BatchStatistics represents batch generation statistics
type BatchStatistics struct {
	TotalBatches           int64                `json:"total_batches"`
	CompletedBatches       int64                `json:"completed_batches"`
	PartialBatches         int64                `json:"partial_batches"`
	FailedBatches          int64                `json:"failed_batches"`
	TotalBarcodesGenerated int64                `json:"total_barcodes_generated"`
	TotalBarcodesRequested int64                `json:"total_barcodes_requested"`
	TotalCollisions        int64                `json:"total_collisions"`
	AverageGenerationTime  float64              `json:"average_generation_time_ms"`
	AverageBatchSize       float64              `json:"average_batch_size"`
	SuccessRate            float64              `json:"success_rate"`
	CollisionRate          float64              `json:"collision_rate"`
	ThroughputPerHour      float64              `json:"throughput_per_hour"`
	MonthlyBatchTrends     []*MonthlyBatchTrend `json:"monthly_batch_trends"`
	TopPerformers          []*BatchPerformer    `json:"top_performers"`
}

// MonthlyBatchTrend represents batch trends by month
type MonthlyBatchTrend struct {
	Month         string  `json:"month"`
	Year          int     `json:"year"`
	TotalBatches  int64   `json:"total_batches"`
	TotalBarcodes int64   `json:"total_barcodes"`
	AvgTime       float64 `json:"average_time"`
	CollisionRate float64 `json:"collision_rate"`
}

// BatchPerformer represents top performing batch generators
type BatchPerformer struct {
	UserID        uuid.UUID `json:"user_id"`
	UserName      string    `json:"user_name"`
	TotalBatches  int64     `json:"total_batches"`
	TotalBarcodes int64     `json:"total_barcodes"`
	SuccessRate   float64   `json:"success_rate"`
}

// CollisionStats represents collision statistics
type CollisionStats struct {
	TotalCollisions  int64               `json:"total_collisions"`
	CollisionRate    float64             `json:"collision_rate"`
	TotalGenerated   int64               `json:"total_generated"`
	MaxRetries       int                 `json:"max_retries"`
	AverageRetries   float64             `json:"average_retries"`
	CollisionsByHour []int64             `json:"collisions_by_hour"`
	CollisionsByDay  []int64             `json:"collisions_by_day"`
	RecentCollisions []*BarcodeCollision `json:"recent_collisions"`
}

// BarcodeCollision represents a barcode collision record
type BarcodeCollision struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	AttemptedBarcode string     `json:"attempted_barcode" db:"attempted_barcode"`
	CollisionAttempt int        `json:"collision_attempt" db:"collision_attempt"`
	BatchID          *uuid.UUID `json:"batch_id,omitempty" db:"batch_id"`
	DetectedAt       time.Time  `json:"detected_at" db:"detected_at"`
	ResolutionMethod *string    `json:"resolution_method,omitempty" db:"resolution_method"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// CollisionTrend represents collision trends over time
type CollisionTrend struct {
	Period        string    `json:"period"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Collisions    int64     `json:"collisions"`
	Generations   int64     `json:"generations"`
	CollisionRate float64   `json:"collision_rate"`
}
