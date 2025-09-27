package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/rs/zerolog"
)

// WarrantyClaimRepositoryImpl implements the WarrantyClaimRepository interface
type WarrantyClaimRepositoryImpl struct {
	*BaseRepository
	logger zerolog.Logger
}

// NewWarrantyClaimRepository creates a new warranty claim repository
func NewWarrantyClaimRepository(
	db *sqlx.DB,
	tenantResolver tenant.TenantResolver,
	logger zerolog.Logger,
) repository.WarrantyClaimRepository {
	return &WarrantyClaimRepositoryImpl{
		BaseRepository: NewBaseRepository(db, tenantResolver),
		logger:         logger.With().Str("repository", "warranty_claim").Logger(),
	}
}

// Create creates a new warranty claim
func (r *WarrantyClaimRepositoryImpl) Create(ctx context.Context, claim *entity.WarrantyClaim) error {
	db, err := r.GetDB(ctx, claim.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		INSERT INTO warranty_claims (
			id, claim_number, barcode_id, customer_id, product_id, storefront_id,
			issue_description, issue_category, issue_date, severity,
			claim_date, validated_at, completed_at,
			status, previous_status, status_updated_at, status_updated_by,
			validated_by, assigned_technician_id, estimated_completion_date, actual_completion_date,
			resolution_type, repair_notes, replacement_product_id, refund_amount,
			repair_cost, shipping_cost, replacement_cost, total_cost,
			customer_name, customer_email, customer_phone,
			pickup_address,
			shipping_provider, tracking_number, estimated_delivery_date, actual_delivery_date, delivery_status,
			customer_notes, admin_notes, rejection_reason, internal_notes,
			priority, tags,
			customer_satisfaction_rating, customer_feedback, processing_time_hours,
			created_at, updated_at
		) VALUES (
			:id, :claim_number, :barcode_id, :customer_id, :product_id, :storefront_id,
			:issue_description, :issue_category, :issue_date, :severity,
			:claim_date, :validated_at, :completed_at,
			:status, :previous_status, :status_updated_at, :status_updated_by,
			:validated_by, :assigned_technician_id, :estimated_completion_date, :actual_completion_date,
			:resolution_type, :repair_notes, :replacement_product_id, :refund_amount,
			:repair_cost, :shipping_cost, :replacement_cost, :total_cost,
			:customer_name, :customer_email, :customer_phone,
			:pickup_address,
			:shipping_provider, :tracking_number, :estimated_delivery_date, :actual_delivery_date, :delivery_status,
			:customer_notes, :admin_notes, :rejection_reason, :internal_notes,
			:priority, :tags,
			:customer_satisfaction_rating, :customer_feedback, :processing_time_hours,
			:created_at, :updated_at
		)`

	_, err = db.NamedExecContext(ctx, query, claim)
	if err != nil {
		r.logger.Error().Err(err).Str("claim_number", claim.ClaimNumber).Msg("Failed to create warranty claim")
		return fmt.Errorf("failed to create warranty claim: %w", err)
	}

	r.logger.Info().Str("claim_number", claim.ClaimNumber).Str("id", claim.ID.String()).Msg("Warranty claim created")
	return nil
}

// GetByID retrieves a warranty claim by its ID
func (r *WarrantyClaimRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.WarrantyClaim, error) {
	// For getting by ID, we need to search across all storefronts first
	query := `SELECT * FROM warranty_claims WHERE id = $1 AND deleted_at IS NULL`

	var claim entity.WarrantyClaim
	err := r.db.GetContext(ctx, &claim, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", id.String()).Msg("Failed to get warranty claim by ID")
		return nil, fmt.Errorf("failed to get warranty claim by ID: %w", err)
	}

	claim.ComputeFields()
	return &claim, nil
}

// GetByClaimNumber retrieves a warranty claim by its claim number
func (r *WarrantyClaimRepositoryImpl) GetByClaimNumber(ctx context.Context, claimNumber string) (*entity.WarrantyClaim, error) {
	query := `SELECT * FROM warranty_claims WHERE claim_number = $1 AND deleted_at IS NULL`

	var claim entity.WarrantyClaim
	err := r.db.GetContext(ctx, &claim, query, claimNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("claim_number", claimNumber).Msg("Failed to get warranty claim by number")
		return nil, fmt.Errorf("failed to get warranty claim by number: %w", err)
	}

	claim.ComputeFields()
	return &claim, nil
}

// GetByBarcodeID retrieves warranty claims for a specific barcode
func (r *WarrantyClaimRepositoryImpl) GetByBarcodeID(ctx context.Context, barcodeID uuid.UUID) ([]*entity.WarrantyClaim, error) {
	query := `SELECT * FROM warranty_claims WHERE barcode_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`

	var claims []*entity.WarrantyClaim
	err := r.db.SelectContext(ctx, &claims, query, barcodeID)
	if err != nil {
		r.logger.Error().Err(err).Str("barcode_id", barcodeID.String()).Msg("Failed to get warranty claims by barcode ID")
		return nil, fmt.Errorf("failed to get warranty claims by barcode ID: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// GetByCustomerID retrieves warranty claims for a specific customer
func (r *WarrantyClaimRepositoryImpl) GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error) {
	query := `
		SELECT * FROM warranty_claims 
		WHERE customer_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	var claims []*entity.WarrantyClaim
	err := r.db.SelectContext(ctx, &claims, query, customerID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Str("customer_id", customerID.String()).Msg("Failed to get warranty claims by customer ID")
		return nil, fmt.Errorf("failed to get warranty claims by customer ID: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// GetByStorefrontID retrieves warranty claims for a specific storefront
func (r *WarrantyClaimRepositoryImpl) GetByStorefrontID(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error) {
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
		From("warranty_claims").
		Where("deleted_at IS NULL").
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var claims []*entity.WarrantyClaim
	err = db.SelectContext(ctx, &claims, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("storefront_id", storefrontID.String()).Msg("Failed to get warranty claims by storefront ID")
		return nil, fmt.Errorf("failed to get warranty claims by storefront ID: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// GetByStatus retrieves warranty claims by status
func (r *WarrantyClaimRepositoryImpl) GetByStatus(ctx context.Context, storefrontID uuid.UUID, status entity.ClaimStatus, limit, offset int) ([]*entity.WarrantyClaim, error) {
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
		From("warranty_claims").
		Where("status = $1 AND deleted_at IS NULL", status).
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var claims []*entity.WarrantyClaim
	err = db.SelectContext(ctx, &claims, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Str("status", string(status)).Msg("Failed to get warranty claims by status")
		return nil, fmt.Errorf("failed to get warranty claims by status: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// GetByTechnician retrieves warranty claims assigned to a specific technician
func (r *WarrantyClaimRepositoryImpl) GetByTechnician(ctx context.Context, technicianID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error) {
	query := `
		SELECT * FROM warranty_claims 
		WHERE assigned_technician_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	var claims []*entity.WarrantyClaim
	err := r.db.SelectContext(ctx, &claims, query, technicianID, limit, offset)
	if err != nil {
		r.logger.Error().Err(err).Str("technician_id", technicianID.String()).Msg("Failed to get warranty claims by technician")
		return nil, fmt.Errorf("failed to get warranty claims by technician: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// GetPendingClaims retrieves all pending warranty claims
func (r *WarrantyClaimRepositoryImpl) GetPendingClaims(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error) {
	return r.GetByStatus(ctx, storefrontID, entity.ClaimStatusPending, limit, offset)
}

// GetOverdueClaims retrieves warranty claims that are overdue
func (r *WarrantyClaimRepositoryImpl) GetOverdueClaims(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error) {
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
		From("warranty_claims").
		Where(`
			estimated_completion_date IS NOT NULL 
			AND estimated_completion_date < NOW() 
			AND status NOT IN ($1, $2, $3, $4) 
			AND deleted_at IS NULL
		`, entity.ClaimStatusCompleted, entity.ClaimStatusCancelled, entity.ClaimStatusRejected, entity.ClaimStatusDelivered).
		TenantWhere(storefrontID).
		OrderBy("estimated_completion_date", "ASC").
		Limit(limit).
		Offset(offset).
		Build()

	var claims []*entity.WarrantyClaim
	err = db.SelectContext(ctx, &claims, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get overdue warranty claims")
		return nil, fmt.Errorf("failed to get overdue warranty claims: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// Update updates an existing warranty claim
func (r *WarrantyClaimRepositoryImpl) Update(ctx context.Context, claim *entity.WarrantyClaim) error {
	db, err := r.GetDB(ctx, claim.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	claim.UpdatedAt = time.Now()

	query := `
		UPDATE warranty_claims SET 
			issue_description = :issue_description,
			issue_category = :issue_category,
			issue_date = :issue_date,
			severity = :severity,
			validated_at = :validated_at,
			completed_at = :completed_at,
			status = :status,
			previous_status = :previous_status,
			status_updated_at = :status_updated_at,
			status_updated_by = :status_updated_by,
			validated_by = :validated_by,
			assigned_technician_id = :assigned_technician_id,
			estimated_completion_date = :estimated_completion_date,
			actual_completion_date = :actual_completion_date,
			resolution_type = :resolution_type,
			repair_notes = :repair_notes,
			replacement_product_id = :replacement_product_id,
			refund_amount = :refund_amount,
			repair_cost = :repair_cost,
			shipping_cost = :shipping_cost,
			replacement_cost = :replacement_cost,
			total_cost = :total_cost,
			customer_phone = :customer_phone,
			pickup_address = :pickup_address,
			shipping_provider = :shipping_provider,
			tracking_number = :tracking_number,
			estimated_delivery_date = :estimated_delivery_date,
			actual_delivery_date = :actual_delivery_date,
			delivery_status = :delivery_status,
			customer_notes = :customer_notes,
			admin_notes = :admin_notes,
			rejection_reason = :rejection_reason,
			internal_notes = :internal_notes,
			priority = :priority,
			tags = :tags,
			customer_satisfaction_rating = :customer_satisfaction_rating,
			customer_feedback = :customer_feedback,
			processing_time_hours = :processing_time_hours,
			updated_at = :updated_at
		WHERE id = :id AND storefront_id = :storefront_id AND deleted_at IS NULL`

	result, err := db.NamedExecContext(ctx, query, claim)
	if err != nil {
		r.logger.Error().Err(err).Str("id", claim.ID.String()).Msg("Failed to update warranty claim")
		return fmt.Errorf("failed to update warranty claim: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("warranty claim not found or not updated")
	}

	r.logger.Info().Str("id", claim.ID.String()).Msg("Warranty claim updated")
	return nil
}

// UpdateStatus updates the status of a warranty claim
func (r *WarrantyClaimRepositoryImpl) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ClaimStatus, updatedBy uuid.UUID, notes string) error {
	// First get the claim to determine storefront
	claim, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return fmt.Errorf("claim not found")
	}

	db, err := r.GetDB(ctx, claim.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	now := time.Now()
	previousStatus := claim.Status.String()

	query := `
		UPDATE warranty_claims SET 
			status = $1,
			previous_status = $2,
			status_updated_at = $3,
			status_updated_by = $4,
			admin_notes = COALESCE($5, admin_notes),
			updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, query, status, previousStatus, now, updatedBy,
		func() *string {
			if notes != "" {
				return &notes
			}
			return nil
		}(), now, id)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.String()).Str("status", string(status)).Msg("Failed to update claim status")
		return fmt.Errorf("failed to update claim status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("claim not found")
	}

	r.logger.Info().Str("id", id.String()).Str("status", string(status)).Str("updated_by", updatedBy.String()).Msg("Claim status updated")
	return nil
}

// Delete soft deletes a warranty claim
func (r *WarrantyClaimRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	// First get the claim to determine storefront
	claim, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return fmt.Errorf("claim not found")
	}

	db, err := r.GetDB(ctx, claim.StorefrontID)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	now := time.Now()
	query := `UPDATE warranty_claims SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`

	result, err := db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		r.logger.Error().Err(err).Str("id", id.String()).Msg("Failed to delete warranty claim")
		return fmt.Errorf("failed to delete warranty claim: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("claim not found")
	}

	r.logger.Info().Str("id", id.String()).Msg("Warranty claim deleted")
	return nil
}

// Count counts warranty claims with optional filters
func (r *WarrantyClaimRepositoryImpl) Count(ctx context.Context, filters *repository.WarrantyClaimFilters) (int, error) {
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
		r.logger.Error().Err(err).Msg("Failed to count warranty claims")
		return 0, fmt.Errorf("failed to count warranty claims: %w", err)
	}

	return count, nil
}

// GetWithFilters retrieves warranty claims with filters and pagination
func (r *WarrantyClaimRepositoryImpl) GetWithFilters(ctx context.Context, filters *repository.WarrantyClaimFilters) ([]*entity.WarrantyClaim, error) {
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

	var claims []*entity.WarrantyClaim
	err = db.SelectContext(ctx, &claims, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get warranty claims with filters")
		return nil, fmt.Errorf("failed to get warranty claims with filters: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// applyFilters applies filters to the query builder
func (r *WarrantyClaimRepositoryImpl) applyFilters(qb QueryBuilder, filters *repository.WarrantyClaimFilters) QueryBuilder {
	qb = qb.Select("*").From("warranty_claims").Where("deleted_at IS NULL")

	if filters.StorefrontID != nil {
		qb = qb.TenantWhere(*filters.StorefrontID)
	}

	if filters.CustomerID != nil {
		qb = qb.Where("customer_id = $1", *filters.CustomerID)
	}

	if filters.ProductID != nil {
		qb = qb.Where("product_id = $1", *filters.ProductID)
	}

	if filters.BarcodeID != nil {
		qb = qb.Where("barcode_id = $1", *filters.BarcodeID)
	}

	if filters.Status != nil {
		qb = qb.Where("status = $1", *filters.Status)
	}

	if filters.Severity != nil {
		qb = qb.Where("severity = $1", *filters.Severity)
	}

	if filters.Priority != nil {
		qb = qb.Where("priority = $1", *filters.Priority)
	}

	if filters.AssignedTechnicianID != nil {
		qb = qb.Where("assigned_technician_id = $1", *filters.AssignedTechnicianID)
	}

	if filters.ValidatedBy != nil {
		qb = qb.Where("validated_by = $1", *filters.ValidatedBy)
	}

	if filters.IssueCategory != nil {
		qb = qb.Where("issue_category = $1", *filters.IssueCategory)
	}

	if filters.ResolutionType != nil {
		qb = qb.Where("resolution_type = $1", *filters.ResolutionType)
	}

	if filters.CreatedAfter != nil {
		qb = qb.Where("created_at >= $1", *filters.CreatedAfter)
	}

	if filters.CreatedBefore != nil {
		qb = qb.Where("created_at <= $1", *filters.CreatedBefore)
	}

	if filters.ValidatedAfter != nil {
		qb = qb.Where("validated_at >= $1", *filters.ValidatedAfter)
	}

	if filters.ValidatedBefore != nil {
		qb = qb.Where("validated_at <= $1", *filters.ValidatedBefore)
	}

	if filters.CompletedAfter != nil {
		qb = qb.Where("completed_at >= $1", *filters.CompletedAfter)
	}

	if filters.CompletedBefore != nil {
		qb = qb.Where("completed_at <= $1", *filters.CompletedBefore)
	}

	if filters.MinCost != nil {
		qb = qb.Where("total_cost >= $1", *filters.MinCost)
	}

	if filters.MaxCost != nil {
		qb = qb.Where("total_cost <= $1", *filters.MaxCost)
	}

	if filters.CustomerEmail != nil {
		qb = qb.Where("LOWER(customer_email) = LOWER($1)", *filters.CustomerEmail)
	}

	if filters.ClaimNumber != nil {
		qb = qb.Where("claim_number = $1", *filters.ClaimNumber)
	}

	if filters.Search != nil && *filters.Search != "" {
		qb = qb.Where(`(
			claim_number ILIKE $1 OR 
			customer_name ILIKE $1 OR 
			customer_email ILIKE $1 OR 
			issue_description ILIKE $1
		)`, "%"+*filters.Search+"%")
	}

	if len(filters.Tags) > 0 {
		qb = qb.Where("tags && $1", filters.Tags)
	}

	if filters.IsOverdue != nil && *filters.IsOverdue {
		qb = qb.Where(`
			estimated_completion_date IS NOT NULL 
			AND estimated_completion_date < NOW() 
			AND status NOT IN ($1, $2, $3, $4)
		`, entity.ClaimStatusCompleted, entity.ClaimStatusCancelled, entity.ClaimStatusRejected, entity.ClaimStatusDelivered)
	}

	return qb
}

// GetClaimsByDateRange retrieves claims within a date range
func (r *WarrantyClaimRepositoryImpl) GetClaimsByDateRange(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*entity.WarrantyClaim, error) {
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
		From("warranty_claims").
		Where("created_at >= $1 AND created_at <= $2 AND deleted_at IS NULL", startDate, endDate).
		TenantWhere(storefrontID).
		OrderBy("created_at", "DESC").
		Limit(limit).
		Offset(offset).
		Build()

	var claims []*entity.WarrantyClaim
	err = db.SelectContext(ctx, &claims, query, args...)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get warranty claims by date range")
		return nil, fmt.Errorf("failed to get warranty claims by date range: %w", err)
	}

	for _, claim := range claims {
		claim.ComputeFields()
	}

	return claims, nil
}

// GetClaimStatistics retrieves claim statistics for analytics
func (r *WarrantyClaimRepositoryImpl) GetClaimStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time) (*repository.ClaimStatistics, error) {
	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Basic statistics query
	query := `
		SELECT 
			COUNT(*) as total_claims,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_claims,
			COUNT(CASE WHEN status = 'validated' THEN 1 END) as validated_claims,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_claims,
			COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected_claims,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_claims,
			AVG(CASE WHEN processing_time_hours IS NOT NULL THEN processing_time_hours END) as avg_resolution_time,
			AVG(repair_cost + shipping_cost + replacement_cost) as avg_repair_cost,
			AVG(CASE WHEN customer_satisfaction_rating IS NOT NULL THEN customer_satisfaction_rating END) as avg_satisfaction,
			SUM(repair_cost) as total_repair_cost,
			SUM(shipping_cost) as total_shipping_cost,
			SUM(replacement_cost) as total_replacement_cost
		FROM warranty_claims 
		WHERE storefront_id = $1 
			AND created_at >= $2 
			AND created_at <= $3 
			AND deleted_at IS NULL`

	var stats repository.ClaimStatistics
	var avgResolutionTime, avgRepairCost, avgSatisfaction sql.NullFloat64

	row := db.QueryRowContext(ctx, query, storefrontID, startDate, endDate)
	err = row.Scan(
		&stats.TotalClaims,
		&stats.PendingClaims,
		&stats.ValidatedClaims,
		&stats.CompletedClaims,
		&stats.RejectedClaims,
		&stats.CancelledClaims,
		&avgResolutionTime,
		&avgRepairCost,
		&avgSatisfaction,
		&stats.TotalRepairCost,
		&stats.TotalShippingCost,
		&stats.TotalReplacementCost,
	)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get claim statistics")
		return nil, fmt.Errorf("failed to get claim statistics: %w", err)
	}

	if avgResolutionTime.Valid {
		stats.AverageResolutionTime = avgResolutionTime.Float64
	}
	if avgRepairCost.Valid {
		stats.AverageRepairCost = avgRepairCost.Float64
	}
	if avgSatisfaction.Valid {
		stats.CustomerSatisfactionAvg = avgSatisfaction.Float64
	}

	// Initialize maps
	stats.ClaimsByCategory = make(map[string]int64)
	stats.ClaimsByStatus = make(map[string]int64)
	stats.ClaimsBySeverity = make(map[string]int64)
	stats.ClaimsByResolutionType = make(map[string]int64)

	// Get claims by category
	categoryQuery := `
		SELECT issue_category, COUNT(*) 
		FROM warranty_claims 
		WHERE storefront_id = $1 AND created_at >= $2 AND created_at <= $3 AND deleted_at IS NULL 
		GROUP BY issue_category`

	rows, err := db.QueryContext(ctx, categoryQuery, storefrontID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get claims by category: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category string
		var count int64
		if err := rows.Scan(&category, &count); err != nil {
			continue
		}
		stats.ClaimsByCategory[category] = count
	}

	return &stats, nil
}

// GenerateClaimNumber generates a unique claim number
func (r *WarrantyClaimRepositoryImpl) GenerateClaimNumber(ctx context.Context, storefrontID uuid.UUID) (string, error) {
	// Generate format: WAR-YYYY-NNNNNN
	year := time.Now().Year()

	db, err := r.GetDB(ctx, storefrontID)
	if err != nil {
		return "", fmt.Errorf("failed to get database connection: %w", err)
	}

	// Get the next sequence number for this year and storefront
	query := `
		SELECT COALESCE(MAX(
			CASE WHEN claim_number ~ '^WAR-' || $1 || '-[0-9]{6}$' 
			THEN CAST(SUBSTRING(claim_number FROM 10) AS INTEGER) 
			ELSE 0 END
		), 0) + 1
		FROM warranty_claims 
		WHERE storefront_id = $2 AND deleted_at IS NULL`

	var nextNum int
	err = db.GetContext(ctx, &nextNum, query, strconv.Itoa(year), storefrontID)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to get next claim number")
		return "", fmt.Errorf("failed to get next claim number: %w", err)
	}

	claimNumber := fmt.Sprintf("WAR-%d-%06d", year, nextNum)
	return claimNumber, nil
}

// GetClaimTimeline retrieves the complete timeline for a claim
func (r *WarrantyClaimRepositoryImpl) GetClaimTimeline(ctx context.Context, claimID uuid.UUID) ([]*entity.ClaimTimeline, error) {
	query := `
		SELECT * FROM claim_timeline 
		WHERE claim_id = $1 
		ORDER BY created_at ASC`

	var timeline []*entity.ClaimTimeline
	err := r.db.SelectContext(ctx, &timeline, query, claimID)
	if err != nil {
		r.logger.Error().Err(err).Str("claim_id", claimID.String()).Msg("Failed to get claim timeline")
		return nil, fmt.Errorf("failed to get claim timeline: %w", err)
	}

	return timeline, nil
}

// AddTimelineEntry adds a new timeline entry for a claim
func (r *WarrantyClaimRepositoryImpl) AddTimelineEntry(ctx context.Context, entry *entity.ClaimTimeline) error {
	query := `
		INSERT INTO claim_timeline (
			id, claim_id, event_type, event_description, handled_by, handled_by_name,
			visible_to_customer, metadata, created_at
		) VALUES (
			:id, :claim_id, :event_type, :event_description, :handled_by, :handled_by_name,
			:visible_to_customer, :metadata, :created_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, entry)
	if err != nil {
		r.logger.Error().Err(err).Str("claim_id", entry.ClaimID.String()).Msg("Failed to add timeline entry")
		return fmt.Errorf("failed to add timeline entry: %w", err)
	}

	return nil
}

// GetClaimAttachments retrieves all attachments for a claim
func (r *WarrantyClaimRepositoryImpl) GetClaimAttachments(ctx context.Context, claimID uuid.UUID) ([]*entity.ClaimAttachment, error) {
	query := `
		SELECT * FROM claim_attachments 
		WHERE claim_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC`

	var attachments []*entity.ClaimAttachment
	err := r.db.SelectContext(ctx, &attachments, query, claimID)
	if err != nil {
		r.logger.Error().Err(err).Str("claim_id", claimID.String()).Msg("Failed to get claim attachments")
		return nil, fmt.Errorf("failed to get claim attachments: %w", err)
	}

	return attachments, nil
}

// AddClaimAttachment adds a new attachment to a claim
func (r *WarrantyClaimRepositoryImpl) AddClaimAttachment(ctx context.Context, attachment *entity.ClaimAttachment) error {
	query := `
		INSERT INTO claim_attachments (
			id, claim_id, file_name, file_path, file_url, file_size, file_type,
			mime_type, attachment_type, description, uploaded_by, security_scan_status,
			security_scan_result, created_at, updated_at
		) VALUES (
			:id, :claim_id, :file_name, :file_path, :file_url, :file_size, :file_type,
			:mime_type, :attachment_type, :description, :uploaded_by, :security_scan_status,
			:security_scan_result, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, attachment)
	if err != nil {
		r.logger.Error().Err(err).Str("claim_id", attachment.ClaimID.String()).Msg("Failed to add claim attachment")
		return fmt.Errorf("failed to add claim attachment: %w", err)
	}

	return nil
}

// UpdateClaimAttachment updates an existing claim attachment
func (r *WarrantyClaimRepositoryImpl) UpdateClaimAttachment(ctx context.Context, attachment *entity.ClaimAttachment) error {
	query := `
		UPDATE claim_attachments SET 
			filename = :filename,
			file_path = :file_path,
			file_url = :file_url,
			description = :description,
			virus_scan_status = :virus_scan_status,
			virus_scan_date = :virus_scan_date,
			is_processed = :is_processed,
			processing_notes = :processing_notes
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, attachment)
	if err != nil {
		r.logger.Error().Err(err).Str("id", attachment.ID.String()).Msg("Failed to update claim attachment")
		return fmt.Errorf("failed to update claim attachment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("attachment not found")
	}

	return nil
}

// DeleteClaimAttachment deletes a claim attachment
func (r *WarrantyClaimRepositoryImpl) DeleteClaimAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE claim_attachments SET deleted_at = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, now, now, attachmentID)
	if err != nil {
		r.logger.Error().Err(err).Str("id", attachmentID.String()).Msg("Failed to delete claim attachment")
		return fmt.Errorf("failed to delete claim attachment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("attachment not found")
	}

	return nil
}

// GetRepairTickets retrieves repair tickets for a claim
func (r *WarrantyClaimRepositoryImpl) GetRepairTickets(ctx context.Context, claimID uuid.UUID) ([]*entity.RepairTicket, error) {
	query := `
		SELECT * FROM repair_tickets 
		WHERE claim_id = $1 AND deleted_at IS NULL 
		ORDER BY created_at DESC`

	var tickets []*entity.RepairTicket
	err := r.db.SelectContext(ctx, &tickets, query, claimID)
	if err != nil {
		r.logger.Error().Err(err).Str("claim_id", claimID.String()).Msg("Failed to get repair tickets")
		return nil, fmt.Errorf("failed to get repair tickets: %w", err)
	}

	return tickets, nil
}

// CreateRepairTicket creates a new repair ticket for a claim
func (r *WarrantyClaimRepositoryImpl) CreateRepairTicket(ctx context.Context, ticket *entity.RepairTicket) error {
	query := `
		INSERT INTO repair_tickets (
			id, ticket_number, claim_id, assigned_technician_id, repair_center_id,
			repair_type, issue_diagnosis, repair_steps, parts_required, labor_hours_estimated,
			labor_hours_actual, repair_cost, parts_cost, total_cost, status, priority,
			estimated_completion_date, actual_completion_date, quality_check_passed,
			quality_check_notes, customer_approval_required, customer_approved,
			technician_notes, internal_notes, created_at, updated_at
		) VALUES (
			:id, :ticket_number, :claim_id, :assigned_technician_id, :repair_center_id,
			:repair_type, :issue_diagnosis, :repair_steps, :parts_required, :labor_hours_estimated,
			:labor_hours_actual, :repair_cost, :parts_cost, :total_cost, :status, :priority,
			:estimated_completion_date, :actual_completion_date, :quality_check_passed,
			:quality_check_notes, :customer_approval_required, :customer_approved,
			:technician_notes, :internal_notes, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, ticket)
	if err != nil {
		r.logger.Error().Err(err).Str("claim_id", ticket.ClaimID.String()).Msg("Failed to create repair ticket")
		return fmt.Errorf("failed to create repair ticket: %w", err)
	}

	return nil
}

// UpdateRepairTicket updates an existing repair ticket
func (r *WarrantyClaimRepositoryImpl) UpdateRepairTicket(ctx context.Context, ticket *entity.RepairTicket) error {
	ticket.UpdatedAt = time.Now()

	query := `
		UPDATE repair_tickets SET 
			assigned_technician_id = :assigned_technician_id,
			repair_center_id = :repair_center_id,
			repair_type = :repair_type,
			issue_diagnosis = :issue_diagnosis,
			repair_steps = :repair_steps,
			parts_required = :parts_required,
			labor_hours_estimated = :labor_hours_estimated,
			labor_hours_actual = :labor_hours_actual,
			repair_cost = :repair_cost,
			parts_cost = :parts_cost,
			total_cost = :total_cost,
			status = :status,
			priority = :priority,
			estimated_completion_date = :estimated_completion_date,
			actual_completion_date = :actual_completion_date,
			quality_check_passed = :quality_check_passed,
			quality_check_notes = :quality_check_notes,
			customer_approval_required = :customer_approval_required,
			customer_approved = :customer_approved,
			technician_notes = :technician_notes,
			internal_notes = :internal_notes,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, ticket)
	if err != nil {
		r.logger.Error().Err(err).Str("id", ticket.ID.String()).Msg("Failed to update repair ticket")
		return fmt.Errorf("failed to update repair ticket: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("repair ticket not found")
	}

	return nil
}

// GetRepairTicketByID retrieves a repair ticket by its ID
func (r *WarrantyClaimRepositoryImpl) GetRepairTicketByID(ctx context.Context, ticketID uuid.UUID) (*entity.RepairTicket, error) {
	query := `SELECT * FROM repair_tickets WHERE id = $1 AND deleted_at IS NULL`

	var ticket entity.RepairTicket
	err := r.db.GetContext(ctx, &ticket, query, ticketID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		r.logger.Error().Err(err).Str("id", ticketID.String()).Msg("Failed to get repair ticket by ID")
		return nil, fmt.Errorf("failed to get repair ticket by ID: %w", err)
	}

	return &ticket, nil
}
