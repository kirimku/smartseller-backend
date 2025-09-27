package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// WarrantyBarcodeRepository defines the interface for warranty barcode data operations
type WarrantyBarcodeRepository interface {
	// Create creates a new warranty barcode
	Create(ctx context.Context, barcode *entity.WarrantyBarcode) error

	// CreateBatch creates multiple warranty barcodes in a single operation
	CreateBatch(ctx context.Context, barcodes []*entity.WarrantyBarcode) error

	// GetByID retrieves a warranty barcode by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WarrantyBarcode, error)

	// GetByBarcodeNumber retrieves a warranty barcode by its barcode number
	GetByBarcodeNumber(ctx context.Context, barcodeNumber string) (*entity.WarrantyBarcode, error)

	// CheckUniqueness checks if a barcode number is unique
	CheckUniqueness(ctx context.Context, barcodeNumber string) (bool, error)

	// GetByProductID retrieves warranty barcodes for a specific product
	GetByProductID(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*entity.WarrantyBarcode, error)

	// GetByStorefrontID retrieves warranty barcodes for a specific storefront
	GetByStorefrontID(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyBarcode, error)

	// GetByBatchID retrieves warranty barcodes from a specific batch
	GetByBatchID(ctx context.Context, batchID uuid.UUID, limit, offset int) ([]*entity.WarrantyBarcode, error)

	// GetByStatus retrieves warranty barcodes by status
	GetByStatus(ctx context.Context, storefrontID uuid.UUID, status entity.BarcodeStatus, limit, offset int) ([]*entity.WarrantyBarcode, error)

	// GetExpiringSoon retrieves warranty barcodes expiring within the specified number of days
	GetExpiringSoon(ctx context.Context, storefrontID uuid.UUID, days int, limit, offset int) ([]*entity.WarrantyBarcode, error)

	// Update updates an existing warranty barcode
	Update(ctx context.Context, barcode *entity.WarrantyBarcode) error

	// Activate activates a warranty barcode
	Activate(ctx context.Context, barcodeNumber string, activatedBy uuid.UUID) error

	// Deactivate deactivates a warranty barcode
	Deactivate(ctx context.Context, barcodeNumber string, reason string) error

	// Delete soft deletes a warranty barcode
	Delete(ctx context.Context, id uuid.UUID) error

	// Count counts warranty barcodes with optional filters
	Count(ctx context.Context, filters *WarrantyBarcodeFilters) (int, error)

	// GetWithFilters retrieves warranty barcodes with filters and pagination
	GetWithFilters(ctx context.Context, filters *WarrantyBarcodeFilters) ([]*entity.WarrantyBarcode, error)

	// GetGenerationStats retrieves generation statistics
	GetGenerationStats(ctx context.Context, req *GenerationStatsRequest) (*GenerationStatsResponse, error)

	// GetUsageStatistics retrieves usage statistics for analytics
	GetUsageStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time) (*WarrantyUsageStats, error)
}

// WarrantyBarcodeFilters represents filters for warranty barcode queries
type WarrantyBarcodeFilters struct {
	StorefrontID     *uuid.UUID            `json:"storefront_id,omitempty"`
	ProductID        *uuid.UUID            `json:"product_id,omitempty"`
	BatchID          *uuid.UUID            `json:"batch_id,omitempty"`
	Status           *entity.BarcodeStatus `json:"status,omitempty"`
	CreatedBy        *uuid.UUID            `json:"created_by,omitempty"`
	ActivatedBy      *uuid.UUID            `json:"activated_by,omitempty"`
	CreatedAfter     *time.Time            `json:"created_after,omitempty"`
	CreatedBefore    *time.Time            `json:"created_before,omitempty"`
	ActivatedAfter   *time.Time            `json:"activated_after,omitempty"`
	ActivatedBefore  *time.Time            `json:"activated_before,omitempty"`
	ExpiresAfter     *time.Time            `json:"expires_after,omitempty"`
	ExpiresBefore    *time.Time            `json:"expires_before,omitempty"`
	GenerationMethod *string               `json:"generation_method,omitempty"`
	BatchNumber      *string               `json:"batch_number,omitempty"`
	Search           *string               `json:"search,omitempty"`
	Page             int                   `json:"page"`
	PageSize         int                   `json:"page_size"`
	SortBy           string                `json:"sort_by"`
	SortDirection    string                `json:"sort_direction"`
}

// GenerationStatsRequest represents a request for generation statistics
type GenerationStatsRequest struct {
	ProductID    *uuid.UUID `json:"product_id,omitempty"`
	StorefrontID *uuid.UUID `json:"storefront_id,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Period       string     `json:"period,omitempty"` // today, week, month, year
}

// GenerationStatsResponse provides generation statistics
type GenerationStatsResponse struct {
	TotalGenerated        int64          `json:"total_generated"`
	GenerationRate        float64        `json:"generation_rate"`
	CollisionCount        int64          `json:"collision_count"`
	CollisionRate         float64        `json:"collision_rate"`
	AverageGenerationTime time.Duration  `json:"average_generation_time"`
	EntropyUtilization    float64        `json:"entropy_utilization"`
	SecurityStatus        string         `json:"security_status"`
	PeriodStatistics      []*PeriodStats `json:"period_statistics"`
}

// PeriodStats represents statistics for a specific period
type PeriodStats struct {
	Period     string        `json:"period"`
	StartDate  time.Time     `json:"start_date"`
	EndDate    time.Time     `json:"end_date"`
	Generated  int64         `json:"generated"`
	Collisions int64         `json:"collisions"`
	AvgTime    time.Duration `json:"average_time"`
}

// WarrantyUsageStats represents warranty usage statistics
type WarrantyUsageStats struct {
	TotalWarranties   int64   `json:"total_warranties"`
	ActiveWarranties  int64   `json:"active_warranties"`
	UsedWarranties    int64   `json:"used_warranties"`
	ExpiredWarranties int64   `json:"expired_warranties"`
	RevokedWarranties int64   `json:"revoked_warranties"`
	ActivationRate    float64 `json:"activation_rate"`
	ExpirationRate    float64 `json:"expiration_rate"`
	AverageLifespan   int     `json:"average_lifespan_days"`
}
