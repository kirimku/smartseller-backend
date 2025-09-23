package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// ProductCategoryFilter defines filtering options for category queries
type ProductCategoryFilter struct {
	// Basic filters
	IDs   []uuid.UUID `json:"ids,omitempty"`
	Names []string    `json:"names,omitempty"`
	Slugs []string    `json:"slugs,omitempty"`

	// Hierarchy filters
	ParentIDs   []uuid.UUID `json:"parent_ids,omitempty"`
	Level       *int        `json:"level,omitempty"`
	IsRoot      *bool       `json:"is_root,omitempty"`      // Categories with no parent
	HasChildren *bool       `json:"has_children,omitempty"` // Categories with subcategories

	// Status filters
	IsActive *bool `json:"is_active,omitempty"`

	// Date filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	UpdatedAfter  *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore *time.Time `json:"updated_before,omitempty"`

	// Text search
	SearchQuery string `json:"search_query,omitempty"` // Search in name, description

	// Sorting
	SortBy    string `json:"sort_by,omitempty"`    // name, level, created_at, updated_at, sort_order
	SortOrder string `json:"sort_order,omitempty"` // asc, desc

	// Pagination
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// ProductCategoryInclude defines what related data to include
type ProductCategoryInclude struct {
	Parent       bool `json:"parent,omitempty"`
	Children     bool `json:"children,omitempty"`
	AllChildren  bool `json:"all_children,omitempty"` // Include all descendants
	ProductCount bool `json:"product_count,omitempty"`

	// Recursive options
	MaxDepth *int `json:"max_depth,omitempty"` // Maximum depth for recursive includes
}

// ProductCategoryRepository defines the interface for category data access
type ProductCategoryRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, category *entity.ProductCategory) error
	GetByID(ctx context.Context, id uuid.UUID, include *ProductCategoryInclude) (*entity.ProductCategory, error)
	GetBySlug(ctx context.Context, slug string, include *ProductCategoryInclude) (*entity.ProductCategory, error)
	Update(ctx context.Context, category *entity.ProductCategory) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Batch operations
	CreateBatch(ctx context.Context, categories []*entity.ProductCategory) error
	GetByIDs(ctx context.Context, ids []uuid.UUID, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)
	UpdateBatch(ctx context.Context, categories []*entity.ProductCategory) error
	DeleteBatch(ctx context.Context, ids []uuid.UUID) error

	// Query operations
	List(ctx context.Context, filter *ProductCategoryFilter, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)
	Count(ctx context.Context, filter *ProductCategoryFilter) (int64, error)

	// Hierarchy operations
	GetRootCategories(ctx context.Context, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)
	GetChildren(ctx context.Context, parentID uuid.UUID, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)
	GetAllChildren(ctx context.Context, parentID uuid.UUID, maxDepth *int) ([]*entity.ProductCategory, error)
	GetParents(ctx context.Context, categoryID uuid.UUID) ([]*entity.ProductCategory, error)
	GetSiblings(ctx context.Context, categoryID uuid.UUID, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)

	// Path operations
	GetByPath(ctx context.Context, path string, include *ProductCategoryInclude) (*entity.ProductCategory, error)
	GetCategoriesByPathPrefix(ctx context.Context, pathPrefix string, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)

	// Tree operations
	GetCategoryTree(ctx context.Context, rootID *uuid.UUID, maxDepth *int) ([]*entity.ProductCategory, error)
	GetFullCategoryTree(ctx context.Context) ([]*entity.ProductCategory, error)
	BuildCategoryBreadcrumb(ctx context.Context, categoryID uuid.UUID) ([]*entity.ProductCategory, error)

	// Status management
	Activate(ctx context.Context, categoryID uuid.UUID) error
	Deactivate(ctx context.Context, categoryID uuid.UUID) error
	BulkActivate(ctx context.Context, categoryIDs []uuid.UUID) error
	BulkDeactivate(ctx context.Context, categoryIDs []uuid.UUID) error

	// Move operations
	MoveCategory(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID) error
	ReorderChildren(ctx context.Context, parentID uuid.UUID, categoryOrders []CategoryOrder) error
	UpdateSortOrder(ctx context.Context, categoryID uuid.UUID, sortOrder int) error

	// Search operations
	Search(ctx context.Context, query string, filter *ProductCategoryFilter, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)
	SearchWithinTree(ctx context.Context, query string, rootID uuid.UUID, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)

	// Product relationship operations
	GetCategoriesWithProducts(ctx context.Context, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)
	GetEmptyCategories(ctx context.Context, include *ProductCategoryInclude) ([]*entity.ProductCategory, error)
	GetCategoryProductCounts(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]int64, error)

	// Slug management
	IsSlugExists(ctx context.Context, slug string) (bool, error)
	IsSlugExistsExcludingCategory(ctx context.Context, slug string, categoryID uuid.UUID) (bool, error)
	GenerateUniqueSlug(ctx context.Context, baseName string) (string, error)

	// Validation operations
	ValidateCategoryHierarchy(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID) error
	CheckCircularReference(ctx context.Context, categoryID uuid.UUID, newParentID uuid.UUID) error
	ValidateMaxDepth(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID, maxDepth int) error

	// Analytics operations
	GetCategoryStatistics(ctx context.Context, categoryID uuid.UUID) (*CategoryStatistics, error)
	GetMostPopularCategories(ctx context.Context, limit int, days int) ([]*entity.ProductCategory, error)
	GetCategoryPerformance(ctx context.Context, categoryID uuid.UUID, days int) (*CategoryPerformance, error)

	// Bulk operations
	BulkMove(ctx context.Context, moves []CategoryMove) error
	BulkDelete(ctx context.Context, categoryIDs []uuid.UUID, deleteProducts bool) error

	// Import/Export operations
	ExportCategoryTree(ctx context.Context, rootID *uuid.UUID) ([]*CategoryExport, error)
	ImportCategoryTree(ctx context.Context, categories []*CategoryImport) error

	// Utility operations
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	GetMaxLevel(ctx context.Context) (int, error)
	GetCategoryCount(ctx context.Context) (int64, error)
	GetActiveChildrenCount(ctx context.Context, parentID uuid.UUID) (int, error)
}

// CategoryOrder defines the sort order for a category
type CategoryOrder struct {
	CategoryID uuid.UUID `json:"category_id"`
	SortOrder  int       `json:"sort_order"`
}

// CategoryMove defines a category move operation
type CategoryMove struct {
	CategoryID  uuid.UUID  `json:"category_id"`
	NewParentID *uuid.UUID `json:"new_parent_id"`
}

// CategoryStatistics contains analytics data for a category
type CategoryStatistics struct {
	CategoryID          uuid.UUID  `json:"category_id"`
	TotalProducts       int64      `json:"total_products"`
	ActiveProducts      int64      `json:"active_products"`
	TotalSubcategories  int        `json:"total_subcategories"`
	ActiveSubcategories int        `json:"active_subcategories"`
	Level               int        `json:"level"`
	MaxDepth            int        `json:"max_depth"`
	TotalViews          int64      `json:"total_views"`
	LastProductAddedAt  *time.Time `json:"last_product_added_at,omitempty"`
}

// CategoryPerformance contains performance metrics for a category
type CategoryPerformance struct {
	CategoryID     uuid.UUID `json:"category_id"`
	Views          int64     `json:"views"`
	ProductViews   int64     `json:"product_views"`
	Sales          int64     `json:"sales"`
	Revenue        string    `json:"revenue"` // Using string for decimal
	ConversionRate *string   `json:"conversion_rate,omitempty"`
	AvgTimeOnPage  *int      `json:"avg_time_on_page,omitempty"` // in seconds
	BounceRate     *string   `json:"bounce_rate,omitempty"`
}

// CategoryExport defines the structure for exporting categories
type CategoryExport struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	ParentPath  *string    `json:"parent_path,omitempty"`
	Level       int        `json:"level"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CategoryImport defines the structure for importing categories
type CategoryImport struct {
	Name        string  `json:"name"`
	Slug        *string `json:"slug,omitempty"` // Auto-generated if not provided
	Description *string `json:"description,omitempty"`
	ParentPath  *string `json:"parent_path,omitempty"` // Path to parent category
	SortOrder   *int    `json:"sort_order,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}
