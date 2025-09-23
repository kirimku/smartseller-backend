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

// PostgreSQLProductCategoryRepository implements the ProductCategoryRepository interface using PostgreSQL
type PostgreSQLProductCategoryRepository struct {
	db *sqlx.DB
}

// NewPostgreSQLProductCategoryRepository creates a new PostgreSQL product category repository
func NewPostgreSQLProductCategoryRepository(db *sqlx.DB) repository.ProductCategoryRepository {
	return &PostgreSQLProductCategoryRepository{
		db: db,
	}
}

// Create creates a new product category in the database
func (r *PostgreSQLProductCategoryRepository) Create(ctx context.Context, category *entity.ProductCategory) error {
	// Validate category before creating
	if err := category.Validate(); err != nil {
		return fmt.Errorf("category validation failed: %w", err)
	}

	// Ensure ID is set
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}

	// Set timestamps
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `
		INSERT INTO product_categories (
			id, name, description, slug, parent_id, sort_order, is_active, created_at, updated_at
		) VALUES (
			:id, :name, :description, :slug, :parent_id, :sort_order, :is_active, :created_at, :updated_at
		)`

	_, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		// Handle unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "slug") {
					return fmt.Errorf("category with slug '%s' already exists", category.Slug)
				}
				return fmt.Errorf("category already exists: %w", err)
			case "23503": // foreign_key_violation
				if strings.Contains(pqErr.Detail, "parent_id") {
					return fmt.Errorf("parent category does not exist")
				}
				return fmt.Errorf("foreign key violation: %w", err)
			}
		}
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// GetByID retrieves a product category by its ID
func (r *PostgreSQLProductCategoryRepository) GetByID(ctx context.Context, id uuid.UUID, include *repository.ProductCategoryInclude) (*entity.ProductCategory, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be nil")
	}

	query := `
		SELECT id, name, description, slug, parent_id, sort_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE id = $1`

	var category entity.ProductCategory
	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category with ID '%s' not found", id)
		}
		return nil, fmt.Errorf("failed to get category by ID: %w", err)
	}

	// Load related data if requested
	if include != nil {
		if err := r.loadCategoryRelations(ctx, &category, include); err != nil {
			return nil, fmt.Errorf("failed to load category relations: %w", err)
		}
	}

	// Compute fields
	category.ComputeFields()

	return &category, nil
}

// GetBySlug retrieves a product category by its slug
func (r *PostgreSQLProductCategoryRepository) GetBySlug(ctx context.Context, slug string, include *repository.ProductCategoryInclude) (*entity.ProductCategory, error) {
	if strings.TrimSpace(slug) == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	query := `
		SELECT id, name, description, slug, parent_id, sort_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE slug = $1`

	var category entity.ProductCategory
	err := r.db.GetContext(ctx, &category, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category with slug '%s' not found", slug)
		}
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	// Load related data if requested
	if include != nil {
		if err := r.loadCategoryRelations(ctx, &category, include); err != nil {
			return nil, fmt.Errorf("failed to load category relations: %w", err)
		}
	}

	// Compute fields
	category.ComputeFields()

	return &category, nil
}

// Update updates an existing product category
func (r *PostgreSQLProductCategoryRepository) Update(ctx context.Context, category *entity.ProductCategory) error {
	// Validate category before updating
	if err := category.Validate(); err != nil {
		return fmt.Errorf("category validation failed: %w", err)
	}

	if category.ID == uuid.Nil {
		return fmt.Errorf("category ID cannot be nil for update")
	}

	// Update timestamp
	category.UpdatedAt = time.Now()

	query := `
		UPDATE product_categories SET
			name = :name,
			description = :description,
			slug = :slug,
			parent_id = :parent_id,
			sort_order = :sort_order,
			is_active = :is_active,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, category)
	if err != nil {
		// Handle unique constraint violations
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Detail, "slug") {
					return fmt.Errorf("category with slug '%s' already exists", category.Slug)
				}
				return fmt.Errorf("category already exists: %w", err)
			case "23503": // foreign_key_violation
				if strings.Contains(pqErr.Detail, "parent_id") {
					return fmt.Errorf("parent category does not exist")
				}
				return fmt.Errorf("foreign key violation: %w", err)
			}
		}
		return fmt.Errorf("failed to update category: %w", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with ID '%s' not found", category.ID)
	}

	return nil
}

// Delete deletes a product category
func (r *PostgreSQLProductCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("category ID cannot be nil")
	}

	// Check if category has children
	hasChildren, err := r.hasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check for children: %w", err)
	}

	if hasChildren {
		return fmt.Errorf("cannot delete category with subcategories")
	}

	query := `DELETE FROM product_categories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with ID '%s' not found", id)
	}

	return nil
}

// GetRootCategories retrieves all root categories (categories without a parent)
func (r *PostgreSQLProductCategoryRepository) GetRootCategories(ctx context.Context, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	query := `
		SELECT id, name, description, slug, parent_id, sort_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE parent_id IS NULL
		ORDER BY sort_order ASC, name ASC`

	var categories []entity.ProductCategory
	err := r.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get root categories: %w", err)
	}

	// Convert to pointer slice and load relations
	result := make([]*entity.ProductCategory, len(categories))
	for i := range categories {
		result[i] = &categories[i]

		// Load related data if requested
		if include != nil {
			if err := r.loadCategoryRelations(ctx, result[i], include); err != nil {
				return nil, fmt.Errorf("failed to load category relations: %w", err)
			}
		}

		// Compute fields
		result[i].ComputeFields()
	}

	return result, nil
}

// GetChildren retrieves all direct children of a category
func (r *PostgreSQLProductCategoryRepository) GetChildren(ctx context.Context, parentID uuid.UUID, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	if parentID == uuid.Nil {
		return nil, fmt.Errorf("parent ID cannot be nil")
	}

	query := `
		SELECT id, name, description, slug, parent_id, sort_order, is_active, created_at, updated_at
		FROM product_categories
		WHERE parent_id = $1
		ORDER BY sort_order ASC, name ASC`

	var categories []entity.ProductCategory
	err := r.db.SelectContext(ctx, &categories, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get children categories: %w", err)
	}

	// Convert to pointer slice and load relations
	result := make([]*entity.ProductCategory, len(categories))
	for i := range categories {
		result[i] = &categories[i]

		// Load related data if requested
		if include != nil {
			if err := r.loadCategoryRelations(ctx, result[i], include); err != nil {
				return nil, fmt.Errorf("failed to load category relations: %w", err)
			}
		}

		// Compute fields
		result[i].ComputeFields()
	}

	return result, nil
}

// Activate activates a category
func (r *PostgreSQLProductCategoryRepository) Activate(ctx context.Context, categoryID uuid.UUID) error {
	if categoryID == uuid.Nil {
		return fmt.Errorf("category ID cannot be nil")
	}

	query := `UPDATE product_categories SET is_active = true, updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, categoryID)
	if err != nil {
		return fmt.Errorf("failed to activate category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with ID '%s' not found", categoryID)
	}

	return nil
}

// Deactivate deactivates a category
func (r *PostgreSQLProductCategoryRepository) Deactivate(ctx context.Context, categoryID uuid.UUID) error {
	if categoryID == uuid.Nil {
		return fmt.Errorf("category ID cannot be nil")
	}

	query := `UPDATE product_categories SET is_active = false, updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, categoryID)
	if err != nil {
		return fmt.Errorf("failed to deactivate category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category with ID '%s' not found", categoryID)
	}

	return nil
}

// IsSlugExists checks if a slug exists
func (r *PostgreSQLProductCategoryRepository) IsSlugExists(ctx context.Context, slug string) (bool, error) {
	if strings.TrimSpace(slug) == "" {
		return false, fmt.Errorf("slug cannot be empty")
	}

	query := `SELECT EXISTS(SELECT 1 FROM product_categories WHERE slug = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, slug)
	if err != nil {
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}

	return exists, nil
}

// Exists checks if a category exists by ID
func (r *PostgreSQLProductCategoryRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	if id == uuid.Nil {
		return false, fmt.Errorf("category ID cannot be nil")
	}

	query := `SELECT EXISTS(SELECT 1 FROM product_categories WHERE id = $1)`

	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check category existence: %w", err)
	}

	return exists, nil
}

// Helper methods

// hasChildren checks if a category has any children
func (r *PostgreSQLProductCategoryRepository) hasChildren(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM product_categories WHERE parent_id = $1)`

	var hasChildren bool
	err := r.db.GetContext(ctx, &hasChildren, query, categoryID)
	if err != nil {
		return false, fmt.Errorf("failed to check for children: %w", err)
	}

	return hasChildren, nil
}

// loadCategoryRelations loads related data based on include options
func (r *PostgreSQLProductCategoryRepository) loadCategoryRelations(ctx context.Context, category *entity.ProductCategory, include *repository.ProductCategoryInclude) error {
	// Load parent if requested
	if include.Parent && category.ParentID != nil {
		parent, err := r.GetByID(ctx, *category.ParentID, nil) // Avoid infinite recursion
		if err == nil {
			category.Parent = parent
		}
		// Don't fail if parent not found - it might have been deleted
	}

	// Load children if requested
	if include.Children {
		children, err := r.GetChildren(ctx, category.ID, nil) // Avoid infinite recursion
		if err != nil {
			return fmt.Errorf("failed to load children: %w", err)
		}
		category.Children = children
	}

	// Load all children recursively if requested
	if include.AllChildren {
		if err := r.loadAllChildren(ctx, category, include.MaxDepth); err != nil {
			return fmt.Errorf("failed to load all children: %w", err)
		}
	}

	return nil
}

// loadAllChildren loads all descendants recursively
func (r *PostgreSQLProductCategoryRepository) loadAllChildren(ctx context.Context, category *entity.ProductCategory, maxDepth *int) error {
	if maxDepth != nil && *maxDepth <= 0 {
		return nil
	}

	children, err := r.GetChildren(ctx, category.ID, nil)
	if err != nil {
		return err
	}

	category.Children = children

	// Recursively load children's children
	var nextMaxDepth *int
	if maxDepth != nil {
		next := *maxDepth - 1
		nextMaxDepth = &next
	}

	for _, child := range children {
		if err := r.loadAllChildren(ctx, child, nextMaxDepth); err != nil {
			return err
		}
	}

	return nil
}

// Placeholder implementations for remaining methods
// These would need full implementation based on business requirements

func (r *PostgreSQLProductCategoryRepository) CreateBatch(ctx context.Context, categories []*entity.ProductCategory) error {
	// TODO: Implement batch create
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetByIDs(ctx context.Context, ids []uuid.UUID, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement batch get by IDs
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) UpdateBatch(ctx context.Context, categories []*entity.ProductCategory) error {
	// TODO: Implement batch update
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) DeleteBatch(ctx context.Context, ids []uuid.UUID) error {
	// TODO: Implement batch delete
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) List(ctx context.Context, filter *repository.ProductCategoryFilter, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement list with filtering
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) Count(ctx context.Context, filter *repository.ProductCategoryFilter) (int64, error) {
	// TODO: Implement count with filtering
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetAllChildren(ctx context.Context, parentID uuid.UUID, maxDepth *int) ([]*entity.ProductCategory, error) {
	// TODO: Implement get all children
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetParents(ctx context.Context, categoryID uuid.UUID) ([]*entity.ProductCategory, error) {
	// TODO: Implement get parents
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetSiblings(ctx context.Context, categoryID uuid.UUID, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement get siblings
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetByPath(ctx context.Context, path string, include *repository.ProductCategoryInclude) (*entity.ProductCategory, error) {
	// TODO: Implement get by path
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetCategoriesByPathPrefix(ctx context.Context, pathPrefix string, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement get categories by path prefix
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetCategoryTree(ctx context.Context, rootID *uuid.UUID, maxDepth *int) ([]*entity.ProductCategory, error) {
	// TODO: Implement get category tree
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetFullCategoryTree(ctx context.Context) ([]*entity.ProductCategory, error) {
	// TODO: Implement get full category tree
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) BuildCategoryBreadcrumb(ctx context.Context, categoryID uuid.UUID) ([]*entity.ProductCategory, error) {
	// TODO: Implement build category breadcrumb
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) BulkActivate(ctx context.Context, categoryIDs []uuid.UUID) error {
	// TODO: Implement bulk activate
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) BulkDeactivate(ctx context.Context, categoryIDs []uuid.UUID) error {
	// TODO: Implement bulk deactivate
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) MoveCategory(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID) error {
	// TODO: Implement move category
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) ReorderChildren(ctx context.Context, parentID uuid.UUID, categoryOrders []repository.CategoryOrder) error {
	// TODO: Implement reorder children
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) UpdateSortOrder(ctx context.Context, categoryID uuid.UUID, sortOrder int) error {
	// TODO: Implement update sort order
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) Search(ctx context.Context, query string, filter *repository.ProductCategoryFilter, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement search
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) SearchWithinTree(ctx context.Context, query string, rootID uuid.UUID, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement search within tree
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetCategoriesWithProducts(ctx context.Context, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement get categories with products
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetEmptyCategories(ctx context.Context, include *repository.ProductCategoryInclude) ([]*entity.ProductCategory, error) {
	// TODO: Implement get empty categories
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetCategoryProductCounts(ctx context.Context, categoryIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	// TODO: Implement get category product counts
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) IsSlugExistsExcludingCategory(ctx context.Context, slug string, categoryID uuid.UUID) (bool, error) {
	// TODO: Implement slug exists excluding category
	return false, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GenerateUniqueSlug(ctx context.Context, baseName string) (string, error) {
	// TODO: Implement generate unique slug
	return "", fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) ValidateCategoryHierarchy(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID) error {
	// TODO: Implement validate category hierarchy
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) CheckCircularReference(ctx context.Context, categoryID uuid.UUID, newParentID uuid.UUID) error {
	// TODO: Implement check circular reference
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) ValidateMaxDepth(ctx context.Context, categoryID uuid.UUID, newParentID *uuid.UUID, maxDepth int) error {
	// TODO: Implement validate max depth
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetCategoryStatistics(ctx context.Context, categoryID uuid.UUID) (*repository.CategoryStatistics, error) {
	// TODO: Implement get category statistics
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetMostPopularCategories(ctx context.Context, limit int, days int) ([]*entity.ProductCategory, error) {
	// TODO: Implement get most popular categories
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetCategoryPerformance(ctx context.Context, categoryID uuid.UUID, days int) (*repository.CategoryPerformance, error) {
	// TODO: Implement get category performance
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) BulkMove(ctx context.Context, moves []repository.CategoryMove) error {
	// TODO: Implement bulk move
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) BulkDelete(ctx context.Context, categoryIDs []uuid.UUID, deleteProducts bool) error {
	// TODO: Implement bulk delete
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) ExportCategoryTree(ctx context.Context, rootID *uuid.UUID) ([]*repository.CategoryExport, error) {
	// TODO: Implement export category tree
	return nil, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) ImportCategoryTree(ctx context.Context, categories []*repository.CategoryImport) error {
	// TODO: Implement import category tree
	return fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	return r.IsSlugExists(ctx, slug)
}

func (r *PostgreSQLProductCategoryRepository) GetMaxLevel(ctx context.Context) (int, error) {
	// TODO: Implement get max level
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetCategoryCount(ctx context.Context) (int64, error) {
	// TODO: Implement get category count
	return 0, fmt.Errorf("not implemented")
}

func (r *PostgreSQLProductCategoryRepository) GetActiveChildrenCount(ctx context.Context, parentID uuid.UUID) (int, error) {
	// TODO: Implement get active children count
	return 0, fmt.Errorf("not implemented")
}
