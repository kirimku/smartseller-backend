package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/rs/zerolog"
)

// WarrantyBarcodeRepositoryImpl implements the WarrantyBarcodeRepository interface
type WarrantyBarcodeRepositoryImpl struct {
	*BaseRepository
	logger zerolog.Logger
}

// NewWarrantyBarcodeRepository creates a new warranty barcode repository
func NewWarrantyBarcodeRepository(
	db *sqlx.DB,
	tenantResolver tenant.TenantResolver,
	logger zerolog.Logger,
) repository.WarrantyBarcodeRepository {
	return &WarrantyBarcodeRepositoryImpl{
		BaseRepository: NewBaseRepository(db, tenantResolver),
		logger:         logger.With().Str("repository", "warranty_barcode").Logger(),
	}
}

// Create creates a new warranty barcode
func (r *WarrantyBarcodeRepositoryImpl) Create(ctx context.Context, barcode *entity.WarrantyBarcode) error {
	db, err := r.GetDB(ctx, barcode.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		INSERT INTO warranty_barcodes (
			id, barcode_number, qr_code_data, product_id, storefront_id,
			warranty_period_months, expiry_date, created_by, batch_id, batch_number, 
			distributed_at, distributed_to, distribution_notes, status, activated_at,
			customer_id, purchase_date, purchase_location, purchase_invoice,
			generation_method, entropy_bits, generation_attempt, collision_checked,
			generated_at, created_at, updated_at
		) VALUES (
			:id, :barcode_number, :qr_code_data, :product_id, :storefront_id,
			:warranty_period_months, :expiry_date, :created_by, :batch_id, :batch_number,
			:distributed_at, :distributed_to, :distribution_notes, :status, :activated_at,
			:customer_id, :purchase_date, :purchase_location, :purchase_invoice,
			:generation_method, :entropy_bits, :generation_attempt, :collision_checked,
			:generated_at, :created_at, :updated_at
		)`

	_, err = db.NamedExecContext(ctx, query, barcode)
	if err != nil {
		r.logger.Error().Err(err).Str("barcode_number", barcode.BarcodeNumber).Msg("Failed to create warranty barcode")
		return fmt.Errorf("failed to create warranty barcode: %w", err)
	}

	r.logger.Info().Str("barcode_number", barcode.BarcodeNumber).Str("id", barcode.ID.String()).Msg("Warranty barcode created")
	return nil
}

// CreateBatch creates multiple warranty barcodes in a single operation
func (r *WarrantyBarcodeRepositoryImpl) CreateBatch(ctx context.Context, barcodes []*entity.WarrantyBarcode) error {
	if len(barcodes) == 0 {
		return nil
	}

	storefrontID := barcodes[0].StorefrontID

	return r.ExecuteInTransaction(ctx, storefrontID, func(tx *sqlx.Tx) error {
		query := `
			INSERT INTO warranty_barcodes (
				id, barcode_number, qr_code_data, product_id, storefront_id,
				warranty_period_months, expiry_date, created_by, batch_id, batch_number,
				distributed_at, distributed_to, distribution_notes, status, activated_at,
				customer_id, purchase_date, purchase_location, purchase_invoice,
				generation_method, entropy_bits, generation_attempt, collision_checked,
				generated_at, created_at, updated_at
			) VALUES (
				:id, :barcode_number, :qr_code_data, :product_id, :storefront_id,
				:warranty_period_months, :expiry_date, :created_by, :batch_id, :batch_number,
				:distributed_at, :distributed_to, :distribution_notes, :status, :activated_at,
				:customer_id, :purchase_date, :purchase_location, :purchase_invoice,
				:generation_method, :entropy_bits, :generation_attempt, :collision_checked,
				:generated_at, :created_at, :updated_at
			)`

		for _, barcode := range barcodes {
			_, err := tx.NamedExecContext(ctx, query, barcode)
			if err != nil {
				r.logger.Error().Err(err).Str("barcode_number", barcode.BarcodeNumber).Msg("Failed to create barcode in batch")
				return fmt.Errorf("failed to create barcode %s in batch: %w", barcode.BarcodeNumber, err)
			}
		}

		r.logger.Info().Int("count", len(barcodes)).Msg("Batch of warranty barcodes created")
		return nil
	})
}

// GetByID retrieves a warranty barcode by its ID
func (r *WarrantyBarcodeRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.WarrantyBarcode, error) {
	// For getting by ID, we need to determine the storefront first
	// This requires a cross-tenant query to find the barcode
	query := `SELECT * FROM warranty_barcodes WHERE id = $1 AND deleted_at IS NULL`

	var barcode entity.WarrantyBarcode
	err := r.db.GetContext(ctx, &barcode, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", id.String()).Msg("Failed to get warranty barcode by ID")
		return nil, fmt.Errorf("failed to get warranty barcode by ID: %w", err)
	}

	barcode.ComputeFields()
	return &barcode, nil
}

// GetByBarcodeNumber retrieves a warranty barcode by its barcode number
func (r *WarrantyBarcodeRepositoryImpl) GetByBarcodeNumber(ctx context.Context, barcodeNumber string) (*entity.WarrantyBarcode, error) {
	// This is a public lookup, so we need to search across all storefronts
	query := `SELECT * FROM warranty_barcodes WHERE barcode_number = $1 AND deleted_at IS NULL`

	var barcode entity.WarrantyBarcode
	err := r.db.GetContext(ctx, &barcode, query, barcodeNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("barcode_number", barcodeNumber).Msg("Failed to get warranty barcode by number")
		return nil, fmt.Errorf("failed to get warranty barcode by number: %w", err)
	}

	barcode.ComputeFields()
	return &barcode, nil
}

// CheckUniqueness checks if a barcode number is unique
func (r *WarrantyBarcodeRepositoryImpl) CheckUniqueness(ctx context.Context, barcodeNumber string) (bool, error) {
	query := `SELECT COUNT(*) FROM warranty_barcodes WHERE barcode_number = $1`

	var count int
	err := r.db.GetContext(ctx, &count, query, barcodeNumber)
	if err != nil {
		r.logger.Error().Err(err).Str("barcode_number", barcodeNumber).Msg("Failed to check barcode uniqueness")
		return false, fmt.Errorf("failed to check barcode uniqueness: %w", err)
	}

	return count == 0, nil
}

// GetByProductID retrieves warranty barcodes for a specific product
func (r *WarrantyBarcodeRepositoryImpl) GetByProductID(ctx context.Context, productID uuid.UUID, limit, offset int) ([]*entity.WarrantyBarcode, error) {
	// Need to get storefront_id for this product first
	var storefrontID uuid.UUID
	err := r.db.GetContext(ctx, &storefrontID, `SELECT storefront_id FROM products WHERE id = $1`, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront for product: %w", err)
	}

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
		From("warranty_barcodes").
		Where("product_id = $1 AND deleted_at IS NULL", productID).
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var barcodes []*entity.WarrantyBarcode
	err = db.SelectContext(ctx, &barcodes, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("product_id", productID.String()).Msg("Failed to get barcodes by product ID")
		return nil, fmt.Errorf("failed to get barcodes by product ID: %w", err)
	}

	for _, barcode := range barcodes {
		barcode.ComputeFields()
	}

	return barcodes, nil
}

// GetByStorefrontID retrieves warranty barcodes for a specific storefront
func (r *WarrantyBarcodeRepositoryImpl) GetByStorefrontID(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyBarcode, error) {
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
		From("warranty_barcodes").
		Where("deleted_at IS NULL").
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var barcodes []*entity.WarrantyBarcode
	err = db.SelectContext(ctx, &barcodes, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("storefront_id", storefrontID.String()).Msg("Failed to get barcodes by storefront ID")
		return nil, fmt.Errorf("failed to get barcodes by storefront ID: %w", err)
	}

	for _, barcode := range barcodes {
		barcode.ComputeFields()
	}

	return barcodes, nil
}

// GetByBatchID retrieves warranty barcodes from a specific batch
func (r *WarrantyBarcodeRepositoryImpl) GetByBatchID(ctx context.Context, batchID uuid.UUID, limit, offset int) ([]*entity.WarrantyBarcode, error) {
	// Get storefront_id from batch
	var storefrontID uuid.UUID
	err := r.db.GetContext(ctx, &storefrontID, `SELECT storefront_id FROM barcode_generation_batches WHERE id = $1`, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storefront for batch: %w", err)
	}

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
		From("warranty_barcodes").
		Where("batch_id = $1 AND deleted_at IS NULL", batchID).
		TenantWhere(storefrontID).
		OrderBy("created_at", "ASC").
		Limit(limit).
		Offset(offset).
		Build()

	var barcodes []*entity.WarrantyBarcode
	err = db.SelectContext(ctx, &barcodes, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("batch_id", batchID.String()).Msg("Failed to get barcodes by batch ID")
		return nil, fmt.Errorf("failed to get barcodes by batch ID: %w", err)
	}

	for _, barcode := range barcodes {
		barcode.ComputeFields()
	}

	return barcodes, nil
}

// GetByStatus retrieves warranty barcodes by status
func (r *WarrantyBarcodeRepositoryImpl) GetByStatus(ctx context.Context, storefrontID uuid.UUID, status entity.BarcodeStatus, limit, offset int) ([]*entity.WarrantyBarcode, error) {
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
		From("warranty_barcodes").
		Where("status = $1 AND deleted_at IS NULL", status).
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var barcodes []*entity.WarrantyBarcode
	err = db.SelectContext(ctx, &barcodes, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("status", string(status)).Msg("Failed to get barcodes by status")
		return nil, fmt.Errorf("failed to get barcodes by status: %w", err)
	}

	for _, barcode := range barcodes {
		barcode.ComputeFields()
	}

	return barcodes, nil
}

// GetExpiringSoon retrieves warranty barcodes expiring within the specified number of days
func (r *WarrantyBarcodeRepositoryImpl) GetExpiringSoon(ctx context.Context, storefrontID uuid.UUID, days int, limit, offset int) ([]*entity.WarrantyBarcode, error) {
	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	qb, err := r.NewQueryBuilder(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to create query builder: %w", err)
	}

	expiryDate := time.Now().AddDate(0, 0, days)
	query, args := qb.
		Select("*").
		From("warranty_barcodes").
		Where("warranty_end_date <= $1 AND warranty_end_date > CURRENT_DATE AND status = $2 AND deleted_at IS NULL", expiryDate, entity.BarcodeStatusActivated).
		TenantWhere(storefrontID).
		OrderBy("warranty_end_date", "ASC").
		Limit(limit).
		Offset(offset).
		Build()

	var barcodes []*entity.WarrantyBarcode
	err = db.SelectContext(ctx, &barcodes, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Int("days", days).Msg("Failed to get expiring barcodes")
		return nil, fmt.Errorf("failed to get expiring barcodes: %w", err)
	}

	for _, barcode := range barcodes {
		barcode.ComputeFields()
	}

	return barcodes, nil
}

// Update updates an existing warranty barcode
func (r *WarrantyBarcodeRepositoryImpl) Update(ctx context.Context, barcode *entity.WarrantyBarcode) error {
	db, err := r.GetDB(ctx, barcode.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	barcode.UpdatedAt = time.Now()

	query := `
		UPDATE warranty_barcodes SET 
			warranty_start_date = :warranty_start_date,
			warranty_end_date = :warranty_end_date,
			warranty_period_months = :warranty_period_months,
			status = :status,
			activated_at = :activated_at,
			activated_by = :activated_by,
			updated_at = :updated_at
		WHERE id = :id AND storefront_id = :storefront_id AND deleted_at IS NULL`

	result, err := db.NamedExecContext(ctx, query, barcode)
	if err != nil {
		r.logger.Error().Err(err).Str("id", barcode.ID.String()).Msg("Failed to update warranty barcode")
		return fmt.Errorf("failed to update warranty barcode: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("warranty barcode not found or not updated")
	}

	r.logger.Info().Str("id", barcode.ID.String()).Msg("Warranty barcode updated")
	return nil
}

// Activate activates a warranty barcode
func (r *WarrantyBarcodeRepositoryImpl) Activate(ctx context.Context, barcodeNumber string, activatedBy uuid.UUID) error {
	// First get the barcode to determine storefront
	barcode, err := r.GetByBarcodeNumber(ctx, barcodeNumber)
	if err != nil {
		return fmt.Errorf("failed to get barcode: %w", err)
	}
	if barcode == nil {
		return fmt.Errorf("barcode not found")
	}

	db, err := r.GetDB(ctx, barcode.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	now := time.Now()
	query := `
		UPDATE warranty_barcodes SET 
			status = $1,
			activated_at = $2,
			updated_at = $3
		WHERE barcode_number = $4 AND status = $5 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, query, entity.BarcodeStatusActivated, now, now, barcodeNumber, entity.BarcodeStatusGenerated)
	if err != nil {
		r.logger.Error().Err(err).Str("barcode_number", barcodeNumber).Msg("Failed to activate warranty barcode")
		return fmt.Errorf("failed to activate warranty barcode: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("barcode not found or already activated")
	}

	r.logger.Info().Str("barcode_number", barcodeNumber).Str("activated_by", activatedBy.String()).Msg("Warranty barcode activated")
	return nil
}

// Deactivate deactivates a warranty barcode
func (r *WarrantyBarcodeRepositoryImpl) Deactivate(ctx context.Context, barcodeNumber string, reason string) error {
	// First get the barcode to determine storefront
	barcode, err := r.GetByBarcodeNumber(ctx, barcodeNumber)
	if err != nil {
		return fmt.Errorf("failed to get barcode: %w", err)
	}
	if barcode == nil {
		return fmt.Errorf("barcode not found")
	}

	db, err := r.GetDB(ctx, barcode.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	now := time.Now()
	query := `
		UPDATE warranty_barcodes SET 
			status = $1,
			updated_at = $2
		WHERE barcode_number = $3 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, query, entity.BarcodeStatusExpired, now, barcodeNumber)
	if err != nil {
		r.logger.Error().Err(err).Str("barcode_number", barcodeNumber).Msg("Failed to deactivate warranty barcode")
		return fmt.Errorf("failed to deactivate warranty barcode: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("barcode not found")
	}

	r.logger.Info().Str("barcode_number", barcodeNumber).Str("reason", reason).Msg("Warranty barcode deactivated")
	return nil
}

// Delete soft deletes a warranty barcode
func (r *WarrantyBarcodeRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// First get the barcode to determine storefront
	barcode, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get barcode: %w", err)
	}
	if barcode == nil {
		return fmt.Errorf("barcode not found")
	}

	db, err := r.GetDB(ctx, barcode.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	now := time.Now()
	query := `UPDATE warranty_barcodes SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.String()).Msg("Failed to delete warranty barcode")
		return fmt.Errorf("failed to delete warranty barcode: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("barcode not found")
	}

	r.logger.Info().Str("id", id.String()).Msg("Warranty barcode deleted")
	return nil
}

// Count counts warranty barcodes with optional filters
func (r *WarrantyBarcodeRepositoryImpl) Count(ctx context.Context, filters *repository.WarrantyBarcodeFilters) (int, error) {
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
		r.logger.Error().Err(err).Msg("Failed to count warranty barcodes")
		return 0, fmt.Errorf("failed to count warranty barcodes: %w", err)
	}

	return count, nil
}

// GetWithFilters retrieves warranty barcodes with filters and pagination
func (r *WarrantyBarcodeRepositoryImpl) GetWithFilters(ctx context.Context, filters *repository.WarrantyBarcodeFilters) ([]*entity.WarrantyBarcode, error) {
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
		sortDirection = strings.ToUpper(filters.SortDirection)
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

	var barcodes []*entity.WarrantyBarcode
	err = db.SelectContext(ctx, &barcodes, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get warranty barcodes with filters")
		return nil, fmt.Errorf("failed to get warranty barcodes with filters: %w", err)
	}

	for _, barcode := range barcodes {
		barcode.ComputeFields()
	}

	return barcodes, nil
}

// applyFilters applies filters to the query builder
func (r *WarrantyBarcodeRepositoryImpl) applyFilters(qb QueryBuilder, filters *repository.WarrantyBarcodeFilters) QueryBuilder {
	qb = qb.Select("*").From("warranty_barcodes").Where("deleted_at IS NULL")

	if filters.StorefrontID != nil {
		qb = qb.TenantWhere(*filters.StorefrontID)
	}

	if filters.ProductID != nil {
		qb = qb.Where("product_id = $1", *filters.ProductID)
	}

	if filters.BatchID != nil {
		qb = qb.Where("batch_id = $1", *filters.BatchID)
	}

	if filters.Status != nil {
		qb = qb.Where("status = $1", *filters.Status)
	}

	if filters.CreatedBy != nil {
		qb = qb.Where("created_by = $1", *filters.CreatedBy)
	}

	if filters.ActivatedBy != nil {
		qb = qb.Where("activated_by = $1", *filters.ActivatedBy)
	}

	if filters.CreatedAfter != nil {
		qb = qb.Where("created_at >= $1", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		qb = qb.Where("created_at <= $1", *filters.CreatedBefore)
	}

	if filters.ActivatedAfter != nil {
		qb = qb.Where("activation_date >= $1", *filters.ActivatedAfter)
	}

	if filters.ActivatedBefore != nil {
		qb = qb.Where("activation_date <= $1", *filters.ActivatedBefore)
	}

	if filters.ExpiresAfter != nil {
		qb = qb.Where("warranty_end_date >= $1", *filters.ExpiresAfter)
	}

	if filters.ExpiresBefore != nil {
		qb = qb.Where("warranty_end_date <= $1", *filters.ExpiresBefore)
	}

	if filters.GenerationMethod != nil {
		qb = qb.Where("generation_method = $1", *filters.GenerationMethod)
	}

	if filters.BatchNumber != nil {
		qb = qb.Where("batch_number = $1", *filters.BatchNumber)
	}

	if filters.Search != nil && *filters.Search != "" {
		qb = qb.Where("(barcode_number ILIKE $1 OR batch_number ILIKE $1)", "%"+*filters.Search+"%")
	}

	return qb
}

// GetGenerationStats retrieves generation statistics (placeholder implementation)
func (r *WarrantyBarcodeRepositoryImpl) GetGenerationStats(ctx context.Context, req *repository.GenerationStatsRequest) (*repository.GenerationStatsResponse, error) {
	// This would require complex aggregation queries
	// For now, return a basic implementation
	return &repository.GenerationStatsResponse{
		TotalGenerated:        0,
		GenerationRate:        0,
		CollisionCount:        0,
		CollisionRate:         0,
		AverageGenerationTime: 0,
		EntropyUtilization:    0,
		SecurityStatus:        "HEALTHY",
		PeriodStatistics:      []*repository.PeriodStats{},
	}, nil
}

// GetUsageStatistics retrieves usage statistics for analytics
func (r *WarrantyBarcodeRepositoryImpl) GetUsageStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time) (*repository.WarrantyUsageStats, error) {
	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT 
			COUNT(*) as total_warranties,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_warranties,
			COUNT(CASE WHEN status = 'used' THEN 1 END) as used_warranties,
			COUNT(CASE WHEN status = 'expired' THEN 1 END) as expired_warranties,
			COUNT(CASE WHEN status = 'revoked' THEN 1 END) as revoked_warranties,
			AVG(EXTRACT(DAY FROM (warranty_end_date - warranty_start_date))) as average_lifespan
		FROM warranty_barcodes 
		WHERE storefront_id = $1 
			AND created_at >= $2 
			AND created_at <= $3 
			AND deleted_at IS NULL`

	var stats repository.WarrantyUsageStats
	row := db.QueryRowContext(ctx, query, storefrontID, startDate, endDate)

	var avgLifespan sql.NullFloat64
	err = row.Scan(
		&stats.TotalWarranties,
		&stats.ActiveWarranties,
		&stats.UsedWarranties,
		&stats.ExpiredWarranties,
		&stats.RevokedWarranties,
		&avgLifespan,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get warranty usage statistics")
		return nil, fmt.Errorf("failed to get warranty usage statistics: %w", err)
	}

	if avgLifespan.Valid {
		stats.AverageLifespan = int(avgLifespan.Float64)
	}

	// Calculate rates
	if stats.TotalWarranties > 0 {
		stats.ActivationRate = float64(stats.ActiveWarranties) / float64(stats.TotalWarranties) * 100
		stats.ExpirationRate = float64(stats.ExpiredWarranties) / float64(stats.TotalWarranties) * 100
	}

	return &stats, nil
}
