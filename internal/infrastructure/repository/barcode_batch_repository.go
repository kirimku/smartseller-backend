package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/rs/zerolog"
)

// BarcodeGenerationBatchRepositoryImpl implements the BarcodeGenerationBatchRepository interface
type BarcodeGenerationBatchRepositoryImpl struct {
	*BaseRepository
	logger zerolog.Logger
}

// NewBarcodeGenerationBatchRepository creates a new barcode generation batch repository
func NewBarcodeGenerationBatchRepository(
	db *sqlx.DB,
	tenantResolver tenant.TenantResolver,
	logger zerolog.Logger,
) repository.BarcodeGenerationBatchRepository {
	return &BarcodeGenerationBatchRepositoryImpl{
		BaseRepository: NewBaseRepository(db, tenantResolver),
		logger:         logger.With().Str("repository", "barcode_generation_batch").Logger(),
	}
}

// CreateBatch creates a new batch generation record
func (r *BarcodeGenerationBatchRepositoryImpl) CreateBatch(ctx context.Context, batch *entity.BarcodeGenerationBatch) error {
	db, err := r.GetDB(ctx, batch.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		INSERT INTO barcode_generation_batches (
			id, batch_number, product_id, storefront_id, requested_quantity, generated_quantity,
			failed_quantity, generation_started_at, generation_completed_at, generation_status,
			average_generation_time_ms, collision_count, retry_count, intended_recipient,
			distribution_notes, requested_by, created_at, updated_at
		) VALUES (
			:id, :batch_number, :product_id, :storefront_id, :requested_quantity, :generated_quantity,
			:failed_quantity, :generation_started_at, :generation_completed_at, :generation_status,
			:average_generation_time_ms, :collision_count, :retry_count, :intended_recipient,
			:distribution_notes, :requested_by, :created_at, :updated_at
		)`

	_, err = db.NamedExecContext(ctx, query, batch)
	if err != nil {
		r.logger.Error().Err(err).Str("batch_number", batch.BatchNumber).Msg("Failed to create batch")
		return fmt.Errorf("failed to create batch: %w", err)
	}

	r.logger.Info().Str("batch_number", batch.BatchNumber).Str("id", batch.ID.String()).Msg("Batch created")
	return nil
}

// GetBatch retrieves a batch by its ID
func (r *BarcodeGenerationBatchRepositoryImpl) GetBatch(ctx context.Context, batchID uuid.UUID) (*entity.BarcodeGenerationBatch, error) {
	query := `SELECT * FROM barcode_generation_batches WHERE id = $1 AND deleted_at IS NULL`

	var batch entity.BarcodeGenerationBatch
	err := r.db.GetContext(ctx, &batch, query, batchID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", batchID.String()).Msg("Failed to get batch by ID")
		return nil, fmt.Errorf("failed to get batch by ID: %w", err)
	}

	batch.ComputeFields()
	return &batch, nil
}

// GetBatchByNumber retrieves a batch by its batch number
func (r *BarcodeGenerationBatchRepositoryImpl) GetBatchByNumber(ctx context.Context, batchNumber string) (*entity.BarcodeGenerationBatch, error) {
	query := `SELECT * FROM barcode_generation_batches WHERE batch_number = $1 AND deleted_at IS NULL`

	var batch entity.BarcodeGenerationBatch
	err := r.db.GetContext(ctx, &batch, query, batchNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("batch_number", batchNumber).Msg("Failed to get batch by number")
		return nil, fmt.Errorf("failed to get batch by number: %w", err)
	}

	batch.ComputeFields()
	return &batch, nil
}

// UpdateBatch updates an existing batch record
func (r *BarcodeGenerationBatchRepositoryImpl) UpdateBatch(ctx context.Context, batch *entity.BarcodeGenerationBatch) error {
	db, err := r.GetDB(ctx, batch.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	batch.UpdatedAt = time.Now()

	query := `
		UPDATE barcode_generation_batches SET 
			generated_quantity = :generated_quantity,
			failed_quantity = :failed_quantity,
			generation_completed_at = :generation_completed_at,
			generation_status = :generation_status,
			average_generation_time_ms = :average_generation_time_ms,
			collision_count = :collision_count,
			retry_count = :retry_count,
			intended_recipient = :intended_recipient,
			distribution_notes = :distribution_notes,
			updated_at = :updated_at
		WHERE id = :id AND storefront_id = :storefront_id AND deleted_at IS NULL`

	result, err := db.NamedExecContext(ctx, query, batch)
	if err != nil {
		r.logger.Error().Err(err).Str("id", batch.ID.String()).Msg("Failed to update batch")
		return fmt.Errorf("failed to update batch: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("batch not found or not updated")
	}

	r.logger.Info().Str("id", batch.ID.String()).Msg("Batch updated")
	return nil
}

// GetBatchesByStorefront retrieves batches for a specific storefront
func (r *BarcodeGenerationBatchRepositoryImpl) GetBatchesByStorefront(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.BarcodeGenerationBatch, error) {
	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	qb, err := r.NewQueryBuilder(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to create query builder: %w", err)
	}

	query, args := qb.
		Select("*").
		From("barcode_generation_batches").
		Where("deleted_at IS NULL").
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var batches []*entity.BarcodeGenerationBatch
	err = db.SelectContext(ctx, &batches, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("storefront_id", storefrontID.String()).Msg("Failed to get batches by storefront")
		return nil, fmt.Errorf("failed to get batches by storefront: %w", err)
	}

	for _, batch := range batches {
		batch.ComputeFields()
	}

	return batches, nil
}

// GetBatchesByProduct retrieves batches for a specific product
func (r *BarcodeGenerationBatchRepositoryImpl) GetBatchesByProduct(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*entity.BarcodeGenerationBatch, error) {
	query := `
		SELECT * FROM barcode_generation_batches 
		WHERE product_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	var batches []*entity.BarcodeGenerationBatch
	err := r.db.SelectContext(ctx, &batches, query, productID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Str("product_id", productID.String()).Msg("Failed to get batches by product")
		return nil, fmt.Errorf("failed to get batches by product: %w", err)
	}

	for _, batch := range batches {
		batch.ComputeFields()
	}

	return batches, nil
}

// GetBatchesByStatus retrieves batches by generation status
func (r *BarcodeGenerationBatchRepositoryImpl) GetBatchesByStatus(ctx context.Context, storefrontID uuid.UUID, status string, limit, offset int) ([]*entity.BarcodeGenerationBatch, error) {
	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	qb, err := r.NewQueryBuilder(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to create query builder: %w", err)
	}

	query, args := qb.
		Select("*").
		From("barcode_generation_batches").
		Where("generation_status = $1 AND deleted_at IS NULL", status).
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var batches []*entity.BarcodeGenerationBatch
	err = db.SelectContext(ctx, &batches, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("status", status).Msg("Failed to get batches by status")
		return nil, fmt.Errorf("failed to get batches by status: %w", err)
	}

	for _, batch := range batches {
		batch.ComputeFields()
	}

	return batches, nil
}

// GetBatchesByDateRange retrieves batches within a date range
func (r *BarcodeGenerationBatchRepositoryImpl) GetBatchesByDateRange(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*entity.BarcodeGenerationBatch, error) {
	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	qb, err := r.NewQueryBuilder(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to create query builder: %w", err)
	}

	query, args := qb.
		Select("*").
		From("barcode_generation_batches").
		Where("created_at >= $1 AND created_at <= $2 AND deleted_at IS NULL", startDate, endDate).
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var batches []*entity.BarcodeGenerationBatch
	err = db.SelectContext(ctx, &batches, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get batches by date range")
		return nil, fmt.Errorf("failed to get batches by date range: %w", err)
	}

	for _, batch := range batches {
		batch.ComputeFields()
	}

	return batches, nil
}

// GetWithFilters retrieves batches with filters and pagination
func (r *BarcodeGenerationBatchRepositoryImpl) GetWithFilters(ctx context.Context, filters *repository.BatchFilters) ([]*entity.BarcodeGenerationBatch, error) {
	if filters.StorefrontID == nil {
		return nil, fmt.Errorf("storefront_id is required")
	}

	db, err := r.GetDB(ctx, *filters.StorefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	qb, err := r.NewQueryBuilder(ctx, *filters.StorefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to create query builder: %w", err)
	}

	qb = r.applyFilters(qb, filters)

	// Apply sorting
	sortBy := "created_at"
	if filters.SortBy != "" {
		sortBy = filters.SortBy
	}
	sortDirection := "DESC"
	if filters.SortDirection != "" {
		sortDirection = filters.SortDirection
	}
	qb = qb.OrderBy(sortBy, sortDirection)

	// Apply pagination
	if filters.PageSize > 0 {
		qb = qb.Limit(filters.PageSize)
		if filters.Page > 1 {
			qb = qb.Offset((filters.Page - 1) * filters.PageSize)
		}
	}

	query, args := qb.Build()

	var batches []*entity.BarcodeGenerationBatch
	err = db.SelectContext(ctx, &batches, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get batches with filters")
		return nil, fmt.Errorf("failed to get batches with filters: %w", err)
	}

	for _, batch := range batches {
		batch.ComputeFields()
	}

	return batches, nil
}

// Count counts batches with optional filters
func (r *BarcodeGenerationBatchRepositoryImpl) Count(ctx context.Context, filters *repository.BatchFilters) (int, error) {
	if filters.StorefrontID == nil {
		return 0, fmt.Errorf("storefront_id is required for count")
	}

	db, err := r.GetDB(ctx, *filters.StorefrontID)
	if err != nil {
		return 0, fmt.Errorf("failed to get database connection: %w", err)
	}

	qb, err := r.NewQueryBuilder(ctx, *filters.StorefrontID)
	if err != nil {
		return 0, fmt.Errorf("failed to create query builder: %w", err)
	}

	qb = r.applyFilters(qb, filters)
	query, args := qb.BuildCount()

	var count int
	err = db.GetContext(ctx, &count, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to count batches")
		return 0, fmt.Errorf("failed to count batches: %w", err)
	}

	return count, nil
}

// Delete soft deletes a batch record
func (r *BarcodeGenerationBatchRepositoryImpl) Delete(ctx context.Context, batchID uuid.UUID) error {
	// First get the batch to determine storefront
	batch, err := r.GetBatch(ctx, batchID)
	if err != nil {
		return fmt.Errorf("failed to get batch: %w", err)
	}
	if batch == nil {
		return fmt.Errorf("batch not found")
	}

	db, err := r.GetDB(ctx, batch.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	now := time.Now()
	query := `UPDATE barcode_generation_batches SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, query, now, now, batchID)
	if err != nil {
		r.logger.Error().Err(err).Str("id", batchID.String()).Msg("Failed to delete batch")
		return fmt.Errorf("failed to delete batch: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("batch not found")
	}

	r.logger.Info().Str("id", batchID.String()).Msg("Batch deleted")
	return nil
}

// GetBatchStatistics retrieves batch generation statistics
func (r *BarcodeGenerationBatchRepositoryImpl) GetBatchStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time) (*repository.BatchStatistics, error) {
	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT 
			COUNT(*) as total_batches,
			COUNT(CASE WHEN generation_status = 'completed' THEN 1 END) as completed_batches,
			COUNT(CASE WHEN generation_status = 'partial' THEN 1 END) as partial_batches,
			COUNT(CASE WHEN generation_status = 'failed' THEN 1 END) as failed_batches,
			SUM(generated_quantity) as total_generated,
			SUM(requested_quantity) as total_requested,
			SUM(collision_count) as total_collisions,
			AVG(average_generation_time_ms) as avg_generation_time,
			AVG(requested_quantity) as avg_batch_size
		FROM barcode_generation_batches 
		WHERE storefront_id = $1 
			AND created_at >= $2 
			AND created_at <= $3 
			AND deleted_at IS NULL`

	var stats repository.BatchStatistics
	var avgGenerationTime, avgBatchSize sql.NullFloat64

	row := db.QueryRowContext(ctx, query, storefrontID, startDate, endDate)
	err = row.Scan(
		&stats.TotalBatches,
		&stats.CompletedBatches,
		&stats.PartialBatches,
		&stats.FailedBatches,
		&stats.TotalBarcodesGenerated,
		&stats.TotalBarcodesRequested,
		&stats.TotalCollisions,
		&avgGenerationTime,
		&avgBatchSize,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get batch statistics")
		return nil, fmt.Errorf("failed to get batch statistics: %w", err)
	}

	if avgGenerationTime.Valid {
		stats.AverageGenerationTime = avgGenerationTime.Float64
	}
	if avgBatchSize.Valid {
		stats.AverageBatchSize = avgBatchSize.Float64
	}

	// Calculate rates
	if stats.TotalBarcodesRequested > 0 {
		stats.SuccessRate = float64(stats.TotalBarcodesGenerated) / float64(stats.TotalBarcodesRequested) * 100
		totalAttempts := stats.TotalBarcodesGenerated + stats.TotalCollisions
		if totalAttempts > 0 {
			stats.CollisionRate = float64(stats.TotalCollisions) / float64(totalAttempts) * 100
		}
	}

	// Calculate throughput (assuming data is for recent period)
	hours := endDate.Sub(startDate).Hours()
	if hours > 0 {
		stats.ThroughputPerHour = float64(stats.TotalBarcodesGenerated) / hours
	}

	return &stats, nil
}

// GenerateBatchNumber generates a unique batch number
func (r *BarcodeGenerationBatchRepositoryImpl) GenerateBatchNumber(ctx context.Context, storefrontID uuid.UUID) (string, error) {
	// Generate format: BCB-YYYYMMDD-NNNN (Barcode Batch - Date - Sequence)
	now := time.Now()
	dateStr := now.Format("20060102")

	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	// Get the next sequence number for today and this storefront
	query := `
		SELECT COALESCE(MAX(
			CASE WHEN batch_number ~ '^BCB-' || $1 || '-[0-9]{4}$' 
			THEN CAST(SUBSTRING(batch_number FROM 15) AS INTEGER) 
			ELSE 0 END
		), 0) + 1
		FROM barcode_generation_batches 
		WHERE storefront_id = $2 AND deleted_at IS NULL`

	var nextNum int
	err = db.GetContext(ctx, &nextNum, query, dateStr, storefrontID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get next batch number")
		return "", fmt.Errorf("failed to get next batch number: %w", err)
	}

	batchNumber := fmt.Sprintf("BCB-%s-%04d", dateStr, nextNum)
	return batchNumber, nil
}

// applyFilters applies filters to the query builder
func (r *BarcodeGenerationBatchRepositoryImpl) applyFilters(qb QueryBuilder, filters *repository.BatchFilters) QueryBuilder {
	qb = qb.Select("*").From("barcode_generation_batches").Where("deleted_at IS NULL")

	if filters.StorefrontID != nil {
		qb = qb.TenantWhere(*filters.StorefrontID)
	}

	if filters.ProductID != nil {
		qb = qb.Where("product_id = $1", *filters.ProductID)
	}

	if filters.RequestedBy != nil {
		qb = qb.Where("requested_by = $1", *filters.RequestedBy)
	}

	if filters.GenerationStatus != nil {
		qb = qb.Where("generation_status = $1", *filters.GenerationStatus)
	}

	if filters.CreatedAfter != nil {
		qb = qb.Where("created_at >= $1", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		qb = qb.Where("created_at <= $1", *filters.CreatedBefore)
	}

	if filters.CompletedAfter != nil {
		qb = qb.Where("generation_completed_at >= $1", *filters.CompletedAfter)
	}

	if filters.CompletedBefore != nil {
		qb = qb.Where("generation_completed_at <= $1", *filters.CompletedBefore)
	}

	if filters.MinQuantity != nil {
		qb = qb.Where("requested_quantity >= $1", *filters.MinQuantity)
	}

	if filters.MaxQuantity != nil {
		qb = qb.Where("requested_quantity <= $1", *filters.MaxQuantity)
	}

	if filters.IntendedRecipient != nil {
		qb = qb.Where("intended_recipient = $1", *filters.IntendedRecipient)
	}

	if filters.BatchNumber != nil {
		qb = qb.Where("batch_number = $1", *filters.BatchNumber)
	}

	if filters.Search != nil && *filters.Search != "" {
		qb = qb.Where("(batch_number ILIKE $1 OR intended_recipient ILIKE $1)", "%"+*filters.Search+"%")
	}

	return qb
}
