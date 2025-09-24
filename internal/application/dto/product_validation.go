package dto

import (
"regexp"
"strings"

"github.com/shopspring/decimal"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationResult contains validation results
type ValidationResult struct {
	Valid  bool              `json:"valid"`
	Errors []ValidationError `json:"errors,omitempty"`
}

var skuPattern = regexp.MustCompile(`^[A-Z0-9]([A-Z0-9_-]){2,99}$`)

// ValidateSKUFormat validates SKU format according to business rules
func ValidateSKUFormat(sku string) *ValidationError {
	if sku == "" {
		return &ValidationError{
			Field:   "sku",
			Message: "SKU is required",
		}
	}
	
	if !skuPattern.MatchString(sku) {
		return &ValidationError{
			Field:   "sku",
			Message: "SKU must be 3-100 characters, uppercase alphanumeric with hyphens/underscores only",
			Value:   sku,
		}
	}
	
	if strings.HasPrefix(sku, "-") || strings.HasPrefix(sku, "_") {
		return &ValidationError{
			Field:   "sku",
			Message: "SKU cannot start with hyphen or underscore",
			Value:   sku,
		}
	}
	
	return nil
}

// ValidatePricing validates pricing with margin checks
func ValidatePricing(basePrice decimal.Decimal, salePrice, costPrice *decimal.Decimal) []ValidationError {
	var errors []ValidationError
	
	if basePrice.IsZero() || basePrice.IsNegative() {
		errors = append(errors, ValidationError{
Field:   "base_price",
Message: "Base price must be greater than 0",
Value:   basePrice.String(),
		})
	}
	
	maxPrice := decimal.NewFromInt(100000000)
	if basePrice.GreaterThan(maxPrice) {
		errors = append(errors, ValidationError{
Field:   "base_price",
Message: "Base price exceeds maximum allowed (100M IDR)",
Value:   basePrice.String(),
		})
	}
	
	if salePrice != nil {
		if salePrice.IsNegative() {
			errors = append(errors, ValidationError{
Field:   "sale_price",
Message: "Sale price cannot be negative",
Value:   salePrice.String(),
			})
		}
		
		if salePrice.GreaterThan(basePrice) {
			errors = append(errors, ValidationError{
Field:   "sale_price",
Message: "Sale price cannot be higher than base price",
Value:   salePrice.String(),
			})
		}
	}
	
	if costPrice != nil && !costPrice.IsZero() {
		effectivePrice := basePrice
		if salePrice != nil {
			effectivePrice = *salePrice
		}
		
		profit := effectivePrice.Sub(*costPrice)
		marginPct := profit.Div(*costPrice).Mul(decimal.NewFromInt(100))
		minMargin := decimal.NewFromFloat(5.0)
		
		if marginPct.LessThan(minMargin) {
			errors = append(errors, ValidationError{
Field:   "pricing",
Message: "Profit margin too low (minimum 5%)",
Value:   marginPct.StringFixed(2) + "%",
})
		}
	}
	
	return errors
}

// ValidateCreateProductRequest performs comprehensive validation
func ValidateCreateProductRequest(req *CreateProductRequest) *ValidationResult {
	result := &ValidationResult{Valid: true, Errors: []ValidationError{}}
	
	if skuError := ValidateSKUFormat(req.SKU); skuError != nil {
		result.Valid = false
		result.Errors = append(result.Errors, *skuError)
	}
	
	if pricingErrors := ValidatePricing(req.BasePrice, req.SalePrice, req.CostPrice); len(pricingErrors) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, pricingErrors...)
	}
	
	return result
}
