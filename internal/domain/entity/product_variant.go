package entity

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ProductVariant represents a specific combination of variant options for a product
// For example: Red/Large T-Shirt, Blue/Medium Jeans
type ProductVariant struct {
	// Primary identification
	ID        uuid.UUID `json:"id" db:"id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`

	// Variant identification
	VariantName string                 `json:"variant_name" db:"variant_name"` // Auto-generated or manual
	VariantSKU  *string                `json:"variant_sku" db:"variant_sku"`   // Optional specific SKU
	Options     map[string]interface{} `json:"options" db:"options"`           // JSONB: {"Color": "Red", "Size": "Large"}

	// Pricing
	Price          decimal.Decimal  `json:"price" db:"price"`                       // Override price
	CompareAtPrice *decimal.Decimal `json:"compare_at_price" db:"compare_at_price"` // MSRP/Original price
	CostPrice      *decimal.Decimal `json:"cost_price" db:"cost_price"`             // Cost for profit calculation

	// Inventory management
	StockQuantity     int  `json:"stock_quantity" db:"stock_quantity"`
	LowStockThreshold *int `json:"low_stock_threshold" db:"low_stock_threshold"`
	TrackQuantity     bool `json:"track_quantity" db:"track_quantity"`

	// Physical properties
	Weight *decimal.Decimal `json:"weight" db:"weight"`
	Length *decimal.Decimal `json:"length" db:"length"`
	Width  *decimal.Decimal `json:"width" db:"width"`
	Height *decimal.Decimal `json:"height" db:"height"`

	// Status and availability
	IsActive  bool `json:"is_active" db:"is_active"`
	IsDefault bool `json:"is_default" db:"is_default"`
	Position  *int `json:"position" db:"position"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields
	IsLowStock     bool             `json:"is_low_stock" db:"-"`
	ProfitMargin   *decimal.Decimal `json:"profit_margin" db:"-"`
	ProfitAmount   *decimal.Decimal `json:"profit_amount" db:"-"`
	FormattedPrice string           `json:"formatted_price" db:"-"`
	OptionCount    int              `json:"option_count" db:"-"`
}

// NewProductVariant creates a new product variant
func NewProductVariant(productID uuid.UUID, options map[string]interface{}, price decimal.Decimal) *ProductVariant {
	variant := &ProductVariant{
		ID:            uuid.New(),
		ProductID:     productID,
		Options:       options,
		Price:         price,
		StockQuantity: 0,
		TrackQuantity: true,
		IsActive:      true,
		IsDefault:     false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Auto-generate variant name
	variant.GenerateVariantName()

	return variant
}

// ValidateOptions validates the variant options
func (pv *ProductVariant) ValidateOptions() error {
	if len(pv.Options) == 0 {
		return fmt.Errorf("at least one option is required")
	}

	if len(pv.Options) > 10 {
		return fmt.Errorf("cannot have more than 10 options")
	}

	// Validate each option
	for optionName, optionValue := range pv.Options {
		// Validate option name
		if optionName == "" {
			return fmt.Errorf("option name cannot be empty")
		}

		if len(optionName) > 100 {
			return fmt.Errorf("option name '%s' cannot exceed 100 characters", optionName)
		}

		// Validate option value
		if optionValue == nil {
			return fmt.Errorf("option value for '%s' cannot be null", optionName)
		}

		// Convert to string for validation
		var valueStr string
		switch v := optionValue.(type) {
		case string:
			valueStr = v
		case float64, int, int64:
			valueStr = fmt.Sprintf("%v", v)
		default:
			// Try to convert to JSON string
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return fmt.Errorf("option value for '%s' is not a valid type", optionName)
			}
			valueStr = string(jsonBytes)
		}

		if strings.TrimSpace(valueStr) == "" {
			return fmt.Errorf("option value for '%s' cannot be empty", optionName)
		}

		if len(valueStr) > 255 {
			return fmt.Errorf("option value for '%s' cannot exceed 255 characters", optionName)
		}
	}

	return nil
}

// ValidatePricing validates all pricing-related fields
func (pv *ProductVariant) ValidatePricing() error {
	// Price must be positive
	if pv.Price.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("price must be greater than zero")
	}

	// Price cannot exceed maximum reasonable amount
	maxPrice := decimal.NewFromInt(999999999) // 999,999,999
	if pv.Price.GreaterThan(maxPrice) {
		return fmt.Errorf("price cannot exceed 999,999,999")
	}

	// Compare at price validation
	if pv.CompareAtPrice != nil {
		if pv.CompareAtPrice.LessThanOrEqual(decimal.Zero) {
			return fmt.Errorf("compare at price must be greater than zero")
		}

		if pv.CompareAtPrice.GreaterThan(maxPrice) {
			return fmt.Errorf("compare at price cannot exceed 999,999,999")
		}

		// Compare at price should typically be higher than selling price
		if pv.CompareAtPrice.LessThan(pv.Price) {
			return fmt.Errorf("compare at price should be higher than selling price")
		}
	}

	// Cost price validation
	if pv.CostPrice != nil {
		if pv.CostPrice.LessThan(decimal.Zero) {
			return fmt.Errorf("cost price cannot be negative")
		}

		if pv.CostPrice.GreaterThan(maxPrice) {
			return fmt.Errorf("cost price cannot exceed 999,999,999")
		}

		// Cost should typically be less than selling price for profit
		if pv.CostPrice.GreaterThan(pv.Price) {
			// This is a warning, not an error - some products may be sold at a loss
			// We'll handle this in business logic layer
		}
	}

	return nil
}

// ValidateInventory validates inventory-related fields
func (pv *ProductVariant) ValidateInventory() error {
	// Stock quantity validation
	if pv.StockQuantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}

	// Maximum reasonable stock quantity
	if pv.StockQuantity > 1000000000 {
		return fmt.Errorf("stock quantity cannot exceed 1,000,000,000")
	}

	// Low stock threshold validation
	if pv.LowStockThreshold != nil {
		if *pv.LowStockThreshold < 0 {
			return fmt.Errorf("low stock threshold cannot be negative")
		}

		if *pv.LowStockThreshold > pv.StockQuantity {
			// This is a warning, not an error - threshold can be higher than current stock
		}
	}

	return nil
}

// ValidateDimensions validates physical dimension fields
func (pv *ProductVariant) ValidateDimensions() error {
	dimensions := map[string]*decimal.Decimal{
		"weight": pv.Weight,
		"length": pv.Length,
		"width":  pv.Width,
		"height": pv.Height,
	}

	for name, dimension := range dimensions {
		if dimension != nil {
			if dimension.LessThan(decimal.Zero) {
				return fmt.Errorf("%s cannot be negative", name)
			}

			// Maximum reasonable dimensions (in centimeters/grams)
			maxDimension := decimal.NewFromInt(100000) // 100,000 cm or 100,000 grams
			if dimension.GreaterThan(maxDimension) {
				return fmt.Errorf("%s cannot exceed 100,000 (cm or grams)", name)
			}
		}
	}

	return nil
}

// Validate performs comprehensive validation of the variant
func (pv *ProductVariant) Validate() error {
	// Validate variant name
	if pv.VariantName == "" {
		return fmt.Errorf("variant name is required")
	}

	if len(pv.VariantName) > 255 {
		return fmt.Errorf("variant name cannot exceed 255 characters")
	}

	// Validate variant SKU if provided
	if pv.VariantSKU != nil {
		sku := strings.TrimSpace(*pv.VariantSKU)
		if sku == "" {
			pv.VariantSKU = nil
		} else {
			if len(sku) > 100 {
				return fmt.Errorf("variant SKU cannot exceed 100 characters")
			}
			pv.VariantSKU = &sku
		}
	}

	// Validate options
	if err := pv.ValidateOptions(); err != nil {
		return fmt.Errorf("options validation failed: %w", err)
	}

	// Validate pricing
	if err := pv.ValidatePricing(); err != nil {
		return fmt.Errorf("pricing validation failed: %w", err)
	}

	// Validate inventory
	if err := pv.ValidateInventory(); err != nil {
		return fmt.Errorf("inventory validation failed: %w", err)
	}

	// Validate dimensions
	if err := pv.ValidateDimensions(); err != nil {
		return fmt.Errorf("dimensions validation failed: %w", err)
	}

	// Position validation
	if pv.Position != nil && *pv.Position < 0 {
		return fmt.Errorf("position cannot be negative")
	}

	return nil
}

// GenerateVariantName generates a human-readable name from options
func (pv *ProductVariant) GenerateVariantName() {
	if len(pv.Options) == 0 {
		pv.VariantName = "Default"
		return
	}

	// Sort option names for consistent naming
	var optionNames []string
	for optionName := range pv.Options {
		optionNames = append(optionNames, optionName)
	}
	sort.Strings(optionNames)

	// Build name from sorted options
	var parts []string
	for _, optionName := range optionNames {
		optionValue := pv.Options[optionName]
		valueStr := fmt.Sprintf("%v", optionValue)
		parts = append(parts, valueStr)
	}

	pv.VariantName = strings.Join(parts, " / ")

	// Truncate if too long
	if len(pv.VariantName) > 255 {
		pv.VariantName = pv.VariantName[:252] + "..."
	}
}

// GetOptionValue gets a specific option value
func (pv *ProductVariant) GetOptionValue(optionName string) (interface{}, bool) {
	value, exists := pv.Options[optionName]
	return value, exists
}

// SetOptionValue sets a specific option value
func (pv *ProductVariant) SetOptionValue(optionName string, optionValue interface{}) error {
	if optionName == "" {
		return fmt.Errorf("option name cannot be empty")
	}

	if optionValue == nil {
		return fmt.Errorf("option value cannot be null")
	}

	// Create copy for validation
	tempOptions := make(map[string]interface{})
	for k, v := range pv.Options {
		tempOptions[k] = v
	}
	tempOptions[optionName] = optionValue

	// Validate with new option
	tempVariant := *pv
	tempVariant.Options = tempOptions
	if err := tempVariant.ValidateOptions(); err != nil {
		return err
	}

	// If validation passes, update
	pv.Options[optionName] = optionValue
	pv.GenerateVariantName()
	pv.UpdatedAt = time.Now()

	return nil
}

// RemoveOption removes an option
func (pv *ProductVariant) RemoveOption(optionName string) error {
	if len(pv.Options) <= 1 {
		return fmt.Errorf("cannot remove the last option")
	}

	delete(pv.Options, optionName)
	pv.GenerateVariantName()
	pv.UpdatedAt = time.Now()

	return nil
}

// UpdatePrice updates the variant price
func (pv *ProductVariant) UpdatePrice(price decimal.Decimal) error {
	tempVariant := *pv
	tempVariant.Price = price
	if err := tempVariant.ValidatePricing(); err != nil {
		return err
	}

	pv.Price = price
	pv.UpdatedAt = time.Now()
	return nil
}

// UpdateStock updates the stock quantity
func (pv *ProductVariant) UpdateStock(quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}

	pv.StockQuantity = quantity
	pv.UpdatedAt = time.Now()
	return nil
}

// DeductStock deducts stock quantity (for sales)
func (pv *ProductVariant) DeductStock(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("deduction quantity must be positive")
	}

	if pv.TrackQuantity && pv.StockQuantity < quantity {
		return fmt.Errorf("insufficient stock: requested %d, available %d", quantity, pv.StockQuantity)
	}

	if pv.TrackQuantity {
		pv.StockQuantity -= quantity
	}
	pv.UpdatedAt = time.Now()
	return nil
}

// RestockInventory adds stock quantity
func (pv *ProductVariant) RestockInventory(quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("restock quantity must be positive")
	}

	pv.StockQuantity += quantity
	pv.UpdatedAt = time.Now()
	return nil
}

// SetAsDefault sets this variant as the default
func (pv *ProductVariant) SetAsDefault() {
	pv.IsDefault = true
	pv.UpdatedAt = time.Now()
}

// UnsetAsDefault removes default status
func (pv *ProductVariant) UnsetAsDefault() {
	pv.IsDefault = false
	pv.UpdatedAt = time.Now()
}

// Activate activates the variant
func (pv *ProductVariant) Activate() {
	pv.IsActive = true
	pv.UpdatedAt = time.Now()
}

// Deactivate deactivates the variant
func (pv *ProductVariant) Deactivate() {
	pv.IsActive = false
	pv.UpdatedAt = time.Now()
}

// IsAvailable checks if the variant is available for purchase
func (pv *ProductVariant) IsAvailable() bool {
	if !pv.IsActive {
		return false
	}

	if pv.TrackQuantity {
		return pv.StockQuantity > 0
	}

	return true
}

// CalculateProfitMargin calculates profit margin percentage
func (pv *ProductVariant) CalculateProfitMargin() *decimal.Decimal {
	if pv.CostPrice == nil || pv.CostPrice.IsZero() {
		return nil
	}

	profit := pv.Price.Sub(*pv.CostPrice)
	margin := profit.Div(pv.Price).Mul(decimal.NewFromInt(100))
	return &margin
}

// CalculateProfitAmount calculates profit amount
func (pv *ProductVariant) CalculateProfitAmount() *decimal.Decimal {
	if pv.CostPrice == nil {
		return nil
	}

	profit := pv.Price.Sub(*pv.CostPrice)
	return &profit
}

// IsLowStockStatus checks if the variant is low on stock
func (pv *ProductVariant) IsLowStockStatus() bool {
	if !pv.TrackQuantity {
		return false
	}

	if pv.LowStockThreshold == nil {
		return false
	}

	return pv.StockQuantity <= *pv.LowStockThreshold
}

// ComputeFields calculates all computed fields
func (pv *ProductVariant) ComputeFields() {
	pv.IsLowStock = pv.IsLowStockStatus()
	pv.ProfitMargin = pv.CalculateProfitMargin()
	pv.ProfitAmount = pv.CalculateProfitAmount()
	pv.FormattedPrice = fmt.Sprintf("Rp %s", pv.Price.StringFixed(0))
	pv.OptionCount = len(pv.Options)
}

// String returns a string representation of the variant
func (pv *ProductVariant) String() string {
	return fmt.Sprintf("ProductVariant{ID: %s, ProductID: %s, Name: %s, Price: %s}",
		pv.ID.String(), pv.ProductID.String(), pv.VariantName, pv.Price.String())
}
