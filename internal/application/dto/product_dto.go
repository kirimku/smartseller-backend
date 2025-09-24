package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// CreateProductRequest represents the request to create a new product
type CreateProductRequest struct {
	Name              string               `json:"name" validate:"required,min=1,max=255" example:"Wireless Bluetooth Headphones"`
	Description       *string              `json:"description,omitempty" validate:"omitempty,max=5000" example:"High-quality wireless headphones with noise cancellation"`
	SKU               string               `json:"sku" validate:"required,min=3,max=100,alphanum_underscore_hyphen" example:"WBH-001"`
	CategoryID        *uuid.UUID           `json:"category_id,omitempty" validate:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Brand             *string              `json:"brand,omitempty" validate:"omitempty,max=255" example:"TechSound"`
	Tags              []string             `json:"tags,omitempty" validate:"omitempty,dive,max=50" example:"wireless,bluetooth,headphones"`
	BasePrice         decimal.Decimal      `json:"base_price" validate:"required,min=0" example:"199.99"`
	SalePrice         *decimal.Decimal     `json:"sale_price,omitempty" validate:"omitempty,min=0" example:"149.99"`
	CostPrice         *decimal.Decimal     `json:"cost_price,omitempty" validate:"omitempty,min=0" example:"80.00"`
	TrackInventory    bool                 `json:"track_inventory" example:"true"`
	StockQuantity     int                  `json:"stock_quantity" validate:"min=0" example:"100"`
	LowStockThreshold *int                 `json:"low_stock_threshold,omitempty" validate:"omitempty,min=0" example:"10"`
	Status            entity.ProductStatus `json:"status" validate:"omitempty,oneof=draft active inactive archived" example:"draft"`
	MetaTitle         *string              `json:"meta_title,omitempty" validate:"omitempty,max=255" example:"Best Wireless Headphones - TechSound"`
	MetaDescription   *string              `json:"meta_description,omitempty" validate:"omitempty,max=500" example:"Discover our premium wireless headphones with superior sound quality"`
	Slug              *string              `json:"slug,omitempty" validate:"omitempty,max=255,slug" example:"wireless-bluetooth-headphones"`
	Weight            *decimal.Decimal     `json:"weight,omitempty" validate:"omitempty,min=0" example:"0.25"`
	DimensionsLength  *decimal.Decimal     `json:"dimensions_length,omitempty" validate:"omitempty,min=0" example:"20.5"`
	DimensionsWidth   *decimal.Decimal     `json:"dimensions_width,omitempty" validate:"omitempty,min=0" example:"15.2"`
	DimensionsHeight  *decimal.Decimal     `json:"dimensions_height,omitempty" validate:"omitempty,min=0" example:"8.5"`
}

// UpdateProductRequest represents the request to update an existing product
type UpdateProductRequest struct {
	Name              *string               `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description       *string               `json:"description,omitempty" validate:"omitempty,max=5000"`
	CategoryID        *uuid.UUID            `json:"category_id,omitempty" validate:"omitempty,uuid4"`
	Brand             *string               `json:"brand,omitempty" validate:"omitempty,max=255"`
	Tags              []string              `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
	BasePrice         *decimal.Decimal      `json:"base_price,omitempty" validate:"omitempty,min=0"`
	SalePrice         *decimal.Decimal      `json:"sale_price,omitempty" validate:"omitempty,min=0"`
	CostPrice         *decimal.Decimal      `json:"cost_price,omitempty" validate:"omitempty,min=0"`
	TrackInventory    *bool                 `json:"track_inventory,omitempty"`
	StockQuantity     *int                  `json:"stock_quantity,omitempty" validate:"omitempty,min=0"`
	LowStockThreshold *int                  `json:"low_stock_threshold,omitempty" validate:"omitempty,min=0"`
	Status            *entity.ProductStatus `json:"status,omitempty" validate:"omitempty,oneof=draft active inactive archived"`
	MetaTitle         *string               `json:"meta_title,omitempty" validate:"omitempty,max=255"`
	MetaDescription   *string               `json:"meta_description,omitempty" validate:"omitempty,max=500"`
	Slug              *string               `json:"slug,omitempty" validate:"omitempty,max=255,slug"`
	Weight            *decimal.Decimal      `json:"weight,omitempty" validate:"omitempty,min=0"`
	DimensionsLength  *decimal.Decimal      `json:"dimensions_length,omitempty" validate:"omitempty,min=0"`
	DimensionsWidth   *decimal.Decimal      `json:"dimensions_width,omitempty" validate:"omitempty,min=0"`
	DimensionsHeight  *decimal.Decimal      `json:"dimensions_height,omitempty" validate:"omitempty,min=0"`
}

// ProductResponse represents the response for a single product
type ProductResponse struct {
	ID                uuid.UUID        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SKU               string           `json:"sku" example:"WBH-001"`
	Name              string           `json:"name" example:"Wireless Bluetooth Headphones"`
	Description       *string          `json:"description,omitempty" example:"High-quality wireless headphones"`
	CategoryID        *uuid.UUID       `json:"category_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	Category          *CategorySummary `json:"category,omitempty"`
	Brand             *string          `json:"brand,omitempty" example:"TechSound"`
	Tags              []string         `json:"tags,omitempty" example:"wireless,bluetooth"`
	BasePrice         decimal.Decimal  `json:"base_price" example:"199.99"`
	SalePrice         *decimal.Decimal `json:"sale_price,omitempty" example:"149.99"`
	CostPrice         *decimal.Decimal `json:"cost_price,omitempty" example:"80.00"`
	EffectivePrice    decimal.Decimal  `json:"effective_price" example:"149.99"`
	ProfitMargin      *decimal.Decimal `json:"profit_margin,omitempty" example:"46.67"`
	TrackInventory    bool             `json:"track_inventory" example:"true"`
	StockQuantity     int              `json:"stock_quantity" example:"100"`
	LowStockThreshold *int             `json:"low_stock_threshold,omitempty" example:"10"`
	IsLowStock        bool             `json:"is_low_stock" example:"false"`
	Status            string           `json:"status" example:"active"`
	MetaTitle         *string          `json:"meta_title,omitempty"`
	MetaDescription   *string          `json:"meta_description,omitempty"`
	Slug              *string          `json:"slug,omitempty" example:"wireless-bluetooth-headphones"`
	Weight            *decimal.Decimal `json:"weight,omitempty" example:"0.25"`
	DimensionsLength  *decimal.Decimal `json:"dimensions_length,omitempty" example:"20.5"`
	DimensionsWidth   *decimal.Decimal `json:"dimensions_width,omitempty" example:"15.2"`
	DimensionsHeight  *decimal.Decimal `json:"dimensions_height,omitempty" example:"8.5"`
	CreatedBy         uuid.UUID        `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440002"`
	CreatedAt         time.Time        `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt         time.Time        `json:"updated_at" example:"2023-01-01T00:00:00Z"`

	// Related entities (optional, loaded based on include parameters)
	Images   []ProductImageSummary   `json:"images,omitempty"`
	Variants []ProductVariantSummary `json:"variants,omitempty"`
}

// ProductSummary represents a product summary for list views
type ProductSummary struct {
	ID             uuid.UUID        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SKU            string           `json:"sku" example:"WBH-001"`
	Name           string           `json:"name" example:"Wireless Bluetooth Headphones"`
	Brand          *string          `json:"brand,omitempty" example:"TechSound"`
	BasePrice      decimal.Decimal  `json:"base_price" example:"199.99"`
	SalePrice      *decimal.Decimal `json:"sale_price,omitempty" example:"149.99"`
	EffectivePrice decimal.Decimal  `json:"effective_price" example:"149.99"`
	StockQuantity  int              `json:"stock_quantity" example:"100"`
	IsLowStock     bool             `json:"is_low_stock" example:"false"`
	Status         string           `json:"status" example:"active"`
	PrimaryImage   *string          `json:"primary_image,omitempty" example:"https://example.com/image.jpg"`
	VariantCount   int              `json:"variant_count" example:"3"`
	CreatedAt      time.Time        `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt      time.Time        `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// ProductListResponse represents a paginated list of products
type ProductListResponse struct {
	Products   []ProductSummary      `json:"products"`
	Pagination PaginationResponse    `json:"pagination"`
	Filters    ProductFiltersApplied `json:"filters_applied,omitempty"`
	Summary    ProductListSummary    `json:"summary,omitempty"`
}

// ProductListSummary provides aggregate information about the product list
type ProductListSummary struct {
	TotalProducts      int             `json:"total_products" example:"1250"`
	ActiveProducts     int             `json:"active_products" example:"980"`
	LowStockProducts   int             `json:"low_stock_products" example:"45"`
	OutOfStockProducts int             `json:"out_of_stock_products" example:"12"`
	TotalValue         decimal.Decimal `json:"total_value" example:"125000.00"`
	AveragePrice       decimal.Decimal `json:"average_price" example:"199.99"`
}

// ProductFilters represents filters for product listing
type ProductFilters struct {
	CategoryIDs   []uuid.UUID            `json:"category_ids,omitempty" validate:"omitempty,dive,uuid4"`
	Status        []entity.ProductStatus `json:"status,omitempty" validate:"omitempty,dive,oneof=draft active inactive archived"`
	Brand         []string               `json:"brand,omitempty" validate:"omitempty,dive,max=255"`
	Tags          []string               `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
	MinPrice      *decimal.Decimal       `json:"min_price,omitempty" validate:"omitempty,min=0"`
	MaxPrice      *decimal.Decimal       `json:"max_price,omitempty" validate:"omitempty,min=0"`
	MinStock      *int                   `json:"min_stock,omitempty" validate:"omitempty,min=0"`
	MaxStock      *int                   `json:"max_stock,omitempty" validate:"omitempty,min=0"`
	IsLowStock    *bool                  `json:"is_low_stock,omitempty"`
	IsOutOfStock  *bool                  `json:"is_out_of_stock,omitempty"`
	HasVariants   *bool                  `json:"has_variants,omitempty"`
	HasImages     *bool                  `json:"has_images,omitempty"`
	CreatedAfter  *time.Time             `json:"created_after,omitempty"`
	CreatedBefore *time.Time             `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time             `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time             `json:"updated_before,omitempty"`
	Search        string                 `json:"search,omitempty" validate:"omitempty,max=255"`
	SearchFields  []string               `json:"search_fields,omitempty" validate:"omitempty,dive,oneof=name description sku brand tags"`
}

// ProductFiltersApplied shows which filters were actually applied
type ProductFiltersApplied struct {
	CategoryIDs  []uuid.UUID            `json:"category_ids,omitempty"`
	Status       []entity.ProductStatus `json:"status,omitempty"`
	Brand        []string               `json:"brand,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	PriceRange   *PriceRange            `json:"price_range,omitempty"`
	StockRange   *StockRange            `json:"stock_range,omitempty"`
	IsLowStock   *bool                  `json:"is_low_stock,omitempty"`
	IsOutOfStock *bool                  `json:"is_out_of_stock,omitempty"`
	HasVariants  *bool                  `json:"has_variants,omitempty"`
	HasImages    *bool                  `json:"has_images,omitempty"`
	DateRange    *DateRange             `json:"date_range,omitempty"`
	Search       string                 `json:"search,omitempty"`
	SearchFields []string               `json:"search_fields,omitempty"`
}

// PriceRange represents a price range filter
type PriceRange struct {
	Min *decimal.Decimal `json:"min,omitempty" example:"10.00"`
	Max *decimal.Decimal `json:"max,omitempty" example:"500.00"`
}

// StockRange represents a stock range filter
type StockRange struct {
	Min *int `json:"min,omitempty" example:"0"`
	Max *int `json:"max,omitempty" example:"1000"`
}

// DateRange represents a date range filter
type DateRange struct {
	After  *time.Time `json:"after,omitempty" example:"2023-01-01T00:00:00Z"`
	Before *time.Time `json:"before,omitempty" example:"2023-12-31T23:59:59Z"`
}

// ProductIncludeOptions specifies which related data to include in the response
type ProductIncludeOptions struct {
	Category bool `json:"category,omitempty" example:"true"`
	Images   bool `json:"images,omitempty" example:"true"`
	Variants bool `json:"variants,omitempty" example:"true"`
	Stats    bool `json:"stats,omitempty" example:"false"`
}

// ProductBulkOperationRequest represents a bulk operation request
type ProductBulkOperationRequest struct {
	ProductIDs      []uuid.UUID            `json:"product_ids" validate:"required,min=1,max=100,dive,uuid4"`
	Operation       string                 `json:"operation" validate:"required,oneof=update_status update_price update_category delete archive activate deactivate"`
	UpdateData      map[string]interface{} `json:"update_data,omitempty"`
	ContinueOnError bool                   `json:"continue_on_error,omitempty" example:"true"`
	BatchSize       int                    `json:"batch_size,omitempty" validate:"omitempty,min=1,max=100" example:"10"`
}

// ProductBulkOperationResponse represents the response from bulk operations
type ProductBulkOperationResponse struct {
	SuccessCount int                       `json:"success_count" example:"8"`
	FailureCount int                       `json:"failure_count" example:"2"`
	Failures     []ProductOperationFailure `json:"failures,omitempty"`
	Metadata     BulkOperationMetadata     `json:"metadata"`
}

// ProductOperationFailure represents a failed operation in bulk processing
type ProductOperationFailure struct {
	ProductID uuid.UUID `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SKU       string    `json:"sku" example:"WBH-001"`
	Error     string    `json:"error" example:"Product not found"`
	ErrorCode string    `json:"error_code" example:"PRODUCT_NOT_FOUND"`
}

// BulkOperationMetadata contains metadata about bulk operations
type BulkOperationMetadata struct {
	TotalDuration    string `json:"total_duration" example:"2.5s"`
	BatchesProcessed int    `json:"batches_processed" example:"1"`
	OperationType    string `json:"operation_type" example:"update_status"`
	ProcessedAt      string `json:"processed_at" example:"2023-01-01T00:00:00Z"`
}

// ProductCloneRequest represents a request to clone a product
type ProductCloneRequest struct {
	SourceProductID     uuid.UUID             `json:"source_product_id" validate:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	NewSKU              string                `json:"new_sku" validate:"required,min=3,max=100,alphanum_underscore_hyphen" example:"WBH-002"`
	NewName             *string               `json:"new_name,omitempty" validate:"omitempty,min=1,max=255" example:"Wireless Bluetooth Headphones v2"`
	CloneVariants       bool                  `json:"clone_variants,omitempty" example:"true"`
	CloneImages         bool                  `json:"clone_images,omitempty" example:"true"`
	CloneVariantOptions bool                  `json:"clone_variant_options,omitempty" example:"true"`
	OverridePrice       *float64              `json:"override_price,omitempty" validate:"omitempty,min=0" example:"249.99"`
	OverrideCategory    *uuid.UUID            `json:"override_category,omitempty" validate:"omitempty,uuid4"`
	OverrideStock       *int                  `json:"override_stock,omitempty" validate:"omitempty,min=0" example:"50"`
	OverrideStatus      *entity.ProductStatus `json:"override_status,omitempty" validate:"omitempty,oneof=draft active inactive archived" example:"draft"`
}

// ProductCloneResponse represents the response from cloning a product
type ProductCloneResponse struct {
	Product        ProductResponse               `json:"product"`
	Category       *CategorySummary              `json:"category,omitempty"`
	VariantOptions []ProductVariantOptionSummary `json:"variant_options,omitempty"`
	Variants       []ProductVariantSummary       `json:"variants,omitempty"`
	Images         []ProductImageSummary         `json:"images,omitempty"`
	Metadata       CloneMetadata                 `json:"metadata"`
}

// CloneMetadata contains metadata about the cloning process
type CloneMetadata struct {
	TotalDuration     string   `json:"total_duration" example:"1.2s"`
	StepsCompleted    []string `json:"steps_completed" example:"product_cloned,images_cloned"`
	Warnings          []string `json:"warnings,omitempty"`
	CreatedEntities   int      `json:"created_entities" example:"5"`
	CategoryCreated   bool     `json:"category_created" example:"false"`
	VariantsGenerated bool     `json:"variants_generated" example:"true"`
}

// ProductStatsResponse represents product statistics
type ProductStatsResponse struct {
	Product         ProductStatsSummary `json:"product"`
	ImageCount      int                 `json:"image_count" example:"5"`
	VariantCount    int                 `json:"variant_count" example:"3"`
	HasPrimaryImage bool                `json:"has_primary_image" example:"true"`
	LastUpdated     time.Time           `json:"last_updated" example:"2023-01-01T00:00:00Z"`
}

// ProductStatsSummary represents summary statistics for a product
type ProductStatsSummary struct {
	ID        uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string          `json:"name" example:"Wireless Bluetooth Headphones"`
	SKU       string          `json:"sku" example:"WBH-001"`
	Status    string          `json:"status" example:"active"`
	BasePrice decimal.Decimal `json:"base_price" example:"199.99"`
}

// Note: PaginationRequest and PaginationResponse are defined in common.go

// Summary DTOs for related entities (referenced in ProductResponse)

// CategorySummary represents a category summary in product responses
type CategorySummary struct {
	ID       uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name     string     `json:"name" example:"Electronics"`
	Path     string     `json:"path" example:"Electronics/Audio/Headphones"`
	ParentID *uuid.UUID `json:"parent_id,omitempty"`
}

// ProductImageSummary represents an image summary in product responses
type ProductImageSummary struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440003"`
	ImageURL  string    `json:"image_url" example:"https://example.com/image.jpg"`
	AltText   *string   `json:"alt_text,omitempty" example:"Product front view"`
	IsPrimary bool      `json:"is_primary" example:"true"`
	SortOrder int       `json:"sort_order" example:"1"`
}

// ProductVariantSummary represents a variant summary in product responses
type ProductVariantSummary struct {
	ID            uuid.UUID        `json:"id" example:"550e8400-e29b-41d4-a716-446655440004"`
	SKU           string           `json:"sku" example:"WBH-001-BLK"`
	VariantName   *string          `json:"variant_name,omitempty" example:"Black"`
	BasePrice     decimal.Decimal  `json:"base_price" example:"199.99"`
	SalePrice     *decimal.Decimal `json:"sale_price,omitempty" example:"149.99"`
	StockQuantity int              `json:"stock_quantity" example:"25"`
	IsActive      bool             `json:"is_active" example:"true"`
}

// ProductVariantOptionSummary represents a variant option summary
type ProductVariantOptionSummary struct {
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440005"`
	OptionName string    `json:"option_name" example:"Color"`
	OptionType string    `json:"option_type" example:"color"`
	Values     []string  `json:"values" example:"Black,White,Gray"`
	IsRequired bool      `json:"is_required" example:"true"`
}
