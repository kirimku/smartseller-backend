package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// ProductImageFilter defines filtering options for image queries
type ProductImageFilter struct {
	// Basic filters
	IDs        []uuid.UUID `json:"ids,omitempty"`
	ProductIDs []uuid.UUID `json:"product_ids,omitempty"`
	VariantIDs []uuid.UUID `json:"variant_ids,omitempty"`

	// Image properties
	IsPrimary *bool    `json:"is_primary,omitempty"`
	MimeTypes []string `json:"mime_types,omitempty"`
	FileNames []string `json:"file_names,omitempty"`

	// File properties
	MinFileSize *int64 `json:"min_file_size,omitempty"`
	MaxFileSize *int64 `json:"max_file_size,omitempty"`
	MinWidth    *int   `json:"min_width,omitempty"`
	MaxWidth    *int   `json:"max_width,omitempty"`
	MinHeight   *int   `json:"min_height,omitempty"`
	MaxHeight   *int   `json:"max_height,omitempty"`

	// External service filters
	HasCloudinaryID  *bool `json:"has_cloudinary_id,omitempty"`
	HasThumbnail     *bool `json:"has_thumbnail,omitempty"`
	HasMultipleSizes *bool `json:"has_multiple_sizes,omitempty"`
	IsOptimized      *bool `json:"is_optimized,omitempty"`

	// Date filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`

	// Text search
	SearchQuery string `json:"search_query,omitempty"` // Search in alt text, file name

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // sort_order, created_at, file_size, width, height
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ProductImageInclude defines what related data to include
type ProductImageInclude struct {
	Product   bool `json:"product,omitempty"`
	Variant   bool `json:"variant,omitempty"`
	FileStats bool `json:"file_stats,omitempty"`
}

// ProductImageRepository defines the interface for image data access
type ProductImageRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, image *entity.ProductImage) error
	GetByID(ctx context.Context, id uuid.UUID, include *ProductImageInclude) (*entity.ProductImage, error)
	Update(ctx context.Context, image *entity.ProductImage) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations
	CreateBatch(ctx context.Context, images []*entity.ProductImage) error
	GetByIDs(ctx context.Context, ids []uuid.UUID, include *ProductImageInclude) ([]*entity.ProductImage, error)
	UpdateBatch(ctx context.Context, images []*entity.ProductImage) error
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error

	// Query operations
	List(ctx context.Context, filter *ProductImageFilter, include *ProductImageInclude) ([]*entity.ProductImage, error)
	Count(ctx context.Context, filter *ProductImageFilter) (int64, error)

	// Product-specific operations
	GetByProduct(ctx context.Context, productID uuid.UUID, include *ProductImageInclude) ([]*entity.ProductImage, error)
	GetProductPrimaryImage(ctx context.Context, productID uuid.UUID, include *ProductImageInclude) (*entity.ProductImage, error)
	SetProductPrimaryImage(ctx context.Context, productID uuid.UUID, imageID uuid.UUID) error
	CountByProduct(ctx context.Context, productID uuid.UUID) (int, error)

	// Variant-specific operations
	GetByVariant(ctx context.Context, variantID uuid.UUID, include *ProductImageInclude) ([]*entity.ProductImage, error)
	GetVariantPrimaryImage(ctx context.Context, variantID uuid.UUID, include *ProductImageInclude) (*entity.ProductImage, error)
	SetVariantPrimaryImage(ctx context.Context, variantID uuid.UUID, imageID uuid.UUID) error
	CountByVariant(ctx context.Context, variantID uuid.UUID) (int, error)

	// Primary image management
	GetPrimaryImages(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]*entity.ProductImage, error)
	GetPrimaryImagesByVariant(ctx context.Context, variantIDs []uuid.UUID) (map[uuid.UUID]*entity.ProductImage, error)
	UnsetPrimaryImages(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) error

	// Sort order management
	UpdateSortOrder(ctx context.Context, imageID uuid.UUID, sortOrder int) error
	ReorderImages(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID, imageOrders []ImageOrder) error
	GetNextSortOrder(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (int, error)

	// URL management
	GetByURL(ctx context.Context, imageURL string, include *ProductImageInclude) (*entity.ProductImage, error)
	IsURLExists(ctx context.Context, imageURL string) (bool, error)
	IsURLExistsExcluding(ctx context.Context, imageURL string, imageID uuid.UUID) (bool, error)
	UpdateImageURL(ctx context.Context, imageID uuid.UUID, newURL string) error
	BulkUpdateURLs(ctx context.Context, urlUpdates []ImageURLUpdate) error

	// File information management
	UpdateFileInfo(ctx context.Context, imageID uuid.UUID, fileName *string, fileSize *int64, mimeType *string, width, height *int) error
	GetImagesByMimeType(ctx context.Context, mimeType string, include *ProductImageInclude) ([]*entity.ProductImage, error)
	GetImagesBySize(ctx context.Context, minSize, maxSize int64, include *ProductImageInclude) ([]*entity.ProductImage, error)
	GetImagesByDimensions(ctx context.Context, minWidth, maxWidth, minHeight, maxHeight int, include *ProductImageInclude) ([]*entity.ProductImage, error)

	// Cloudinary management
	UpdateCloudinaryInfo(ctx context.Context, imageID uuid.UUID, cloudinaryID, cloudinaryURL, thumbnailURL, mediumURL, largeURL *string) error
	GetByCloudinaryID(ctx context.Context, cloudinaryID string, include *ProductImageInclude) (*entity.ProductImage, error)
	GetImagesWithoutCloudinary(ctx context.Context, limit int) ([]*entity.ProductImage, error)
	GetImagesWithoutOptimization(ctx context.Context, limit int) ([]*entity.ProductImage, error)

	// Alt text management
	UpdateAltText(ctx context.Context, imageID uuid.UUID, altText string) error
	GetImagesWithoutAltText(ctx context.Context, limit int) ([]*entity.ProductImage, error)
	BulkUpdateAltText(ctx context.Context, altTextUpdates []ImageAltTextUpdate) error

	// Search operations
	Search(ctx context.Context, query string, filter *ProductImageFilter, include *ProductImageInclude) ([]*entity.ProductImage, error)
	SearchByAltText(ctx context.Context, query string, filter *ProductImageFilter, include *ProductImageInclude) ([]*entity.ProductImage, error)

	// Analytics operations
	GetImageStatistics(ctx context.Context, imageID uuid.UUID) (*ImageStatistics, error)
	GetMostViewedImages(ctx context.Context, limit int, days int) ([]*entity.ProductImage, error)
	GetImageUsageReport(ctx context.Context) (*ImageUsageReport, error)
	GetImageSizeDistribution(ctx context.Context) (map[string]int64, error)
	GetMimeTypeDistribution(ctx context.Context) (map[string]int64, error)

	// Validation operations
	ValidateImageURL(ctx context.Context, imageURL string) error
	CheckImageAccessibility(ctx context.Context, imageURL string) error
	ValidateImageDimensions(ctx context.Context, width, height int) error

	// Cleanup operations
	CleanupOrphanedImages(ctx context.Context) (int64, error)
	RemoveDeadLinks(ctx context.Context) (int64, error)
	CleanupDuplicateImages(ctx context.Context) (int64, error)
	RemoveUnusedImages(ctx context.Context, daysUnused int) (int64, error)

	// Import/Export operations
	ExportImages(ctx context.Context, productID uuid.UUID) ([]*ImageExport, error)
	ImportImages(ctx context.Context, productID uuid.UUID, images []*ImageImport) error

	// Bulk operations
	BulkDelete(ctx context.Context, imageIDs []uuid.UUID) error
	BulkUpdateMetadata(ctx context.Context, updates []ImageMetadataUpdate) error
	BulkOptimize(ctx context.Context, imageIDs []uuid.UUID) error

	// Utility operations
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetImageCount(ctx context.Context, productID uuid.UUID, variantID *uuid.UUID) (int, error)
	GetTotalImageSize(ctx context.Context, productID uuid.UUID) (int64, error)
	GetImageTypes(ctx context.Context, productID uuid.UUID) ([]string, error)
}

// ImageOrder defines the sort order for an image
type ImageOrder struct {
	ImageID   uuid.UUID `json:"image_id"`
	SortOrder int       `json:"sort_order"`
}

// ImageURLUpdate defines a URL update operation
type ImageURLUpdate struct {
	ImageID uuid.UUID `json:"image_id"`
	NewURL  string    `json:"new_url"`
}

// ImageAltTextUpdate defines an alt text update operation
type ImageAltTextUpdate struct {
	ImageID uuid.UUID `json:"image_id"`
	AltText string    `json:"alt_text"`
}

// ImageMetadataUpdate defines a metadata update operation
type ImageMetadataUpdate struct {
	ImageID  uuid.UUID `json:"image_id"`
	FileName *string   `json:"file_name,omitempty"`
	AltText  *string   `json:"alt_text,omitempty"`
	MimeType *string   `json:"mime_type,omitempty"`
}

// ImageStatistics contains analytics data for an image
type ImageStatistics struct {
	ImageID         uuid.UUID  `json:"image_id"`
	TotalViews      int64      `json:"total_views"`
	UniqueViews     int64      `json:"unique_views"`
	ClickThroughs   int64      `json:"click_throughs"`
	LoadTime        *int       `json:"avg_load_time_ms,omitempty"`
	ErrorCount      int64      `json:"error_count"`
	LastViewedAt    *time.Time `json:"last_viewed_at,omitempty"`
	ConversionRate  *string    `json:"conversion_rate,omitempty"`
	PopularityScore *int       `json:"popularity_score,omitempty"`
}

// ImageUsageReport contains overall image usage statistics
type ImageUsageReport struct {
	TotalImages        int64            `json:"total_images"`
	ProductImages      int64            `json:"product_images"`
	VariantImages      int64            `json:"variant_images"`
	PrimaryImages      int64            `json:"primary_images"`
	OptimizedImages    int64            `json:"optimized_images"`
	ImagesWithAltText  int64            `json:"images_with_alt_text"`
	TotalFileSize      int64            `json:"total_file_size_bytes"`
	AverageFileSize    int64            `json:"average_file_size_bytes"`
	MimeTypeBreakdown  map[string]int64 `json:"mime_type_breakdown"`
	SizeRangeBreakdown map[string]int64 `json:"size_range_breakdown"`
	OrphanedImages     int64            `json:"orphaned_images"`
	DeadLinks          int64            `json:"dead_links"`
}

// ImageExport defines the structure for exporting images
type ImageExport struct {
	ID            uuid.UUID  `json:"id"`
	ProductID     uuid.UUID  `json:"product_id"`
	VariantID     *uuid.UUID `json:"variant_id,omitempty"`
	ImageURL      string     `json:"image_url"`
	AltText       *string    `json:"alt_text,omitempty"`
	IsPrimary     bool       `json:"is_primary"`
	SortOrder     int        `json:"sort_order"`
	FileName      *string    `json:"file_name,omitempty"`
	FileSize      *int64     `json:"file_size,omitempty"`
	MimeType      *string    `json:"mime_type,omitempty"`
	Width         *int       `json:"width,omitempty"`
	Height        *int       `json:"height,omitempty"`
	CloudinaryID  *string    `json:"cloudinary_id,omitempty"`
	CloudinaryURL *string    `json:"cloudinary_url,omitempty"`
	ThumbnailURL  *string    `json:"thumbnail_url,omitempty"`
	MediumURL     *string    `json:"medium_url,omitempty"`
	LargeURL      *string    `json:"large_url,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ImageImport defines the structure for importing images
type ImageImport struct {
	VariantID       *uuid.UUID `json:"variant_id,omitempty"`
	ImageURL        string     `json:"image_url"`
	AltText         *string    `json:"alt_text,omitempty"`
	IsPrimary       *bool      `json:"is_primary,omitempty"`
	SortOrder       *int       `json:"sort_order,omitempty"`
	FileName        *string    `json:"file_name,omitempty"`
	AutoOptimize    *bool      `json:"auto_optimize,omitempty"`     // Whether to automatically optimize the image
	GenerateAltText *bool      `json:"generate_alt_text,omitempty"` // Whether to auto-generate alt text
}
