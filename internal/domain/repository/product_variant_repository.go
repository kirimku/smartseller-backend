package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// ProductVariantFilter defines filtering options for variant queries
type ProductVariantFilter struct {
	// Basic filters
	IDs         []uuid.UUID `json:"ids,omitempty"`
	ProductIDs  []uuid.UUID `json:"product_ids,omitempty"`
	VariantSKUs []string    `json:"variant_skus,omitempty"`

	// Option filters
	Options      map[string]interface{} `json:"options,omitempty"`       // Specific option combinations
	OptionNames  []string               `json:"option_names,omitempty"`  // Filter by option names
	OptionValues []string               `json:"option_values,omitempty"` // Filter by option values

	// Status filters
	IsActive  *bool `json:"is_active,omitempty"`
	IsDefault *bool `json:"is_default,omitempty"`

	// Pricing filters
	MinPrice *decimal.Decimal `json:"min_price,omitempty"`
	MaxPrice *decimal.Decimal `json:"max_price,omitempty"`

	// Inventory filters
	MinStock      *int  `json:"min_stock,omitempty"`
	MaxStock      *int  `json:"max_stock,omitempty"`
	IsLowStock    *bool `json:"is_low_stock,omitempty"`
	TrackQuantity *bool `json:"track_quantity,omitempty"`
	IsAvailable   *bool `json:"is_available,omitempty"`

	// Physical properties
	HasDimensions *bool `json:"has_dimensions,omitempty"`
	HasWeight     *bool `json:"has_weight,omitempty"`

	// Date filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`

	// Text search
	SearchQuery string `json:"search_query,omitempty"` // Search in variant name, SKU

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // variant_name, price, stock_quantity, created_at, position
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ProductVariantInclude defines what related data to include
type ProductVariantInclude struct {
	Product    bool `json:"product,omitempty"`
	Images     bool `json:"images,omitempty"`
	Statistics bool `json:"statistics,omitempty"`
}

// ProductVariantRepository defines the interface for variant data access
type ProductVariantRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, variant *entity.ProductVariant) error
	GetByID(ctx context.Context, id uuid.UUID, include *ProductVariantInclude) (*entity.ProductVariant, error)
	GetBySKU(ctx context.Context, sku string, include *ProductVariantInclude) (*entity.ProductVariant, error)
	Update(ctx context.Context, variant *entity.ProductVariant) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations
	CreateBatch(ctx context.Context, variants []*entity.ProductVariant) error
	GetByIDs(ctx context.Context, ids []uuid.UUID, include *ProductVariantInclude) ([]*entity.ProductVariant, error)
	UpdateBatch(ctx context.Context, variants []*entity.ProductVariant) error
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error

	// Query operations
	List(ctx context.Context, filter *ProductVariantFilter, include *ProductVariantInclude) ([]*entity.ProductVariant, error)
	Count(ctx context.Context, filter *ProductVariantFilter) (int64, error)

	// Product-specific operations
	GetByProduct(ctx context.Context, productID uuid.UUID, include *ProductVariantInclude) ([]*entity.ProductVariant, error)
	GetDefaultVariant(ctx context.Context, productID uuid.UUID, include *ProductVariantInclude) (*entity.ProductVariant, error)
	SetDefaultVariant(ctx context.Context, productID uuid.UUID, variantID uuid.UUID) error
	CountByProduct(ctx context.Context, productID uuid.UUID) (int, error)

	// Option-based operations
	GetByOptions(ctx context.Context, productID uuid.UUID, options map[string]interface{}, include *ProductVariantInclude) (*entity.ProductVariant, error)
	GetByOptionValue(ctx context.Context, productID uuid.UUID, optionName string, optionValue string, include *ProductVariantInclude) ([]*entity.ProductVariant, error)
	GetVariantsWithOption(ctx context.Context, optionName string, optionValue string, include *ProductVariantInclude) ([]*entity.ProductVariant, error)

	// Stock management
	UpdateStock(ctx context.Context, variantID uuid.UUID, quantity int) error
	DeductStock(ctx context.Context, variantID uuid.UUID, quantity int) error
	RestockInventory(ctx context.Context, variantID uuid.UUID, quantity int) error
	GetLowStockVariants(ctx context.Context, threshold int, include *ProductVariantInclude) ([]*entity.ProductVariant, error)
	BulkUpdateStock(ctx context.Context, updates []VariantStockUpdate) error

	// Pricing operations
	UpdatePrice(ctx context.Context, variantID uuid.UUID, price decimal.Decimal) error
	BulkUpdatePrices(ctx context.Context, updates []VariantPriceUpdate) error
	GetVariantsInPriceRange(ctx context.Context, minPrice, maxPrice decimal.Decimal, include *ProductVariantInclude) ([]*entity.ProductVariant, error)

	// Status management
	Activate(ctx context.Context, variantID uuid.UUID) error
	Deactivate(ctx context.Context, variantID uuid.UUID) error
	BulkActivate(ctx context.Context, variantIDs []uuid.UUID) error
	BulkDeactivate(ctx context.Context, variantIDs []uuid.UUID) error

	// Position management
	UpdatePosition(ctx context.Context, variantID uuid.UUID, position int) error
	ReorderVariants(ctx context.Context, productID uuid.UUID, variantOrders []VariantOrder) error
	GetNextPosition(ctx context.Context, productID uuid.UUID) (int, error)

	// Search operations
	Search(ctx context.Context, query string, filter *ProductVariantFilter, include *ProductVariantInclude) ([]*entity.ProductVariant, error)
	SearchByOptions(ctx context.Context, optionQueries map[string]string, filter *ProductVariantFilter, include *ProductVariantInclude) ([]*entity.ProductVariant, error)

	// SKU management
	IsSkuExists(ctx context.Context, sku string) (bool, error)
	IsSkuExistsExcluding(ctx context.Context, sku string, variantID uuid.UUID) (bool, error)
	GenerateVariantSKU(ctx context.Context, productID uuid.UUID, options map[string]interface{}) (string, error)

	// Validation operations
	ValidateVariantOptions(ctx context.Context, productID uuid.UUID, options map[string]interface{}) error
	CheckDuplicateVariant(ctx context.Context, productID uuid.UUID, options map[string]interface{}, excludeID *uuid.UUID) error
	ValidateVariantData(ctx context.Context, variant *entity.ProductVariant) error

	// Analytics operations
	GetVariantStatistics(ctx context.Context, variantID uuid.UUID) (*VariantStatistics, error)
	GetTopSellingVariants(ctx context.Context, limit int, days int) ([]*entity.ProductVariant, error)
	GetVariantPerformance(ctx context.Context, variantID uuid.UUID, days int) (*VariantPerformance, error)
	GetOptionPopularity(ctx context.Context, productID uuid.UUID) (map[string]map[string]int64, error)

	// Cleanup operations
	CleanupOrphanedVariants(ctx context.Context) (int64, error)
	RemoveInactiveVariants(ctx context.Context, daysInactive int) (int64, error)

	// Utility operations
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsBySKU(ctx context.Context, sku string) (bool, error)
	GetAvailableVariantCount(ctx context.Context, productID uuid.UUID) (int, error)
}

// VariantStockUpdate defines a stock update operation
type VariantStockUpdate struct {
	VariantID uuid.UUID `json:"variant_id"`
	Quantity  int       `json:"quantity"`
}

// VariantPriceUpdate defines a price update operation
type VariantPriceUpdate struct {
	VariantID uuid.UUID       `json:"variant_id"`
	Price     decimal.Decimal `json:"price"`
}

// VariantOrder defines the sort order for a variant
type VariantOrder struct {
	VariantID uuid.UUID `json:"variant_id"`
	Position  int       `json:"position"`
}

// VariantStatistics contains analytics data for a variant
type VariantStatistics struct {
	VariantID      uuid.UUID        `json:"variant_id"`
	TotalViews     int64            `json:"total_views"`
	TotalSales     int64            `json:"total_sales"`
	TotalRevenue   decimal.Decimal  `json:"total_revenue"`
	ConversionRate *decimal.Decimal `json:"conversion_rate,omitempty"`
	LastSoldAt     *time.Time       `json:"last_sold_at,omitempty"`
	ProfitMargin   *decimal.Decimal `json:"profit_margin,omitempty"`
	ProfitAmount   *decimal.Decimal `json:"profit_amount,omitempty"`
}

// VariantPerformance contains performance metrics for a variant
type VariantPerformance struct {
	VariantID      uuid.UUID        `json:"variant_id"`
	Views          int64            `json:"views"`
	Sales          int64            `json:"sales"`
	Revenue        decimal.Decimal  `json:"revenue"`
	ConversionRate *decimal.Decimal `json:"conversion_rate,omitempty"`
	AvgOrderValue  *decimal.Decimal `json:"avg_order_value,omitempty"`
	StockTurnover  *decimal.Decimal `json:"stock_turnover,omitempty"`
	ProfitMargin   *decimal.Decimal `json:"profit_margin,omitempty"`
	PopularityRank *int             `json:"popularity_rank,omitempty"`
}
