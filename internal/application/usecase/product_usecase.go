package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// ProductUseCase handles all product-related business operations
type ProductUseCase struct {
	productRepo       repository.ProductRepository
	categoryRepo      repository.ProductCategoryRepository
	variantRepo       repository.ProductVariantRepository
	variantOptionRepo repository.ProductVariantOptionRepository
	imageRepo         repository.ProductImageRepository
	logger            *slog.Logger
}

// NewProductUseCase creates a new instance of ProductUseCase
func NewProductUseCase(
	productRepo repository.ProductRepository,
	categoryRepo repository.ProductCategoryRepository,
	variantRepo repository.ProductVariantRepository,
	variantOptionRepo repository.ProductVariantOptionRepository,
	imageRepo repository.ProductImageRepository,
	logger *slog.Logger,
) *ProductUseCase {
	return &ProductUseCase{
		productRepo:       productRepo,
		categoryRepo:      categoryRepo,
		variantRepo:       variantRepo,
		variantOptionRepo: variantOptionRepo,
		imageRepo:         imageRepo,
		logger:            logger,
	}
}

// CreateProductRequest represents the data needed to create a new product
type CreateProductRequest struct {
	Name              string               `json:"name" validate:"required,min=1,max=255"`
	Description       *string              `json:"description" validate:"omitempty,max=5000"`
	SKU               string               `json:"sku" validate:"required,min=3,max=100"`
	CategoryID        *uuid.UUID           `json:"category_id" validate:"omitempty"`
	Brand             *string              `json:"brand" validate:"omitempty,max=255"`
	Tags              []string             `json:"tags" validate:"omitempty,dive,max=50"`
	BasePrice         decimal.Decimal      `json:"base_price" validate:"required,min=0"`
	SalePrice         *decimal.Decimal     `json:"sale_price" validate:"omitempty,min=0"`
	CostPrice         *decimal.Decimal     `json:"cost_price" validate:"omitempty,min=0"`
	TrackInventory    bool                 `json:"track_inventory"`
	StockQuantity     int                  `json:"stock_quantity" validate:"min=0"`
	LowStockThreshold *int                 `json:"low_stock_threshold" validate:"omitempty,min=0"`
	Status            entity.ProductStatus `json:"status" validate:"omitempty,oneof=draft active inactive archived"`
	MetaTitle         *string              `json:"meta_title" validate:"omitempty,max=255"`
	MetaDescription   *string              `json:"meta_description" validate:"omitempty,max=500"`
	Slug              *string              `json:"slug" validate:"omitempty,max=255"`
	Weight            *decimal.Decimal     `json:"weight" validate:"omitempty,min=0"`
	DimensionsLength  *decimal.Decimal     `json:"dimensions_length" validate:"omitempty,min=0"`
	DimensionsWidth   *decimal.Decimal     `json:"dimensions_width" validate:"omitempty,min=0"`
	DimensionsHeight  *decimal.Decimal     `json:"dimensions_height" validate:"omitempty,min=0"`
	CreatedBy         uuid.UUID            `json:"created_by" validate:"required"`
}

// UpdateProductRequest represents the data needed to update a product
type UpdateProductRequest struct {
	Name              *string               `json:"name" validate:"omitempty,min=1,max=255"`
	Description       *string               `json:"description" validate:"omitempty,max=5000"`
	CategoryID        *uuid.UUID            `json:"category_id" validate:"omitempty"`
	Brand             *string               `json:"brand" validate:"omitempty,max=255"`
	Tags              []string              `json:"tags" validate:"omitempty,dive,max=50"`
	BasePrice         *decimal.Decimal      `json:"base_price" validate:"omitempty,min=0"`
	SalePrice         *decimal.Decimal      `json:"sale_price" validate:"omitempty,min=0"`
	CostPrice         *decimal.Decimal      `json:"cost_price" validate:"omitempty,min=0"`
	TrackInventory    *bool                 `json:"track_inventory"`
	StockQuantity     *int                  `json:"stock_quantity" validate:"omitempty,min=0"`
	LowStockThreshold *int                  `json:"low_stock_threshold" validate:"omitempty,min=0"`
	Status            *entity.ProductStatus `json:"status" validate:"omitempty,oneof=draft active inactive archived"`
	MetaTitle         *string               `json:"meta_title" validate:"omitempty,max=255"`
	MetaDescription   *string               `json:"meta_description" validate:"omitempty,max=500"`
	Slug              *string               `json:"slug" validate:"omitempty,max=255"`
	Weight            *decimal.Decimal      `json:"weight" validate:"omitempty,min=0"`
	DimensionsLength  *decimal.Decimal      `json:"dimensions_length" validate:"omitempty,min=0"`
	DimensionsWidth   *decimal.Decimal      `json:"dimensions_width" validate:"omitempty,min=0"`
	DimensionsHeight  *decimal.Decimal      `json:"dimensions_height" validate:"omitempty,min=0"`
}

// ProductListFilter represents filtering options for listing products
type ProductListFilter struct {
	CategoryIDs    []uuid.UUID            `json:"category_ids"`
	Status         []entity.ProductStatus `json:"status"`
	MinPrice       *decimal.Decimal       `json:"min_price"`
	MaxPrice       *decimal.Decimal       `json:"max_price"`
	MinStock       *int                   `json:"min_stock"`
	MaxStock       *int                   `json:"max_stock"`
	IsLowStock     *bool                  `json:"is_low_stock"`
	TrackInventory *bool                  `json:"track_inventory"`
	SearchQuery    string                 `json:"search_query"`
	Tags           []string               `json:"tags"`
	CreatedAfter   *time.Time             `json:"created_after"`
	CreatedBefore  *time.Time             `json:"created_before"`
	UpdatedAfter   *time.Time             `json:"updated_after"`
	UpdatedBefore  *time.Time             `json:"updated_before"`
}

// ListProductsRequest represents pagination and filtering for product listing
type ListProductsRequest struct {
	Filter   ProductListFilter          `json:"filter"`
	Page     int                        `json:"page" validate:"min=1"`
	PageSize int                        `json:"page_size" validate:"min=1,max=100"`
	SortBy   string                     `json:"sort_by" validate:"omitempty,oneof=name created_at updated_at base_price stock_quantity status"`
	SortDesc bool                       `json:"sort_desc"`
	Include  *repository.ProductInclude `json:"include"`
}

// CreateProduct creates a new product with business validation
func (uc *ProductUseCase) CreateProduct(ctx context.Context, req CreateProductRequest) (*entity.Product, error) {
	// Validate business rules
	if err := uc.validateCreateProductRequest(ctx, req); err != nil {
		uc.logger.Error("Product creation validation failed",
			"sku", req.SKU,
			"error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if SKU already exists
	existing, err := uc.productRepo.GetBySKU(ctx, req.SKU, nil)
	if err == nil && existing != nil {
		uc.logger.Warn("Attempt to create product with duplicate SKU",
			"sku", req.SKU,
			"existing_id", existing.ID)
		return nil, fmt.Errorf("product with SKU %s already exists", req.SKU)
	}

	// Verify category exists if provided
	if req.CategoryID != nil {
		category, err := uc.categoryRepo.GetByID(ctx, *req.CategoryID, nil)
		if err != nil {
			uc.logger.Error("Category not found for product creation",
				"category_id", *req.CategoryID,
				"error", err)
			return nil, fmt.Errorf("category not found: %w", err)
		}
		uc.logger.Debug("Category validated for product creation",
			"category_id", category.ID,
			"category_name", category.Name)
	}

	// Create product entity
	product := &entity.Product{
		ID:                uuid.New(),
		SKU:               req.SKU,
		Name:              req.Name,
		Description:       req.Description,
		CategoryID:        req.CategoryID,
		Brand:             req.Brand,
		Tags:              pq.StringArray(req.Tags),
		BasePrice:         req.BasePrice,
		SalePrice:         req.SalePrice,
		CostPrice:         req.CostPrice,
		TrackInventory:    req.TrackInventory,
		StockQuantity:     req.StockQuantity,
		LowStockThreshold: req.LowStockThreshold,
		Status:            req.Status,
		MetaTitle:         req.MetaTitle,
		MetaDescription:   req.MetaDescription,
		Slug:              req.Slug,
		Weight:            req.Weight,
		DimensionsLength:  req.DimensionsLength,
		DimensionsWidth:   req.DimensionsWidth,
		DimensionsHeight:  req.DimensionsHeight,
		CreatedBy:         req.CreatedBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Set default status if not provided
	if product.Status == "" {
		product.Status = entity.ProductStatusDraft
	}

	// Set default low stock threshold if tracking inventory but no threshold provided
	if product.TrackInventory && product.LowStockThreshold == nil {
		defaultThreshold := 10
		product.LowStockThreshold = &defaultThreshold
	}

	// Validate the product entity
	if err := product.Validate(); err != nil {
		uc.logger.Error("Product entity validation failed",
			"sku", req.SKU,
			"error", err)
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Create product in repository
	if err := uc.productRepo.Create(ctx, product); err != nil {
		uc.logger.Error("Failed to create product in repository",
			"sku", req.SKU,
			"error", err)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	uc.logger.Info("Product created successfully",
		"product_id", product.ID,
		"sku", product.SKU,
		"name", product.Name)

	return product, nil
}

// GetProduct retrieves a product by ID with permission validation
func (uc *ProductUseCase) GetProduct(ctx context.Context, productID uuid.UUID, include *repository.ProductInclude) (*entity.Product, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	product, err := uc.productRepo.GetByID(ctx, productID, include)
	if err != nil {
		uc.logger.Error("Failed to get product",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	uc.logger.Debug("Product retrieved successfully",
		"product_id", productID,
		"sku", product.SKU)

	return product, nil
}

// GetProductBySKU retrieves a product by SKU
func (uc *ProductUseCase) GetProductBySKU(ctx context.Context, sku string, include *repository.ProductInclude) (*entity.Product, error) {
	if sku == "" {
		return nil, fmt.Errorf("SKU cannot be empty")
	}

	product, err := uc.productRepo.GetBySKU(ctx, sku, include)
	if err != nil {
		uc.logger.Error("Failed to get product by SKU",
			"sku", sku,
			"error", err)
		return nil, fmt.Errorf("failed to get product by SKU: %w", err)
	}

	uc.logger.Debug("Product retrieved by SKU successfully",
		"sku", sku,
		"product_id", product.ID)

	return product, nil
}

// UpdateProduct updates an existing product with authorization check
func (uc *ProductUseCase) UpdateProduct(ctx context.Context, productID uuid.UUID, req UpdateProductRequest) (*entity.Product, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	// Get existing product
	existingProduct, err := uc.productRepo.GetByID(ctx, productID, nil)
	if err != nil {
		uc.logger.Error("Product not found for update",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Validate business rules for update
	if err := uc.validateUpdateProductRequest(ctx, req, existingProduct); err != nil {
		uc.logger.Error("Product update validation failed",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Apply updates to the product
	updatedProduct := uc.applyProductUpdates(existingProduct, req)
	updatedProduct.UpdatedAt = time.Now()

	// Validate the updated product entity
	if err := updatedProduct.Validate(); err != nil {
		uc.logger.Error("Updated product entity validation failed",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("product validation failed: %w", err)
	}

	// Update product in repository
	if err := uc.productRepo.Update(ctx, updatedProduct); err != nil {
		uc.logger.Error("Failed to update product in repository",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	uc.logger.Info("Product updated successfully",
		"product_id", productID,
		"sku", updatedProduct.SKU)

	return updatedProduct, nil
}

// ListProducts retrieves products with filtering and pagination
func (uc *ProductUseCase) ListProducts(ctx context.Context, req ListProductsRequest) ([]*entity.Product, int64, error) {
	// Validate request
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Build search filter
	filter := uc.buildProductFilter(req)

	// Get products from repository
	products, err := uc.productRepo.List(ctx, filter, req.Include)
	if err != nil {
		uc.logger.Error("Failed to list products",
			"filter", req.Filter,
			"error", err)
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	// Get total count for pagination
	totalCount, err := uc.productRepo.Count(ctx, filter)
	if err != nil {
		uc.logger.Error("Failed to count products",
			"filter", req.Filter,
			"error", err)
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	uc.logger.Debug("Products listed successfully",
		"count", len(products),
		"total", totalCount,
		"page", req.Page)

	return products, totalCount, nil
}

// SearchProducts searches products with full-text search
func (uc *ProductUseCase) SearchProducts(ctx context.Context, query string, req ListProductsRequest) ([]*entity.Product, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Build search filter
	filter := uc.buildProductFilter(req)

	// Search products from repository
	products, err := uc.productRepo.Search(ctx, query, filter, req.Include)
	if err != nil {
		uc.logger.Error("Failed to search products",
			"query", query,
			"filter", req.Filter,
			"error", err)
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	uc.logger.Debug("Products searched successfully",
		"query", query,
		"count", len(products))

	return products, nil
}

// DeleteProduct soft deletes a product with dependency check
func (uc *ProductUseCase) DeleteProduct(ctx context.Context, productID uuid.UUID) error {
	if productID == uuid.Nil {
		return fmt.Errorf("product ID cannot be empty")
	}

	// Check if product exists
	product, err := uc.productRepo.GetByID(ctx, productID, nil)
	if err != nil {
		uc.logger.Error("Product not found for deletion",
			"product_id", productID,
			"error", err)
		return fmt.Errorf("product not found: %w", err)
	}

	// Check for dependencies (variants, orders, etc.)
	if err := uc.validateProductDeletion(ctx, productID); err != nil {
		uc.logger.Error("Product deletion validation failed",
			"product_id", productID,
			"error", err)
		return fmt.Errorf("cannot delete product: %w", err)
	}

	// Soft delete the product
	if err := uc.productRepo.Delete(ctx, productID); err != nil {
		uc.logger.Error("Failed to delete product",
			"product_id", productID,
			"error", err)
		return fmt.Errorf("failed to delete product: %w", err)
	}

	uc.logger.Info("Product deleted successfully",
		"product_id", productID,
		"sku", product.SKU)

	return nil
}

// Helper methods for validation and business logic

func (uc *ProductUseCase) validateCreateProductRequest(ctx context.Context, req CreateProductRequest) error {
	// Create temporary product to use validation method
	tempProduct := &entity.Product{SKU: req.SKU}
	if err := tempProduct.ValidateSKU(); err != nil {
		return fmt.Errorf("invalid SKU format: %w", err)
	}

	// Validate pricing
	if req.SalePrice != nil && req.SalePrice.GreaterThanOrEqual(req.BasePrice) {
		return fmt.Errorf("sale price must be less than base price")
	}

	if req.CostPrice != nil && req.CostPrice.GreaterThan(req.BasePrice) {
		return fmt.Errorf("cost price cannot be greater than base price")
	}

	// Validate inventory settings
	if req.StockQuantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}

	if req.LowStockThreshold != nil && *req.LowStockThreshold < 0 {
		return fmt.Errorf("low stock threshold cannot be negative")
	}

	return nil
}

func (uc *ProductUseCase) validateUpdateProductRequest(ctx context.Context, req UpdateProductRequest, existing *entity.Product) error {
	// Validate pricing if provided
	currentBasePrice := existing.BasePrice
	if req.BasePrice != nil {
		currentBasePrice = *req.BasePrice
	}

	if req.SalePrice != nil && req.SalePrice.GreaterThanOrEqual(currentBasePrice) {
		return fmt.Errorf("sale price must be less than base price")
	}

	if req.CostPrice != nil && req.CostPrice.GreaterThan(currentBasePrice) {
		return fmt.Errorf("cost price cannot be greater than base price")
	}

	// Validate inventory settings
	if req.StockQuantity != nil && *req.StockQuantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}

	if req.LowStockThreshold != nil && *req.LowStockThreshold < 0 {
		return fmt.Errorf("low stock threshold cannot be negative")
	}

	return nil
}

func (uc *ProductUseCase) validateProductDeletion(ctx context.Context, productID uuid.UUID) error {
	// Check if product has variants
	variants, err := uc.variantRepo.GetByProduct(ctx, productID, nil)
	if err != nil {
		return fmt.Errorf("failed to check product variants: %w", err)
	}

	if len(variants) > 0 {
		return fmt.Errorf("cannot delete product with variants")
	}

	// Additional checks can be added here (e.g., orders, reservations)

	return nil
}

func (uc *ProductUseCase) applyProductUpdates(product *entity.Product, req UpdateProductRequest) *entity.Product {
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = req.Description
	}
	if req.CategoryID != nil {
		product.CategoryID = req.CategoryID
	}
	if req.Brand != nil {
		product.Brand = req.Brand
	}
	if req.Tags != nil {
		product.Tags = pq.StringArray(req.Tags)
	}
	if req.BasePrice != nil {
		product.BasePrice = *req.BasePrice
	}
	if req.SalePrice != nil {
		product.SalePrice = req.SalePrice
	}
	if req.CostPrice != nil {
		product.CostPrice = req.CostPrice
	}
	if req.TrackInventory != nil {
		product.TrackInventory = *req.TrackInventory
	}
	if req.StockQuantity != nil {
		product.StockQuantity = *req.StockQuantity
	}
	if req.LowStockThreshold != nil {
		product.LowStockThreshold = req.LowStockThreshold
	}
	if req.Status != nil {
		product.Status = *req.Status
	}
	if req.MetaTitle != nil {
		product.MetaTitle = req.MetaTitle
	}
	if req.MetaDescription != nil {
		product.MetaDescription = req.MetaDescription
	}
	if req.Slug != nil {
		product.Slug = req.Slug
	}
	if req.Weight != nil {
		product.Weight = req.Weight
	}
	if req.DimensionsLength != nil {
		product.DimensionsLength = req.DimensionsLength
	}
	if req.DimensionsWidth != nil {
		product.DimensionsWidth = req.DimensionsWidth
	}
	if req.DimensionsHeight != nil {
		product.DimensionsHeight = req.DimensionsHeight
	}

	return product
}

func (uc *ProductUseCase) buildProductFilter(req ListProductsRequest) *repository.ProductFilter {
	filter := &repository.ProductFilter{
		CategoryIDs:   req.Filter.CategoryIDs,
		Status:        req.Filter.Status,
		MinPrice:      req.Filter.MinPrice,
		MaxPrice:      req.Filter.MaxPrice,
		MinStock:      req.Filter.MinStock,
		MaxStock:      req.Filter.MaxStock,
		IsLowStock:    req.Filter.IsLowStock,
		TrackQuantity: req.Filter.TrackInventory,
		SearchQuery:   req.Filter.SearchQuery,
		CreatedAfter:  req.Filter.CreatedAfter,
		CreatedBefore: req.Filter.CreatedBefore,
		UpdatedAfter:  req.Filter.UpdatedAfter,
		UpdatedBefore: req.Filter.UpdatedBefore,
	}

	// Set pagination
	filter.Limit = req.PageSize
	filter.Offset = (req.Page - 1) * req.PageSize

	// Set sorting
	if req.SortBy != "" {
		filter.SortBy = req.SortBy
		if req.SortDesc {
			filter.SortOrder = "desc"
		} else {
			filter.SortOrder = "asc"
		}
	} else {
		filter.SortBy = "created_at"
		filter.SortOrder = "desc"
	}

	return filter
}
