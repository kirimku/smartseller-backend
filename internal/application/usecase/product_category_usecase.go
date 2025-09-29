package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// ProductCategoryUseCase handles all product category-related business operations
type ProductCategoryUseCase struct {
	categoryRepo repository.ProductCategoryRepository
	productRepo  repository.ProductRepository
	logger       *slog.Logger
}

// NewProductCategoryUseCase creates a new instance of ProductCategoryUseCase
func NewProductCategoryUseCase(
	categoryRepo repository.ProductCategoryRepository,
	productRepo repository.ProductRepository,
	logger *slog.Logger,
) *ProductCategoryUseCase {
	return &ProductCategoryUseCase{
		categoryRepo: categoryRepo,
		productRepo:  productRepo,
		logger:       logger,
	}
}

// CreateCategoryRequest represents the data needed to create a category
type CreateCategoryRequest struct {
	Name        string     `json:"name" validate:"required,min=1,max=255"`
	Description *string    `json:"description" validate:"omitempty,max=1000"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty"`
	IsActive    bool       `json:"is_active"`
	SortOrder   int        `json:"sort_order" validate:"min=0"`
}

// UpdateCategoryRequest represents the data needed to update a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order" validate:"omitempty,min=0"`
}

// MoveCategoryRequest represents a category move request
type MoveCategoryRequest struct {
	CategoryID   uuid.UUID  `json:"category_id" validate:"required"`
	NewParentID  *uuid.UUID `json:"new_parent_id" validate:"omitempty"`
	NewSortOrder *int       `json:"new_sort_order" validate:"omitempty,min=0"`
}

// CategoryTreeNode represents a node in the category tree
type CategoryTreeNode struct {
	*entity.ProductCategory
	Children []*CategoryTreeNode `json:"children"`
	Level    int                 `json:"level"`
}

// CreateCategory creates a new category with hierarchy validation
func (uc *ProductCategoryUseCase) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*entity.ProductCategory, error) {
	// Validate hierarchy if parent is specified
	if req.ParentID != nil {
		if err := uc.validateParentCategory(ctx, *req.ParentID); err != nil {
			uc.logger.Error("Parent category validation failed",
				"parent_id", *req.ParentID,
				"error", err)
			return nil, fmt.Errorf("invalid parent category: %w", err)
		}
	}

	// Generate slug from name
	slug := uc.generateSlug(req.Name)

	// Check if slug is unique using the repository method
	exists, err := uc.categoryRepo.IsSlugExists(ctx, slug)
	if err != nil {
		uc.logger.Error("Failed to check slug uniqueness",
			"slug", slug,
			"error", err)
		return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("slug '%s' already exists", slug)
	}

	// Create category entity
	category := &entity.ProductCategory{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		Slug:        slug,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create category in repository
	if err := uc.categoryRepo.Create(ctx, category); err != nil {
		uc.logger.Error("Failed to create category in repository",
			"name", req.Name,
			"error", err)
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	uc.logger.Info("Category created successfully",
		"category_id", category.ID,
		"name", category.Name,
		"slug", category.Slug)

	return category, nil
}

// UpdateCategory updates an existing category
func (uc *ProductCategoryUseCase) UpdateCategory(ctx context.Context, categoryID uuid.UUID, req UpdateCategoryRequest) (*entity.ProductCategory, error) {
	if categoryID == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	// Get existing category
	existingCategory, err := uc.categoryRepo.GetByID(ctx, categoryID, nil)
	if err != nil {
		uc.logger.Error("Category not found for update",
			"category_id", categoryID,
			"error", err)
		return nil, fmt.Errorf("category not found: %w", err)
	}

	// Apply updates
	updatedCategory := uc.applyCategoryUpdates(existingCategory, req)

	// Regenerate slug if name changed
	if req.Name != nil && *req.Name != existingCategory.Name {
		newSlug := uc.generateSlug(*req.Name)
		exists, err := uc.categoryRepo.IsSlugExistsExcludingCategory(ctx, newSlug, categoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to check slug uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("slug '%s' already exists", newSlug)
		}
		updatedCategory.Slug = newSlug
	}

	updatedCategory.UpdatedAt = time.Now()

	// Update category in repository
	if err := uc.categoryRepo.Update(ctx, updatedCategory); err != nil {
		uc.logger.Error("Failed to update category in repository",
			"category_id", categoryID,
			"error", err)
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	uc.logger.Info("Category updated successfully",
		"category_id", categoryID,
		"name", updatedCategory.Name)

	return updatedCategory, nil
}

// DeleteCategory deletes a category with product reassignment
func (uc *ProductCategoryUseCase) DeleteCategory(ctx context.Context, categoryID uuid.UUID, reassignToID *uuid.UUID) error {
	if categoryID == uuid.Nil {
		return fmt.Errorf("category ID cannot be empty")
	}

	// Get existing category
	category, err := uc.categoryRepo.GetByID(ctx, categoryID, nil)
	if err != nil {
		uc.logger.Error("Category not found for deletion",
			"category_id", categoryID,
			"error", err)
		return fmt.Errorf("category not found: %w", err)
	}

	// Check for child categories
	children, err := uc.categoryRepo.GetChildren(ctx, categoryID, nil)
	if err != nil {
		uc.logger.Error("Failed to check for child categories",
			"category_id", categoryID,
			"error", err)
		return fmt.Errorf("failed to check child categories: %w", err)
	}

	if len(children) > 0 {
		return fmt.Errorf("cannot delete category with child categories")
	}

	// Handle product reassignment
	if err := uc.handleProductReassignment(ctx, categoryID, reassignToID); err != nil {
		return fmt.Errorf("failed to reassign products: %w", err)
	}

	// Delete category
	if err := uc.categoryRepo.Delete(ctx, categoryID); err != nil {
		uc.logger.Error("Failed to delete category",
			"category_id", categoryID,
			"error", err)
		return fmt.Errorf("failed to delete category: %w", err)
	}

	uc.logger.Info("Category deleted successfully",
		"category_id", categoryID,
		"name", category.Name)

	return nil
}

// GetCategoryTree returns the complete category tree for navigation
func (uc *ProductCategoryUseCase) GetCategoryTree(ctx context.Context, rootID *uuid.UUID) ([]*CategoryTreeNode, error) {
	// Use the repository method to get category tree
	categories, err := uc.categoryRepo.GetCategoryTree(ctx, rootID, nil)
	if err != nil {
		uc.logger.Error("Failed to get category tree",
			"root_id", rootID,
			"error", err)
		return nil, fmt.Errorf("failed to get category tree: %w", err)
	}

	// Build tree structure
	tree := uc.buildCategoryTree(categories, rootID, 0)

	uc.logger.Debug("Category tree retrieved",
		"root_count", len(tree))

	return tree, nil
}

// MoveCategoryToParent moves a category to a new parent with validation
func (uc *ProductCategoryUseCase) MoveCategoryToParent(ctx context.Context, req MoveCategoryRequest) (*entity.ProductCategory, error) {
	if req.CategoryID == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be empty")
	}

	// Validate the move using repository validation
	if req.NewParentID != nil {
		if err := uc.categoryRepo.ValidateCategoryHierarchy(ctx, req.CategoryID, req.NewParentID); err != nil {
			return nil, fmt.Errorf("invalid category move: %w", err)
		}
	}

	// Use repository method to move category
	if err := uc.categoryRepo.MoveCategory(ctx, req.CategoryID, req.NewParentID); err != nil {
		uc.logger.Error("Failed to move category",
			"category_id", req.CategoryID,
			"new_parent_id", req.NewParentID,
			"error", err)
		return nil, fmt.Errorf("failed to move category: %w", err)
	}

	// Get updated category
	category, err := uc.categoryRepo.GetByID(ctx, req.CategoryID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated category: %w", err)
	}

	uc.logger.Info("Category moved successfully",
		"category_id", req.CategoryID,
		"new_parent_id", req.NewParentID)

	return category, nil
}

// Helper methods

func (uc *ProductCategoryUseCase) validateParentCategory(ctx context.Context, parentID uuid.UUID) error {
	parent, err := uc.categoryRepo.GetByID(ctx, parentID, nil)
	if err != nil {
		return fmt.Errorf("parent category not found")
	}

	if !parent.IsActive {
		return fmt.Errorf("parent category is not active")
	}

	return nil
}

func (uc *ProductCategoryUseCase) generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")
	// Remove special characters (keep only alphanumeric and hyphens)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func (uc *ProductCategoryUseCase) applyCategoryUpdates(category *entity.ProductCategory, req UpdateCategoryRequest) *entity.ProductCategory {
	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Description != nil {
		category.Description = req.Description
	}
	if req.IsActive != nil {
		category.IsActive = *req.IsActive
	}
	if req.SortOrder != nil {
		category.SortOrder = *req.SortOrder
	}

	return category
}

func (uc *ProductCategoryUseCase) handleProductReassignment(ctx context.Context, categoryID uuid.UUID, reassignToID *uuid.UUID) error {
	// Get products in this category
	filter := &repository.ProductFilter{
		CategoryIDs: []uuid.UUID{categoryID},
	}

	products, err := uc.productRepo.List(ctx, filter, nil)
	if err != nil {
		return fmt.Errorf("failed to get products in category: %w", err)
	}

	if len(products) == 0 {
		return nil // No products to reassign
	}

	// Update products to new category
	for _, product := range products {
		product.CategoryID = reassignToID
		if err := uc.productRepo.Update(ctx, product); err != nil {
			return fmt.Errorf("failed to reassign product %s: %w", product.SKU, err)
		}
	}

	uc.logger.Info("Products reassigned successfully",
		"count", len(products),
		"from_category", categoryID,
		"to_category", reassignToID)

	return nil
}

func (uc *ProductCategoryUseCase) buildCategoryTree(categories []*entity.ProductCategory, parentID *uuid.UUID, level int) []*CategoryTreeNode {
	var nodes []*CategoryTreeNode

	for _, category := range categories {
		// Check if this category belongs to the current parent level
		isMatch := (parentID == nil && category.ParentID == nil) ||
			(parentID != nil && category.ParentID != nil && *category.ParentID == *parentID)

		if isMatch {
			node := &CategoryTreeNode{
				ProductCategory: category,
				Level:           level,
				Children:        uc.buildCategoryTree(categories, &category.ID, level+1),
			}
			nodes = append(nodes, node)
		}
	}

	return nodes
}

// GetCategoryByID retrieves a category by its ID
func (uc *ProductCategoryUseCase) GetCategoryByID(ctx context.Context, categoryID uuid.UUID) (*entity.ProductCategory, error) {
	include := &repository.ProductCategoryInclude{}
	category, err := uc.categoryRepo.GetByID(ctx, categoryID, include)
	if err != nil {
		uc.logger.Error("Failed to get category by ID",
			"category_id", categoryID,
			"error", err)
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// GetAllCategories retrieves all categories with optional filtering
func (uc *ProductCategoryUseCase) GetAllCategories(ctx context.Context, filters repository.ProductCategoryFilter) ([]*entity.ProductCategory, error) {
	include := &repository.ProductCategoryInclude{}
	categories, err := uc.categoryRepo.List(ctx, &filters, include)
	if err != nil {
		uc.logger.Error("Failed to get all categories",
			"filters", filters,
			"error", err)
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}
