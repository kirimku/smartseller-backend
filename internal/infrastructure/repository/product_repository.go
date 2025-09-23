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
	"github.com/shopspring/decimal"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// PostgreSQLProductRepository implements the ProductRepository interface using PostgreSQL
type PostgreSQLProductRepository struct {
	db *sqlx.DB
}

// NewPostgreSQLProductRepository creates a new PostgreSQL product repository
func NewPostgreSQLProductRepository(db *sqlx.DB) repository.ProductRepository {
	return &PostgreSQLProductRepository{
		db: db,
	}
}

// Create creates a new product in the database
func (r *PostgreSQLProductRepository) Create(ctx context.Context, product *entity.Product) error {
	// Validate product before creating
	if err := product.Validate(); err != nil {
		return fmt.Errorf("product validation failed: %w", err)
	}

	// Ensure ID is set
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	query := `
		INSERT INTO products (
			id, sku, name, description, category_id, brand, tags,
			base_price, sale_price, cost_price,
			track_inventory, stock_quantity, low_stock_threshold,
			status, meta_title, meta_description, slug,
			weight, dimensions_length, dimensions_width, dimensions_height,
			created_by, created_at, updated_at
		) VALUES (
			:id, :sku, :name, :description, :category_id, :brand, :tags,
			:base_price, :sale_price, :cost_price,
			:track_inventory, :stock_quantity, :low_stock_threshold,
			:status, :meta_title, :meta_description, :slug,
			:weight, :dimensions_length, :dimensions_width, :dimensions_height,
			:created_by, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		// Create context for error mapping
		context := map[string]interface{}{
			"id":          product.ID,
			"sku":         product.SKU,
			"slug":        product.Slug,
			"category_id": product.CategoryID,
		}

		// Map PostgreSQL error to repository error
		mappedErr := MapPostgreSQLError(err, "Product", context)
		return WrapWithContext(mappedErr, "CreateProduct", context)
	}

	return nil
}

// GetByID retrieves a product by its ID
func (r *PostgreSQLProductRepository) GetByID(ctx context.Context, id uuid.UUID, include *repository.ProductInclude) (*entity.Product, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be nil")
	}

	query := `
		SELECT 
			p.id, p.sku, p.name, p.description, p.category_id, p.brand, p.tags,
			p.base_price, p.sale_price, p.cost_price,
			p.track_inventory, p.stock_quantity, p.low_stock_threshold,
			p.status, p.meta_title, p.meta_description, p.slug,
			p.weight, p.dimensions_length, p.dimensions_width, p.dimensions_height,
			p.created_by, p.created_at, p.updated_at, p.deleted_at
		FROM products p
		WHERE p.id = $1 AND p.deleted_at IS NULL`

	var product entity.Product
	err := r.db.GetContext(ctx, &product, query, id)
	if err != nil {
		context := map[string]interface{}{"id": id}

		if err == sql.ErrNoRows {
			return nil, NewNotFoundError("Product", id)
		}

		mappedErr := MapPostgreSQLError(err, "Product", context)
		return nil, WrapWithContext(mappedErr, "GetProductByID", context)
	}

	// Load related data if requested
	if include != nil {
		if err := r.loadProductRelations(ctx, &product, include); err != nil {
			return nil, fmt.Errorf("failed to load product relations: %w", err)
		}
	}

	// Compute fields
	product.ComputeFields()

	return &product, nil
}

// GetBySKU retrieves a product by its SKU
func (r *PostgreSQLProductRepository) GetBySKU(ctx context.Context, sku string, include *repository.ProductInclude) (*entity.Product, error) {
	if strings.TrimSpace(sku) == "" {
		return nil, fmt.Errorf("SKU cannot be empty")
	}

	query := `
		SELECT 
			p.id, p.sku, p.name, p.description, p.category_id, p.brand, p.tags,
			p.base_price, p.sale_price, p.cost_price,
			p.track_inventory, p.stock_quantity, p.low_stock_threshold,
			p.status, p.meta_title, p.meta_description, p.slug,
			p.weight, p.dimensions_length, p.dimensions_width, p.dimensions_height,
			p.created_by, p.created_at, p.updated_at, p.deleted_at
		FROM products p
		WHERE p.sku = $1 AND p.deleted_at IS NULL`

	var product entity.Product
	err := r.db.GetContext(ctx, &product, query, sku)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product with SKU '%s' not found", sku)
		}
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	// Load related data if requested
	if include != nil {
		if err := r.loadProductRelations(ctx, &product, include); err != nil {
			return nil, fmt.Errorf("failed to load product relations: %w", err)
		}
	}

	// Compute fields
	product.ComputeFields()

	return &product, nil
}

// Update updates an existing product
func (r *PostgreSQLProductRepository) Update(ctx context.Context, product *entity.Product) error {
	// Validate product before updating
	if err := product.Validate(); err != nil {
		return fmt.Errorf("product validation failed: %w", err)
	}

	if product.ID == uuid.Nil {
		return fmt.Errorf("product ID cannot be nil for update")
	}

	// Update timestamp
	product.UpdatedAt = time.Now()

	query := `
		UPDATE products SET
			sku = :sku,
			name = :name,
			description = :description,
			category_id = :category_id,
			brand = :brand,
			tags = :tags,
			base_price = :base_price,
			sale_price = :sale_price,
			cost_price = :cost_price,
			track_inventory = :track_inventory,
			stock_quantity = :stock_quantity,
			low_stock_threshold = :low_stock_threshold,
			status = :status,
			meta_title = :meta_title,
			meta_description = :meta_description,
			slug = :slug,
			weight = :weight,
			dimensions_length = :dimensions_length,
			dimensions_width = :dimensions_width,
			dimensions_height = :dimensions_height,
			updated_at = :updated_at
		WHERE id = :id AND deleted_at IS NULL`

	result, err := r.db.NamedExecContext(ctx, query, product)
	if err != nil {
		// Handle unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "sku") {
					return fmt.Errorf("product with SKU '%s' already exists", product.SKU)
				}
				if strings.Contains(pqErr.Detail, "slug") {
					return fmt.Errorf("product with slug '%s' already exists", product.Slug)
				}
				return fmt.Errorf("product already exists: %w", err)
			case "23503": // foreign_key_violation
				if strings.Contains(pqErr.Detail, "category_id") {
					return fmt.Errorf("category with ID '%s' does not exist", product.CategoryID)
				}
				return fmt.Errorf("foreign key violation: %w", err)
			}
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with ID '%s' not found or already deleted", product.ID)
	}

	return nil
}

// Delete performs a soft delete on a product
func (r *PostgreSQLProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("product ID cannot be nil")
	}

	query := `
		UPDATE products 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with ID '%s' not found or already deleted", id)
	}

	return nil
}

// RestoreProduct restores a soft-deleted product
func (r *PostgreSQLProductRepository) RestoreProduct(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("product ID cannot be nil")
	}

	query := `
		UPDATE products 
		SET deleted_at = NULL, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NOT NULL`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to restore product: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with ID '%s' not found or not deleted", id)
	}

	return nil
}

// loadProductRelations loads related data based on include options
func (r *PostgreSQLProductRepository) loadProductRelations(ctx context.Context, product *entity.Product, include *repository.ProductInclude) error {
	// Load category if requested
	if include.Category {
		if err := r.loadProductCategory(ctx, product); err != nil {
			return fmt.Errorf("failed to load category: %w", err)
		}
	}

	// Load images if requested
	if include.Images {
		if err := r.loadProductImages(ctx, product, include); err != nil {
			return fmt.Errorf("failed to load images: %w", err)
		}
	}

	// Load variants if requested
	if include.Variants {
		if err := r.loadProductVariants(ctx, product, include); err != nil {
			return fmt.Errorf("failed to load variants: %w", err)
		}
	}

	// Load variant options if requested
	if include.VariantOptions {
		if err := r.loadProductVariantOptions(ctx, product); err != nil {
			return fmt.Errorf("failed to load variant options: %w", err)
		}
	}

	return nil
}

// loadProductCategory loads the product's category
func (r *PostgreSQLProductRepository) loadProductCategory(ctx context.Context, product *entity.Product) error {
	if product.CategoryID == nil {
		return nil // No category to load
	}

	query := `
		SELECT id, parent_id, name, slug, description, path, level, sort_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE id = $1`

	var category entity.ProductCategory
	err := r.db.GetContext(ctx, &category, query, *product.CategoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Category not found, but don't fail the product load
		}
		return fmt.Errorf("failed to load category: %w", err)
	}

	// Note: In a real implementation, you might want to set this on a Category field
	// For now, we'll store it in a map or similar structure if needed
	return nil
}

// loadProductImages loads the product's images
func (r *PostgreSQLProductRepository) loadProductImages(ctx context.Context, product *entity.Product, include *repository.ProductInclude) error {
	query := `
		SELECT 
			id, product_id, variant_id, image_url, alt_text, is_primary, sort_order,
			file_name, file_size, mime_type, width, height,
			cloudinary_id, cloudinary_url, thumbnail_url, medium_url, large_url,
			created_at, updated_at
		FROM product_images
		WHERE product_id = $1 AND variant_id IS NULL`

	// Add active filter if requested
	if include.OnlyActiveImages {
		// Assuming there's an is_active column, adjust as needed
		// query += " AND is_active = true"
	}

	query += " ORDER BY sort_order ASC, created_at ASC"

	var images []entity.ProductImage
	err := r.db.SelectContext(ctx, &images, query, product.ID)
	if err != nil {
		return fmt.Errorf("failed to load images: %w", err)
	}

	// Compute fields for each image
	for i := range images {
		images[i].ComputeFields()
	}

	// Note: In a real implementation, you might want to set this on an Images field
	return nil
}

// loadProductVariants loads the product's variants
func (r *PostgreSQLProductRepository) loadProductVariants(ctx context.Context, product *entity.Product, include *repository.ProductInclude) error {
	query := `
		SELECT 
			id, product_id, variant_name, variant_sku, options,
			price, compare_at_price, cost_price,
			stock_quantity, low_stock_threshold, track_quantity,
			weight, length, width, height,
			is_active, is_default, position,
			created_at, updated_at
		FROM product_variants
		WHERE product_id = $1`

	// Add active filter if requested
	if include.OnlyActiveVariants {
		query += " AND is_active = true"
	}

	query += " ORDER BY position ASC, created_at ASC"

	var variants []entity.ProductVariant
	err := r.db.SelectContext(ctx, &variants, query, product.ID)
	if err != nil {
		return fmt.Errorf("failed to load variants: %w", err)
	}

	// Compute fields for each variant
	for i := range variants {
		variants[i].ComputeFields()
	}

	// Note: In a real implementation, you might want to set this on a Variants field
	return nil
}

// loadProductVariantOptions loads the product's variant options
func (r *PostgreSQLProductRepository) loadProductVariantOptions(ctx context.Context, product *entity.Product) error {
	query := `
		SELECT 
			id, product_id, option_name, option_values, display_name, 
			sort_order, is_required, created_at, updated_at
		FROM product_variant_options
		WHERE product_id = $1
		ORDER BY sort_order ASC, created_at ASC`

	var options []entity.ProductVariantOption
	err := r.db.SelectContext(ctx, &options, query, product.ID)
	if err != nil {
		return fmt.Errorf("failed to load variant options: %w", err)
	}

	// Compute fields for each option
	for i := range options {
		options[i].ComputeFields()
	}

	// Note: In a real implementation, you might want to set this on a VariantOptions field
	return nil
}

// Exists checks if a product exists by ID
func (r *PostgreSQLProductRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	if id == uuid.Nil {
		return false, fmt.Errorf("product ID cannot be nil")
	}

	query := `SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check product existence: %w", err)
	}

	return exists, nil
}

// ExistsBySKU checks if a product exists by SKU
func (r *PostgreSQLProductRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	if strings.TrimSpace(sku) == "" {
		return false, fmt.Errorf("SKU cannot be empty")
	}

	query := `SELECT EXISTS(SELECT 1 FROM products WHERE sku = $1 AND deleted_at IS NULL)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, sku)
	if err != nil {
		return false, fmt.Errorf("failed to check product SKU existence: %w", err)
	}

	return exists, nil
}

// Status management methods

// Activate activates a product
func (r *PostgreSQLProductRepository) Activate(ctx context.Context, productID uuid.UUID) error {
	return r.UpdateStatus(ctx, productID, entity.ProductStatusActive)
}

// Deactivate deactivates a product
func (r *PostgreSQLProductRepository) Deactivate(ctx context.Context, productID uuid.UUID) error {
	return r.UpdateStatus(ctx, productID, entity.ProductStatusInactive)
}

// UpdateStatus updates the status of a product
func (r *PostgreSQLProductRepository) UpdateStatus(ctx context.Context, productID uuid.UUID, status entity.ProductStatus) error {
	if productID == uuid.Nil {
		return fmt.Errorf("product ID cannot be nil")
	}

	if !status.Valid() {
		return fmt.Errorf("invalid product status: %s", status)
	}

	query := `UPDATE products SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, status, productID)
	if err != nil {
		return fmt.Errorf("failed to update product status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with ID '%s' not found", productID)
	}

	return nil
}

// BulkActivate activates multiple products
func (r *PostgreSQLProductRepository) BulkActivate(ctx context.Context, productIDs []uuid.UUID) error {
	return r.BulkUpdateStatus(ctx, productIDs, entity.ProductStatusActive)
}

// BulkDeactivate deactivates multiple products
func (r *PostgreSQLProductRepository) BulkDeactivate(ctx context.Context, productIDs []uuid.UUID) error {
	return r.BulkUpdateStatus(ctx, productIDs, entity.ProductStatusInactive)
}

// BulkUpdateStatus updates the status of multiple products
func (r *PostgreSQLProductRepository) BulkUpdateStatus(ctx context.Context, productIDs []uuid.UUID, status entity.ProductStatus) error {
	if len(productIDs) == 0 {
		return fmt.Errorf("product IDs cannot be empty")
	}

	if !status.Valid() {
		return fmt.Errorf("invalid product status: %s", status)
	}

	query := `UPDATE products SET status = $1, updated_at = NOW() WHERE id = ANY($2) AND deleted_at IS NULL`

	_, err := r.db.ExecContext(ctx, query, status, pq.Array(productIDs))
	if err != nil {
		return fmt.Errorf("failed to bulk update product status: %w", err)
	}

	return nil
}

// Placeholder implementations for remaining methods
// These would need full implementation based on business requirements

func (r *PostgreSQLProductRepository) CreateBatch(ctx context.Context, products []*entity.Product) error {
	if len(products) == 0 {
		return nil
	}

	// Start a transaction for atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the bulk insert query
	query := `
		INSERT INTO products (
			id, sku, name, description, category_id, brand, 
			base_price, sale_price, cost_price, weight, 
			dimensions_length, dimensions_width, dimensions_height,
			status, track_inventory, stock_quantity, low_stock_threshold,
			meta_title, meta_description, slug, created_by,
			created_at, updated_at
		) VALUES `

	// Build values placeholders
	var values []string
	var args []interface{}
	argIndex := 1

	for _, product := range products {
		// Validate product before insertion
		if err := product.Validate(); err != nil {
			return fmt.Errorf("invalid product data: %w", err)
		}

		// Create placeholder for this product
		placeholder := fmt.Sprintf(`(
			$%d, $%d, $%d, $%d, $%d, $%d, 
			$%d, $%d, $%d, $%d, 
			$%d, $%d, $%d,
			$%d, $%d, $%d, $%d,
			$%d, $%d, $%d, $%d,
			$%d, $%d
		)`, argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4, argIndex+5,
			argIndex+6, argIndex+7, argIndex+8, argIndex+9,
			argIndex+10, argIndex+11, argIndex+12,
			argIndex+13, argIndex+14, argIndex+15, argIndex+16,
			argIndex+17, argIndex+18, argIndex+19, argIndex+20,
			argIndex+21, argIndex+22)

		values = append(values, placeholder)

		// Add arguments
		args = append(args,
			product.ID, product.SKU, product.Name, product.Description, product.CategoryID, product.Brand,
			product.BasePrice, product.SalePrice, product.CostPrice, product.Weight,
			product.DimensionsLength, product.DimensionsWidth, product.DimensionsHeight,
			product.Status, product.TrackInventory, product.StockQuantity, product.LowStockThreshold,
			product.MetaTitle, product.MetaDescription, product.Slug, product.CreatedBy,
			product.CreatedAt, product.UpdatedAt,
		)

		argIndex += 23
	}

	// Combine query
	finalQuery := query + strings.Join(values, ", ")

	// Execute the batch insert
	_, err = tx.ExecContext(ctx, finalQuery, args...)
	if err != nil {
		// Check for specific constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				return fmt.Errorf("duplicate SKU or other unique constraint violation: %w", err)
			case "23503": // foreign_key_violation
				return fmt.Errorf("invalid category reference: %w", err)
			case "23514": // check_violation
				return fmt.Errorf("constraint violation: %w", err)
			}
		}
		return fmt.Errorf("failed to batch create products: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch create transaction: %w", err)
	}

	return nil
}

func (r *PostgreSQLProductRepository) GetByIDs(ctx context.Context, ids []uuid.UUID, include *repository.ProductInclude) ([]*entity.Product, error) {
	if len(ids) == 0 {
		return []*entity.Product{}, nil
	}

	// Create a filter with the specific IDs
	filter := &repository.ProductFilter{
		IDs: ids,
	}

	// Use the Search method with ID filter
	return r.Search(ctx, "", filter, include)
}

func (r *PostgreSQLProductRepository) UpdateBatch(ctx context.Context, products []*entity.Product) error {
	if len(products) == 0 {
		return nil
	}

	// Start a transaction for atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the update statement
	updateQuery := `
		UPDATE products SET 
			sku = $2, name = $3, description = $4, category_id = $5, brand = $6,
			base_price = $7, sale_price = $8, cost_price = $9, weight = $10,
			dimensions_length = $11, dimensions_width = $12, dimensions_height = $13,
			status = $14, track_inventory = $15, stock_quantity = $16, low_stock_threshold = $17,
			meta_title = $18, meta_description = $19, slug = $20, updated_at = $21
		WHERE id = $1 AND deleted_at IS NULL`

	// Execute update for each product
	for _, product := range products {
		// Validate product before update
		if err := product.Validate(); err != nil {
			return fmt.Errorf("invalid product data for ID %s: %w", product.ID, err)
		}

		// Set updated timestamp
		product.UpdatedAt = time.Now()

		_, err := tx.ExecContext(ctx, updateQuery,
			product.ID, product.SKU, product.Name, product.Description, product.CategoryID, product.Brand,
			product.BasePrice, product.SalePrice, product.CostPrice, product.Weight,
			product.DimensionsLength, product.DimensionsWidth, product.DimensionsHeight,
			product.Status, product.TrackInventory, product.StockQuantity, product.LowStockThreshold,
			product.MetaTitle, product.MetaDescription, product.Slug, product.UpdatedAt,
		)
		if err != nil {
			// Check for specific constraint violations
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505": // unique_violation
					return fmt.Errorf("duplicate SKU for product %s: %w", product.ID, err)
				case "23503": // foreign_key_violation
					return fmt.Errorf("invalid category reference for product %s: %w", product.ID, err)
				case "23514": // check_violation
					return fmt.Errorf("constraint violation for product %s: %w", product.ID, err)
				}
			}
			return fmt.Errorf("failed to update product %s: %w", product.ID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit batch update transaction: %w", err)
	}

	return nil
}

func (r *PostgreSQLProductRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	// Use soft delete by setting deleted_at timestamp
	query := `
		UPDATE products 
		SET deleted_at = NOW(), updated_at = NOW() 
		WHERE id = ANY($1) AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to batch delete products: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected count: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no products found with the provided IDs or already deleted")
	}

	return nil
}

func (r *PostgreSQLProductRepository) List(ctx context.Context, filter *repository.ProductFilter, include *repository.ProductInclude) ([]*entity.Product, error) {
	// Use the Search method with empty query to get filtered list
	return r.Search(ctx, "", filter, include)
}

func (r *PostgreSQLProductRepository) Count(ctx context.Context, filter *repository.ProductFilter) (int64, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Apply filters
	if filter != nil {
		// IDs filter (specific product IDs)
		if len(filter.IDs) > 0 {
			placeholders := make([]string, len(filter.IDs))
			for i, id := range filter.IDs {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, id)
				argIndex++
			}
			whereConditions = append(whereConditions, fmt.Sprintf("id IN (%s)", strings.Join(placeholders, ",")))
		}

		// Status filter (slice of statuses)
		if len(filter.Status) > 0 {
			placeholders := make([]string, len(filter.Status))
			for i, status := range filter.Status {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, status)
				argIndex++
			}
			whereConditions = append(whereConditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
		}

		// Category IDs filter
		if len(filter.CategoryIDs) > 0 {
			placeholders := make([]string, len(filter.CategoryIDs))
			for i, categoryID := range filter.CategoryIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, categoryID)
				argIndex++
			}
			whereConditions = append(whereConditions, fmt.Sprintf("category_id IN (%s)", strings.Join(placeholders, ",")))
		}

		if filter.MinPrice != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("base_price >= $%d", argIndex))
			args = append(args, *filter.MinPrice)
			argIndex++
		}

		if filter.MaxPrice != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("base_price <= $%d", argIndex))
			args = append(args, *filter.MaxPrice)
			argIndex++
		}

		if filter.MinStock != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("stock_quantity >= $%d", argIndex))
			args = append(args, *filter.MinStock)
			argIndex++
		}

		if filter.MaxStock != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("stock_quantity <= $%d", argIndex))
			args = append(args, *filter.MaxStock)
			argIndex++
		}

		if filter.SearchQuery != "" {
			searchCondition := fmt.Sprintf(`(
				name ILIKE $%d OR 
				description ILIKE $%d OR 
				sku ILIKE $%d OR 
				brand ILIKE $%d
			)`, argIndex, argIndex, argIndex, argIndex)
			whereConditions = append(whereConditions, searchCondition)
			args = append(args, "%"+filter.SearchQuery+"%")
			argIndex++
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build query
	query := "SELECT COUNT(*) FROM products" + whereClause

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}

func (r *PostgreSQLProductRepository) Search(ctx context.Context, query string, filter *repository.ProductFilter, include *repository.ProductInclude) ([]*entity.Product, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Base SELECT with actual product fields
	selectQuery := `
		SELECT p.id, p.sku, p.name, p.description, p.category_id, p.brand, 
		       p.base_price, p.sale_price, p.cost_price, p.weight, 
		       p.dimensions_length, p.dimensions_width, p.dimensions_height,
		       p.status, p.track_inventory, p.stock_quantity, p.low_stock_threshold,
		       p.meta_title, p.meta_description, p.slug, p.created_by,
		       p.created_at, p.updated_at, p.deleted_at`

	fromQuery := ` FROM products p`

	// Include category if requested
	if include != nil && include.Category {
		selectQuery += `, c.id as cat_id, c.name as cat_name, c.description as cat_description`
		fromQuery += ` LEFT JOIN product_categories c ON p.category_id = c.id`
	}

	// Text search condition
	if query != "" {
		searchCondition := fmt.Sprintf(`(
			p.name ILIKE $%d OR 
			p.description ILIKE $%d OR 
			p.sku ILIKE $%d OR 
			p.brand ILIKE $%d
		)`, argIndex, argIndex, argIndex, argIndex)
		whereConditions = append(whereConditions, searchCondition)
		args = append(args, "%"+query+"%")
		argIndex++
	}

	// Apply filters
	if filter != nil {
		// Status filter (slice of statuses)
		if len(filter.Status) > 0 {
			placeholders := make([]string, len(filter.Status))
			for i, status := range filter.Status {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, status)
				argIndex++
			}
			whereConditions = append(whereConditions, fmt.Sprintf("p.status IN (%s)", strings.Join(placeholders, ",")))
		}

		// Category IDs filter
		if len(filter.CategoryIDs) > 0 {
			placeholders := make([]string, len(filter.CategoryIDs))
			for i, categoryID := range filter.CategoryIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIndex)
				args = append(args, categoryID)
				argIndex++
			}
			whereConditions = append(whereConditions, fmt.Sprintf("p.category_id IN (%s)", strings.Join(placeholders, ",")))
		}

		if filter.MinPrice != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("p.base_price >= $%d", argIndex))
			args = append(args, *filter.MinPrice)
			argIndex++
		}

		if filter.MaxPrice != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("p.base_price <= $%d", argIndex))
			args = append(args, *filter.MaxPrice)
			argIndex++
		}

		if filter.MinStock != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("p.stock_quantity >= $%d", argIndex))
			args = append(args, *filter.MinStock)
			argIndex++
		}

		if filter.MaxStock != nil {
			whereConditions = append(whereConditions, fmt.Sprintf("p.stock_quantity <= $%d", argIndex))
			args = append(args, *filter.MaxStock)
			argIndex++
		}

		if filter.SearchQuery != "" && filter.SearchQuery != query {
			searchCondition := fmt.Sprintf(`(
				p.name ILIKE $%d OR 
				p.description ILIKE $%d OR 
				p.sku ILIKE $%d
			)`, argIndex, argIndex, argIndex)
			whereConditions = append(whereConditions, searchCondition)
			args = append(args, "%"+filter.SearchQuery+"%")
			argIndex++
		}
	}

	// Build WHERE clause
	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Build ORDER BY clause
	orderBy := " ORDER BY p.updated_at DESC"
	if filter != nil && filter.SortBy != "" {
		switch filter.SortBy {
		case "name":
			orderBy = " ORDER BY p.name"
		case "price":
			orderBy = " ORDER BY p.base_price"
		case "created_at":
			orderBy = " ORDER BY p.created_at"
		case "updated_at":
			orderBy = " ORDER BY p.updated_at"
		case "stock_quantity":
			orderBy = " ORDER BY p.stock_quantity"
		}

		if filter.SortOrder == "desc" {
			orderBy += " DESC"
		} else {
			orderBy += " ASC"
		}
	}

	// Build LIMIT and OFFSET
	limitOffset := ""
	if filter != nil {
		if filter.Limit > 0 {
			limitOffset += fmt.Sprintf(" LIMIT $%d", argIndex)
			args = append(args, filter.Limit)
			argIndex++
		}
		if filter.Offset > 0 {
			limitOffset += fmt.Sprintf(" OFFSET $%d", argIndex)
			args = append(args, filter.Offset)
			argIndex++
		}
	}

	// Combine query
	finalQuery := selectQuery + fromQuery + whereClause + orderBy + limitOffset

	rows, err := r.db.QueryContext(ctx, finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		product := &entity.Product{}
		var description, brand sql.NullString
		var salePrice, costPrice, weight, dimensionsLength, dimensionsWidth, dimensionsHeight sql.NullString
		var lowStockThreshold sql.NullInt32
		var metaTitle, metaDescription, slug sql.NullString
		var deletedAt sql.NullTime

		scanArgs := []interface{}{
			&product.ID, &product.SKU, &product.Name, &description, &product.CategoryID, &brand,
			&product.BasePrice, &salePrice, &costPrice, &weight,
			&dimensionsLength, &dimensionsWidth, &dimensionsHeight,
			&product.Status, &product.TrackInventory, &product.StockQuantity, &lowStockThreshold,
			&metaTitle, &metaDescription, &slug, &product.CreatedBy,
			&product.CreatedAt, &product.UpdatedAt, &deletedAt,
		}

		// Add category fields if included
		var category *entity.ProductCategory
		if include != nil && include.Category {
			category = &entity.ProductCategory{}
			var catDescription sql.NullString
			scanArgs = append(scanArgs, &category.ID, &category.Name, &catDescription)
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			product.Description = &description.String
		}
		if brand.Valid {
			product.Brand = &brand.String
		}
		if salePrice.Valid {
			if price, err := decimal.NewFromString(salePrice.String); err == nil {
				product.SalePrice = &price
			}
		}
		if costPrice.Valid {
			if price, err := decimal.NewFromString(costPrice.String); err == nil {
				product.CostPrice = &price
			}
		}
		if weight.Valid {
			if w, err := decimal.NewFromString(weight.String); err == nil {
				product.Weight = &w
			}
		}
		if dimensionsLength.Valid {
			if dim, err := decimal.NewFromString(dimensionsLength.String); err == nil {
				product.DimensionsLength = &dim
			}
		}
		if dimensionsWidth.Valid {
			if dim, err := decimal.NewFromString(dimensionsWidth.String); err == nil {
				product.DimensionsWidth = &dim
			}
		}
		if dimensionsHeight.Valid {
			if dim, err := decimal.NewFromString(dimensionsHeight.String); err == nil {
				product.DimensionsHeight = &dim
			}
		}
		if lowStockThreshold.Valid {
			threshold := int(lowStockThreshold.Int32)
			product.LowStockThreshold = &threshold
		}
		if metaTitle.Valid {
			product.MetaTitle = &metaTitle.String
		}
		if metaDescription.Valid {
			product.MetaDescription = &metaDescription.String
		}
		if slug.Valid {
			product.Slug = &slug.String
		}
		if deletedAt.Valid {
			product.DeletedAt = &deletedAt.Time
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over product rows: %w", err)
	}

	return products, nil
}

func (r *PostgreSQLProductRepository) SearchWithHighlight(ctx context.Context, query string, filter *repository.ProductFilter) ([]*entity.Product, error) {
	// For now, just use regular search without highlighting
	// TODO: Implement full-text search with highlighting using PostgreSQL ts_headline
	return r.Search(ctx, query, filter, nil)
}

func (r *PostgreSQLProductRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID, filter *repository.ProductFilter, include *repository.ProductInclude) ([]*entity.Product, error) {
	// Create a modified filter with the specific category ID
	categoryFilter := &repository.ProductFilter{
		CategoryIDs: []uuid.UUID{categoryID},
	}

	// Merge with existing filter if provided
	if filter != nil {
		categoryFilter.Status = filter.Status
		categoryFilter.MinPrice = filter.MinPrice
		categoryFilter.MaxPrice = filter.MaxPrice
		categoryFilter.MinStock = filter.MinStock
		categoryFilter.MaxStock = filter.MaxStock
		categoryFilter.SearchQuery = filter.SearchQuery
		categoryFilter.SortBy = filter.SortBy
		categoryFilter.SortOrder = filter.SortOrder
		categoryFilter.Limit = filter.Limit
		categoryFilter.Offset = filter.Offset
	}

	// Use the Search method with category filter
	return r.Search(ctx, "", categoryFilter, include)
}

func (r *PostgreSQLProductRepository) GetByCategoryPath(ctx context.Context, categoryPath string, filter *repository.ProductFilter, include *repository.ProductInclude) ([]*entity.Product, error) {
	// First, find the category ID by path
	var categoryID uuid.UUID
	query := `
		WITH RECURSIVE category_tree AS (
			-- Base case: find root categories
			SELECT id, name, parent_id, path, 1 as level
			FROM product_categories 
			WHERE parent_id IS NULL
			
			UNION ALL
			
			-- Recursive case: find child categories
			SELECT c.id, c.name, c.parent_id, 
			       CASE 
			           WHEN ct.path IS NULL THEN c.name
			           ELSE ct.path || '/' || c.name
			       END as path,
			       ct.level + 1
			FROM product_categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
		SELECT id FROM category_tree WHERE path = $1`

	err := r.db.QueryRowContext(ctx, query, categoryPath).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*entity.Product{}, nil // No category found with this path
		}
		return nil, fmt.Errorf("failed to find category by path: %w", err)
	}

	// Use GetByCategory with the found category ID
	return r.GetByCategory(ctx, categoryID, filter, include)
}

func (r *PostgreSQLProductRepository) UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	// TODO: Implement stock update
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	// TODO: Implement stock deduction
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) RestockInventory(ctx context.Context, productID uuid.UUID, quantity int) error {
	// TODO: Implement restock
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) GetLowStockProducts(ctx context.Context, threshold int, include *repository.ProductInclude) ([]*entity.Product, error) {
	query := `
		SELECT p.id, p.sku, p.name, p.description, p.category_id, p.brand, 
		       p.base_price, p.sale_price, p.cost_price, p.weight, 
		       p.dimensions_length, p.dimensions_width, p.dimensions_height,
		       p.status, p.track_inventory, p.stock_quantity, p.low_stock_threshold,
		       p.meta_title, p.meta_description, p.slug, p.created_by,
		       p.created_at, p.updated_at, p.deleted_at
		FROM products p
		WHERE p.track_inventory = true 
		AND (
			(p.low_stock_threshold IS NOT NULL AND p.stock_quantity <= p.low_stock_threshold) OR
			(p.low_stock_threshold IS NULL AND p.stock_quantity <= $1)
		)
		AND p.status IN ('active', 'inactive')
		ORDER BY p.stock_quantity ASC, p.name ASC`

	rows, err := r.db.QueryContext(ctx, query, threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		product := &entity.Product{}
		var description, brand sql.NullString
		var salePrice, costPrice, weight, dimensionsLength, dimensionsWidth, dimensionsHeight sql.NullString
		var lowStockThreshold sql.NullInt32
		var metaTitle, metaDescription, slug sql.NullString
		var deletedAt sql.NullTime

		if err := rows.Scan(
			&product.ID, &product.SKU, &product.Name, &description, &product.CategoryID, &brand,
			&product.BasePrice, &salePrice, &costPrice, &weight,
			&dimensionsLength, &dimensionsWidth, &dimensionsHeight,
			&product.Status, &product.TrackInventory, &product.StockQuantity, &lowStockThreshold,
			&metaTitle, &metaDescription, &slug, &product.CreatedBy,
			&product.CreatedAt, &product.UpdatedAt, &deletedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan low stock product: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			product.Description = &description.String
		}
		if brand.Valid {
			product.Brand = &brand.String
		}
		if salePrice.Valid {
			if price, err := decimal.NewFromString(salePrice.String); err == nil {
				product.SalePrice = &price
			}
		}
		if costPrice.Valid {
			if price, err := decimal.NewFromString(costPrice.String); err == nil {
				product.CostPrice = &price
			}
		}
		if weight.Valid {
			if w, err := decimal.NewFromString(weight.String); err == nil {
				product.Weight = &w
			}
		}
		if dimensionsLength.Valid {
			if dim, err := decimal.NewFromString(dimensionsLength.String); err == nil {
				product.DimensionsLength = &dim
			}
		}
		if dimensionsWidth.Valid {
			if dim, err := decimal.NewFromString(dimensionsWidth.String); err == nil {
				product.DimensionsWidth = &dim
			}
		}
		if dimensionsHeight.Valid {
			if dim, err := decimal.NewFromString(dimensionsHeight.String); err == nil {
				product.DimensionsHeight = &dim
			}
		}
		if lowStockThreshold.Valid {
			threshold := int(lowStockThreshold.Int32)
			product.LowStockThreshold = &threshold
		}
		if metaTitle.Valid {
			product.MetaTitle = &metaTitle.String
		}
		if metaDescription.Valid {
			product.MetaDescription = &metaDescription.String
		}
		if slug.Valid {
			product.Slug = &slug.String
		}
		if deletedAt.Valid {
			product.DeletedAt = &deletedAt.Time
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over low stock product rows: %w", err)
	}

	return products, nil
}

func (r *PostgreSQLProductRepository) SetFeatured(ctx context.Context, productID uuid.UUID) error {
	// TODO: Implement set featured
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) UnsetFeatured(ctx context.Context, productID uuid.UUID) error {
	// TODO: Implement unset featured
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) GetFeaturedProducts(ctx context.Context, limit int, include *repository.ProductInclude) ([]*entity.Product, error) {
	// TODO: Implement get featured products
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) UpdatePrice(ctx context.Context, productID uuid.UUID, price decimal.Decimal) error {
	// TODO: Implement price update
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) BulkUpdatePrices(ctx context.Context, updates []struct {
	ProductID uuid.UUID
	Price     decimal.Decimal
}) error {
	if len(updates) == 0 {
		return nil
	}

	// Start a transaction for atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare the update statement
	updateQuery := `
		UPDATE products 
		SET base_price = $2, updated_at = NOW() 
		WHERE id = $1 AND deleted_at IS NULL`

	// Execute update for each price update
	for _, update := range updates {
		_, err := tx.ExecContext(ctx, updateQuery, update.ProductID, update.Price)
		if err != nil {
			return fmt.Errorf("failed to update price for product %s: %w", update.ProductID, err)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk price update transaction: %w", err)
	}

	return nil
}

func (r *PostgreSQLProductRepository) GetProductsInPriceRange(ctx context.Context, minPrice, maxPrice decimal.Decimal, include *repository.ProductInclude) ([]*entity.Product, error) {
	// TODO: Implement get products in price range
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) IsSkuExists(ctx context.Context, sku string) (bool, error) {
	return r.ExistsBySKU(ctx, sku)
}

func (r *PostgreSQLProductRepository) IsSkuExistsExcludingProduct(ctx context.Context, sku string, productID uuid.UUID) (bool, error) {
	// TODO: Implement SKU exists excluding product
	return false, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) GenerateNextSKU(ctx context.Context, prefix string) (string, error) {
	// TODO: Implement SKU generation
	return "", fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) GetProductStatistics(ctx context.Context, productID uuid.UUID) (*repository.ProductStatistics, error) {
	// TODO: Implement get product statistics
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) GetTopSellingProducts(ctx context.Context, limit int, days int) ([]*entity.Product, error) {
	// TODO: Implement get top selling products
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) GetRecentlyUpdatedProducts(ctx context.Context, limit int, include *repository.ProductInclude) ([]*entity.Product, error) {
	// Create a filter for recently updated products
	filter := &repository.ProductFilter{
		SortBy:    "updated_at",
		SortOrder: "desc",
		Limit:     limit,
	}

	// Use the Search method with empty query to get all products, sorted by updated_at
	return r.Search(ctx, "", filter, include)
}

func (r *PostgreSQLProductRepository) GetProductsByPriceRange(ctx context.Context, ranges []repository.PriceRange) (map[string]int64, error) {
	// TODO: Implement get products by price range
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductRepository) GetProductCountByCategory(ctx context.Context) (map[uuid.UUID]int64, error) {
	query := `
		SELECT COALESCE(category_id, '00000000-0000-0000-0000-000000000000'::uuid) as category_id, COUNT(*) as count
		FROM products 
		WHERE deleted_at IS NULL
		GROUP BY category_id`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get product count by category: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]int64)
	for rows.Next() {
		var categoryID uuid.UUID
		var count int64
		if err := rows.Scan(&categoryID, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category count: %w", err)
		}
		result[categoryID] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over category count rows: %w", err)
	}

	return result, nil
}

func (r *PostgreSQLProductRepository) GetProductCountByStatus(ctx context.Context) (map[entity.ProductStatus]int64, error) {
	query := `
		SELECT status, COUNT(*) as count
		FROM products 
		WHERE deleted_at IS NULL
		GROUP BY status`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get product count by status: %w", err)
	}
	defer rows.Close()

	result := make(map[entity.ProductStatus]int64)
	for rows.Next() {
		var status entity.ProductStatus
		var count int64
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}
		result[status] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over status count rows: %w", err)
	}

	return result, nil
}

func (r *PostgreSQLProductRepository) ValidateProductData(ctx context.Context, product *entity.Product) error {
	// TODO: Implement product data validation
	return product.Validate()
}

func (r *PostgreSQLProductRepository) CheckDuplicateFields(ctx context.Context, fields map[string]interface{}, excludeID *uuid.UUID) error {
	// TODO: Implement duplicate field checking
	return fmt.Errorf("not implemented")
}
