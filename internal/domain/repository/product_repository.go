package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// ProductFilter defines filtering options for product queries
type ProductFilter struct {
	// Basic filters
	IDs         []uuid.UUID `json:"ids,omitempty"`
	SKUs        []string    `json:"skus,omitempty"`
	Names       []string    `json:"names,omitempty"`
	CategoryIDs []uuid.UUID `json:"category_ids,omitempty"`

	// Status filters
	Status     []entity.ProductStatus `json:"status,omitempty"`
	IsActive   *bool                  `json:"is_active,omitempty"`
	IsFeatured *bool                  `json:"is_featured,omitempty"`

	// Pricing filters
	MinPrice *decimal.Decimal `json:"min_price,omitempty"`
	MaxPrice *decimal.Decimal `json:"max_price,omitempty"`

	// Inventory filters
	MinStock         *int  `json:"min_stock,omitempty"`
	MaxStock         *int  `json:"max_stock,omitempty"`
	IsLowStock       *bool `json:"is_low_stock,omitempty"`
	TrackQuantity    *bool `json:"track_quantity,omitempty"`
	RequiresShipping *bool `json:"requires_shipping,omitempty"`

	// Date filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`

	// Text search
	SearchQuery string `json:"search_query,omitempty"` // Search in name, description, SKU

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // name, price, created_at, updated_at, stock_quantity
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ProductInclude defines what related data to include
type ProductInclude struct {
	Category       bool `json:"category,omitempty"`
	Images         bool `json:"images,omitempty"`
	Variants       bool `json:"variants,omitempty"`
	VariantOptions bool `json:"variant_options,omitempty"`

	// For performance optimization
	OnlyActiveImages    bool `json:"only_active_images,omitempty"`
	OnlyActiveVariants  bool `json:"only_active_variants,omitempty"`
	IncludeImageDetails bool `json:"include_image_details,omitempty"`
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uuid.UUID, include *ProductInclude) (*entity.Product, error)
	GetBySKU(ctx context.Context, sku string, include *ProductInclude) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations
	CreateBatch(ctx context.Context, products []*entity.Product) error
	GetByIDs(ctx context.Context, ids []uuid.UUID, include *ProductInclude) ([]*entity.Product, error)
	UpdateBatch(ctx context.Context, products []*entity.Product) error
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error

	// Query operations
	List(ctx context.Context, filter *ProductFilter, include *ProductInclude) ([]*entity.Product, error)
	Count(ctx context.Context, filter *ProductFilter) (int64, error)

	// Search operations
	Search(ctx context.Context, query string, filter *ProductFilter, include *ProductInclude) ([]*entity.Product, error)
	SearchWithHighlight(ctx context.Context, query string, filter *ProductFilter) ([]*entity.Product, error)

	// Category-related operations
	GetByCategory(ctx context.Context, categoryID uuid.UUID, filter *ProductFilter, include *ProductInclude) ([]*entity.Product, error)
	GetByCategoryPath(ctx context.Context, categoryPath string, filter *ProductFilter, include *ProductInclude) ([]*entity.Product, error)

	// Stock management
	UpdateStock(ctx context.Context, productID uuid.UUID, quantity int) error
	DeductStock(ctx context.Context, productID uuid.UUID, quantity int) error
	RestockInventory(ctx context.Context, productID uuid.UUID, quantity int) error
	GetLowStockProducts(ctx context.Context, threshold int, include *ProductInclude) ([]*entity.Product, error)

	// Status management
	UpdateStatus(ctx context.Context, productID uuid.UUID, status entity.ProductStatus) error
	BulkUpdateStatus(ctx context.Context, productIDs []uuid.UUID, status entity.ProductStatus) error
	Activate(ctx context.Context, productID uuid.UUID) error
	Deactivate(ctx context.Context, productID uuid.UUID) error
	BulkActivate(ctx context.Context, productIDs []uuid.UUID) error
	BulkDeactivate(ctx context.Context, productIDs []uuid.UUID) error

	// Featured products
	SetFeatured(ctx context.Context, productID uuid.UUID) error
	UnsetFeatured(ctx context.Context, productID uuid.UUID) error
	GetFeaturedProducts(ctx context.Context, limit int, include *ProductInclude) ([]*entity.Product, error)

	// Pricing operations
	UpdatePrice(ctx context.Context, productID uuid.UUID, price decimal.Decimal) error
	BulkUpdatePrices(ctx context.Context, updates []struct {
		ProductID uuid.UUID
		Price     decimal.Decimal
	}) error
	GetProductsInPriceRange(ctx context.Context, minPrice, maxPrice decimal.Decimal, include *ProductInclude) ([]*entity.Product, error)

	// SKU management
	IsSkuExists(ctx context.Context, sku string) (bool, error)
	IsSkuExistsExcludingProduct(ctx context.Context, sku string, productID uuid.UUID) (bool, error)
	GenerateNextSKU(ctx context.Context, prefix string) (string, error)

	// Analytics and reporting
	GetProductStatistics(ctx context.Context, productID uuid.UUID) (*ProductStatistics, error)
	GetTopSellingProducts(ctx context.Context, limit int, days int) ([]*entity.Product, error)
	GetRecentlyUpdatedProducts(ctx context.Context, limit int, include *ProductInclude) ([]*entity.Product, error)
	GetProductsByPriceRange(ctx context.Context, ranges []PriceRange) (map[string]int64, error)
	GetProductCountByCategory(ctx context.Context) (map[uuid.UUID]int64, error)
	GetProductCountByStatus(ctx context.Context) (map[entity.ProductStatus]int64, error)

	// Validation operations
	ValidateProductData(ctx context.Context, product *entity.Product) error
	CheckDuplicateFields(ctx context.Context, fields map[string]interface{}, excludeID *uuid.UUID) error

	// Utility operations
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsBySKU(ctx context.Context, sku string) (bool, error)
}

// ProductStatistics contains analytics data for a product
type ProductStatistics struct {
	ProductID      uuid.UUID        `json:"product_id"`
	TotalViews     int64            `json:"total_views"`
	TotalSales     int64            `json:"total_sales"`
	TotalRevenue   decimal.Decimal  `json:"total_revenue"`
	AverageRating  *decimal.Decimal `json:"average_rating,omitempty"`
	ReviewCount    int64            `json:"review_count"`
	VariantCount   int              `json:"variant_count"`
	ImageCount     int              `json:"image_count"`
	LastSoldAt     *time.Time       `json:"last_sold_at,omitempty"`
	LastViewedAt   *time.Time       `json:"last_viewed_at,omitempty"`
	ConversionRate *decimal.Decimal `json:"conversion_rate,omitempty"`
}

// PriceRange defines a price range for analytics
type PriceRange struct {
	Label    string          `json:"label"`
	MinPrice decimal.Decimal `json:"min_price"`
	MaxPrice decimal.Decimal `json:"max_price"`
}
