package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// PostgreSQLProductVariantOptionRepository implements the ProductVariantOptionRepository interface using PostgreSQL
type PostgreSQLProductVariantOptionRepository struct {
	db *sqlx.DB
}

// NewPostgreSQLProductVariantOptionRepository creates a new PostgreSQL product variant option repository
func NewPostgreSQLProductVariantOptionRepository(db *sqlx.DB) repository.ProductVariantOptionRepository {
	return &PostgreSQLProductVariantOptionRepository{
		db: db,
	}
}

// Create creates a new product variant option in the database
func (r *PostgreSQLProductVariantOptionRepository) Create(ctx context.Context, option *entity.ProductVariantOption) error {
	// Validate option before creating
	if err := option.Validate(); err != nil {
		return fmt.Errorf("variant option validation failed: %w", err)
	}

	// Ensure ID is set
	if option.ID == uuid.Nil {
		option.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	option.CreatedAt = now
	option.UpdatedAt = now

	query := `
		INSERT INTO product_variant_options (
			id, product_id, option_name, option_values, display_name, sort_order, is_required, created_at, updated_at
		) VALUES (
			:id, :product_id, :option_name, :option_values, :display_name, :sort_order, :is_required, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, option)
	if err != nil {
		// Handle unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "option_name") {
					return fmt.Errorf("option with name '%s' already exists for this product", option.OptionName)
				}
				return fmt.Errorf("variant option already exists: %w", err)
			case "23503": // foreign_key_violation
				if strings.Contains(pqErr.Detail, "product_id") {
					return fmt.Errorf("product with ID '%s' does not exist", option.ProductID)
				}
				return fmt.Errorf("foreign key violation: %w", err)
			}
		}
		return fmt.Errorf("failed to create variant option: %w", err)
	}

	return nil
}

// GetByID retrieves a product variant option by its ID
func (r *PostgreSQLProductVariantOptionRepository) GetByID(ctx context.Context, id uuid.UUID, include *repository.ProductVariantOptionInclude) (*entity.ProductVariantOption, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("variant option ID cannot be nil")
	}

	query := `
		SELECT id, product_id, option_name, option_values, display_name, sort_order, is_required, created_at, updated_at
		FROM product_variant_options
		WHERE id = $1`

	var option entity.ProductVariantOption
	err := r.db.GetContext(ctx, &option, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("variant option with ID '%s' not found", id)
		}
		return nil, fmt.Errorf("failed to get variant option by ID: %w", err)
	}

	// Load related data if requested
	if include != nil {
		if err := r.loadVariantOptionRelations(ctx, &option, include); err != nil {
			return nil, fmt.Errorf("failed to load variant option relations: %w", err)
		}
	}

	// Compute fields
	option.ComputeFields()

	return &option, nil
}

// Update updates an existing product variant option
func (r *PostgreSQLProductVariantOptionRepository) Update(ctx context.Context, option *entity.ProductVariantOption) error {
	// Validate option before updating
	if err := option.Validate(); err != nil {
		return fmt.Errorf("variant option validation failed: %w", err)
	}

	if option.ID == uuid.Nil {
		return fmt.Errorf("variant option ID cannot be nil for update")
	}

	// Update timestamp
	option.UpdatedAt = time.Now()

	query := `
		UPDATE product_variant_options SET
			option_name = :option_name,
			option_values = :option_values,
			display_name = :display_name,
			sort_order = :sort_order,
			is_required = :is_required,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, option)
	if err != nil {
		// Handle unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "option_name") {
					return fmt.Errorf("option with name '%s' already exists for this product", option.OptionName)
				}
				return fmt.Errorf("variant option already exists: %w", err)
			}
		}
		return fmt.Errorf("failed to update variant option: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variant option with ID '%s' not found", option.ID)
	}

	return nil
}

// Delete deletes a product variant option
func (r *PostgreSQLProductVariantOptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("variant option ID cannot be nil")
	}

	query := `DELETE FROM product_variant_options WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete variant option: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("variant option with ID '%s' not found", id)
	}

	return nil
}

// GetByProduct retrieves all variant options for a product
func (r *PostgreSQLProductVariantOptionRepository) GetByProduct(ctx context.Context, productID uuid.UUID, include *repository.ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be nil")
	}

	query := `
		SELECT id, product_id, option_name, option_values, display_name, sort_order, is_required, created_at, updated_at
		FROM product_variant_options
		WHERE product_id = $1
		ORDER BY sort_order ASC, option_name ASC`

	var options []entity.ProductVariantOption
	err := r.db.SelectContext(ctx, &options, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variant options by product: %w", err)
	}

	// Convert to pointer slice and load relations
	result := make([]*entity.ProductVariantOption, len(options))
	for i := range options {
		result[i] = &options[i]

		// Load related data if requested
		if include != nil {
			if err := r.loadVariantOptionRelations(ctx, result[i], include); err != nil {
				return nil, fmt.Errorf("failed to load variant option relations: %w", err)
			}
		}

		// Compute fields
		result[i].ComputeFields()
	}

	return result, nil
}

// GetByProductAndName retrieves a variant option by product and option name
func (r *PostgreSQLProductVariantOptionRepository) GetByProductAndName(ctx context.Context, productID uuid.UUID, optionName string, include *repository.ProductVariantOptionInclude) (*entity.ProductVariantOption, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be nil")
	}

	if strings.TrimSpace(optionName) == "" {
		return nil, fmt.Errorf("option name cannot be empty")
	}

	query := `
		SELECT id, product_id, option_name, option_values, display_name, sort_order, is_required, created_at, updated_at
		FROM product_variant_options
		WHERE product_id = $1 AND option_name = $2`

	var option entity.ProductVariantOption
	err := r.db.GetContext(ctx, &option, query, productID, optionName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("variant option with name '%s' not found for product '%s'", optionName, productID)
		}
		return nil, fmt.Errorf("failed to get variant option by product and name: %w", err)
	}

	// Load related data if requested
	if include != nil {
		if err := r.loadVariantOptionRelations(ctx, &option, include); err != nil {
			return nil, fmt.Errorf("failed to load variant option relations: %w", err)
		}
	}

	// Compute fields
	option.ComputeFields()

	return &option, nil
}

// IsOptionNameExists checks if an option name exists for a product
func (r *PostgreSQLProductVariantOptionRepository) IsOptionNameExists(ctx context.Context, productID uuid.UUID, optionName string) (bool, error) {
	if productID == uuid.Nil {
		return false, fmt.Errorf("product ID cannot be nil")
	}

	if strings.TrimSpace(optionName) == "" {
		return false, fmt.Errorf("option name cannot be empty")
	}

	query := `SELECT EXISTS(SELECT 1 FROM product_variant_options WHERE product_id = $1 AND option_name = $2)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, productID, optionName)
	if err != nil {
		return false, fmt.Errorf("failed to check option name existence: %w", err)
	}

	return exists, nil
}

// Exists checks if a variant option exists by ID
func (r *PostgreSQLProductVariantOptionRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	if id == uuid.Nil {
		return false, fmt.Errorf("variant option ID cannot be nil")
	}

	query := `SELECT EXISTS(SELECT 1 FROM product_variant_options WHERE id = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check variant option existence: %w", err)
	}

	return exists, nil
}

// Helper methods

// loadVariantOptionRelations loads related data based on include options
func (r *PostgreSQLProductVariantOptionRepository) loadVariantOptionRelations(ctx context.Context, option *entity.ProductVariantOption, include *repository.ProductVariantOptionInclude) error {
	// Load product if requested
	if include.Product {
		// In a real implementation, you might inject the product repository
		// For now, we'll skip this to avoid circular dependencies
	}

	// Load usage stats if requested
	if include.UsageStats {
		// TODO: Implement usage stats loading
	}

	// Load variant count if requested
	if include.VariantCount {
		// TODO: Implement variant count loading
	}

	return nil
}

// Placeholder implementations for remaining methods
// These would need full implementation based on business requirements

func (r *PostgreSQLProductVariantOptionRepository) CreateBatch(ctx context.Context, options []*entity.ProductVariantOption) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetByIDs(ctx context.Context, ids []uuid.UUID, include *repository.ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) UpdateBatch(ctx context.Context, options []*entity.ProductVariantOption) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) List(ctx context.Context, filter *repository.ProductVariantOptionFilter, include *repository.ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) Count(ctx context.Context, filter *repository.ProductVariantOptionFilter) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) CountByProduct(ctx context.Context, productID uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetOptionNamesByProduct(ctx context.Context, productID uuid.UUID) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) IsOptionNameExistsExcluding(ctx context.Context, productID uuid.UUID, optionName string, excludeID uuid.UUID) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetAllValuesForOption(ctx context.Context, optionID uuid.UUID) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetUniqueValuesForProduct(ctx context.Context, productID uuid.UUID, optionName string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) IsValueExists(ctx context.Context, optionID uuid.UUID, value string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) AddValueToOption(ctx context.Context, optionID uuid.UUID, value string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) RemoveValueFromOption(ctx context.Context, optionID uuid.UUID, value string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) UpdateOptionValue(ctx context.Context, optionID uuid.UUID, oldValue, newValue string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) AddValuesToOption(ctx context.Context, optionID uuid.UUID, values []string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) RemoveValuesFromOption(ctx context.Context, optionID uuid.UUID, values []string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) ReplaceOptionValues(ctx context.Context, optionID uuid.UUID, values []string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) UpdateSortOrder(ctx context.Context, optionID uuid.UUID, sortOrder int) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) ReorderOptions(ctx context.Context, productID uuid.UUID, optionOrders []repository.OptionOrder) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetNextSortOrder(ctx context.Context, productID uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) Search(ctx context.Context, query string, filter *repository.ProductVariantOptionFilter, include *repository.ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) SearchByValue(ctx context.Context, value string, productID *uuid.UUID) ([]*entity.ProductVariantOption, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) ValidateOptionName(ctx context.Context, productID uuid.UUID, optionName string, excludeID *uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) ValidateOptionValues(ctx context.Context, values []string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) CheckOptionUsage(ctx context.Context, optionID uuid.UUID) (*repository.OptionUsage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetOptionStatistics(ctx context.Context, optionID uuid.UUID) (*repository.OptionStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetMostUsedOptions(ctx context.Context, limit int) ([]*entity.ProductVariantOption, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetValueUsageStatistics(ctx context.Context, optionID uuid.UUID) (map[string]int64, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetOptionUsageByProduct(ctx context.Context, productID uuid.UUID) ([]*repository.OptionUsageInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) ExportOptions(ctx context.Context, productID uuid.UUID) ([]*repository.OptionExport, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) ImportOptions(ctx context.Context, productID uuid.UUID, options []*repository.OptionImport) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) CleanupUnusedOptions(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) RemoveEmptyOptions(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetOptionCount(ctx context.Context, productID uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductVariantOptionRepository) GetTotalValueCount(ctx context.Context, productID uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}
