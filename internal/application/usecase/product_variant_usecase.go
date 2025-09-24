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
	"github.com/shopspring/decimal"
)

// ProductVariantUseCase handles all product variant-related business operations
type ProductVariantUseCase struct {
	variantRepo       repository.ProductVariantRepository
	variantOptionRepo repository.ProductVariantOptionRepository
	productRepo       repository.ProductRepository
	logger            *slog.Logger
}

// NewProductVariantUseCase creates a new instance of ProductVariantUseCase
func NewProductVariantUseCase(
	variantRepo repository.ProductVariantRepository,
	variantOptionRepo repository.ProductVariantOptionRepository,
	productRepo repository.ProductRepository,
	logger *slog.Logger,
) *ProductVariantUseCase {
	return &ProductVariantUseCase{
		variantRepo:       variantRepo,
		variantOptionRepo: variantOptionRepo,
		productRepo:       productRepo,
		logger:            logger,
	}
}

// CreateVariantOptionRequest represents the data needed to create a variant option
type CreateVariantOptionRequest struct {
	ProductID   uuid.UUID `json:"product_id" validate:"required"`
	OptionName  string    `json:"option_name" validate:"required,min=1,max=100"`
	DisplayName *string   `json:"display_name" validate:"omitempty,max=200"`
	Values      []string  `json:"values" validate:"required,min=1,dive,required,max=100"`
	IsRequired  bool      `json:"is_required"`
	SortOrder   int       `json:"sort_order" validate:"min=0"`
}

// UpdateVariantOptionRequest represents the data needed to update a variant option
type UpdateVariantOptionRequest struct {
	OptionName  *string  `json:"option_name" validate:"omitempty,min=1,max=100"`
	DisplayName *string  `json:"display_name" validate:"omitempty,max=200"`
	Values      []string `json:"values" validate:"omitempty,min=1,dive,required,max=100"`
	IsRequired  *bool    `json:"is_required"`
	SortOrder   *int     `json:"sort_order" validate:"omitempty,min=0"`
}

// CreateVariantRequest represents the data needed to create a variant
type CreateVariantRequest struct {
	ProductID         uuid.UUID              `json:"product_id" validate:"required"`
	VariantName       *string                `json:"variant_name" validate:"omitempty,max=255"`
	VariantSKU        *string                `json:"variant_sku" validate:"omitempty,max=100"`
	Options           map[string]interface{} `json:"options" validate:"required"`
	Price             *decimal.Decimal       `json:"price" validate:"omitempty,gte=0"`
	CompareAtPrice    *decimal.Decimal       `json:"compare_at_price" validate:"omitempty,gte=0"`
	CostPrice         *decimal.Decimal       `json:"cost_price" validate:"omitempty,gte=0"`
	StockQuantity     int                    `json:"stock_quantity" validate:"min=0"`
	LowStockThreshold int                    `json:"low_stock_threshold" validate:"min=0"`
	TrackQuantity     bool                   `json:"track_quantity"`
	IsDefault         bool                   `json:"is_default"`
	Weight            *decimal.Decimal       `json:"weight" validate:"omitempty,gte=0"`
	Length            *decimal.Decimal       `json:"length" validate:"omitempty,gte=0"`
	Width             *decimal.Decimal       `json:"width" validate:"omitempty,gte=0"`
	Height            *decimal.Decimal       `json:"height" validate:"omitempty,gte=0"`
	IsActive          bool                   `json:"is_active"`
}

// UpdateVariantRequest represents the data needed to update a variant
type UpdateVariantRequest struct {
	VariantName       *string                `json:"variant_name" validate:"omitempty,max=255"`
	VariantSKU        *string                `json:"variant_sku" validate:"omitempty,max=100"`
	Options           map[string]interface{} `json:"options" validate:"omitempty"`
	Price             *decimal.Decimal       `json:"price" validate:"omitempty,gte=0"`
	CompareAtPrice    *decimal.Decimal       `json:"compare_at_price" validate:"omitempty,gte=0"`
	CostPrice         *decimal.Decimal       `json:"cost_price" validate:"omitempty,gte=0"`
	StockQuantity     *int                   `json:"stock_quantity" validate:"omitempty,min=0"`
	LowStockThreshold *int                   `json:"low_stock_threshold" validate:"omitempty,min=0"`
	TrackQuantity     *bool                  `json:"track_quantity"`
	Weight            *decimal.Decimal       `json:"weight" validate:"omitempty,gte=0"`
	Length            *decimal.Decimal       `json:"length" validate:"omitempty,gte=0"`
	Width             *decimal.Decimal       `json:"width" validate:"omitempty,gte=0"`
	Height            *decimal.Decimal       `json:"height" validate:"omitempty,gte=0"`
	IsActive          *bool                  `json:"is_active"`
}

// VariantCombinationRequest represents a request to generate variant combinations
type VariantCombinationRequest struct {
	ProductID uuid.UUID                     `json:"product_id" validate:"required"`
	Options   map[string][]string           `json:"options" validate:"required,min=1"`
	Defaults  *CreateVariantDefaultsRequest `json:"defaults" validate:"omitempty"`
}

// CreateVariantDefaultsRequest represents default values for generated variants
type CreateVariantDefaultsRequest struct {
	Price             *decimal.Decimal `json:"price" validate:"omitempty,gte=0"`
	CompareAtPrice    *decimal.Decimal `json:"compare_at_price" validate:"omitempty,gte=0"`
	CostPrice         *decimal.Decimal `json:"cost_price" validate:"omitempty,gte=0"`
	StockQuantity     int              `json:"stock_quantity" validate:"min=0"`
	LowStockThreshold int              `json:"low_stock_threshold" validate:"min=0"`
	TrackQuantity     bool             `json:"track_quantity"`
	Weight            *decimal.Decimal `json:"weight" validate:"omitempty,gte=0"`
	Length            *decimal.Decimal `json:"length" validate:"omitempty,gte=0"`
	Width             *decimal.Decimal `json:"width" validate:"omitempty,gte=0"`
	Height            *decimal.Decimal `json:"height" validate:"omitempty,gte=0"`
	IsActive          bool             `json:"is_active"`
}

// CreateVariantOption creates a new variant option for a product
func (uc *ProductVariantUseCase) CreateVariantOption(ctx context.Context, req CreateVariantOptionRequest) (*entity.ProductVariantOption, error) {
	// Validate product exists
	product, err := uc.productRepo.GetByID(ctx, req.ProductID, nil)
	if err != nil {
		uc.logger.Error("Product not found for variant option creation",
			"product_id", req.ProductID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Check if option name already exists for this product
	exists, err := uc.variantOptionRepo.IsOptionNameExists(ctx, req.ProductID, req.OptionName)
	if err != nil {
		uc.logger.Error("Failed to check option name existence",
			"product_id", req.ProductID,
			"option_name", req.OptionName,
			"error", err)
		return nil, fmt.Errorf("failed to check option name: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("option name '%s' already exists for this product", req.OptionName)
	}

	// Generate display name if not provided
	displayName := req.DisplayName
	if displayName == nil {
		name := uc.generateDisplayName(req.OptionName)
		displayName = &name
	}

	// Get next sort order if not provided
	sortOrder := req.SortOrder
	if sortOrder == 0 {
		nextOrder, err := uc.variantOptionRepo.GetNextSortOrder(ctx, req.ProductID)
		if err != nil {
			uc.logger.Warn("Failed to get next sort order, using default",
				"product_id", req.ProductID,
				"error", err)
			sortOrder = 1
		} else {
			sortOrder = nextOrder
		}
	}

	// Create variant option entity
	variantOption := &entity.ProductVariantOption{
		ID:           uuid.New(),
		ProductID:    req.ProductID,
		OptionName:   req.OptionName,
		DisplayName:  displayName,
		OptionValues: req.Values,
		IsRequired:   req.IsRequired,
		SortOrder:    sortOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Validate entity
	if err := variantOption.Validate(); err != nil {
		uc.logger.Error("Variant option validation failed",
			"product_id", req.ProductID,
			"option_name", req.OptionName,
			"error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create variant option in repository
	if err := uc.variantOptionRepo.Create(ctx, variantOption); err != nil {
		uc.logger.Error("Failed to create variant option in repository",
			"product_id", req.ProductID,
			"option_name", req.OptionName,
			"error", err)
		return nil, fmt.Errorf("failed to create variant option: %w", err)
	}

	uc.logger.Info("Variant option created successfully",
		"variant_option_id", variantOption.ID,
		"product_id", req.ProductID,
		"option_name", req.OptionName,
		"product_name", product.Name)

	return variantOption, nil
}

// UpdateVariantOption updates an existing variant option
func (uc *ProductVariantUseCase) UpdateVariantOption(ctx context.Context, optionID uuid.UUID, req UpdateVariantOptionRequest) (*entity.ProductVariantOption, error) {
	if optionID == uuid.Nil {
		return nil, fmt.Errorf("option ID cannot be empty")
	}

	// Get existing variant option
	existingOption, err := uc.variantOptionRepo.GetByID(ctx, optionID, nil)
	if err != nil {
		uc.logger.Error("Variant option not found for update",
			"option_id", optionID,
			"error", err)
		return nil, fmt.Errorf("variant option not found: %w", err)
	}

	// Check option name uniqueness if being changed
	if req.OptionName != nil && *req.OptionName != existingOption.OptionName {
		exists, err := uc.variantOptionRepo.IsOptionNameExistsExcluding(ctx, existingOption.ProductID, *req.OptionName, optionID)
		if err != nil {
			return nil, fmt.Errorf("failed to check option name uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("option name '%s' already exists for this product", *req.OptionName)
		}
	}

	// Apply updates
	updatedOption := uc.applyVariantOptionUpdates(existingOption, req)
	updatedOption.UpdatedAt = time.Now()

	// Validate updated entity
	if err := updatedOption.Validate(); err != nil {
		uc.logger.Error("Updated variant option validation failed",
			"option_id", optionID,
			"error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update variant option in repository
	if err := uc.variantOptionRepo.Update(ctx, updatedOption); err != nil {
		uc.logger.Error("Failed to update variant option in repository",
			"option_id", optionID,
			"error", err)
		return nil, fmt.Errorf("failed to update variant option: %w", err)
	}

	uc.logger.Info("Variant option updated successfully",
		"option_id", optionID,
		"product_id", updatedOption.ProductID,
		"option_name", updatedOption.OptionName)

	return updatedOption, nil
}

// DeleteVariantOption deletes a variant option
func (uc *ProductVariantUseCase) DeleteVariantOption(ctx context.Context, optionID uuid.UUID) error {
	if optionID == uuid.Nil {
		return fmt.Errorf("option ID cannot be empty")
	}

	// Get existing variant option
	existingOption, err := uc.variantOptionRepo.GetByID(ctx, optionID, nil)
	if err != nil {
		uc.logger.Error("Variant option not found for deletion",
			"option_id", optionID,
			"error", err)
		return fmt.Errorf("variant option not found: %w", err)
	}

	// Check if option is used by any variants
	variants, err := uc.variantRepo.GetByProduct(ctx, existingOption.ProductID, nil)
	if err != nil {
		uc.logger.Error("Failed to check variant usage",
			"option_id", optionID,
			"product_id", existingOption.ProductID,
			"error", err)
		return fmt.Errorf("failed to check variant usage: %w", err)
	}

	// Check if any variant uses this option
	for _, variant := range variants {
		if variant.Options != nil {
			if _, hasOption := variant.Options[existingOption.OptionName]; hasOption {
				return fmt.Errorf("cannot delete option '%s' as it is used by existing variants", existingOption.OptionName)
			}
		}
	}

	// Delete variant option
	if err := uc.variantOptionRepo.Delete(ctx, optionID); err != nil {
		uc.logger.Error("Failed to delete variant option",
			"option_id", optionID,
			"error", err)
		return fmt.Errorf("failed to delete variant option: %w", err)
	}

	uc.logger.Info("Variant option deleted successfully",
		"option_id", optionID,
		"product_id", existingOption.ProductID,
		"option_name", existingOption.OptionName)

	return nil
}

// CreateVariant creates a new product variant
func (uc *ProductVariantUseCase) CreateVariant(ctx context.Context, req CreateVariantRequest) (*entity.ProductVariant, error) {
	// Validate product exists
	product, err := uc.productRepo.GetByID(ctx, req.ProductID, nil)
	if err != nil {
		uc.logger.Error("Product not found for variant creation",
			"product_id", req.ProductID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Validate variant options against defined options
	if err := uc.validateVariantOptions(ctx, req.ProductID, req.Options); err != nil {
		uc.logger.Error("Variant options validation failed",
			"product_id", req.ProductID,
			"options", req.Options,
			"error", err)
		return nil, fmt.Errorf("invalid variant options: %w", err)
	}

	// Check for duplicate variant
	if err := uc.variantRepo.CheckDuplicateVariant(ctx, req.ProductID, req.Options, nil); err != nil {
		uc.logger.Error("Duplicate variant check failed",
			"product_id", req.ProductID,
			"options", req.Options,
			"error", err)
		return nil, fmt.Errorf("variant with these options already exists: %w", err)
	}

	// Generate SKU if not provided
	sku := req.VariantSKU
	if sku == nil || *sku == "" {
		generatedSKU, err := uc.variantRepo.GenerateVariantSKU(ctx, req.ProductID, req.Options)
		if err != nil {
			uc.logger.Warn("Failed to generate SKU, creating manual one",
				"product_id", req.ProductID,
				"error", err)
			manualSKU := uc.generateManualSKU(product.SKU, req.Options)
			sku = &manualSKU
		} else {
			sku = &generatedSKU
		}
	} else {
		// Check SKU uniqueness
		exists, err := uc.variantRepo.IsSkuExists(ctx, *sku)
		if err != nil {
			return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("SKU '%s' already exists", *sku)
		}
	}

	// Generate variant name if not provided
	variantName := req.VariantName
	if variantName == nil || *variantName == "" {
		generatedName := uc.generateVariantName(product.Name, req.Options)
		variantName = &generatedName
	}

	// Use product pricing if variant pricing not provided
	price := req.Price
	if price == nil {
		price = &product.BasePrice
	}

	comparePrice := req.CompareAtPrice
	if comparePrice == nil {
		comparePrice = product.SalePrice
	}

	costPrice := req.CostPrice
	if costPrice == nil {
		costPrice = product.CostPrice
	}

	// Handle default variant logic
	isDefault := req.IsDefault
	if isDefault {
		// Check if there's already a default variant
		existingDefault, err := uc.variantRepo.GetDefaultVariant(ctx, req.ProductID, nil)
		if err == nil && existingDefault != nil {
			// Unset existing default
			isDefault = false
			uc.logger.Warn("Default variant already exists, creating as non-default",
				"product_id", req.ProductID,
				"existing_default_id", existingDefault.ID)
		}
	}

	// Create variant entity
	variant := &entity.ProductVariant{
		ID:                uuid.New(),
		ProductID:         req.ProductID,
		VariantName:       *variantName,
		VariantSKU:        sku,
		Options:           req.Options,
		Price:             *price,
		CompareAtPrice:    comparePrice,
		CostPrice:         costPrice,
		StockQuantity:     req.StockQuantity,
		LowStockThreshold: &req.LowStockThreshold,
		TrackQuantity:     req.TrackQuantity,
		IsDefault:         isDefault,
		Weight:            req.Weight,
		Length:            req.Length,
		Width:             req.Width,
		Height:            req.Height,
		IsActive:          req.IsActive,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Validate entity
	if err := variant.Validate(); err != nil {
		uc.logger.Error("Variant validation failed",
			"product_id", req.ProductID,
			"sku", *sku,
			"error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create variant in repository
	if err := uc.variantRepo.Create(ctx, variant); err != nil {
		uc.logger.Error("Failed to create variant in repository",
			"product_id", req.ProductID,
			"sku", *sku,
			"error", err)
		return nil, fmt.Errorf("failed to create variant: %w", err)
	}

	// Set as default variant if requested and no existing default
	if req.IsDefault {
		if err := uc.variantRepo.SetDefaultVariant(ctx, req.ProductID, variant.ID); err != nil {
			uc.logger.Warn("Failed to set variant as default",
				"variant_id", variant.ID,
				"product_id", req.ProductID,
				"error", err)
		}
	}

	uc.logger.Info("Variant created successfully",
		"variant_id", variant.ID,
		"product_id", req.ProductID,
		"sku", *sku,
		"variant_name", *variantName,
		"product_name", product.Name,
		"is_default", isDefault)

	return variant, nil
}

// UpdateVariant updates an existing product variant
func (uc *ProductVariantUseCase) UpdateVariant(ctx context.Context, variantID uuid.UUID, req UpdateVariantRequest) (*entity.ProductVariant, error) {
	if variantID == uuid.Nil {
		return nil, fmt.Errorf("variant ID cannot be empty")
	}

	// Get existing variant
	existingVariant, err := uc.variantRepo.GetByID(ctx, variantID, nil)
	if err != nil {
		uc.logger.Error("Variant not found for update",
			"variant_id", variantID,
			"error", err)
		return nil, fmt.Errorf("variant not found: %w", err)
	}

	// Validate options if being changed
	if req.Options != nil {
		if err := uc.validateVariantOptions(ctx, existingVariant.ProductID, req.Options); err != nil {
			return nil, fmt.Errorf("invalid variant options: %w", err)
		}

		// Check for duplicate variant
		if err := uc.variantRepo.CheckDuplicateVariant(ctx, existingVariant.ProductID, req.Options, &variantID); err != nil {
			return nil, fmt.Errorf("variant with these options already exists: %w", err)
		}
	}

	// Check SKU uniqueness if being changed
	if req.VariantSKU != nil && (existingVariant.VariantSKU == nil || *req.VariantSKU != *existingVariant.VariantSKU) {
		exists, err := uc.variantRepo.IsSkuExistsExcluding(ctx, *req.VariantSKU, variantID)
		if err != nil {
			return nil, fmt.Errorf("failed to check SKU uniqueness: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("SKU '%s' already exists", *req.VariantSKU)
		}
	}

	// Apply updates
	updatedVariant := uc.applyVariantUpdates(existingVariant, req)
	updatedVariant.UpdatedAt = time.Now()

	// Regenerate variant name if options changed
	if req.Options != nil {
		product, err := uc.productRepo.GetByID(ctx, existingVariant.ProductID, nil)
		if err == nil {
			newName := uc.generateVariantName(product.Name, req.Options)
			updatedVariant.VariantName = newName
		}
	}

	// Validate updated entity
	if err := updatedVariant.Validate(); err != nil {
		uc.logger.Error("Updated variant validation failed",
			"variant_id", variantID,
			"error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Update variant in repository
	if err := uc.variantRepo.Update(ctx, updatedVariant); err != nil {
		uc.logger.Error("Failed to update variant in repository",
			"variant_id", variantID,
			"error", err)
		return nil, fmt.Errorf("failed to update variant: %w", err)
	}

	uc.logger.Info("Variant updated successfully",
		"variant_id", variantID,
		"product_id", updatedVariant.ProductID,
		"variant_sku", updatedVariant.VariantSKU)

	return updatedVariant, nil
}

// DeleteVariant deletes a product variant
func (uc *ProductVariantUseCase) DeleteVariant(ctx context.Context, variantID uuid.UUID) error {
	if variantID == uuid.Nil {
		return fmt.Errorf("variant ID cannot be empty")
	}

	// Get existing variant
	existingVariant, err := uc.variantRepo.GetByID(ctx, variantID, nil)
	if err != nil {
		uc.logger.Error("Variant not found for deletion",
			"variant_id", variantID,
			"error", err)
		return fmt.Errorf("variant not found: %w", err)
	}

	// Check if it's the default variant
	if existingVariant.IsDefault {
		// Count total variants for the product
		count, err := uc.variantRepo.CountByProduct(ctx, existingVariant.ProductID)
		if err != nil {
			return fmt.Errorf("failed to count variants: %w", err)
		}

		if count > 1 {
			return fmt.Errorf("cannot delete default variant when other variants exist. Please set another variant as default first")
		}
	}

	// TODO: Check for variant dependencies (orders, inventory transactions, etc.)
	// This would be implemented based on business requirements

	// Delete variant
	if err := uc.variantRepo.Delete(ctx, variantID); err != nil {
		uc.logger.Error("Failed to delete variant",
			"variant_id", variantID,
			"error", err)
		return fmt.Errorf("failed to delete variant: %w", err)
	}

	uc.logger.Info("Variant deleted successfully",
		"variant_id", variantID,
		"product_id", existingVariant.ProductID,
		"variant_sku", existingVariant.VariantSKU)

	return nil
}

// GenerateVariantCombinations generates all possible variant combinations
func (uc *ProductVariantUseCase) GenerateVariantCombinations(ctx context.Context, req VariantCombinationRequest) ([]*entity.ProductVariant, error) {
	// Validate product exists
	product, err := uc.productRepo.GetByID(ctx, req.ProductID, nil)
	if err != nil {
		uc.logger.Error("Product not found for variant generation",
			"product_id", req.ProductID,
			"error", err)
		return nil, fmt.Errorf("product not found: %w", err)
	}

	// Validate that all option names exist for the product
	for optionName := range req.Options {
		exists, err := uc.variantOptionRepo.IsOptionNameExists(ctx, req.ProductID, optionName)
		if err != nil {
			return nil, fmt.Errorf("failed to validate option '%s': %w", optionName, err)
		}
		if !exists {
			return nil, fmt.Errorf("option '%s' does not exist for this product", optionName)
		}
	}

	// Generate all combinations
	combinations := uc.generateOptionCombinations(req.Options)

	var variants []*entity.ProductVariant
	for i, combination := range combinations {
		// Check if variant already exists
		err := uc.variantRepo.CheckDuplicateVariant(ctx, req.ProductID, combination, nil)
		if err == nil {
			// Variant doesn't exist, create it
			createReq := uc.buildCreateVariantRequest(req.ProductID, combination, req.Defaults, i == 0)

			variant, err := uc.CreateVariant(ctx, createReq)
			if err != nil {
				uc.logger.Warn("Failed to create variant combination",
					"product_id", req.ProductID,
					"combination", combination,
					"error", err)
				continue
			}
			variants = append(variants, variant)
		}
	}

	uc.logger.Info("Variant combinations generated",
		"product_id", req.ProductID,
		"total_combinations", len(combinations),
		"created_variants", len(variants),
		"product_name", product.Name)

	return variants, nil
}

// Helper methods

func (uc *ProductVariantUseCase) generateDisplayName(optionName string) string {
	// Convert snake_case or kebab-case to Title Case
	words := strings.FieldsFunc(optionName, func(c rune) bool {
		return c == '_' || c == '-' || c == ' '
	})

	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(words, " ")
}

func (uc *ProductVariantUseCase) validateVariantOptions(ctx context.Context, productID uuid.UUID, options map[string]interface{}) error {
	if len(options) == 0 {
		return fmt.Errorf("variant must have at least one option")
	}

	// Get all defined options for the product
	definedOptions, err := uc.variantOptionRepo.GetByProduct(ctx, productID, nil)
	if err != nil {
		return fmt.Errorf("failed to get product options: %w", err)
	}

	// Check required options
	for _, definedOption := range definedOptions {
		if definedOption.IsRequired {
			if _, exists := options[definedOption.OptionName]; !exists {
				return fmt.Errorf("required option '%s' is missing", definedOption.OptionName)
			}
		}
	}

	// Validate option values
	for optionName, optionValue := range options {
		// Find the defined option
		var definedOption *entity.ProductVariantOption
		for _, opt := range definedOptions {
			if opt.OptionName == optionName {
				definedOption = opt
				break
			}
		}

		if definedOption == nil {
			return fmt.Errorf("option '%s' is not defined for this product", optionName)
		}

		// Validate option value
		valueStr := fmt.Sprintf("%v", optionValue)
		if !uc.isValueInArray(valueStr, definedOption.OptionValues) {
			return fmt.Errorf("value '%s' is not valid for option '%s'", valueStr, optionName)
		}
	}

	return nil
}

func (uc *ProductVariantUseCase) generateManualSKU(baseSKU string, options map[string]interface{}) string {
	sku := baseSKU
	for _, value := range options {
		valueStr := fmt.Sprintf("%v", value)
		// Take first 3 characters and convert to uppercase
		if len(valueStr) >= 3 {
			sku += "-" + strings.ToUpper(valueStr[:3])
		} else {
			sku += "-" + strings.ToUpper(valueStr)
		}
	}
	return sku
}

func (uc *ProductVariantUseCase) generateVariantName(productName string, options map[string]interface{}) string {
	name := productName
	var optionParts []string

	for _, optionValue := range options {
		optionParts = append(optionParts, fmt.Sprintf("%v", optionValue))
	}

	if len(optionParts) > 0 {
		name += " (" + strings.Join(optionParts, ", ") + ")"
	}

	return name
}

func (uc *ProductVariantUseCase) generateOptionCombinations(options map[string][]string) []map[string]interface{} {
	if len(options) == 0 {
		return nil
	}

	// Convert to slice for easier processing
	var names []string
	var values [][]string

	for name, vals := range options {
		names = append(names, name)
		values = append(values, vals)
	}

	// Generate combinations recursively
	var combinations []map[string]interface{}
	uc.generateCombinationsRecursive(names, values, 0, make(map[string]interface{}), &combinations)

	return combinations
}

func (uc *ProductVariantUseCase) generateCombinationsRecursive(names []string, values [][]string, index int, current map[string]interface{}, result *[]map[string]interface{}) {
	if index == len(names) {
		// Copy current combination
		combination := make(map[string]interface{})
		for k, v := range current {
			combination[k] = v
		}
		*result = append(*result, combination)
		return
	}

	// Try each value for current option
	for _, value := range values[index] {
		current[names[index]] = value
		uc.generateCombinationsRecursive(names, values, index+1, current, result)
	}
}

func (uc *ProductVariantUseCase) buildCreateVariantRequest(productID uuid.UUID, combination map[string]interface{}, defaults *CreateVariantDefaultsRequest, isFirst bool) CreateVariantRequest {
	req := CreateVariantRequest{
		ProductID:         productID,
		Options:           combination,
		StockQuantity:     0,
		LowStockThreshold: 1,
		TrackQuantity:     true,
		IsDefault:         isFirst, // First combination becomes default
		IsActive:          true,
	}

	if defaults != nil {
		req.Price = defaults.Price
		req.CompareAtPrice = defaults.CompareAtPrice
		req.CostPrice = defaults.CostPrice
		req.StockQuantity = defaults.StockQuantity
		req.LowStockThreshold = defaults.LowStockThreshold
		req.TrackQuantity = defaults.TrackQuantity
		req.Weight = defaults.Weight
		req.Length = defaults.Length
		req.Width = defaults.Width
		req.Height = defaults.Height
		req.IsActive = defaults.IsActive
	}

	return req
}

func (uc *ProductVariantUseCase) applyVariantOptionUpdates(option *entity.ProductVariantOption, req UpdateVariantOptionRequest) *entity.ProductVariantOption {
	if req.OptionName != nil {
		option.OptionName = *req.OptionName
	}
	if req.DisplayName != nil {
		option.DisplayName = req.DisplayName
	}
	if req.Values != nil {
		option.OptionValues = req.Values
	}
	if req.IsRequired != nil {
		option.IsRequired = *req.IsRequired
	}
	if req.SortOrder != nil {
		option.SortOrder = *req.SortOrder
	}
	return option
}

func (uc *ProductVariantUseCase) applyVariantUpdates(variant *entity.ProductVariant, req UpdateVariantRequest) *entity.ProductVariant {
	if req.VariantName != nil {
		variant.VariantName = *req.VariantName
	}
	if req.VariantSKU != nil {
		variant.VariantSKU = req.VariantSKU
	}
	if req.Options != nil {
		variant.Options = req.Options
	}
	if req.Price != nil {
		variant.Price = *req.Price
	}
	if req.CompareAtPrice != nil {
		variant.CompareAtPrice = req.CompareAtPrice
	}
	if req.CostPrice != nil {
		variant.CostPrice = req.CostPrice
	}
	if req.StockQuantity != nil {
		variant.StockQuantity = *req.StockQuantity
	}
	if req.LowStockThreshold != nil {
		variant.LowStockThreshold = req.LowStockThreshold
	}
	if req.TrackQuantity != nil {
		variant.TrackQuantity = *req.TrackQuantity
	}
	if req.Weight != nil {
		variant.Weight = req.Weight
	}
	if req.Length != nil {
		variant.Length = req.Length
	}
	if req.Width != nil {
		variant.Width = req.Width
	}
	if req.Height != nil {
		variant.Height = req.Height
	}
	if req.IsActive != nil {
		variant.IsActive = *req.IsActive
	}
	return variant
}

func (uc *ProductVariantUseCase) isValueInArray(value string, array []string) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}
