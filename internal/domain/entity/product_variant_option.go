package entity

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ProductVariantOption defines the available options for a product's variants
// For example: Color options ["Red", "Blue", "Green"] or Size options ["S", "M", "L", "XL"]
type ProductVariantOption struct {
	// Primary identification
	ID        uuid.UUID `json:"id" db:"id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`

	// Option definition
	OptionName   string         `json:"option_name" db:"option_name"`     // e.g., "Color", "Size", "Material"
	OptionValues pq.StringArray `json:"option_values" db:"option_values"` // e.g., ["Red", "Blue", "Green"]

	// Display properties
	DisplayName *string `json:"display_name" db:"display_name"` // User-friendly name
	SortOrder   int     `json:"sort_order" db:"sort_order"`
	IsRequired  bool    `json:"is_required" db:"is_required"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields
	ValueCount int `json:"value_count" db:"-"`
}

// NewProductVariantOption creates a new product variant option
func NewProductVariantOption(productID uuid.UUID, optionName string, optionValues []string) *ProductVariantOption {
	return &ProductVariantOption{
		ID:           uuid.New(),
		ProductID:    productID,
		OptionName:   optionName,
		OptionValues: optionValues,
		SortOrder:    0,
		IsRequired:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ValidateOptionName validates the option name
func (pvo *ProductVariantOption) ValidateOptionName() error {
	if pvo.OptionName == "" {
		return fmt.Errorf("option name is required")
	}

	if len(pvo.OptionName) > 100 {
		return fmt.Errorf("option name cannot exceed 100 characters")
	}

	// Option name should contain only letters, numbers, spaces, and common punctuation
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.]+$`)
	if !validName.MatchString(pvo.OptionName) {
		return fmt.Errorf("option name can only contain letters, numbers, spaces, hyphens, underscores, and dots")
	}

	// Should not start or end with whitespace
	if strings.TrimSpace(pvo.OptionName) != pvo.OptionName {
		return fmt.Errorf("option name cannot start or end with whitespace")
	}

	return nil
}

// ValidateOptionValues validates the option values
func (pvo *ProductVariantOption) ValidateOptionValues() error {
	if len(pvo.OptionValues) == 0 {
		return fmt.Errorf("at least one option value is required")
	}

	if len(pvo.OptionValues) > 100 {
		return fmt.Errorf("cannot have more than 100 option values")
	}

	// Check for duplicates and validate each value
	seen := make(map[string]bool)
	for i, value := range pvo.OptionValues {
		// Trim whitespace
		trimmedValue := strings.TrimSpace(value)
		if trimmedValue == "" {
			return fmt.Errorf("option value at index %d cannot be empty", i)
		}

		if len(trimmedValue) > 255 {
			return fmt.Errorf("option value at index %d cannot exceed 255 characters", i)
		}

		// Check for duplicates (case-insensitive)
		lowerValue := strings.ToLower(trimmedValue)
		if seen[lowerValue] {
			return fmt.Errorf("duplicate option value: %s", trimmedValue)
		}
		seen[lowerValue] = true

		// Update the value with trimmed version
		pvo.OptionValues[i] = trimmedValue
	}

	return nil
}

// Validate performs comprehensive validation of the variant option
func (pvo *ProductVariantOption) Validate() error {
	// Validate option name
	if err := pvo.ValidateOptionName(); err != nil {
		return fmt.Errorf("option name validation failed: %w", err)
	}

	// Validate option values
	if err := pvo.ValidateOptionValues(); err != nil {
		return fmt.Errorf("option values validation failed: %w", err)
	}

	// Sort order should not be negative
	if pvo.SortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	// Validate display name if provided
	if pvo.DisplayName != nil {
		displayName := strings.TrimSpace(*pvo.DisplayName)
		if displayName == "" {
			pvo.DisplayName = nil // Clear empty display name
		} else if len(displayName) > 255 {
			return fmt.Errorf("display name cannot exceed 255 characters")
		} else {
			pvo.DisplayName = &displayName
		}
	}

	return nil
}

// IsValidValue checks if a value is valid for this option
func (pvo *ProductVariantOption) IsValidValue(value string) bool {
	trimmedValue := strings.TrimSpace(value)
	for _, optionValue := range pvo.OptionValues {
		if strings.EqualFold(optionValue, trimmedValue) {
			return true
		}
	}
	return false
}

// HasValue checks if the option has a specific value
func (pvo *ProductVariantOption) HasValue(value string) bool {
	return pvo.IsValidValue(value)
}

// AddValue adds a new value to the option (if not already present)
func (pvo *ProductVariantOption) AddValue(value string) error {
	trimmedValue := strings.TrimSpace(value)
	if trimmedValue == "" {
		return fmt.Errorf("value cannot be empty")
	}

	if len(trimmedValue) > 255 {
		return fmt.Errorf("value cannot exceed 255 characters")
	}

	// Check if value already exists (case-insensitive)
	if pvo.IsValidValue(trimmedValue) {
		return fmt.Errorf("value already exists: %s", trimmedValue)
	}

	// Check limit
	if len(pvo.OptionValues) >= 100 {
		return fmt.Errorf("cannot have more than 100 option values")
	}

	pvo.OptionValues = append(pvo.OptionValues, trimmedValue)
	pvo.UpdatedAt = time.Now()

	return nil
}

// RemoveValue removes a value from the option
func (pvo *ProductVariantOption) RemoveValue(value string) error {
	if len(pvo.OptionValues) <= 1 {
		return fmt.Errorf("cannot remove the last option value")
	}

	trimmedValue := strings.TrimSpace(value)
	for i, optionValue := range pvo.OptionValues {
		if strings.EqualFold(optionValue, trimmedValue) {
			// Remove from slice
			pvo.OptionValues = append(pvo.OptionValues[:i], pvo.OptionValues[i+1:]...)
			pvo.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("value not found: %s", trimmedValue)
}

// UpdateValue replaces an old value with a new value
func (pvo *ProductVariantOption) UpdateValue(oldValue, newValue string) error {
	trimmedOldValue := strings.TrimSpace(oldValue)
	trimmedNewValue := strings.TrimSpace(newValue)

	if trimmedNewValue == "" {
		return fmt.Errorf("new value cannot be empty")
	}

	if len(trimmedNewValue) > 255 {
		return fmt.Errorf("new value cannot exceed 255 characters")
	}

	// Check if new value already exists (but not the same as old value)
	if !strings.EqualFold(trimmedOldValue, trimmedNewValue) && pvo.IsValidValue(trimmedNewValue) {
		return fmt.Errorf("new value already exists: %s", trimmedNewValue)
	}

	// Find and update the old value
	for i, optionValue := range pvo.OptionValues {
		if strings.EqualFold(optionValue, trimmedOldValue) {
			pvo.OptionValues[i] = trimmedNewValue
			pvo.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("old value not found: %s", trimmedOldValue)
}

// SetValues replaces all values with new ones
func (pvo *ProductVariantOption) SetValues(values []string) error {
	if len(values) == 0 {
		return fmt.Errorf("at least one option value is required")
	}

	// Create a copy to validate without modifying original
	tempOption := *pvo
	tempOption.OptionValues = make([]string, len(values))
	copy(tempOption.OptionValues, values)

	// Validate the new values
	if err := tempOption.ValidateOptionValues(); err != nil {
		return err
	}

	// If validation passes, update the actual values
	pvo.OptionValues = tempOption.OptionValues
	pvo.UpdatedAt = time.Now()

	return nil
}

// GetDisplayName returns the display name if set, otherwise returns the option name
func (pvo *ProductVariantOption) GetDisplayName() string {
	if pvo.DisplayName != nil && *pvo.DisplayName != "" {
		return *pvo.DisplayName
	}
	return pvo.OptionName
}

// SetDisplayName sets the display name
func (pvo *ProductVariantOption) SetDisplayName(displayName string) error {
	trimmedName := strings.TrimSpace(displayName)
	if trimmedName == "" {
		pvo.DisplayName = nil
		pvo.UpdatedAt = time.Now()
		return nil
	}

	if len(trimmedName) > 255 {
		return fmt.Errorf("display name cannot exceed 255 characters")
	}

	pvo.DisplayName = &trimmedName
	pvo.UpdatedAt = time.Now()
	return nil
}

// UpdateSortOrder updates the sort order
func (pvo *ProductVariantOption) UpdateSortOrder(sortOrder int) error {
	if sortOrder < 0 {
		return fmt.Errorf("sort order cannot be negative")
	}

	pvo.SortOrder = sortOrder
	pvo.UpdatedAt = time.Now()
	return nil
}

// SetRequired sets whether this option is required
func (pvo *ProductVariantOption) SetRequired(isRequired bool) {
	pvo.IsRequired = isRequired
	pvo.UpdatedAt = time.Now()
}

// GetValueCount returns the number of option values
func (pvo *ProductVariantOption) GetValueCount() int {
	return len(pvo.OptionValues)
}

// ComputeFields calculates computed fields for the variant option
func (pvo *ProductVariantOption) ComputeFields() {
	pvo.ValueCount = pvo.GetValueCount()
}

// String returns a string representation of the variant option
func (pvo *ProductVariantOption) String() string {
	return fmt.Sprintf("ProductVariantOption{ID: %s, ProductID: %s, Name: %s, Values: %v}",
		pvo.ID.String(), pvo.ProductID.String(), pvo.OptionName, pvo.OptionValues)
}
