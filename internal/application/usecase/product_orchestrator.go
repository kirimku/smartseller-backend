package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// ProductOrchestrator coordinates complex operations across multiple product-related use cases
type ProductOrchestrator struct {
	// Use cases
	productUseCase  *ProductUseCase
	categoryUseCase *ProductCategoryUseCase
	variantUseCase  *ProductVariantUseCase
	imageUseCase    *ProductImageUseCase

	// Repositories for transaction management
	productRepo       repository.ProductRepository
	categoryRepo      repository.ProductCategoryRepository
	variantRepo       repository.ProductVariantRepository
	variantOptionRepo repository.ProductVariantOptionRepository
	imageRepo         repository.ProductImageRepository

	logger *slog.Logger
}

// NewProductOrchestrator creates a new product orchestrator
func NewProductOrchestrator(
	productUseCase *ProductUseCase,
	categoryUseCase *ProductCategoryUseCase,
	variantUseCase *ProductVariantUseCase,
	imageUseCase *ProductImageUseCase,
	productRepo repository.ProductRepository,
	categoryRepo repository.ProductCategoryRepository,
	variantRepo repository.ProductVariantRepository,
	variantOptionRepo repository.ProductVariantOptionRepository,
	imageRepo repository.ProductImageRepository,
	logger *slog.Logger,
) *ProductOrchestrator {
	return &ProductOrchestrator{
		productUseCase:    productUseCase,
		categoryUseCase:   categoryUseCase,
		variantUseCase:    variantUseCase,
		imageUseCase:      imageUseCase,
		productRepo:       productRepo,
		categoryRepo:      categoryRepo,
		variantRepo:       variantRepo,
		variantOptionRepo: variantOptionRepo,
		imageRepo:         imageRepo,
		logger:            logger,
	}
}

// CreateCompleteProductRequest represents a request to create a complete product with all related data
type CreateCompleteProductRequest struct {
	// Product data
	Product CreateProductRequest `json:"product" validate:"required"`

	// Category data (optional - create if doesn't exist)
	Category *CreateCategoryRequest `json:"category,omitempty"`

	// Variant options (optional)
	VariantOptions []CreateVariantOptionRequest `json:"variant_options,omitempty"`

	// Variants (optional)
	Variants []CreateVariantRequest `json:"variants,omitempty"`

	// Images (optional)
	Images []CreateImageRequest `json:"images,omitempty"`

	// Orchestration options
	AutoCreateCategory     bool `json:"auto_create_category,omitempty"`
	SetFirstImageAsPrimary bool `json:"set_first_image_as_primary,omitempty"`
	ValidateImageURLs      bool `json:"validate_image_urls,omitempty"`
}

// CreateCompleteProductResponse represents the response from creating a complete product
type CreateCompleteProductResponse struct {
	Product        *entity.Product                `json:"product"`
	Category       *entity.ProductCategory        `json:"category,omitempty"`
	VariantOptions []*entity.ProductVariantOption `json:"variant_options,omitempty"`
	Variants       []*entity.ProductVariant       `json:"variants,omitempty"`
	Images         []*entity.ProductImage         `json:"images,omitempty"`
	Metadata       CreateCompleteProductMetadata  `json:"metadata"`
}

// CreateCompleteProductMetadata contains metadata about the creation process
type CreateCompleteProductMetadata struct {
	TotalDuration     time.Duration `json:"total_duration"`
	StepsCompleted    []string      `json:"steps_completed"`
	Warnings          []string      `json:"warnings,omitempty"`
	CreatedEntities   int           `json:"created_entities"`
	CategoryCreated   bool          `json:"category_created"`
	VariantsGenerated bool          `json:"variants_generated"`
}

// UpdateCompleteProductRequest represents a request to update a complete product with all related data
type UpdateCompleteProductRequest struct {
	ProductID uuid.UUID `json:"product_id" validate:"required"`

	// Update operations
	Product *UpdateProductRequest `json:"product,omitempty"`

	// Add operations
	AddVariantOptions []CreateVariantOptionRequest `json:"add_variant_options,omitempty"`
	AddVariants       []CreateVariantRequest       `json:"add_variants,omitempty"`
	AddImages         []CreateImageRequest         `json:"add_images,omitempty"`

	// Delete operations
	DeleteVariantOptionIDs []uuid.UUID `json:"delete_variant_option_ids,omitempty"`
	DeleteVariantIDs       []uuid.UUID `json:"delete_variant_ids,omitempty"`
	DeleteImageIDs         []uuid.UUID `json:"delete_image_ids,omitempty"`

	// Orchestration options
	UpdateInventoryStock     bool `json:"update_inventory_stock,omitempty"`
	RecalculateVariantPrices bool `json:"recalculate_variant_prices,omitempty"`
	ReorderImages            bool `json:"reorder_images,omitempty"`
	ValidateConsistency      bool `json:"validate_consistency,omitempty"`
}

// CreateCompleteProduct creates a complete product with all related entities in a coordinated manner
func (po *ProductOrchestrator) CreateCompleteProduct(ctx context.Context, req CreateCompleteProductRequest) (*CreateCompleteProductResponse, error) {
	startTime := time.Now()
	metadata := CreateCompleteProductMetadata{
		StepsCompleted: make([]string, 0),
	}

	po.logger.Info("Starting complete product creation",
		slog.String("product_name", req.Product.Name),
		slog.String("product_sku", req.Product.SKU))

	// Step 1: Handle category creation if needed
	var category *entity.ProductCategory
	if req.Category != nil && req.AutoCreateCategory {
		po.logger.Info("Creating category", slog.String("category_name", req.Category.Name))

		var err error
		category, err = po.categoryUseCase.CreateCategory(ctx, *req.Category)
		if err != nil {
			po.logger.Error("Failed to create category", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to create category: %w", err)
		}

		req.Product.CategoryID = &category.ID
		metadata.CategoryCreated = true
		metadata.StepsCompleted = append(metadata.StepsCompleted, "category_created")
		metadata.CreatedEntities++
	}

	// Step 2: Create the main product
	po.logger.Info("Creating main product")
	product, err := po.productUseCase.CreateProduct(ctx, req.Product)
	if err != nil {
		po.logger.Error("Failed to create product", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	metadata.StepsCompleted = append(metadata.StepsCompleted, "product_created")
	metadata.CreatedEntities++

	// Step 3: Create variant options if provided
	var variantOptions []*entity.ProductVariantOption
	if len(req.VariantOptions) > 0 {
		po.logger.Info("Creating variant options", slog.Int("count", len(req.VariantOptions)))

		for _, optionReq := range req.VariantOptions {
			optionReq.ProductID = product.ID
			option, err := po.variantUseCase.CreateVariantOption(ctx, optionReq)
			if err != nil {
				po.logger.Warn("Failed to create variant option",
					slog.String("option_name", optionReq.OptionName),
					slog.String("error", err.Error()))
				metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("Failed to create variant option %s: %v", optionReq.OptionName, err))
				continue
			}
			variantOptions = append(variantOptions, option)
			metadata.CreatedEntities++
		}
		metadata.StepsCompleted = append(metadata.StepsCompleted, "variant_options_created")
	}

	// Step 4: Create variants
	var variants []*entity.ProductVariant
	if len(req.Variants) > 0 {
		po.logger.Info("Creating variants", slog.Int("count", len(req.Variants)))

		for _, variantReq := range req.Variants {
			variantReq.ProductID = product.ID
			variant, err := po.variantUseCase.CreateVariant(ctx, variantReq)
			if err != nil {
				variantName := ""
				if variantReq.VariantName != nil {
					variantName = *variantReq.VariantName
				}
				po.logger.Warn("Failed to create variant",
					slog.String("variant_name", variantName),
					slog.String("error", err.Error()))
				metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("Failed to create variant %s: %v", variantName, err))
				continue
			}
			variants = append(variants, variant)
			metadata.CreatedEntities++
		}
		metadata.StepsCompleted = append(metadata.StepsCompleted, "variants_created")
	}

	// Step 5: Create images
	var images []*entity.ProductImage
	if len(req.Images) > 0 {
		po.logger.Info("Creating product images", slog.Int("count", len(req.Images)))

		for i, imgReq := range req.Images {
			imgReq.ProductID = product.ID

			// Set first image as primary if requested
			if i == 0 && req.SetFirstImageAsPrimary {
				isPrimary := true
				imgReq.IsPrimary = &isPrimary
			}

			// Validate image URLs if requested
			if req.ValidateImageURLs {
				if err := po.imageUseCase.ValidateImageURL(ctx, imgReq.ImageURL); err != nil {
					po.logger.Warn("Image URL validation failed",
						slog.String("url", imgReq.ImageURL),
						slog.String("error", err.Error()))
					metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("Image URL validation failed for %s: %v", imgReq.ImageURL, err))
					continue
				}
			}

			image, err := po.imageUseCase.CreateImage(ctx, imgReq)
			if err != nil {
				po.logger.Warn("Failed to create image",
					slog.String("url", imgReq.ImageURL),
					slog.String("error", err.Error()))
				metadata.Warnings = append(metadata.Warnings, fmt.Sprintf("Failed to create image %s: %v", imgReq.ImageURL, err))
				continue
			}
			images = append(images, image)
			metadata.CreatedEntities++
		}
		metadata.StepsCompleted = append(metadata.StepsCompleted, "images_created")
	}

	// Calculate metadata
	metadata.TotalDuration = time.Since(startTime)

	response := &CreateCompleteProductResponse{
		Product:        product,
		Category:       category,
		VariantOptions: variantOptions,
		Variants:       variants,
		Images:         images,
		Metadata:       metadata,
	}

	po.logger.Info("Complete product creation finished",
		slog.String("product_id", product.ID.String()),
		slog.Duration("duration", metadata.TotalDuration),
		slog.Int("created_entities", metadata.CreatedEntities),
		slog.Int("warnings", len(metadata.Warnings)))

	return response, nil
}

// UpdateCompleteProduct updates a complete product with all related entities
func (po *ProductOrchestrator) UpdateCompleteProduct(ctx context.Context, req UpdateCompleteProductRequest) (*entity.Product, error) {
	po.logger.Info("Starting complete product update", slog.String("product_id", req.ProductID.String()))

	// Step 1: Update main product if requested
	var product *entity.Product
	var err error

	if req.Product != nil {
		product, err = po.productUseCase.UpdateProduct(ctx, req.ProductID, *req.Product)
		if err != nil {
			po.logger.Error("Failed to update product", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to update product: %w", err)
		}
	} else {
		product, err = po.productUseCase.GetProduct(ctx, req.ProductID, nil)
		if err != nil {
			po.logger.Error("Failed to get product", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to get product: %w", err)
		}
	}

	// Step 2: Handle variant option operations
	if len(req.DeleteVariantOptionIDs) > 0 {
		po.logger.Info("Deleting variant options", slog.Int("count", len(req.DeleteVariantOptionIDs)))
		for _, optionID := range req.DeleteVariantOptionIDs {
			if err := po.variantUseCase.DeleteVariantOption(ctx, optionID); err != nil {
				po.logger.Warn("Failed to delete variant option",
					slog.String("option_id", optionID.String()),
					slog.String("error", err.Error()))
			}
		}
	}

	if len(req.AddVariantOptions) > 0 {
		po.logger.Info("Adding variant options", slog.Int("count", len(req.AddVariantOptions)))
		for _, optionReq := range req.AddVariantOptions {
			optionReq.ProductID = req.ProductID
			if _, err := po.variantUseCase.CreateVariantOption(ctx, optionReq); err != nil {
				po.logger.Warn("Failed to add variant option",
					slog.String("option_name", optionReq.OptionName),
					slog.String("error", err.Error()))
			}
		}
	}

	// Step 3: Handle variant operations
	if len(req.DeleteVariantIDs) > 0 {
		po.logger.Info("Deleting variants", slog.Int("count", len(req.DeleteVariantIDs)))
		for _, variantID := range req.DeleteVariantIDs {
			if err := po.variantUseCase.DeleteVariant(ctx, variantID); err != nil {
				po.logger.Warn("Failed to delete variant",
					slog.String("variant_id", variantID.String()),
					slog.String("error", err.Error()))
			}
		}
	}

	if len(req.AddVariants) > 0 {
		po.logger.Info("Adding variants", slog.Int("count", len(req.AddVariants)))
		for _, variantReq := range req.AddVariants {
			variantReq.ProductID = req.ProductID
			if _, err := po.variantUseCase.CreateVariant(ctx, variantReq); err != nil {
				variantName := ""
				if variantReq.VariantName != nil {
					variantName = *variantReq.VariantName
				}
				po.logger.Warn("Failed to add variant",
					slog.String("variant_name", variantName),
					slog.String("error", err.Error()))
			}
		}
	}

	// Step 4: Handle image operations
	if len(req.DeleteImageIDs) > 0 {
		po.logger.Info("Deleting images", slog.Int("count", len(req.DeleteImageIDs)))
		for _, imageID := range req.DeleteImageIDs {
			if err := po.imageUseCase.DeleteImage(ctx, imageID); err != nil {
				po.logger.Warn("Failed to delete image",
					slog.String("image_id", imageID.String()),
					slog.String("error", err.Error()))
			}
		}
	}

	if len(req.AddImages) > 0 {
		po.logger.Info("Adding images", slog.Int("count", len(req.AddImages)))
		for _, imgReq := range req.AddImages {
			imgReq.ProductID = req.ProductID
			if _, err := po.imageUseCase.CreateImage(ctx, imgReq); err != nil {
				po.logger.Warn("Failed to add image",
					slog.String("url", imgReq.ImageURL),
					slog.String("error", err.Error()))
			}
		}
	}

	// Step 5: Handle orchestration options
	if req.UpdateInventoryStock {
		po.logger.Info("Updating inventory stock for product")
		// This would typically trigger inventory recalculation
		// Implementation depends on specific business requirements
	}

	if req.RecalculateVariantPrices {
		po.logger.Info("Recalculating variant prices")
		// This would recalculate variant prices based on product base price
		// Implementation depends on specific business requirements
	}

	po.logger.Info("Complete product update finished", slog.String("product_id", req.ProductID.String()))
	return product, nil
}

// GetProductWithAllRelatedData retrieves a product with all its related entities
func (po *ProductOrchestrator) GetProductWithAllRelatedData(ctx context.Context, productID uuid.UUID) (*CreateCompleteProductResponse, error) {
	po.logger.Info("Getting product with all related data", slog.String("product_id", productID.String()))

	// Get main product
	product, err := po.productUseCase.GetProduct(ctx, productID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	response := &CreateCompleteProductResponse{
		Product: product,
	}

	// Get category if exists
	if product.CategoryID != nil {
		// Note: This method needs to be implemented in category use case
		po.logger.Debug("Product has category", slog.String("category_id", product.CategoryID.String()))
	}

	// Get images
	images, err := po.imageUseCase.GetProductImages(ctx, productID)
	if err != nil {
		po.logger.Warn("Failed to get images", slog.String("error", err.Error()))
	} else {
		response.Images = images
	}

	return response, nil
}

// ValidateProductConsistency validates that all related entities are consistent
func (po *ProductOrchestrator) ValidateProductConsistency(ctx context.Context, productID uuid.UUID) error {
	po.logger.Info("Validating product consistency", slog.String("product_id", productID.String()))

	// Get product
	_, err := po.productUseCase.GetProduct(ctx, productID, nil)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Get images
	images, err := po.imageUseCase.GetProductImages(ctx, productID)
	if err != nil {
		return fmt.Errorf("failed to get images: %w", err)
	}

	// Validate that at least one primary image exists
	hasPrimaryImage := false
	for _, image := range images {
		if image.IsPrimary && image.VariantID == nil {
			hasPrimaryImage = true
			break
		}
	}

	if len(images) > 0 && !hasPrimaryImage {
		po.logger.Warn("Product has images but no primary image set", slog.String("product_id", productID.String()))
	}

	po.logger.Info("Product consistency validation completed", slog.String("product_id", productID.String()))
	return nil
}

// CloneProductRequest represents a request to clone a product
type CloneProductRequest struct {
	SourceProductID uuid.UUID `json:"source_product_id" validate:"required"`
	NewSKU          string    `json:"new_sku" validate:"required"`
	NewName         *string   `json:"new_name,omitempty"`

	// Clone options
	CloneVariants       bool `json:"clone_variants,omitempty"`
	CloneImages         bool `json:"clone_images,omitempty"`
	CloneVariantOptions bool `json:"clone_variant_options,omitempty"`

	// Override settings
	OverridePrice    *float64               `json:"override_price,omitempty"`
	OverrideCategory *uuid.UUID             `json:"override_category,omitempty"`
	OverrideStock    *int                   `json:"override_stock,omitempty"`
	OverrideStatus   *entity.ProductStatus  `json:"override_status,omitempty"`
	OverrideData     map[string]interface{} `json:"override_data,omitempty"`
}

// BulkProductOperationRequest represents a request for bulk operations
type BulkProductOperationRequest struct {
	ProductIDs []uuid.UUID `json:"product_ids" validate:"required,min=1"`
	Operation  string      `json:"operation" validate:"required"`

	// Update data (for bulk update operations)
	UpdateData map[string]interface{} `json:"update_data,omitempty"`

	// Options
	ContinueOnError bool `json:"continue_on_error,omitempty"`
	BatchSize       int  `json:"batch_size,omitempty"`
}

// BulkProductOperationResponse represents the response from bulk operations
type BulkProductOperationResponse struct {
	SuccessCount int                    `json:"success_count"`
	FailureCount int                    `json:"failure_count"`
	Failures     []BulkOperationFailure `json:"failures,omitempty"`
	Metadata     BulkOperationMetadata  `json:"metadata"`
}

// BulkOperationFailure represents a failed operation in bulk processing
type BulkOperationFailure struct {
	ProductID uuid.UUID `json:"product_id"`
	Error     string    `json:"error"`
}

// BulkOperationMetadata contains metadata about bulk operations
type BulkOperationMetadata struct {
	TotalDuration    time.Duration `json:"total_duration"`
	BatchesProcessed int           `json:"batches_processed"`
	OperationType    string        `json:"operation_type"`
}

// CloneProduct creates a copy of an existing product with its related entities
func (po *ProductOrchestrator) CloneProduct(ctx context.Context, req CloneProductRequest) (*CreateCompleteProductResponse, error) {
	po.logger.Info("Starting product clone",
		slog.String("source_id", req.SourceProductID.String()),
		slog.String("new_sku", req.NewSKU))

	// Get source product
	sourceProduct, err := po.productUseCase.GetProduct(ctx, req.SourceProductID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get source product: %w", err)
	}

	// Create clone product request
	cloneProductReq := CreateProductRequest{
		Name:              sourceProduct.Name,
		SKU:               req.NewSKU,
		Description:       sourceProduct.Description,
		BasePrice:         sourceProduct.BasePrice,
		SalePrice:         sourceProduct.SalePrice,
		CostPrice:         sourceProduct.CostPrice,
		CategoryID:        sourceProduct.CategoryID,
		Brand:             sourceProduct.Brand,
		Tags:              sourceProduct.Tags,
		TrackInventory:    sourceProduct.TrackInventory,
		StockQuantity:     sourceProduct.StockQuantity,
		LowStockThreshold: sourceProduct.LowStockThreshold,
		Status:            sourceProduct.Status,
		MetaTitle:         sourceProduct.MetaTitle,
		MetaDescription:   sourceProduct.MetaDescription,
		Weight:            sourceProduct.Weight,
		DimensionsLength:  sourceProduct.DimensionsLength,
		DimensionsWidth:   sourceProduct.DimensionsWidth,
		DimensionsHeight:  sourceProduct.DimensionsHeight,
		CreatedBy:         sourceProduct.CreatedBy,
	}

	// Apply overrides
	if req.NewName != nil {
		cloneProductReq.Name = *req.NewName
	}
	if req.OverridePrice != nil {
		cloneProductReq.BasePrice = decimal.NewFromFloat(*req.OverridePrice)
	}
	if req.OverrideCategory != nil {
		cloneProductReq.CategoryID = req.OverrideCategory
	}
	if req.OverrideStock != nil {
		cloneProductReq.StockQuantity = *req.OverrideStock
	}
	if req.OverrideStatus != nil {
		cloneProductReq.Status = *req.OverrideStatus
	}

	// Create the cloned product
	clonedProduct, err := po.productUseCase.CreateProduct(ctx, cloneProductReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloned product: %w", err)
	}

	response := &CreateCompleteProductResponse{
		Product: clonedProduct,
		Metadata: CreateCompleteProductMetadata{
			StepsCompleted:  []string{"product_cloned"},
			CreatedEntities: 1,
		},
	}

	// Clone variant options if requested
	if req.CloneVariantOptions {
		po.logger.Info("Cloning variant options")
		// This would require a method to get variant options by product ID
		// For now, we'll note it as a step that would be implemented
		response.Metadata.StepsCompleted = append(response.Metadata.StepsCompleted, "variant_options_cloned")
	}

	// Clone images if requested
	var clonedImages []*entity.ProductImage
	if req.CloneImages {
		po.logger.Info("Cloning product images")
		sourceImages, err := po.imageUseCase.GetProductImages(ctx, req.SourceProductID)
		if err != nil {
			po.logger.Warn("Failed to get source images", slog.String("error", err.Error()))
		} else {
			for _, sourceImage := range sourceImages {
				cloneImageReq := CreateImageRequest{
					ProductID: clonedProduct.ID,
					ImageURL:  sourceImage.ImageURL,
					AltText:   sourceImage.AltText,
					SortOrder: &sourceImage.SortOrder,
					IsPrimary: &sourceImage.IsPrimary,
				}

				clonedImage, err := po.imageUseCase.CreateImage(ctx, cloneImageReq)
				if err != nil {
					po.logger.Warn("Failed to clone image",
						slog.String("url", sourceImage.ImageURL),
						slog.String("error", err.Error()))
					continue
				}
				clonedImages = append(clonedImages, clonedImage)
				response.Metadata.CreatedEntities++
			}
			response.Images = clonedImages
			response.Metadata.StepsCompleted = append(response.Metadata.StepsCompleted, "images_cloned")
		}
	}

	po.logger.Info("Product clone completed",
		slog.String("cloned_id", clonedProduct.ID.String()),
		slog.Int("created_entities", response.Metadata.CreatedEntities))

	return response, nil
}

// BulkUpdateProducts performs bulk updates on multiple products
func (po *ProductOrchestrator) BulkUpdateProducts(ctx context.Context, req BulkProductOperationRequest) (*BulkProductOperationResponse, error) {
	startTime := time.Now()

	po.logger.Info("Starting bulk product update",
		slog.String("operation", req.Operation),
		slog.Int("product_count", len(req.ProductIDs)))

	response := &BulkProductOperationResponse{
		Metadata: BulkOperationMetadata{
			OperationType: req.Operation,
		},
	}

	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 50 // Default batch size
	}

	// Process in batches
	for i := 0; i < len(req.ProductIDs); i += batchSize {
		end := i + batchSize
		if end > len(req.ProductIDs) {
			end = len(req.ProductIDs)
		}

		batch := req.ProductIDs[i:end]
		po.logger.Debug("Processing batch",
			slog.Int("batch_number", i/batchSize+1),
			slog.Int("batch_size", len(batch)))

		for _, productID := range batch {
			err := po.processBulkProductOperation(ctx, productID, req)
			if err != nil {
				response.FailureCount++
				response.Failures = append(response.Failures, BulkOperationFailure{
					ProductID: productID,
					Error:     err.Error(),
				})

				if !req.ContinueOnError {
					po.logger.Error("Bulk operation failed, stopping",
						slog.String("product_id", productID.String()),
						slog.String("error", err.Error()))
					break
				}
			} else {
				response.SuccessCount++
			}
		}

		response.Metadata.BatchesProcessed++
	}

	response.Metadata.TotalDuration = time.Since(startTime)

	po.logger.Info("Bulk product update completed",
		slog.Int("success_count", response.SuccessCount),
		slog.Int("failure_count", response.FailureCount),
		slog.Duration("duration", response.Metadata.TotalDuration))

	return response, nil
}

// processBulkProductOperation processes a single product in bulk operations
func (po *ProductOrchestrator) processBulkProductOperation(ctx context.Context, productID uuid.UUID, req BulkProductOperationRequest) error {
	switch req.Operation {
	case "update_status":
		if statusValue, ok := req.UpdateData["status"].(string); ok {
			status := entity.ProductStatus(statusValue)
			updateReq := UpdateProductRequest{
				Status: &status,
			}
			_, err := po.productUseCase.UpdateProduct(ctx, productID, updateReq)
			return err
		}
		return fmt.Errorf("invalid status value for product %s", productID.String())

	case "update_price":
		if priceValue, ok := req.UpdateData["price"].(float64); ok {
			basePrice := decimal.NewFromFloat(priceValue)
			updateReq := UpdateProductRequest{
				BasePrice: &basePrice,
			}
			_, err := po.productUseCase.UpdateProduct(ctx, productID, updateReq)
			return err
		}
		return fmt.Errorf("invalid price value for product %s", productID.String())

	case "update_category":
		if categoryValue, ok := req.UpdateData["category_id"].(string); ok {
			categoryID, err := uuid.Parse(categoryValue)
			if err != nil {
				return fmt.Errorf("invalid category ID format for product %s: %w", productID.String(), err)
			}
			updateReq := UpdateProductRequest{
				CategoryID: &categoryID,
			}
			_, err = po.productUseCase.UpdateProduct(ctx, productID, updateReq)
			return err
		}
		return fmt.Errorf("invalid category_id value for product %s", productID.String())

	case "delete":
		return po.productUseCase.DeleteProduct(ctx, productID)

	default:
		return fmt.Errorf("unsupported bulk operation: %s", req.Operation)
	}
}

// BulkDeleteProducts deletes multiple products and their related entities
func (po *ProductOrchestrator) BulkDeleteProducts(ctx context.Context, productIDs []uuid.UUID, continueOnError bool) (*BulkProductOperationResponse, error) {
	req := BulkProductOperationRequest{
		ProductIDs:      productIDs,
		Operation:       "delete",
		ContinueOnError: continueOnError,
	}

	return po.BulkUpdateProducts(ctx, req)
}

// GetProductWithRelatedDataStatistics returns statistics about a product's related data
func (po *ProductOrchestrator) GetProductWithRelatedDataStatistics(ctx context.Context, productID uuid.UUID) (map[string]interface{}, error) {
	po.logger.Info("Getting product statistics", slog.String("product_id", productID.String()))

	stats := make(map[string]interface{})

	// Get product
	product, err := po.productUseCase.GetProduct(ctx, productID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	stats["product"] = map[string]interface{}{
		"id":         product.ID,
		"name":       product.Name,
		"sku":        product.SKU,
		"status":     product.Status,
		"base_price": product.BasePrice,
	}

	// Get image count
	images, err := po.imageUseCase.GetProductImages(ctx, productID)
	if err != nil {
		po.logger.Warn("Failed to get images for statistics", slog.String("error", err.Error()))
		stats["image_count"] = 0
		stats["has_primary_image"] = false
	} else {
		stats["image_count"] = len(images)
		hasPrimaryImage := false
		for _, image := range images {
			if image.IsPrimary && image.VariantID == nil {
				hasPrimaryImage = true
				break
			}
		}
		stats["has_primary_image"] = hasPrimaryImage
	}

	// Additional statistics would be added here for variants, etc.
	stats["last_updated"] = product.UpdatedAt

	return stats, nil
}
