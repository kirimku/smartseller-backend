package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// PostgreSQLProductImageRepository implements the ProductImageRepository interface using PostgreSQL
type PostgreSQLProductImageRepository struct {
	db *sqlx.DB
}

// NewPostgreSQLProductImageRepository creates a new PostgreSQL product image repository
func NewPostgreSQLProductImageRepository(db *sqlx.DB) repository.ProductImageRepository {
	return &PostgreSQLProductImageRepository{
		db: db,
	}
}

// Create creates a new product image in the database
func (r *PostgreSQLProductImageRepository) Create(ctx context.Context, image *entity.ProductImage) error {
	// Validate image before creating
	if err := image.Validate(); err != nil {
		return fmt.Errorf("image validation failed: %w", err)
	}

	// Ensure ID is set
	if image.ID == uuid.Nil {
		image.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	image.CreatedAt = now
	image.UpdatedAt = now

	query := `
		INSERT INTO product_images (
			id, product_id, variant_id, image_url, cloudinary_url, cloudinary_public_id, 
			alt_text, width, height, file_size, mime_type, is_primary, sort_order, created_at, updated_at
		) VALUES (
			:id, :product_id, :variant_id, :image_url, :cloudinary_url, :cloudinary_public_id,
			:alt_text, :width, :height, :file_size, :mime_type, :is_primary, :sort_order, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, image)
	if err != nil {
		// Handle foreign key constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // foreign_key_violation
				if err.(*pq.Error).Constraint == "product_images_product_id_fkey" {
					return fmt.Errorf("product with ID '%s' does not exist", image.ProductID)
				}
				if err.(*pq.Error).Constraint == "product_images_variant_id_fkey" {
					return fmt.Errorf("variant with ID '%s' does not exist", *image.VariantID)
				}
				return fmt.Errorf("foreign key violation: %w", err)
			}
		}
		return fmt.Errorf("failed to create image: %w", err)
	}

	return nil
}

// GetByID retrieves a product image by its ID
func (r *PostgreSQLProductImageRepository) GetByID(ctx context.Context, id uuid.UUID, include *repository.ProductImageInclude) (*entity.ProductImage, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("image ID cannot be nil")
	}

	query := `
		SELECT id, product_id, variant_id, image_url, cloudinary_url, cloudinary_public_id,
			   alt_text, width, height, file_size, mime_type, is_primary, sort_order, created_at, updated_at
		FROM product_images
		WHERE id = $1`

	var image entity.ProductImage
	err := r.db.GetContext(ctx, &image, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("image with ID '%s' not found", id)
		}
		return nil, fmt.Errorf("failed to get image by ID: %w", err)
	}

	// Compute fields
	image.ComputeFields()

	return &image, nil
}

// Update updates an existing product image
func (r *PostgreSQLProductImageRepository) Update(ctx context.Context, image *entity.ProductImage) error {
	// Validate image before updating
	if err := image.Validate(); err != nil {
		return fmt.Errorf("image validation failed: %w", err)
	}

	if image.ID == uuid.Nil {
		return fmt.Errorf("image ID cannot be nil for update")
	}

	// Update timestamp
	image.UpdatedAt = time.Now()

	query := `
		UPDATE product_images SET
			image_url = :image_url,
			cloudinary_url = :cloudinary_url,
			cloudinary_public_id = :cloudinary_public_id,
			alt_text = :alt_text,
			width = :width,
			height = :height,
			file_size = :file_size,
			mime_type = :mime_type,
			is_primary = :is_primary,
			sort_order = :sort_order,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, image)
	if err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with ID '%s' not found", image.ID)
	}

	return nil
}

// Delete deletes a product image
func (r *PostgreSQLProductImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("image ID cannot be nil")
	}

	query := `DELETE FROM product_images WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with ID '%s' not found", id)
	}

	return nil
}

// GetByProduct retrieves all images for a product
func (r *PostgreSQLProductImageRepository) GetByProduct(ctx context.Context, productID uuid.UUID, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be nil")
	}

	query := `
		SELECT id, product_id, variant_id, image_url, cloudinary_url, cloudinary_public_id,
			   alt_text, width, height, file_size, mime_type, is_primary, sort_order, created_at, updated_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY sort_order ASC, created_at ASC`

	var images []*entity.ProductImage
	err := r.db.SelectContext(ctx, &images, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images by product: %w", err)
	}

	// Compute fields for all images
	for _, image := range images {
		image.ComputeFields()
	}

	return images, nil
}

// GetByVariant retrieves all images for a product variant
func (r *PostgreSQLProductImageRepository) GetByVariant(ctx context.Context, variantID uuid.UUID, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	if variantID == uuid.Nil {
		return nil, fmt.Errorf("variant ID cannot be nil")
	}

	query := `
		SELECT id, product_id, variant_id, image_url, cloudinary_url, cloudinary_public_id,
			   alt_text, width, height, file_size, mime_type, is_primary, sort_order, created_at, updated_at
		FROM product_images
		WHERE variant_id = $1
		ORDER BY is_primary DESC, sort_order ASC`

	var images []entity.ProductImage
	err := r.db.SelectContext(ctx, &images, query, variantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images by variant: %w", err)
	}

	// Convert to pointer slice and compute fields
	result := make([]*entity.ProductImage, len(images))
	for i := range images {
		result[i] = &images[i]
		result[i].ComputeFields()
	}

	return result, nil
}

// SetPrimaryImage sets an image as primary for its product/variant
func (r *PostgreSQLProductImageRepository) SetPrimaryImage(ctx context.Context, imageID uuid.UUID) error {
	if imageID == uuid.Nil {
		return fmt.Errorf("image ID cannot be nil")
	}

	// Begin transaction to ensure consistency
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the product_id and variant_id for this image
	var productID uuid.UUID
	var variantID *uuid.UUID
	err = tx.GetContext(ctx, &struct {
		ProductID uuid.UUID  `db:"product_id"`
		VariantID *uuid.UUID `db:"variant_id"`
	}{ProductID: productID, VariantID: variantID},
		"SELECT product_id, variant_id FROM product_images WHERE id = $1", imageID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("image with ID '%s' not found", imageID)
		}
		return fmt.Errorf("failed to get image info: %w", err)
	}

	// Clear primary flag from all images of the same product/variant
	var clearQuery string
	var clearArgs []interface{}
	if variantID != nil {
		clearQuery = "UPDATE product_images SET is_primary = false, updated_at = $1 WHERE variant_id = $2"
		clearArgs = []interface{}{time.Now(), *variantID}
	} else {
		clearQuery = "UPDATE product_images SET is_primary = false, updated_at = $1 WHERE product_id = $2 AND variant_id IS NULL"
		clearArgs = []interface{}{time.Now(), productID}
	}

	_, err = tx.ExecContext(ctx, clearQuery, clearArgs...)
	if err != nil {
		return fmt.Errorf("failed to clear primary flags: %w", err)
	}

	// Set the specified image as primary
	result, err := tx.ExecContext(ctx,
		"UPDATE product_images SET is_primary = true, updated_at = $1 WHERE id = $2",
		time.Now(), imageID)
	if err != nil {
		return fmt.Errorf("failed to set primary image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with ID '%s' not found", imageID)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateSortOrder updates the sort order of images
func (r *PostgreSQLProductImageRepository) UpdateSortOrder(ctx context.Context, imageID uuid.UUID, sortOrder int) error {
	if imageID == uuid.Nil {
		return fmt.Errorf("image ID cannot be nil")
	}

	query := `UPDATE product_images SET sort_order = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, sortOrder, time.Now(), imageID)
	if err != nil {
		return fmt.Errorf("failed to update sort order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("image with ID '%s' not found", imageID)
	}

	return nil
}

// GetPrimaryImage retrieves the primary image for a product or variant
func (r *PostgreSQLProductImageRepository) GetPrimaryImage(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (*entity.ProductImage, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be nil")
	}

	var query string
	var args []interface{}

	if variantID != nil {
		query = `
			SELECT id, product_id, variant_id, image_url, cloudinary_url, cloudinary_public_id,
				   alt_text, width, height, file_size, mime_type, is_primary, sort_order, created_at, updated_at
			FROM product_images
			WHERE variant_id = $1 AND is_primary = true`
		args = []interface{}{*variantID}
	} else {
		query = `
			SELECT id, product_id, variant_id, image_url, cloudinary_url, cloudinary_public_id,
				   alt_text, width, height, file_size, mime_type, is_primary, sort_order, created_at, updated_at
			FROM product_images
			WHERE product_id = $1 AND variant_id IS NULL AND is_primary = true`
		args = []interface{}{productID}
	}

	var image entity.ProductImage
	err := r.db.GetContext(ctx, &image, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("primary image not found")
		}
		return nil, fmt.Errorf("failed to get primary image: %w", err)
	}

	// Compute fields
	image.ComputeFields()

	return &image, nil
}

// Exists checks if an image exists by ID
func (r *PostgreSQLProductImageRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	if id == uuid.Nil {
		return false, fmt.Errorf("image ID cannot be nil")
	}

	query := `SELECT EXISTS(SELECT 1 FROM product_images WHERE id = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check image existence: %w", err)
	}

	return exists, nil
}

// Placeholder implementations for remaining methods
// These would need full implementation based on business requirements

func (r *PostgreSQLProductImageRepository) CreateBatch(ctx context.Context, images []*entity.ProductImage) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetByIDs(ctx context.Context, ids []uuid.UUID, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) UpdateBatch(ctx context.Context, images []*entity.ProductImage) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return fmt.Errorf("image IDs cannot be empty")
	}

	query := `DELETE FROM product_images WHERE id = ANY($1)`

	result, err := r.db.ExecContext(ctx, query, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to bulk delete images: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no images found with provided IDs")
	}

	return nil
}

func (r *PostgreSQLProductImageRepository) List(ctx context.Context, filter *repository.ProductImageFilter, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) Count(ctx context.Context, filter *repository.ProductImageFilter) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetByCloudinaryID(ctx context.Context, cloudinaryID string, include *repository.ProductImageInclude) (*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetByURL(ctx context.Context, url string, include *repository.ProductImageInclude) (*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) DeleteByProduct(ctx context.Context, productID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) DeleteByVariant(ctx context.Context, variantID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) DeleteByCloudinaryPublicID(ctx context.Context, publicID string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) ReorderImages(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID, imageOrders []repository.ImageOrder) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetNextSortOrder(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) CopyImagesToVariant(ctx context.Context, sourceProductID, targetVariantID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImagesByMimeType(ctx context.Context, mimeType string, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImagesBySize(ctx context.Context, minSize, maxSize int64, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetLargeImages(ctx context.Context, minSizeBytes int64) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetBrokenImages(ctx context.Context) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImagesWithoutCloudinary(ctx context.Context, limit int) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) UpdateCloudinaryInfo(ctx context.Context, imageID uuid.UUID, cloudinaryID, cloudinaryURL, thumbnailURL, mediumURL, largeURL *string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) BulkUpdateCloudinaryInfo(ctx context.Context, updates []interface{}) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) BulkUpdateAltText(ctx context.Context, altTextUpdates []repository.ImageAltTextUpdate) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) BulkUpdateMetadata(ctx context.Context, updates []repository.ImageMetadataUpdate) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) BulkUpdateURLs(ctx context.Context, urlUpdates []repository.ImageURLUpdate) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) CheckImageAccessibility(ctx context.Context, imageURL string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) BulkOptimize(ctx context.Context, ids []uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) CleanupOrphanedImages(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImageStatistics(ctx context.Context, imageID uuid.UUID) (*repository.ImageStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) Search(ctx context.Context, query string, filter *repository.ProductImageFilter, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) SearchByAltText(ctx context.Context, query string, filter *repository.ProductImageFilter, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) CountByProduct(ctx context.Context, productID uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) CountByVariant(ctx context.Context, variantID uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetTotalFileSize(ctx context.Context, productID uuid.UUID) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) ValidateImageConstraints(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImagesByDimensions(ctx context.Context, minWidth, maxWidth, minHeight, maxHeight int, include *repository.ProductImageInclude) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) OptimizeImageOrder(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

// Missing interface methods - placeholder implementations
func (r *PostgreSQLProductImageRepository) CleanupDuplicateImages(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) ExportImages(ctx context.Context, productID uuid.UUID) ([]*repository.ImageExport, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImageCount(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (int, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImageSizeDistribution(ctx context.Context) (map[string]int64, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImageTypes(ctx context.Context, productID uuid.UUID) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImageUsageReport(ctx context.Context) (*repository.ImageUsageReport, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImagesWithoutAltText(ctx context.Context, limit int) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetImagesWithoutOptimization(ctx context.Context, limit int) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetMimeTypeDistribution(ctx context.Context) (map[string]int64, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetMostViewedImages(ctx context.Context, limit int, days int) ([]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetPrimaryImages(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetPrimaryImagesByVariant(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID]*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetProductPrimaryImage(ctx context.Context, productID uuid.UUID, include *repository.ProductImageInclude) (*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetVariantPrimaryImage(ctx context.Context, variantID uuid.UUID, include *repository.ProductImageInclude) (*entity.ProductImage, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) SetProductPrimaryImage(ctx context.Context, productID uuid.UUID, imageID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) SetVariantPrimaryImage(ctx context.Context, variantID uuid.UUID, imageID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) UnsetPrimaryImages(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) GetTotalImageSize(ctx context.Context, productID uuid.UUID) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) ImportImages(ctx context.Context, productID uuid.UUID, images []*repository.ImageImport) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) IsURLExists(ctx context.Context, imageURL string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) IsURLExistsExcluding(ctx context.Context, imageURL string, imageID uuid.UUID) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) RemoveDeadLinks(ctx context.Context) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) RemoveUnusedImages(ctx context.Context, daysUnused int) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) UpdateAltText(ctx context.Context, imageID uuid.UUID, altText string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) UpdateFileInfo(ctx context.Context, imageID uuid.UUID, fileName *string, fileSize *int64, mimeType *string, width, height *int) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) UpdateImageURL(ctx context.Context, imageID uuid.UUID, newURL string) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) ValidateImageDimensions(ctx context.Context, width, height int) error {
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductImageRepository) ValidateImageURL(ctx context.Context, imageURL string) error {
	return fmt.Errorf("not implemented")
}
