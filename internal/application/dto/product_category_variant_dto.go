package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ========================================
// Product Category DTOs
// ========================================

// CreateCategoryRequest represents the request to create a new product category
type CreateCategoryRequest struct {
	Name            string     `json:"name" validate:"required,min=1,max=255" example:"Electronics"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=1000" example:"All electronic products"`
	ParentID        *uuid.UUID `json:"parent_id,omitempty" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Slug            *string    `json:"slug,omitempty" validate:"omitempty,max=255,slug" example:"electronics"`
	ImageURL        *string    `json:"image_url,omitempty" validate:"omitempty,url" example:"https://example.com/category.jpg"`
	IsActive        *bool      `json:"is_active,omitempty" example:"true"`
	SortOrder       *int       `json:"sort_order,omitempty" validate:"omitempty,min=0" example:"1"`
	MetaTitle       *string    `json:"meta_title,omitempty" validate:"omitempty,max=255"`
	MetaDescription *string    `json:"meta_description,omitempty" validate:"omitempty,max=500"`
}

// UpdateCategoryRequest represents the request to update an existing product category
type UpdateCategoryRequest struct {
	Name            *string    `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description     *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	ParentID        *uuid.UUID `json:"parent_id,omitempty" validate:"omitempty,uuid4"`
	Slug            *string    `json:"slug,omitempty" validate:"omitempty,max=255,slug"`
	ImageURL        *string    `json:"image_url,omitempty" validate:"omitempty,url"`
	IsActive        *bool      `json:"is_active,omitempty"`
	SortOrder       *int       `json:"sort_order,omitempty" validate:"omitempty,min=0"`
	MetaTitle       *string    `json:"meta_title,omitempty" validate:"omitempty,max=255"`
	MetaDescription *string    `json:"meta_description,omitempty" validate:"omitempty,max=500"`
}

// CategoryResponse represents the response for a single product category
type CategoryResponse struct {
	ID              uuid.UUID         `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name            string            `json:"name" example:"Electronics"`
	Description     *string           `json:"description,omitempty" example:"All electronic products"`
	ParentID        *uuid.UUID        `json:"parent_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	Parent          *CategorySummary  `json:"parent,omitempty"`
	Children        []CategorySummary `json:"children,omitempty"`
	Path            string            `json:"path" example:"Electronics"`
	Level           int               `json:"level" example:"1"`
	Slug            *string           `json:"slug,omitempty" example:"electronics"`
	ImageURL        *string           `json:"image_url,omitempty" example:"https://example.com/category.jpg"`
	IsActive        bool              `json:"is_active" example:"true"`
	SortOrder       int               `json:"sort_order" example:"1"`
	ProductCount    int               `json:"product_count" example:"150"`
	MetaTitle       *string           `json:"meta_title,omitempty"`
	MetaDescription *string           `json:"meta_description,omitempty"`
	CreatedAt       time.Time         `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt       time.Time         `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// CategoryTreeResponse represents a hierarchical category tree
type CategoryTreeResponse struct {
	Categories []CategoryTreeNode `json:"categories"`
	TotalCount int                `json:"total_count" example:"25"`
}

// CategoryTreeNode represents a node in the category tree
type CategoryTreeNode struct {
	ID           uuid.UUID          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name         string             `json:"name" example:"Electronics"`
	Description  *string            `json:"description,omitempty"`
	Path         string             `json:"path" example:"Electronics"`
	Level        int                `json:"level" example:"1"`
	ProductCount int                `json:"product_count" example:"150"`
	IsActive     bool               `json:"is_active" example:"true"`
	SortOrder    int                `json:"sort_order" example:"1"`
	Children     []CategoryTreeNode `json:"children,omitempty"`
}

// CategoryListResponse represents a paginated list of categories
type CategoryListResponse struct {
	Categories []CategorySummary   `json:"categories"`
	Pagination PaginationResponse  `json:"pagination"`
	Summary    CategoryListSummary `json:"summary,omitempty"`
}

// CategoryListSummary provides aggregate information about categories
type CategoryListSummary struct {
	TotalCategories  int `json:"total_categories" example:"25"`
	ActiveCategories int `json:"active_categories" example:"22"`
	RootCategories   int `json:"root_categories" example:"5"`
	MaxDepthLevel    int `json:"max_depth_level" example:"3"`
}

// MoveCategoryRequest represents a request to move a category to a new parent
type MoveCategoryRequest struct {
	CategoryID   uuid.UUID  `json:"category_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	NewParentID  *uuid.UUID `json:"new_parent_id,omitempty" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440001"`
	NewSortOrder *int       `json:"new_sort_order,omitempty" validate:"omitempty,min=0" example:"2"`
}

// ========================================
// Product Variant DTOs
// ========================================

// CreateVariantOptionRequest represents the request to create a variant option
type CreateVariantOptionRequest struct {
	ProductID   uuid.UUID `json:"product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	OptionName  string    `json:"option_name" validate:"required,min=1,max=100" example:"Color"`
	OptionType  string    `json:"option_type" validate:"required,oneof=text select color size" example:"color"`
	Values      []string  `json:"values" validate:"required,min=1,dive,max=50" example:"Red,Blue,Green"`
	IsRequired  bool      `json:"is_required" example:"true"`
	SortOrder   *int      `json:"sort_order,omitempty" validate:"omitempty,min=0" example:"1"`
	DisplayName *string   `json:"display_name,omitempty" validate:"omitempty,max=100" example:"Choose Color"`
}

// UpdateVariantOptionRequest represents the request to update a variant option
type UpdateVariantOptionRequest struct {
	OptionName  *string  `json:"option_name,omitempty" validate:"omitempty,min=1,max=100"`
	OptionType  *string  `json:"option_type,omitempty" validate:"omitempty,oneof=text select color size"`
	Values      []string `json:"values,omitempty" validate:"omitempty,min=1,dive,max=50"`
	IsRequired  *bool    `json:"is_required,omitempty"`
	SortOrder   *int     `json:"sort_order,omitempty" validate:"omitempty,min=0"`
	DisplayName *string  `json:"display_name,omitempty" validate:"omitempty,max=100"`
}

// VariantOptionResponse represents the response for a variant option
type VariantOptionResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductID   uuid.UUID `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ProductName string    `json:"product_name,omitempty" example:"Wireless Headphones"`
	OptionName  string    `json:"option_name" example:"Color"`
	OptionType  string    `json:"option_type" example:"color"`
	Values      []string  `json:"values" example:"Red,Blue,Green"`
	IsRequired  bool      `json:"is_required" example:"true"`
	SortOrder   int       `json:"sort_order" example:"1"`
	DisplayName *string   `json:"display_name,omitempty" example:"Choose Color"`
	CreatedAt   time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// CreateVariantRequest represents the request to create a product variant
type CreateVariantRequest struct {
	ProductID         uuid.UUID              `json:"product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	SKU               string                 `json:"sku" validate:"required,min=3,max=100,alphanum_underscore_hyphen" example:"WBH-001-RED"`
	VariantName       *string                `json:"variant_name,omitempty" validate:"omitempty,max=255" example:"Red Wireless Headphones"`
	Options           map[string]interface{} `json:"options" validate:"required" example:"{\"color\":\"red\",\"size\":\"medium\"}"`
	BasePrice         *decimal.Decimal       `json:"base_price,omitempty" validate:"omitempty,min=0" example:"199.99"`
	SalePrice         *decimal.Decimal       `json:"sale_price,omitempty" validate:"omitempty,min=0" example:"149.99"`
	CostPrice         *decimal.Decimal       `json:"cost_price,omitempty" validate:"omitempty,min=0" example:"80.00"`
	StockQuantity     int                    `json:"stock_quantity" validate:"min=0" example:"25"`
	LowStockThreshold *int                   `json:"low_stock_threshold,omitempty" validate:"omitempty,min=0" example:"5"`
	Weight            *decimal.Decimal       `json:"weight,omitempty" validate:"omitempty,min=0" example:"0.25"`
	DimensionsLength  *decimal.Decimal       `json:"dimensions_length,omitempty" validate:"omitempty,min=0" example:"20.5"`
	DimensionsWidth   *decimal.Decimal       `json:"dimensions_width,omitempty" validate:"omitempty,min=0" example:"15.2"`
	DimensionsHeight  *decimal.Decimal       `json:"dimensions_height,omitempty" validate:"omitempty,min=0" example:"8.5"`
	IsActive          *bool                  `json:"is_active,omitempty" example:"true"`
	IsDefault         *bool                  `json:"is_default,omitempty" example:"false"`
}

// UpdateVariantRequest represents the request to update a product variant
type UpdateVariantRequest struct {
	VariantName       *string                `json:"variant_name,omitempty" validate:"omitempty,max=255"`
	Options           map[string]interface{} `json:"options,omitempty"`
	BasePrice         *decimal.Decimal       `json:"base_price,omitempty" validate:"omitempty,min=0"`
	SalePrice         *decimal.Decimal       `json:"sale_price,omitempty" validate:"omitempty,min=0"`
	CostPrice         *decimal.Decimal       `json:"cost_price,omitempty" validate:"omitempty,min=0"`
	StockQuantity     *int                   `json:"stock_quantity,omitempty" validate:"omitempty,min=0"`
	LowStockThreshold *int                   `json:"low_stock_threshold,omitempty" validate:"omitempty,min=0"`
	Weight            *decimal.Decimal       `json:"weight,omitempty" validate:"omitempty,min=0"`
	DimensionsLength  *decimal.Decimal       `json:"dimensions_length,omitempty" validate:"omitempty,min=0"`
	DimensionsWidth   *decimal.Decimal       `json:"dimensions_width,omitempty" validate:"omitempty,min=0"`
	DimensionsHeight  *decimal.Decimal       `json:"dimensions_height,omitempty" validate:"omitempty,min=0"`
	IsActive          *bool                  `json:"is_active,omitempty"`
	IsDefault         *bool                  `json:"is_default,omitempty"`
}

// VariantResponse represents the response for a product variant
type VariantResponse struct {
	ID                uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductID         uuid.UUID              `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ProductName       string                 `json:"product_name,omitempty" example:"Wireless Headphones"`
	SKU               string                 `json:"sku" example:"WBH-001-RED"`
	VariantName       *string                `json:"variant_name,omitempty" example:"Red Wireless Headphones"`
	Options           map[string]interface{} `json:"options" example:"{\"color\":\"red\",\"size\":\"medium\"}"`
	BasePrice         decimal.Decimal        `json:"base_price" example:"199.99"`
	SalePrice         *decimal.Decimal       `json:"sale_price,omitempty" example:"149.99"`
	CostPrice         *decimal.Decimal       `json:"cost_price,omitempty" example:"80.00"`
	EffectivePrice    decimal.Decimal        `json:"effective_price" example:"149.99"`
	ProfitMargin      *decimal.Decimal       `json:"profit_margin,omitempty" example:"46.67"`
	StockQuantity     int                    `json:"stock_quantity" example:"25"`
	LowStockThreshold *int                   `json:"low_stock_threshold,omitempty" example:"5"`
	IsLowStock        bool                   `json:"is_low_stock" example:"false"`
	Weight            *decimal.Decimal       `json:"weight,omitempty" example:"0.25"`
	DimensionsLength  *decimal.Decimal       `json:"dimensions_length,omitempty" example:"20.5"`
	DimensionsWidth   *decimal.Decimal       `json:"dimensions_width,omitempty" example:"15.2"`
	DimensionsHeight  *decimal.Decimal       `json:"dimensions_height,omitempty" example:"8.5"`
	IsActive          bool                   `json:"is_active" example:"true"`
	IsDefault         bool                   `json:"is_default" example:"false"`
	CreatedAt         time.Time              `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt         time.Time              `json:"updated_at" example:"2023-01-01T00:00:00Z"`

	// Related data
	Images []ProductImageSummary `json:"images,omitempty"`
}

// VariantListResponse represents a paginated list of variants
type VariantListResponse struct {
	Variants   []VariantSummary   `json:"variants"`
	Pagination PaginationResponse `json:"pagination"`
	Summary    VariantListSummary `json:"summary,omitempty"`
}

// VariantSummary represents a variant summary for list views
type VariantSummary struct {
	ID             uuid.UUID              `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductID      uuid.UUID              `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	SKU            string                 `json:"sku" example:"WBH-001-RED"`
	VariantName    *string                `json:"variant_name,omitempty" example:"Red Wireless Headphones"`
	Options        map[string]interface{} `json:"options" example:"{\"color\":\"red\"}"`
	BasePrice      decimal.Decimal        `json:"base_price" example:"199.99"`
	SalePrice      *decimal.Decimal       `json:"sale_price,omitempty" example:"149.99"`
	EffectivePrice decimal.Decimal        `json:"effective_price" example:"149.99"`
	StockQuantity  int                    `json:"stock_quantity" example:"25"`
	IsLowStock     bool                   `json:"is_low_stock" example:"false"`
	IsActive       bool                   `json:"is_active" example:"true"`
	IsDefault      bool                   `json:"is_default" example:"false"`
	PrimaryImage   *string                `json:"primary_image,omitempty" example:"https://example.com/image.jpg"`
	CreatedAt      time.Time              `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt      time.Time              `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// VariantListSummary provides aggregate information about variants
type VariantListSummary struct {
	TotalVariants      int             `json:"total_variants" example:"15"`
	ActiveVariants     int             `json:"active_variants" example:"12"`
	LowStockVariants   int             `json:"low_stock_variants" example:"3"`
	OutOfStockVariants int             `json:"out_of_stock_variants" example:"1"`
	TotalValue         decimal.Decimal `json:"total_value" example:"3750.00"`
	AveragePrice       decimal.Decimal `json:"average_price" example:"199.99"`
}

// GenerateVariantsRequest represents a request to generate variants from options
type GenerateVariantsRequest struct {
	ProductID          uuid.UUID                  `json:"product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	OptionCombinations []VariantOptionCombination `json:"option_combinations" validate:"required,min=1"`
	SKUTemplate        string                     `json:"sku_template" validate:"required" example:"WBH-001-{color}-{size}"`
	BasePrice          *decimal.Decimal           `json:"base_price,omitempty" validate:"omitempty,min=0" example:"199.99"`
	StockQuantity      *int                       `json:"stock_quantity,omitempty" validate:"omitempty,min=0" example:"10"`
	PriceAdjustments   map[string]decimal.Decimal `json:"price_adjustments,omitempty" example:"{\"large\":\"10.00\"}"`
}

// VariantOptionCombination represents a combination of variant options
type VariantOptionCombination struct {
	Options map[string]string `json:"options" validate:"required" example:"{\"color\":\"red\",\"size\":\"medium\"}"`
	SKU     *string           `json:"sku,omitempty" validate:"omitempty,min=3,max=100" example:"WBH-001-RED-M"`
	Price   *decimal.Decimal  `json:"price,omitempty" validate:"omitempty,min=0" example:"199.99"`
	Stock   *int              `json:"stock,omitempty" validate:"omitempty,min=0" example:"10"`
}

// GenerateVariantsResponse represents the response from variant generation
type GenerateVariantsResponse struct {
	GeneratedVariants []VariantSummary           `json:"generated_variants"`
	SkippedCount      int                        `json:"skipped_count" example:"2"`
	FailedCount       int                        `json:"failed_count" example:"1"`
	Failures          []VariantGenerationFailure `json:"failures,omitempty"`
	Metadata          VariantGenerationMetadata  `json:"metadata"`
}

// VariantGenerationFailure represents a failed variant generation
type VariantGenerationFailure struct {
	Options   map[string]string `json:"options" example:"{\"color\":\"red\",\"size\":\"large\"}"`
	Error     string            `json:"error" example:"SKU already exists"`
	ErrorCode string            `json:"error_code" example:"VARIANT_ALREADY_EXISTS"`
}

// VariantGenerationMetadata contains metadata about variant generation
type VariantGenerationMetadata struct {
	TotalCombinations int    `json:"total_combinations" example:"12"`
	ProcessingTime    string `json:"processing_time" example:"500ms"`
	GeneratedAt       string `json:"generated_at" example:"2023-01-01T00:00:00Z"`
}

// ========================================
// Product Image DTOs
// ========================================

// CreateImageRequest represents the request to create a product image
type CreateImageRequest struct {
	ProductID          uuid.UUID  `json:"product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	VariantID          *uuid.UUID `json:"variant_id,omitempty" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440001"`
	ImageURL           string     `json:"image_url" validate:"required,url" example:"https://example.com/image.jpg"`
	CloudinaryURL      *string    `json:"cloudinary_url,omitempty" validate:"omitempty,url"`
	CloudinaryPublicID *string    `json:"cloudinary_public_id,omitempty" validate:"omitempty,max=255"`
	AltText            *string    `json:"alt_text,omitempty" validate:"omitempty,max=255" example:"Product front view"`
	Width              *int       `json:"width,omitempty" validate:"omitempty,min=1" example:"800"`
	Height             *int       `json:"height,omitempty" validate:"omitempty,min=1" example:"600"`
	FileSize           *int64     `json:"file_size,omitempty" validate:"omitempty,min=1" example:"1048576"`
	MimeType           *string    `json:"mime_type,omitempty" validate:"omitempty,max=50" example:"image/jpeg"`
	IsPrimary          *bool      `json:"is_primary,omitempty" example:"true"`
	SortOrder          *int       `json:"sort_order,omitempty" validate:"omitempty,min=0" example:"1"`
}

// UpdateImageRequest represents the request to update an image
type UpdateImageRequest struct {
	ImageURL           *string `json:"image_url,omitempty" validate:"omitempty,url"`
	CloudinaryURL      *string `json:"cloudinary_url,omitempty" validate:"omitempty,url"`
	CloudinaryPublicID *string `json:"cloudinary_public_id,omitempty" validate:"omitempty,max=255"`
	AltText            *string `json:"alt_text,omitempty" validate:"omitempty,max=255"`
	Width              *int    `json:"width,omitempty" validate:"omitempty,min=1"`
	Height             *int    `json:"height,omitempty" validate:"omitempty,min=1"`
	FileSize           *int64  `json:"file_size,omitempty" validate:"omitempty,min=1"`
	MimeType           *string `json:"mime_type,omitempty" validate:"omitempty,max=50"`
	IsPrimary          *bool   `json:"is_primary,omitempty"`
	SortOrder          *int    `json:"sort_order,omitempty" validate:"omitempty,min=0"`
}

// ImageResponse represents the response for a product image
type ImageResponse struct {
	ID                 uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductID          uuid.UUID  `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	VariantID          *uuid.UUID `json:"variant_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	ImageURL           string     `json:"image_url" example:"https://example.com/image.jpg"`
	CloudinaryURL      *string    `json:"cloudinary_url,omitempty" example:"https://res.cloudinary.com/demo/image.jpg"`
	CloudinaryPublicID *string    `json:"cloudinary_public_id,omitempty" example:"sample_id"`
	AltText            *string    `json:"alt_text,omitempty" example:"Product front view"`
	Width              *int       `json:"width,omitempty" example:"800"`
	Height             *int       `json:"height,omitempty" example:"600"`
	FileSize           *int64     `json:"file_size,omitempty" example:"1048576"`
	MimeType           *string    `json:"mime_type,omitempty" example:"image/jpeg"`
	IsPrimary          bool       `json:"is_primary" example:"true"`
	SortOrder          int        `json:"sort_order" example:"1"`
	CreatedAt          time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt          time.Time  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// BulkImageUploadRequest represents a request to upload multiple images
type BulkImageUploadRequest struct {
	ProductID uuid.UUID            `json:"product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	VariantID *uuid.UUID           `json:"variant_id,omitempty" validate:"omitempty,uuid4"`
	Images    []CreateImageRequest `json:"images" validate:"required,min=1,max=10"`
}

// BulkImageUploadResponse represents the response from bulk image upload
type BulkImageUploadResponse struct {
	UploadedImages []ImageResponse      `json:"uploaded_images"`
	FailedCount    int                  `json:"failed_count" example:"1"`
	Failures       []ImageUploadFailure `json:"failures,omitempty"`
	Metadata       ImageUploadMetadata  `json:"metadata"`
}

// ImageUploadFailure represents a failed image upload
type ImageUploadFailure struct {
	ImageURL  string `json:"image_url" example:"https://example.com/invalid.jpg"`
	Error     string `json:"error" example:"Invalid image format"`
	ErrorCode string `json:"error_code" example:"IMAGE_INVALID_FORMAT"`
}

// ImageUploadMetadata contains metadata about image uploads
type ImageUploadMetadata struct {
	TotalImages    int    `json:"total_images" example:"5"`
	ProcessingTime string `json:"processing_time" example:"2.5s"`
	UploadedAt     string `json:"uploaded_at" example:"2023-01-01T00:00:00Z"`
}

// ReorderImagesRequest represents a request to reorder images
type ReorderImagesRequest struct {
	ProductID   uuid.UUID    `json:"product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	VariantID   *uuid.UUID   `json:"variant_id,omitempty" validate:"omitempty,uuid4"`
	ImageOrders []ImageOrder `json:"image_orders" validate:"required,min=1"`
}

// ImageOrder represents the new order for an image
type ImageOrder struct {
	ImageID   uuid.UUID `json:"image_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440001"`
	SortOrder int       `json:"sort_order" validate:"min=0" example:"1"`
}

// SetPrimaryImageRequest represents a request to set primary image
type SetPrimaryImageRequest struct {
	ProductID uuid.UUID  `json:"product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	VariantID *uuid.UUID `json:"variant_id,omitempty" validate:"omitempty,uuid4"`
	ImageID   uuid.UUID  `json:"image_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440001"`
}
