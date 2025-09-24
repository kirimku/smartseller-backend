package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// StockUpdateRequest represents a stock update request
type StockUpdateRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`
	Quantity  int       `json:"quantity" validate:"required"`
	Reason    string    `json:"reason" validate:"required,max=255"`
	UpdatedBy uuid.UUID `json:"updated_by" validate:"required"`
}

// StockReservationRequest represents a stock reservation request
type StockReservationRequest struct {
	ProductID  uuid.UUID  `json:"product_id" validate:"required"`
	Quantity   int        `json:"quantity" validate:"required,min=1"`
	OrderID    *uuid.UUID `json:"order_id" validate:"omitempty"`
	ReservedBy uuid.UUID  `json:"reserved_by" validate:"required"`
	ExpiresAt  *time.Time `json:"expires_at" validate:"omitempty"`
}

// LowStockAlert represents a low stock alert
type LowStockAlert struct {
	ProductID         uuid.UUID  `json:"product_id"`
	ProductName       string     `json:"product_name"`
	SKU               string     `json:"sku"`
	CurrentStock      int        `json:"current_stock"`
	LowStockThreshold int        `json:"low_stock_threshold"`
	CategoryID        *uuid.UUID `json:"category_id"`
	LastUpdated       time.Time  `json:"last_updated"`
}

// UpdateStock updates product stock with reason tracking
func (uc *ProductUseCase) UpdateStock(ctx context.Context, req StockUpdateRequest) (*entity.Product, error) {
	if req.ProductID == uuid.Nil {
		return nil, fmt.Errorf("product ID cannot be empty")
	}

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, req.ProductID, nil)
	if err != nil {
		uc.logger.Error("Product not found for stock update",
			"product_id", req.ProductID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Validate stock update
	if err := uc.validateStockUpdate(ctx, product, req); err != nil {
		uc.logger.Error("Stock update validation failed",
			"product_id", req.ProductID,
			"quantity", req.Quantity,
			"error", err)
		return nil, fmt.Errorf("stock update validation failed: %w", err)
	}

	// Calculate new stock quantity
	newStockQuantity := product.StockQuantity + req.Quantity
	if newStockQuantity < 0 {
		return nil, fmt.Errorf("insufficient stock: current=%d, requested=%d", product.StockQuantity, -req.Quantity)
	}

	// Update stock using repository method
	if err := uc.productRepo.UpdateStock(ctx, req.ProductID, newStockQuantity); err != nil {
		uc.logger.Error("Failed to update stock in repository",
			"product_id", req.ProductID,
			"new_quantity", newStockQuantity,
			"error", err)
		return nil, fmt.Errorf("failed to update stock: %w", err)
	}

	// Get updated product
	updatedProduct, err := uc.productRepo.GetByID(ctx, req.ProductID, nil)
	if err != nil {
		uc.logger.Error("Failed to get updated product",
			"product_id", req.ProductID,
			"error", err)
		return nil, fmt.Errorf("failed to get updated product: %w", err)
	}

	uc.logger.Info("Stock updated successfully",
		"product_id", req.ProductID,
		"sku", product.SKU,
		"old_quantity", product.StockQuantity,
		"new_quantity", updatedProduct.StockQuantity,
		"change", req.Quantity,
		"reason", req.Reason)

	return updatedProduct, nil
}

// GetLowStockProducts retrieves products with low stock
func (uc *ProductUseCase) GetLowStockProducts(ctx context.Context, customThreshold *int) ([]*LowStockAlert, error) {
	threshold := 10 // Default threshold
	if customThreshold != nil {
		threshold = *customThreshold
	}

	// Get low stock products from repository
	products, err := uc.productRepo.GetLowStockProducts(ctx, threshold, &repository.ProductInclude{
		Category: true,
	})
	if err != nil {
		uc.logger.Error("Failed to get low stock products",
			"threshold", threshold,
			"error", err)
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}

	// Convert to alerts
	alerts := make([]*LowStockAlert, len(products))
	for i, product := range products {
		thresholdValue := threshold
		if product.LowStockThreshold != nil {
			thresholdValue = *product.LowStockThreshold
		}

		alerts[i] = &LowStockAlert{
			ProductID:         product.ID,
			ProductName:       product.Name,
			SKU:               product.SKU,
			CurrentStock:      product.StockQuantity,
			LowStockThreshold: thresholdValue,
			CategoryID:        product.CategoryID,
			LastUpdated:       product.UpdatedAt,
		}
	}

	uc.logger.Debug("Low stock products retrieved",
		"count", len(alerts),
		"threshold", threshold)

	return alerts, nil
}

// ReserveStock reserves stock for orders
func (uc *ProductUseCase) ReserveStock(ctx context.Context, req StockReservationRequest) error {
	if req.ProductID == uuid.Nil {
		return fmt.Errorf("product ID cannot be empty")
	}

	if req.Quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, req.ProductID, nil)
	if err != nil {
		uc.logger.Error("Product not found for stock reservation",
			"product_id", req.ProductID,
			"error", err)
		return fmt.Errorf("product not found: %w", err)
	}

	// Check if product tracks inventory
	if !product.TrackInventory {
		uc.logger.Debug("Stock reservation skipped for non-tracked product",
			"product_id", req.ProductID,
			"sku", product.SKU)
		return nil // No need to reserve stock for products that don't track inventory
	}

	// Check if sufficient stock is available
	if product.StockQuantity < req.Quantity {
		uc.logger.Warn("Insufficient stock for reservation",
			"product_id", req.ProductID,
			"sku", product.SKU,
			"available", product.StockQuantity,
			"requested", req.Quantity)
		return fmt.Errorf("insufficient stock: available=%d, requested=%d", product.StockQuantity, req.Quantity)
	}

	// Deduct stock using repository method
	if err := uc.productRepo.DeductStock(ctx, req.ProductID, req.Quantity); err != nil {
		uc.logger.Error("Failed to deduct stock for reservation",
			"product_id", req.ProductID,
			"quantity", req.Quantity,
			"error", err)
		return fmt.Errorf("failed to reserve stock: %w", err)
	}

	uc.logger.Info("Stock reserved successfully",
		"product_id", req.ProductID,
		"sku", product.SKU,
		"quantity", req.Quantity,
		"order_id", req.OrderID)

	// TODO: Store reservation record for tracking and expiration
	// This would require a separate reservations table/repository

	return nil
}

// ReleaseStock releases reserved stock (e.g., from cancelled orders)
func (uc *ProductUseCase) ReleaseStock(ctx context.Context, productID uuid.UUID, quantity int, reason string) error {
	if productID == uuid.Nil {
		return fmt.Errorf("product ID cannot be empty")
	}

	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, productID, nil)
	if err != nil {
		uc.logger.Error("Product not found for stock release",
			"product_id", productID,
			"error", err)
		return fmt.Errorf("product not found: %w", err)
	}

	// Check if product tracks inventory
	if !product.TrackInventory {
		uc.logger.Debug("Stock release skipped for non-tracked product",
			"product_id", productID,
			"sku", product.SKU)
		return nil // No need to release stock for products that don't track inventory
	}

	// Add stock back using repository method
	if err := uc.productRepo.RestockInventory(ctx, productID, quantity); err != nil {
		uc.logger.Error("Failed to release stock",
			"product_id", productID,
			"quantity", quantity,
			"error", err)
		return fmt.Errorf("failed to release stock: %w", err)
	}

	uc.logger.Info("Stock released successfully",
		"product_id", productID,
		"sku", product.SKU,
		"quantity", quantity,
		"reason", reason)

	return nil
}

// BulkStockUpdateRequest represents a bulk stock update request
type BulkStockUpdateRequest struct {
	Updates []StockUpdateRequest `json:"updates" validate:"required,min=1,dive"`
}

// BulkUpdateStock updates stock for multiple products
func (uc *ProductUseCase) BulkUpdateStock(ctx context.Context, req BulkStockUpdateRequest) ([]*entity.Product, error) {
	if len(req.Updates) == 0 {
		return nil, fmt.Errorf("updates cannot be empty")
	}

	var updatedProducts []*entity.Product
	var errors []error

	// Process each update
	for _, updateReq := range req.Updates {
		product, err := uc.UpdateStock(ctx, updateReq)
		if err != nil {
			uc.logger.Error("Failed to update stock in bulk operation",
				"product_id", updateReq.ProductID,
				"error", err)
			errors = append(errors, fmt.Errorf("product %s: %w", updateReq.ProductID, err))
			continue
		}
		updatedProducts = append(updatedProducts, product)
	}

	// If there were any errors, return them
	if len(errors) > 0 {
		return updatedProducts, fmt.Errorf("bulk stock update completed with %d errors", len(errors))
	}

	uc.logger.Info("Bulk stock update completed successfully",
		"count", len(updatedProducts))

	return updatedProducts, nil
}

// GetStockMovementHistory gets stock movement history for a product
// TODO: This would require a stock_movements table to track all changes
func (uc *ProductUseCase) GetStockMovementHistory(ctx context.Context, productID uuid.UUID, limit int) ([]interface{}, error) {
	// Placeholder implementation
	// In a real implementation, this would query a stock_movements table
	uc.logger.Debug("Stock movement history requested",
		"product_id", productID,
		"limit", limit)

	return []interface{}{}, fmt.Errorf("stock movement history not implemented yet")
}

// Validation helper methods

func (uc *ProductUseCase) validateStockUpdate(ctx context.Context, product *entity.Product, req StockUpdateRequest) error {
	// Check if product tracks inventory
	if !product.TrackInventory {
		return fmt.Errorf("product does not track inventory")
	}

	// Check if the reason is provided
	if req.Reason == "" {
		return fmt.Errorf("reason is required for stock updates")
	}

	// Validate quantity change
	if req.Quantity == 0 {
		return fmt.Errorf("quantity change cannot be zero")
	}

	// Check for negative stock (if reducing stock)
	if req.Quantity < 0 {
		newStock := product.StockQuantity + req.Quantity
		if newStock < 0 {
			return fmt.Errorf("insufficient stock for reduction: current=%d, reduction=%d",
				product.StockQuantity, -req.Quantity)
		}
	}

	return nil
}
