package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// ProductVariantOptionFilter defines filtering options for variant option queries
type ProductVariantOptionFilter struct {
	// Basic filters
	IDs         []uuid.UUID `json:"ids,omitempty"`
	ProductIDs  []uuid.UUID `json:"product_ids,omitempty"`
	OptionNames []string    `json:"option_names,omitempty"`

	// Option properties
	IsRequired *bool `json:"is_required,omitempty"`
	HasValues  *bool `json:"has_values,omitempty"`
	MinValues  *int  `json:"min_values,omitempty"`
	MaxValues  *int  `json:"max_values,omitempty"`

	// Value filters
	OptionValues []string `json:"option_values,omitempty"` // Filter by specific values
	ValueCount   *int     `json:"value_count,omitempty"`   // Exact value count

	// Date filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`

	// Text search
	SearchQuery string `json:"search_query,omitempty"` // Search in option name, display name

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // option_name, sort_order, created_at, value_count
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ProductVariantOptionInclude defines what related data to include
type ProductVariantOptionInclude struct {
	Product      bool `json:"product,omitempty"`
	UsageStats   bool `json:"usage_stats,omitempty"`   // Include usage statistics
	VariantCount bool `json:"variant_count,omitempty"` // Count of variants using this option
}

// ProductVariantOptionRepository defines the interface for variant option data access
type ProductVariantOptionRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, option *entity.ProductVariantOption) error
	GetByID(ctx context.Context, id uuid.UUID, include *ProductVariantOptionInclude) (*entity.ProductVariantOption, error)
	Update(ctx context.Context, option *entity.ProductVariantOption) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations
	CreateBatch(ctx context.Context, options []*entity.ProductVariantOption) error
	GetByIDs(ctx context.Context, ids []uuid.UUID, include *ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error)
	UpdateBatch(ctx context.Context, options []*entity.ProductVariantOption) error
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error

	// Query operations
	List(ctx context.Context, filter *ProductVariantOptionFilter, include *ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error)
	Count(ctx context.Context, filter *ProductVariantOptionFilter) (int64, error)

	// Product-specific operations
	GetByProduct(ctx context.Context, productID uuid.UUID, include *ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error)
	GetByProductAndName(ctx context.Context, productID uuid.UUID, optionName string, include *ProductVariantOptionInclude) (*entity.ProductVariantOption, error)
	CountByProduct(ctx context.Context, productID uuid.UUID) (int, error)

	// Option name operations
	GetOptionNamesByProduct(ctx context.Context, productID uuid.UUID) ([]string, error)
	IsOptionNameExists(ctx context.Context, productID uuid.UUID, optionName string) (bool, error)
	IsOptionNameExistsExcluding(ctx context.Context, productID uuid.UUID, optionName string, excludeID uuid.UUID) (bool, error)

	// Value operations
	GetAllValuesForOption(ctx context.Context, optionID uuid.UUID) ([]string, error)
	GetUniqueValuesForProduct(ctx context.Context, productID uuid.UUID, optionName string) ([]string, error)
	IsValueExists(ctx context.Context, optionID uuid.UUID, value string) (bool, error)
	AddValueToOption(ctx context.Context, optionID uuid.UUID, value string) error
	RemoveValueFromOption(ctx context.Context, optionID uuid.UUID, value string) error
	UpdateOptionValue(ctx context.Context, optionID uuid.UUID, oldValue, newValue string) error

	// Bulk value operations
	AddValuesToOption(ctx context.Context, optionID uuid.UUID, values []string) error
	RemoveValuesFromOption(ctx context.Context, optionID uuid.UUID, values []string) error
	ReplaceOptionValues(ctx context.Context, optionID uuid.UUID, values []string) error

	// Sort order operations
	UpdateSortOrder(ctx context.Context, optionID uuid.UUID, sortOrder int) error
	ReorderOptions(ctx context.Context, productID uuid.UUID, optionOrders []OptionOrder) error
	GetNextSortOrder(ctx context.Context, productID uuid.UUID) (int, error)

	// Search operations
	Search(ctx context.Context, query string, filter *ProductVariantOptionFilter, include *ProductVariantOptionInclude) ([]*entity.ProductVariantOption, error)
	SearchByValue(ctx context.Context, value string, productID *uuid.UUID) ([]*entity.ProductVariantOption, error)

	// Validation operations
	ValidateOptionName(ctx context.Context, productID uuid.UUID, optionName string, excludeID *uuid.UUID) error
	ValidateOptionValues(ctx context.Context, values []string) error
	CheckOptionUsage(ctx context.Context, optionID uuid.UUID) (*OptionUsage, error)

	// Analytics operations
	GetOptionStatistics(ctx context.Context, optionID uuid.UUID) (*OptionStatistics, error)
	GetMostUsedOptions(ctx context.Context, limit int) ([]*entity.ProductVariantOption, error)
	GetValueUsageStatistics(ctx context.Context, optionID uuid.UUID) (map[string]int64, error)
	GetOptionUsageByProduct(ctx context.Context, productID uuid.UUID) ([]*OptionUsageInfo, error)

	// Import/Export operations
	ExportOptions(ctx context.Context, productID uuid.UUID) ([]*OptionExport, error)
	ImportOptions(ctx context.Context, productID uuid.UUID, options []*OptionImport) error

	// Cleanup operations
	CleanupUnusedOptions(ctx context.Context) (int64, error)
	RemoveEmptyOptions(ctx context.Context) (int64, error)

	// Utility operations
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetOptionCount(ctx context.Context, productID uuid.UUID) (int, error)
	GetTotalValueCount(ctx context.Context, productID uuid.UUID) (int, error)
}

// OptionOrder defines the sort order for an option
type OptionOrder struct {
	OptionID  uuid.UUID `json:"option_id"`
	SortOrder int       `json:"sort_order"`
}

// OptionUsage contains information about how an option is used
type OptionUsage struct {
	OptionID          uuid.UUID `json:"option_id"`
	VariantCount      int64     `json:"variant_count"`   // Number of variants using this option
	ActiveVariants    int64     `json:"active_variants"` // Number of active variants using this option
	TotalSales        int64     `json:"total_sales"`     // Total sales from variants using this option
	MostUsedValue     *string   `json:"most_used_value,omitempty"`
	LeastUsedValue    *string   `json:"least_used_value,omitempty"`
	CanBeDeleted      bool      `json:"can_be_deleted"`               // Whether it's safe to delete
	DependentEntities []string  `json:"dependent_entities,omitempty"` // Types of entities depending on this option
}

// OptionStatistics contains analytics data for an option
type OptionStatistics struct {
	OptionID            uuid.UUID           `json:"option_id"`
	ValueCount          int                 `json:"value_count"`
	VariantCount        int64               `json:"variant_count"`
	ActiveVariantCount  int64               `json:"active_variant_count"`
	TotalViews          int64               `json:"total_views"`
	TotalSales          int64               `json:"total_sales"`
	ValueDistribution   map[string]int64    `json:"value_distribution"` // Value -> count mapping
	PopularCombinations []OptionCombination `json:"popular_combinations,omitempty"`
}

// OptionCombination represents a popular combination of option values
type OptionCombination struct {
	Values map[string]string `json:"values"` // option_name -> value mapping
	Count  int64             `json:"count"`  // Number of variants with this combination
	Sales  int64             `json:"sales"`  // Number of sales with this combination
}

// OptionUsageInfo contains usage information for an option within a product
type OptionUsageInfo struct {
	OptionID     uuid.UUID `json:"option_id"`
	OptionName   string    `json:"option_name"`
	ValueCount   int       `json:"value_count"`
	VariantCount int64     `json:"variant_count"`
	IsRequired   bool      `json:"is_required"`
	SortOrder    int       `json:"sort_order"`
}

// OptionExport defines the structure for exporting options
type OptionExport struct {
	ID           uuid.UUID `json:"id"`
	OptionName   string    `json:"option_name"`
	DisplayName  *string   `json:"display_name,omitempty"`
	OptionValues []string  `json:"option_values"`
	SortOrder    int       `json:"sort_order"`
	IsRequired   bool      `json:"is_required"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// OptionImport defines the structure for importing options
type OptionImport struct {
	OptionName   string   `json:"option_name"`
	DisplayName  *string  `json:"display_name,omitempty"`
	OptionValues []string `json:"option_values"`
	SortOrder    *int     `json:"sort_order,omitempty"`
	IsRequired   *bool    `json:"is_required,omitempty"`
}
