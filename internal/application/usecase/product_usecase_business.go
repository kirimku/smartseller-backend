package usecase

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// ActivateProduct activates a product with validation
func (uc *ProductUseCase) ActivateProduct(ctx context.Context, productID uuid.UUID) (*entity.Product, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, productID, nil)
	if err != nil {
		uc.logger.Error("Product not found for activation",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check if product can be activated
	if err := uc.validateProductActivation(ctx, product); err != nil {
		uc.logger.Error("Product activation validation failed",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("cannot activate product: %w", err)
	}

	// Update status to active
	product.Status = entity.ProductStatusActive

	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.Error("Failed to activate product",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("failed to activate product: %w", err)
	}

	uc.logger.Info("Product activated successfully",
		"product_id", productID,
		"sku", product.SKU)

	return product, nil
}

// DeactivateProduct deactivates a product with inventory check
func (uc *ProductUseCase) DeactivateProduct(ctx context.Context, productID uuid.UUID) (*entity.Product, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, productID, nil)
	if err != nil {
		uc.logger.Error("Product not found for deactivation",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check if product can be deactivated
	if err := uc.validateProductDeactivation(ctx, product); err != nil {
		uc.logger.Error("Product deactivation validation failed",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("cannot deactivate product: %w", err)
	}

	// Update status to inactive
	product.Status = entity.ProductStatusInactive

	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.Error("Failed to deactivate product",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("failed to deactivate product: %w", err)
	}

	uc.logger.Info("Product deactivated successfully",
		"product_id", productID,
		"sku", product.SKU)

	return product, nil
}

// ArchiveProduct archives a product with cleanup
func (uc *ProductUseCase) ArchiveProduct(ctx context.Context, productID uuid.UUID) (*entity.Product, error) {
	if productID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, productID, nil)
	if err != nil {
		uc.logger.Error("Product not found for archiving",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check if product can be archived
	if err := uc.validateProductArchiving(ctx, product); err != nil {
		uc.logger.Error("Product archiving validation failed",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("cannot archive product: %w", err)
	}

	// Update status to archived
	product.Status = entity.ProductStatusArchived

	if err := uc.productRepo.Update(ctx, product); err != nil {
		uc.logger.Error("Failed to archive product",
			"product_id", productID,
			"error", err)
		return nil, fmt.Errorf("failed to archive product: %w", err)
	}

	uc.logger.Info("Product archived successfully",
		"product_id", productID,
		"sku", product.SKU)

	return product, nil
}

// DuplicateProductRequest represents the data needed to duplicate a product
type DuplicateProductRequest struct {
	SourceProductID uuid.UUID             `json:"source_product_id" validate:"required"`
	NewSKU          string                `json:"new_sku" validate:"required,min=3,max=100"`
	NewName         *string               `json:"new_name" validate:"omitempty,min=1,max=255"`
	CopyVariants    bool                  `json:"copy_variants"`
	CopyImages      bool                  `json:"copy_images"`
	NewStatus       *entity.ProductStatus `json:"new_status" validate:"omitempty,oneof=draft active inactive archived"`
	CreatedBy       uuid.UUID             `json:"created_by" validate:"required"`
}

// DuplicateProduct duplicates a product with SKU generation
func (uc *ProductUseCase) DuplicateProduct(ctx context.Context, req DuplicateProductRequest) (*entity.Product, error) {
	// Validate request
	if req.SourceProductID == uuid.Nil {
		return nil, fmt.Errorf("source product ID cannot be empty")
	}

	// Check if new SKU already exists
	existing, err := uc.productRepo.GetBySKU(ctx, req.NewSKU, nil)
	if err == nil && existing != nil {
		uc.logger.Warn("Attempt to duplicate product with existing SKU",
			"new_sku", req.NewSKU,
			"existing_id", existing.ID)
		return nil, fmt.Errorf("product with SKU %s already exists", req.NewSKU)
	}

	// Get source product
	sourceProduct, err := uc.productRepo.GetByID(ctx, req.SourceProductID, &repository.ProductInclude{
		Category:       true,
		Images:         req.CopyImages,
		Variants:       req.CopyVariants,
		VariantOptions: req.CopyVariants,
	})
	if err != nil {
		uc.logger.Error("Source product not found for duplication",
			"source_product_id", req.SourceProductID,
			"error", err)
		return nil, fmt.Errorf("source product not found: %w", err)
	}

	// Create new product based on source
	newProduct := &entity.Product{
		ID:                uuid.New(),
		SKU:               req.NewSKU,
		Name:              sourceProduct.Name,
		Description:       sourceProduct.Description,
		CategoryID:        sourceProduct.CategoryID,
		Brand:             sourceProduct.Brand,
		Tags:              sourceProduct.Tags,
		BasePrice:         sourceProduct.BasePrice,
		SalePrice:         sourceProduct.SalePrice,
		CostPrice:         sourceProduct.CostPrice,
		TrackInventory:    sourceProduct.TrackInventory,
		StockQuantity:     0, // Reset stock quantity for new product
		LowStockThreshold: sourceProduct.LowStockThreshold,
		Status:            entity.ProductStatusDraft, // Default to draft
		MetaTitle:         sourceProduct.MetaTitle,
		MetaDescription:   sourceProduct.MetaDescription,
		Slug:              nil, // Will be regenerated based on new name
		Weight:            sourceProduct.Weight,
		DimensionsLength:  sourceProduct.DimensionsLength,
		DimensionsWidth:   sourceProduct.DimensionsWidth,
		DimensionsHeight:  sourceProduct.DimensionsHeight,
		CreatedBy:         req.CreatedBy,
	}

	// Override with new values if provided
	if req.NewName != nil {
		newProduct.Name = *req.NewName
	}
	if req.NewStatus != nil {
		newProduct.Status = *req.NewStatus
	}

	// Create the duplicated product
	if err := uc.productRepo.Create(ctx, newProduct); err != nil {
		uc.logger.Error("Failed to create duplicated product",
			"source_product_id", req.SourceProductID,
			"new_sku", req.NewSKU,
			"error", err)
		return nil, fmt.Errorf("failed to create duplicated product: %w", err)
	}

	// TODO: Copy variants and images if requested
	// This would be implemented when variant and image use cases are ready

	uc.logger.Info("Product duplicated successfully",
		"source_product_id", req.SourceProductID,
		"new_product_id", newProduct.ID,
		"new_sku", newProduct.SKU)

	return newProduct, nil
}

// BulkUpdateStatusRequest represents a bulk status update request
type BulkUpdateStatusRequest struct {
	ProductIDs []uuid.UUID          `json:"product_ids" validate:"required,min=1"`
	NewStatus  entity.ProductStatus `json:"new_status" validate:"required,oneof=draft active inactive archived"`
	UpdatedBy  uuid.UUID            `json:"updated_by" validate:"required"`
}

// BulkUpdateStatus updates the status of multiple products
func (uc *ProductUseCase) BulkUpdateStatus(ctx context.Context, req BulkUpdateStatusRequest) ([]*entity.Product, error) {
	if len(req.ProductIDs) == 0 {
		return nil, fmt.Errorf("product IDs cannot be empty")
	}

	// Get all products to validate they exist
	products, err := uc.productRepo.GetByIDs(ctx, req.ProductIDs, nil)
	if err != nil {
		uc.logger.Error("Failed to get products for bulk status update",
			"product_ids", req.ProductIDs,
			"error", err)
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	if len(products) != len(req.ProductIDs) {
		return nil, fmt.Errorf("some products were not found")
	}

	// Validate each product can have its status updated
	for _, product := range products {
		if err := uc.validateStatusTransition(ctx, product, req.NewStatus); err != nil {
			uc.logger.Error("Invalid status transition for product",
				"product_id", product.ID,
				"current_status", product.Status,
				"new_status", req.NewStatus,
				"error", err)
			return nil, fmt.Errorf("invalid status transition for product %s: %w", product.SKU, err)
		}
	}

	// Update all products
	for _, product := range products {
		product.Status = req.NewStatus
	}

	// Bulk update in repository
	if err := uc.productRepo.UpdateBatch(ctx, products); err != nil {
		uc.logger.Error("Failed to bulk update product statuses",
			"product_ids", req.ProductIDs,
			"new_status", req.NewStatus,
			"error", err)
		return nil, fmt.Errorf("failed to update product statuses: %w", err)
	}

	uc.logger.Info("Product statuses updated successfully",
		"count", len(products),
		"new_status", req.NewStatus)

	return products, nil
}

// Business validation helper methods

func (uc *ProductUseCase) validateProductActivation(ctx context.Context, product *entity.Product) error {
	// Check if product is already active
	if product.Status == entity.ProductStatusActive {
		return fmt.Errorf("product is already active")
	}

	// Check if product has required information for activation
	if product.BasePrice.IsZero() {
		return fmt.Errorf("product must have a base price to be activated")
	}

	// Additional business rules can be added here
	return nil
}

func (uc *ProductUseCase) validateProductDeactivation(ctx context.Context, product *entity.Product) error {
	// Check if product is already inactive
	if product.Status == entity.ProductStatusInactive {
		return fmt.Errorf("product is already inactive")
	}

	// Check if product is archived (cannot deactivate archived products)
	if product.Status == entity.ProductStatusArchived {
		return fmt.Errorf("cannot deactivate archived product")
	}

	// Additional business rules can be added here (e.g., check for pending orders)
	return nil
}

func (uc *ProductUseCase) validateProductArchiving(ctx context.Context, product *entity.Product) error {
	// Check if product is already archived
	if product.Status == entity.ProductStatusArchived {
		return fmt.Errorf("product is already archived")
	}

	// Additional business rules can be added here (e.g., check for recent sales)
	return nil
}

func (uc *ProductUseCase) validateStatusTransition(ctx context.Context, product *entity.Product, newStatus entity.ProductStatus) error {
	// Same status check
	if product.Status == newStatus {
		return fmt.Errorf("product is already in %s status", newStatus)
	}

	// Business rules for status transitions
	switch newStatus {
	case entity.ProductStatusActive:
		return uc.validateProductActivation(ctx, product)
	case entity.ProductStatusInactive:
		return uc.validateProductDeactivation(ctx, product)
	case entity.ProductStatusArchived:
		return uc.validateProductArchiving(ctx, product)
	case entity.ProductStatusDraft:
		// Draft status is usually allowed from any status
		return nil
	default:
		return fmt.Errorf("invalid product status: %s", newStatus)
	}
}
