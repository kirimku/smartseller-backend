package entity

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// ProductStatus represents the lifecycle status of a product
type ProductStatus string

const (
	// ProductStatusDraft represents a product in draft state (not visible to customers)
	ProductStatusDraft ProductStatus = "draft"

	// ProductStatusActive represents an active product (visible and purchasable)
	ProductStatusActive ProductStatus = "active"

	// ProductStatusInactive represents an inactive product (visible but not purchasable)
	ProductStatusInactive ProductStatus = "inactive"

	// ProductStatusArchived represents an archived product (not visible, for historical purposes)
	ProductStatusArchived ProductStatus = "archived"
)

// Valid validates the product status
func (ps ProductStatus) Valid() bool {
	switch ps {
	case ProductStatusDraft, ProductStatusActive, ProductStatusInactive, ProductStatusArchived:
		return true
	default:
		return false
	}
}

// String returns the string representation of ProductStatus
func (ps ProductStatus) String() string {
	return string(ps)
}

// Value implements the driver.Valuer interface for database storage
func (ps ProductStatus) Value() (driver.Value, error) {
	return string(ps), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (ps *ProductStatus) Scan(value interface{}) error {
	if value == nil {
		*ps = ProductStatusDraft
		return nil
	}
	if str, ok := value.(string); ok {
		*ps = ProductStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ProductStatus", value)
}

// Product represents a product in the SmartSeller e-commerce system
type Product struct {
	// Primary identification
	ID  uuid.UUID `json:"id" db:"id"`
	SKU string    `json:"sku" db:"sku"`

	// Basic information
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description" db:"description"`

	// Organization
	CategoryID *uuid.UUID     `json:"category_id" db:"category_id"`
	Brand      *string        `json:"brand" db:"brand"`
	Tags       pq.StringArray `json:"tags" db:"tags"`

	// Pricing (using decimal for precise monetary calculations)
	BasePrice decimal.Decimal  `json:"base_price" db:"base_price"`
	SalePrice *decimal.Decimal `json:"sale_price" db:"sale_price"`
	CostPrice *decimal.Decimal `json:"cost_price" db:"cost_price"`

	// Inventory management
	TrackInventory    bool `json:"track_inventory" db:"track_inventory"`
	StockQuantity     int  `json:"stock_quantity" db:"stock_quantity"`
	LowStockThreshold *int `json:"low_stock_threshold" db:"low_stock_threshold"`

	// Product status
	Status ProductStatus `json:"status" db:"status"`

	// SEO and marketing
	MetaTitle       *string `json:"meta_title" db:"meta_title"`
	MetaDescription *string `json:"meta_description" db:"meta_description"`
	Slug            *string `json:"slug" db:"slug"`

	// Physical attributes
	Weight           *decimal.Decimal `json:"weight" db:"weight"`                       // in kg
	DimensionsLength *decimal.Decimal `json:"dimensions_length" db:"dimensions_length"` // in cm
	DimensionsWidth  *decimal.Decimal `json:"dimensions_width" db:"dimensions_width"`   // in cm
	DimensionsHeight *decimal.Decimal `json:"dimensions_height" db:"dimensions_height"` // in cm

	// Ownership and audit
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`

	// Computed fields (not stored in database)
	IsLowStock     bool             `json:"is_low_stock" db:"-"`
	ProfitMargin   *decimal.Decimal `json:"profit_margin" db:"-"`
	EffectivePrice decimal.Decimal  `json:"effective_price" db:"-"`
}

// NewProduct creates a new product with default values
func NewProduct(name, sku string, basePrice decimal.Decimal, createdBy uuid.UUID) *Product {
	return &Product{
		ID:                uuid.New(),
		SKU:               sku,
		Name:              name,
		BasePrice:         basePrice,
		TrackInventory:    true,
		StockQuantity:     0,
		LowStockThreshold: newInt(10), // Default low stock threshold
		Status:            ProductStatusDraft,
		CreatedBy:         createdBy,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// ValidateSKU validates the SKU format
func (p *Product) ValidateSKU() error {
	if p.SKU == "" {
		return fmt.Errorf("SKU is required")
	}

	// SKU should be alphanumeric with optional hyphens and underscores
	// Length should be between 3 and 100 characters
	if len(p.SKU) < 3 || len(p.SKU) > 100 {
		return fmt.Errorf("SKU must be between 3 and 100 characters")
	}

	// Only allow alphanumeric characters, hyphens, and underscores
	validSKU := regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)
	if !validSKU.MatchString(p.SKU) {
		return fmt.Errorf("SKU can only contain letters, numbers, hyphens, and underscores")
	}

	return nil
}

// ValidatePricing validates product pricing rules
func (p *Product) ValidatePricing() error {
	// Base price must be non-negative
	if p.BasePrice.IsNegative() {
		return fmt.Errorf("base price cannot be negative")
	}

	// Sale price validation
	if p.SalePrice != nil {
		if p.SalePrice.IsNegative() {
			return fmt.Errorf("sale price cannot be negative")
		}
		// Sale price should not be higher than base price
		if p.SalePrice.GreaterThan(p.BasePrice) {
			return fmt.Errorf("sale price cannot be higher than base price")
		}
	}

	// Cost price validation
	if p.CostPrice != nil {
		if p.CostPrice.IsNegative() {
			return fmt.Errorf("cost price cannot be negative")
		}
	}

	return nil
}

// ValidateDimensions validates product physical dimensions
func (p *Product) ValidateDimensions() error {
	// All dimensions must be positive if provided
	if p.DimensionsLength != nil && p.DimensionsLength.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("length must be positive")
	}
	if p.DimensionsWidth != nil && p.DimensionsWidth.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("width must be positive")
	}
	if p.DimensionsHeight != nil && p.DimensionsHeight.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("height must be positive")
	}
	if p.Weight != nil && p.Weight.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("weight must be positive")
	}

	return nil
}

// Validate performs comprehensive validation of the product
func (p *Product) Validate() error {
	// Required fields
	if p.Name == "" {
		return fmt.Errorf("product name is required")
	}
	if len(p.Name) > 500 {
		return fmt.Errorf("product name cannot exceed 500 characters")
	}

	// Validate SKU
	if err := p.ValidateSKU(); err != nil {
		return fmt.Errorf("SKU validation failed: %w", err)
	}

	// Validate pricing
	if err := p.ValidatePricing(); err != nil {
		return fmt.Errorf("pricing validation failed: %w", err)
	}

	// Validate dimensions
	if err := p.ValidateDimensions(); err != nil {
		return fmt.Errorf("dimensions validation failed: %w", err)
	}

	// Validate status
	if !p.Status.Valid() {
		return fmt.Errorf("invalid product status: %s", p.Status)
	}

	// Validate stock
	if p.StockQuantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}

	if p.LowStockThreshold != nil && *p.LowStockThreshold < 0 {
		return fmt.Errorf("low stock threshold cannot be negative")
	}

	return nil
}

// CalculateProfitMargin calculates the profit margin percentage
func (p *Product) CalculateProfitMargin() *decimal.Decimal {
	if p.CostPrice == nil || p.CostPrice.IsZero() {
		return nil
	}

	effectivePrice := p.GetEffectivePrice()
	if effectivePrice.IsZero() {
		return nil
	}

	// Profit margin = ((Selling Price - Cost Price) / Selling Price) * 100
	profit := effectivePrice.Sub(*p.CostPrice)
	margin := profit.Div(effectivePrice).Mul(decimal.NewFromInt(100))

	return &margin
}

// GetEffectivePrice returns the effective selling price (sale price if available, otherwise base price)
func (p *Product) GetEffectivePrice() decimal.Decimal {
	if p.SalePrice != nil && p.SalePrice.GreaterThan(decimal.Zero) {
		return *p.SalePrice
	}
	return p.BasePrice
}

// IsLowStockLevel checks if the product is at or below the low stock threshold
func (p *Product) IsLowStockLevel() bool {
	if !p.TrackInventory || p.LowStockThreshold == nil {
		return false
	}
	return p.StockQuantity <= *p.LowStockThreshold
}

// CanTransitionTo checks if the product can transition to the specified status
func (p *Product) CanTransitionTo(newStatus ProductStatus) bool {
	switch p.Status {
	case ProductStatusDraft:
		// Draft can go to any status
		return true
	case ProductStatusActive:
		// Active can go to inactive or archived
		return newStatus == ProductStatusInactive || newStatus == ProductStatusArchived
	case ProductStatusInactive:
		// Inactive can go to active or archived
		return newStatus == ProductStatusActive || newStatus == ProductStatusArchived
	case ProductStatusArchived:
		// Archived products cannot change status
		return false
	default:
		return false
	}
}

// UpdateStatus updates the product status with validation
func (p *Product) UpdateStatus(newStatus ProductStatus) error {
	if !newStatus.Valid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	if !p.CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", p.Status, newStatus)
	}

	p.Status = newStatus
	p.UpdatedAt = time.Now()
	return nil
}

// UpdateStock updates the stock quantity
func (p *Product) UpdateStock(quantity int) error {
	if quantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}

	p.StockQuantity = quantity
	p.UpdatedAt = time.Now()
	return nil
}

// AdjustStock adjusts the stock by a delta (can be positive or negative)
func (p *Product) AdjustStock(delta int) error {
	newQuantity := p.StockQuantity + delta
	if newQuantity < 0 {
		return fmt.Errorf("insufficient stock: current=%d, adjustment=%d", p.StockQuantity, delta)
	}

	p.StockQuantity = newQuantity
	p.UpdatedAt = time.Now()
	return nil
}

// GenerateSlug generates a URL-friendly slug from the product name
func (p *Product) GenerateSlug() {
	if p.Name == "" {
		return
	}

	// Convert to lowercase
	slug := strings.ToLower(p.Name)

	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	// Limit length to 255 characters
	if len(slug) > 255 {
		slug = slug[:255]
	}

	p.Slug = &slug
}

// SoftDelete marks the product as deleted
func (p *Product) SoftDelete() {
	now := time.Now()
	p.DeletedAt = &now
	p.UpdatedAt = now
}

// Restore restores a soft-deleted product
func (p *Product) Restore() {
	p.DeletedAt = nil
	p.UpdatedAt = time.Now()
}

// IsDeleted checks if the product is soft-deleted
func (p *Product) IsDeleted() bool {
	return p.DeletedAt != nil
}

// ComputeFields calculates computed fields for the product
func (p *Product) ComputeFields() {
	p.IsLowStock = p.IsLowStockLevel()
	p.ProfitMargin = p.CalculateProfitMargin()
	p.EffectivePrice = p.GetEffectivePrice()
}

// String returns a string representation of the product
func (p *Product) String() string {
	return fmt.Sprintf("Product{ID: %s, SKU: %s, Name: %s, Status: %s}",
		p.ID.String(), p.SKU, p.Name, p.Status)
}

// Helper function to create a pointer to an int
func newInt(i int) *int {
	return &i
}
