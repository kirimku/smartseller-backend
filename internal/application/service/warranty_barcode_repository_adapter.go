package service

import (
	"context"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// WarrantyBarcodeRepositoryAdapter adapts domain repository to service interface
type WarrantyBarcodeRepositoryAdapter struct {
	domainRepo repository.WarrantyBarcodeRepository
}

// NewWarrantyBarcodeRepositoryAdapter creates a new adapter
func NewWarrantyBarcodeRepositoryAdapter(domainRepo repository.WarrantyBarcodeRepository) WarrantyBarcodeRepository {
	return &WarrantyBarcodeRepositoryAdapter{
		domainRepo: domainRepo,
	}
}

// Create creates a new warranty barcode
func (a *WarrantyBarcodeRepositoryAdapter) Create(ctx context.Context, barcode *entity.WarrantyBarcode) error {
	return a.domainRepo.Create(ctx, barcode)
}

// GetByBarcodeNumber retrieves a warranty barcode by its barcode number
func (a *WarrantyBarcodeRepositoryAdapter) GetByBarcodeNumber(ctx context.Context, barcodeNumber string) (*entity.WarrantyBarcode, error) {
	return a.domainRepo.GetByBarcodeNumber(ctx, barcodeNumber)
}

// CheckUniqueness checks if a barcode number is unique
func (a *WarrantyBarcodeRepositoryAdapter) CheckUniqueness(ctx context.Context, barcodeNumber string) (bool, error) {
	return a.domainRepo.CheckUniqueness(ctx, barcodeNumber)
}

// CreateBatch creates multiple warranty barcodes in a single operation
func (a *WarrantyBarcodeRepositoryAdapter) CreateBatch(ctx context.Context, barcodes []*entity.WarrantyBarcode) error {
	return a.domainRepo.CreateBatch(ctx, barcodes)
}

// GetGenerationStats provides generation statistics (mock implementation for now)
func (a *WarrantyBarcodeRepositoryAdapter) GetGenerationStats(ctx context.Context, req *GenerationStatsRequest) (*GenerationStatsResponse, error) {
	// TODO: Implement actual statistics when domain repository supports it
	// For now, return mock data
	return &GenerationStatsResponse{
		TotalGenerated:        0,
		GenerationRate:        0.0,
		CollisionCount:        0,
		CollisionRate:         0.0,
		AverageGenerationTime: 0,
		EntropyUtilization:    0.0,
		RecommendedAction:     "No action required",
		SecurityStatus:        "Good",
		PeriodStatistics:      []*PeriodStats{},
	}, nil
}