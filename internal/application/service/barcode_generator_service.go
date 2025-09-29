package service

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/rs/zerolog"
)

// BarcodeGeneratorService defines the interface for secure barcode generation
type BarcodeGeneratorService interface {
	// Single barcode generation
	GenerateBarcode(ctx context.Context, productID, storefrontID, createdBy uuid.UUID, warrantyPeriodMonths int) (*entity.WarrantyBarcode, error)

	// Batch generation
	GenerateBatch(ctx context.Context, req *BatchGenerationRequest) (*BatchGenerationResult, error)

	// Validation
	ValidateBarcodeFormat(barcode string) error
	CheckUniqueness(ctx context.Context, barcodeNumber string) (bool, error)

	// Statistics
	GetGenerationStats(ctx context.Context, req *GenerationStatsRequest) (*GenerationStatsResponse, error)

	// Configuration
	GetConfiguration() *GeneratorConfiguration
}

// BatchGenerationRequest represents a request to generate multiple barcodes
type BatchGenerationRequest struct {
	ProductID            uuid.UUID `json:"product_id" validate:"required"`
	StorefrontID         uuid.UUID `json:"storefront_id" validate:"required"`
	Quantity             int       `json:"quantity" validate:"required,min=1,max=10000"`
	WarrantyPeriodMonths int       `json:"warranty_period_months" validate:"required,min=1"`
	BatchNumber          *string   `json:"batch_number,omitempty"`
	IntendedRecipient    *string   `json:"intended_recipient,omitempty"`
	DistributionNotes    *string   `json:"distribution_notes,omitempty"`
	CreatedBy            uuid.UUID `json:"created_by" validate:"required"`
}

// BatchGenerationResult represents the result of batch generation
type BatchGenerationResult struct {
	BatchID           uuid.UUID                  `json:"batch_id"`
	BatchNumber       string                     `json:"batch_number"`
	RequestedQuantity int                        `json:"requested_quantity"`
	GeneratedQuantity int                        `json:"generated_quantity"`
	FailedQuantity    int                        `json:"failed_quantity"`
	CollisionCount    int                        `json:"collision_count"`
	GenerationTime    time.Duration              `json:"generation_time"`
	Barcodes          []*entity.WarrantyBarcode  `json:"barcodes"`
	Statistics        *BatchGenerationStatistics `json:"statistics"`
	DownloadURL       *string                    `json:"download_url,omitempty"`
}

// BatchGenerationStatistics provides detailed statistics about the generation
type BatchGenerationStatistics struct {
	TotalPossibleCombinations *big.Int      `json:"total_possible_combinations"`
	UtilizedCombinations      int64         `json:"utilized_combinations"`
	EntropyUtilizationRate    float64       `json:"entropy_utilization_rate"`
	AverageGenerationTime     time.Duration `json:"average_generation_time"`
	CollisionRate             float64       `json:"collision_rate"`
	SuccessRate               float64       `json:"success_rate"`
	SecurityScore             string        `json:"security_score"`
	RecommendedAction         string        `json:"recommended_action"`
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
	EstimatedCapacity     *big.Int       `json:"estimated_capacity"`
	RecommendedAction     string         `json:"recommended_action"`
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

// GeneratorConfiguration provides configuration details
type GeneratorConfiguration struct {
	Format             string  `json:"format"`
	RandomLength       int     `json:"random_length"`
	CharacterSet       string  `json:"character_set"`
	EntropyBitsPerChar int     `json:"entropy_bits_per_char"`
	TotalEntropy       int     `json:"total_entropy"`
	MaxRetries         int     `json:"max_retries"`
	CollisionThreshold float64 `json:"collision_threshold"`
	SecurityLevel      string  `json:"security_level"`
}

// Repository interfaces needed by the service
type WarrantyBarcodeRepository interface {
	Create(ctx context.Context, barcode *entity.WarrantyBarcode) error
	GetByBarcodeNumber(ctx context.Context, barcodeNumber string) (*entity.WarrantyBarcode, error)
	CheckUniqueness(ctx context.Context, barcodeNumber string) (bool, error)
	CreateBatch(ctx context.Context, barcodes []*entity.WarrantyBarcode) error
	GetGenerationStats(ctx context.Context, req *GenerationStatsRequest) (*GenerationStatsResponse, error)
}

type BarcodeCollisionRepository interface {
	LogCollision(ctx context.Context, attemptedBarcode string, attempt int, batchID *uuid.UUID) error
	GetCollisionStats(ctx context.Context, startDate, endDate *time.Time) (*CollisionStats, error)
}

type BatchRepository interface {
	CreateBatch(ctx context.Context, batch *entity.BarcodeGenerationBatch) error
	UpdateBatch(ctx context.Context, batch *entity.BarcodeGenerationBatch) error
	GetBatch(ctx context.Context, batchID uuid.UUID) (*entity.BarcodeGenerationBatch, error)
}

// CollisionStats represents collision statistics
type CollisionStats struct {
	TotalCollisions int64   `json:"total_collisions"`
	CollisionRate   float64 `json:"collision_rate"`
	TotalGenerated  int64   `json:"total_generated"`
	MaxRetries      int     `json:"max_retries"`
}

// Constants for the barcode generator
const (
	// Character set excludes confusing characters (I, O, 1, 0)
	BarcodeCharacterSet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	BarcodeRandomLength = 12
	MaxRetries          = 3

	// Security thresholds
	CollisionThresholdWarning  = 0.01 // 0.01%
	CollisionThresholdCritical = 0.1  // 0.1%
	EntropyThresholdWarning    = 10.0 // 10% of total entropy used
	EntropyThresholdCritical   = 50.0 // 50% of total entropy used
)

// barcodeGeneratorService implements the BarcodeGeneratorService interface
type barcodeGeneratorService struct {
	barcodeRepo   WarrantyBarcodeRepository
	collisionRepo BarcodeCollisionRepository
	batchRepo     BatchRepository
	logger        zerolog.Logger
	config        *GeneratorConfiguration
}

// NewBarcodeGeneratorService creates a new barcode generator service
func NewBarcodeGeneratorService(
	barcodeRepo WarrantyBarcodeRepository,
	collisionRepo BarcodeCollisionRepository,
	batchRepo BatchRepository,
	logger zerolog.Logger,
) BarcodeGeneratorService {
	config := &GeneratorConfiguration{
		Format:             "REX[YY][RANDOM_12]",
		RandomLength:       BarcodeRandomLength,
		CharacterSet:       BarcodeCharacterSet,
		EntropyBitsPerChar: 5,  // log2(32) = 5 bits per character
		TotalEntropy:       60, // 12 chars * 5 bits = 60 bits
		MaxRetries:         MaxRetries,
		CollisionThreshold: CollisionThresholdWarning,
		SecurityLevel:      "HIGH",
	}

	return &barcodeGeneratorService{
		barcodeRepo:   barcodeRepo,
		collisionRepo: collisionRepo,
		batchRepo:     batchRepo,
		logger:        logger.With().Str("service", "barcode_generator").Logger(),
		config:        config,
	}
}

// GenerateBarcode generates a single secure warranty barcode
func (s *barcodeGeneratorService) GenerateBarcode(ctx context.Context, productID, storefrontID, createdBy uuid.UUID, warrantyPeriodMonths int) (*entity.WarrantyBarcode, error) {
	// Debug logging
	s.logger.Info().Int("warrantyPeriodMonths", warrantyPeriodMonths).Msg("DEBUG: Service received warranty period")

	// Create new warranty barcode entity
	barcode := entity.NewWarrantyBarcode(productID, storefrontID, createdBy, warrantyPeriodMonths)
	
	// Debug logging after entity creation
	s.logger.Info().Int("WarrantyPeriodMonths", barcode.WarrantyPeriodMonths).Msg("DEBUG: Entity created with warranty period")

	// Generate barcode number
	if err := barcode.GenerateBarcodeNumber(); err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate barcode number")
		return nil, fmt.Errorf("failed to generate barcode number: %w", err)
	}

	// Generate unique barcode number
	if err := s.generateUniqueBarcodeNumber(ctx, barcode, nil); err != nil {
		return nil, fmt.Errorf("failed to generate unique barcode: %w", err)
	}
	barcode.CollisionChecked = true

	// Debug logging before repository save
	s.logger.Info().Int("WarrantyPeriodMonths", barcode.WarrantyPeriodMonths).Str("BarcodeNumber", barcode.BarcodeNumber).Msg("DEBUG: Saving barcode to repository")

	// Save to repository
	if err := s.barcodeRepo.Create(ctx, barcode); err != nil {
		s.logger.Error().Err(err).Str("barcode", barcode.BarcodeNumber).Msg("Failed to save barcode")
		return nil, fmt.Errorf("failed to save barcode: %w", err)
	}

	// Debug logging after repository save
	s.logger.Info().Int("WarrantyPeriodMonths", barcode.WarrantyPeriodMonths).Str("BarcodeNumber", barcode.BarcodeNumber).Msg("DEBUG: Barcode saved successfully")

	s.logger.Info().
		Str("barcode_id", barcode.ID.String()).
		Str("barcode_number", barcode.BarcodeNumber).
		Str("product_id", productID.String()).
		Str("storefront_id", storefrontID.String()).
		Int("warranty_period_months", warrantyPeriodMonths).
		Msg("Barcode generated successfully")

	return barcode, nil
}

// GenerateBatch generates multiple barcodes in a batch
func (s *barcodeGeneratorService) GenerateBatch(
	ctx context.Context,
	req *BatchGenerationRequest,
) (*BatchGenerationResult, error) {
	start := time.Now()

	// Create batch record
	batchNumber := s.generateBatchNumber(req.BatchNumber)
	batch := &entity.BarcodeGenerationBatch{
		ID:                uuid.New(),
		BatchNumber:       batchNumber,
		ProductID:         req.ProductID,
		StorefrontID:      req.StorefrontID,
		RequestedQuantity: req.Quantity,
		RequestedBy:       req.CreatedBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if req.IntendedRecipient != nil {
		batch.IntendedRecipient = *req.IntendedRecipient
	}
	if req.DistributionNotes != nil {
		batch.DistributionNotes = req.DistributionNotes
	}

	err := s.batchRepo.CreateBatch(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("failed to create batch record: %w", err)
	}

	// Generate barcodes
	barcodes := make([]*entity.WarrantyBarcode, 0, req.Quantity)
	collisionCount := 0
	failedCount := 0

	for i := 0; i < req.Quantity; i++ {
		barcode := entity.NewWarrantyBarcode(
			req.ProductID,
			req.StorefrontID,
			req.CreatedBy,
			req.WarrantyPeriodMonths,
		)
		barcode.BatchID = &batch.ID
		barcode.BatchNumber = &batchNumber

		err := s.generateUniqueBarcodeNumber(ctx, barcode, &batch.ID)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("batch_id", batch.ID.String()).
				Int("barcode_index", i).
				Msg("Failed to generate barcode in batch")
			failedCount++
			continue
		}

		if barcode.GenerationAttempt > 1 {
			collisionCount += barcode.GenerationAttempt - 1
		}

		barcodes = append(barcodes, barcode)
	}

	// Save barcodes in batch
	if len(barcodes) > 0 {
		err = s.barcodeRepo.CreateBatch(ctx, barcodes)
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("batch_id", batch.ID.String()).
				Int("barcode_count", len(barcodes)).
				Msg("Failed to save barcodes batch")
			return nil, fmt.Errorf("failed to save barcodes batch: %w", err)
		}
	}

	// Update batch record
	batch.GeneratedQuantity = len(barcodes)
	batch.FailedQuantity = failedCount
	batch.CollisionCount = collisionCount
	batch.GenerationCompletedAt = &time.Time{}
	*batch.GenerationCompletedAt = time.Now()

	generationTime := time.Since(start)
	avgTime := int(generationTime.Milliseconds())
	if len(barcodes) > 0 {
		avgTime = int(generationTime.Milliseconds()) / len(barcodes)
	}
	batch.AverageGenerationTimeMs = &avgTime

	if failedCount == 0 {
		batch.GenerationStatus = "completed"
	} else if len(barcodes) > 0 {
		batch.GenerationStatus = "partial"
	} else {
		batch.GenerationStatus = "failed"
	}

	err = s.batchRepo.UpdateBatch(ctx, batch)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Str("batch_id", batch.ID.String()).
			Msg("Failed to update batch record")
	}

	// Calculate statistics
	statistics := s.calculateBatchStatistics(req.Quantity, len(barcodes), failedCount, collisionCount, generationTime)

	result := &BatchGenerationResult{
		BatchID:           batch.ID,
		BatchNumber:       batchNumber,
		RequestedQuantity: req.Quantity,
		GeneratedQuantity: len(barcodes),
		FailedQuantity:    failedCount,
		CollisionCount:    collisionCount,
		GenerationTime:    generationTime,
		Barcodes:          barcodes,
		Statistics:        statistics,
	}

	s.logger.Info().
		Str("batch_id", batch.ID.String()).
		Int("requested", req.Quantity).
		Int("generated", len(barcodes)).
		Int("failed", failedCount).
		Int("collisions", collisionCount).
		Dur("total_time", generationTime).
		Msg("Batch generation completed")

	return result, nil
}

// ValidateBarcodeFormat validates the barcode format
func (s *barcodeGeneratorService) ValidateBarcodeFormat(barcode string) error {
	if len(barcode) != 17 { // REX + 2 digits + 12 random chars
		return fmt.Errorf("invalid barcode length: expected 17, got %d", len(barcode))
	}

	if barcode[:3] != "REX" {
		return fmt.Errorf("invalid barcode prefix: expected REX, got %s", barcode[:3])
	}

	// Validate year part (positions 3-4)
	yearPart := barcode[3:5]
	for _, char := range yearPart {
		if char < '0' || char > '9' {
			return fmt.Errorf("invalid year part: must be numeric, got %s", yearPart)
		}
	}

	// Validate random part (positions 5-16)
	randomPart := barcode[5:]
	for _, char := range randomPart {
		found := false
		for _, allowedChar := range BarcodeCharacterSet {
			if char == allowedChar {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("invalid character in random part: %c", char)
		}
	}

	return nil
}

// CheckUniqueness checks if a barcode number is unique
func (s *barcodeGeneratorService) CheckUniqueness(ctx context.Context, barcodeNumber string) (bool, error) {
	return s.barcodeRepo.CheckUniqueness(ctx, barcodeNumber)
}

// GetGenerationStats retrieves generation statistics
func (s *barcodeGeneratorService) GetGenerationStats(
	ctx context.Context,
	req *GenerationStatsRequest,
) (*GenerationStatsResponse, error) {
	return s.barcodeRepo.GetGenerationStats(ctx, req)
}

// GetConfiguration returns the generator configuration
func (s *barcodeGeneratorService) GetConfiguration() *GeneratorConfiguration {
	return s.config
}

// generateUniqueBarcodeNumber generates a unique barcode number with collision detection
func (s *barcodeGeneratorService) generateUniqueBarcodeNumber(
	ctx context.Context,
	barcode *entity.WarrantyBarcode,
	batchID *uuid.UUID,
) error {
	for attempt := 1; attempt <= MaxRetries; attempt++ {
		// Generate the barcode number
		err := barcode.GenerateBarcodeNumber()
		if err != nil {
			return fmt.Errorf("failed to generate barcode number on attempt %d: %w", attempt, err)
		}

		// Check uniqueness
		isUnique, err := s.barcodeRepo.CheckUniqueness(ctx, barcode.BarcodeNumber)
		if err != nil {
			return fmt.Errorf("failed to check uniqueness on attempt %d: %w", attempt, err)
		}

		if isUnique {
			barcode.GenerationAttempt = attempt
			barcode.CollisionChecked = true
			return nil
		}

		// Log collision
		err = s.collisionRepo.LogCollision(ctx, barcode.BarcodeNumber, attempt, batchID)
		if err != nil {
			s.logger.Warn().
				Err(err).
				Str("attempted_barcode", barcode.BarcodeNumber).
				Int("attempt", attempt).
				Msg("Failed to log collision")
		}

		s.logger.Debug().
			Str("attempted_barcode", barcode.BarcodeNumber).
			Int("attempt", attempt).
			Msg("Barcode collision detected, retrying")
	}

	return fmt.Errorf("failed to generate unique barcode after %d attempts", MaxRetries)
}

// generateBatchNumber generates a batch number
func (s *barcodeGeneratorService) generateBatchNumber(provided *string) string {
	if provided != nil && *provided != "" {
		return *provided
	}

	// Generate format: BATCH-YYYY-MM-DD-HHMMSS
	now := time.Now()
	return fmt.Sprintf("BATCH-%s", now.Format("2006-01-02-150405"))
}

// calculateBatchStatistics calculates detailed statistics for a batch generation
func (s *barcodeGeneratorService) calculateBatchStatistics(
	requested, generated, failed, collisions int,
	generationTime time.Duration,
) *BatchGenerationStatistics {
	// Calculate total possible combinations (32^12)
	base := big.NewInt(32)
	exponent := big.NewInt(12)
	totalCombinations := new(big.Int).Exp(base, exponent, nil)

	// Calculate rates
	successRate := float64(generated) / float64(requested) * 100
	var collisionRate float64
	if generated+failed > 0 {
		collisionRate = float64(collisions) / float64(generated+failed+collisions) * 100
	}

	avgGenerationTime := time.Duration(0)
	if generated > 0 {
		avgGenerationTime = generationTime / time.Duration(generated)
	}

	// Determine security score and recommendation
	securityScore := "EXCELLENT"
	recommendedAction := "continue"

	if collisionRate > CollisionThresholdCritical {
		securityScore = "POOR"
		recommendedAction = "review_algorithm"
	} else if collisionRate > CollisionThresholdWarning {
		securityScore = "GOOD"
		recommendedAction = "monitor"
	}

	return &BatchGenerationStatistics{
		TotalPossibleCombinations: totalCombinations,
		UtilizedCombinations:      int64(generated), // This would need to be tracked globally
		EntropyUtilizationRate:    0.0,              // Would need global tracking
		AverageGenerationTime:     avgGenerationTime,
		CollisionRate:             collisionRate,
		SuccessRate:               successRate,
		SecurityScore:             securityScore,
		RecommendedAction:         recommendedAction,
	}
}
