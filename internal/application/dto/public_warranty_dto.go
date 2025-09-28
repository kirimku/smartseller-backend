package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// ===== PUBLIC WARRANTY VALIDATION DTOs =====

// PublicWarrantyValidationRequest represents a request to validate a warranty barcode
type PublicWarrantyValidationRequest struct {
	BarcodeValue string `json:"barcode_value" validate:"required,min=8,max=50" example:"WB-2024-001234567"`
	ProductSKU   string `json:"product_sku,omitempty" validate:"omitempty,max=100" example:"SKU-PHONE-001"`
}

// PublicWarrantyValidationResponse represents the response for warranty validation
type PublicWarrantyValidationResponse struct {
	Valid           bool                        `json:"valid" example:"true"`
	BarcodeValue    string                      `json:"barcode_value" example:"WB-2024-001234567"`
	Status          string                      `json:"status" example:"active"`
	Message         string                      `json:"message" example:"Warranty is valid and active"`
	Product         *PublicProductInfo          `json:"product,omitempty"`
	Warranty        *PublicWarrantyInfo         `json:"warranty,omitempty"`
	Coverage        *PublicWarrantyCoverage     `json:"coverage,omitempty"`
	ValidationTime  time.Time                   `json:"validation_time" example:"2024-01-15T10:30:00Z"`
}

// PublicProductInfo represents basic product information for public API
type PublicProductInfo struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	SKU         string  `json:"sku" example:"SKU-PHONE-001"`
	Name        string  `json:"name" example:"Smartphone Pro Max 256GB"`
	Brand       string  `json:"brand" example:"TechBrand"`
	Model       string  `json:"model" example:"Pro Max"`
	Category    string  `json:"category" example:"Electronics"`
	Description *string `json:"description,omitempty" example:"Latest flagship smartphone with advanced features"`
	ImageURL    *string `json:"image_url,omitempty" example:"https://example.com/images/phone.jpg"`
}

// PublicWarrantyInfo represents warranty information for public API
type PublicWarrantyInfo struct {
	ID              string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	BarcodeValue    string     `json:"barcode_value" example:"WB-2024-001234567"`
	Status          string     `json:"status" example:"active"`
	IsActive        bool       `json:"is_active" example:"true"`
	ActivatedAt     *time.Time `json:"activated_at,omitempty" example:"2024-01-10T14:30:00Z"`
	ExpiryDate      time.Time  `json:"expiry_date" example:"2026-01-10T23:59:59Z"`
	DaysRemaining   int        `json:"days_remaining" example:"730"`
	WarrantyPeriod  string     `json:"warranty_period" example:"24 months"`
	IsExpired       bool       `json:"is_expired" example:"false"`
	CanClaim        bool       `json:"can_claim" example:"true"`
}

// PublicWarrantyCoverage represents warranty coverage details for public API
type PublicWarrantyCoverage struct {
	CoverageType        string   `json:"coverage_type" example:"comprehensive"`
	CoveredComponents   []string `json:"covered_components" example:"hardware,software,battery,screen"`
	ExcludedComponents  []string `json:"excluded_components,omitempty" example:"water_damage,physical_abuse"`
	RepairCoverage      bool     `json:"repair_coverage" example:"true"`
	ReplacementCoverage bool     `json:"replacement_coverage" example:"true"`
	LaborCoverage       bool     `json:"labor_coverage" example:"true"`
	PartsCoverage       bool     `json:"parts_coverage" example:"true"`
	MaxClaimAmount      *decimal.Decimal `json:"max_claim_amount,omitempty" example:"500.00"`
	ClaimsRemaining     *int     `json:"claims_remaining,omitempty" example:"2"`
	MaxClaims           *int     `json:"max_claims,omitempty" example:"3"`
	Terms               []string `json:"terms,omitempty" example:"Must provide proof of purchase,Damage must be reported within 30 days"`
}

// PublicWarrantyLookupRequest represents a request to lookup warranty by product
type PublicWarrantyLookupRequest struct {
	ProductSKU      string `json:"product_sku" validate:"required,max=100" example:"SKU-PHONE-001"`
	SerialNumber    string `json:"serial_number,omitempty" validate:"omitempty,max=100" example:"SN123456789"`
	PurchaseDate    string `json:"purchase_date,omitempty" validate:"omitempty" example:"2024-01-10"`
	CustomerEmail   string `json:"customer_email,omitempty" validate:"omitempty,email" example:"customer@example.com"`
}

// PublicWarrantyLookupResponse represents the response for warranty lookup
type PublicWarrantyLookupResponse struct {
	Found       bool                    `json:"found" example:"true"`
	Warranties  []PublicWarrantyInfo    `json:"warranties"`
	Product     *PublicProductInfo      `json:"product,omitempty"`
	Message     string                  `json:"message" example:"Found 2 active warranties for this product"`
	SearchTime  time.Time               `json:"search_time" example:"2024-01-15T10:30:00Z"`
}

// PublicProductInfoRequest represents a request to get product information
type PublicProductInfoRequest struct {
	ProductSKU   string `json:"product_sku,omitempty" validate:"omitempty,max=100" example:"SKU-PHONE-001"`
	ProductID    string `json:"product_id,omitempty" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	BarcodeValue string `json:"barcode_value,omitempty" validate:"omitempty,min=8,max=50" example:"WB-2024-001234567"`
}

// PublicProductInfoResponse represents the response for product information
type PublicProductInfoResponse struct {
	Product         *PublicProductInfo      `json:"product"`
	WarrantyOptions []PublicWarrantyOption  `json:"warranty_options,omitempty"`
	Specifications  map[string]interface{}  `json:"specifications,omitempty" example:"{\"screen_size\":\"6.7 inches\",\"storage\":\"256GB\",\"color\":\"Space Gray\"}"`
	Documentation   []PublicDocumentInfo    `json:"documentation,omitempty"`
	SupportInfo     *PublicSupportInfo      `json:"support_info,omitempty"`
	RetrievedAt     time.Time               `json:"retrieved_at" example:"2024-01-15T10:30:00Z"`
}

// PublicWarrantyOption represents available warranty options for a product
type PublicWarrantyOption struct {
	Type            string           `json:"type" example:"standard"`
	Name            string           `json:"name" example:"Standard Warranty"`
	Duration        string           `json:"duration" example:"24 months"`
	Coverage        string           `json:"coverage" example:"Hardware defects and manufacturing issues"`
	Price           *decimal.Decimal `json:"price,omitempty" example:"0.00"`
	IsDefault       bool             `json:"is_default" example:"true"`
	Description     string           `json:"description" example:"Covers all manufacturing defects for 2 years"`
}

// PublicDocumentInfo represents product documentation information
type PublicDocumentInfo struct {
	Type        string `json:"type" example:"manual"`
	Title       string `json:"title" example:"User Manual"`
	Description string `json:"description" example:"Complete user guide and setup instructions"`
	URL         string `json:"url" example:"https://example.com/docs/user-manual.pdf"`
	Language    string `json:"language" example:"en"`
	FileSize    string `json:"file_size,omitempty" example:"2.5MB"`
}

// PublicSupportInfo represents product support information
type PublicSupportInfo struct {
	SupportEmail    string   `json:"support_email" example:"support@techbrand.com"`
	SupportPhone    string   `json:"support_phone" example:"+1-800-123-4567"`
	SupportHours    string   `json:"support_hours" example:"Mon-Fri 9AM-6PM EST"`
	OnlineSupport   string   `json:"online_support" example:"https://support.techbrand.com"`
	ChatSupport     bool     `json:"chat_support" example:"true"`
	SupportLanguages []string `json:"support_languages" example:"en,es,fr"`
	FAQ             string   `json:"faq,omitempty" example:"https://support.techbrand.com/faq"`
}

// PublicWarrantyCoverageCheckRequest represents a request to check warranty coverage
type PublicWarrantyCoverageCheckRequest struct {
	BarcodeValue    string `json:"barcode_value" validate:"required,min=8,max=50" example:"WB-2024-001234567"`
	IssueType       string `json:"issue_type" validate:"required,max=100" example:"hardware_failure"`
	IssueCategory   string `json:"issue_category,omitempty" validate:"omitempty,max=100" example:"screen"`
	Description     string `json:"description,omitempty" validate:"omitempty,max=500" example:"Screen has dead pixels in the upper right corner"`
	PurchaseDate    string `json:"purchase_date,omitempty" validate:"omitempty" example:"2024-01-10"`
}

// PublicWarrantyCoverageCheckResponse represents the response for warranty coverage check
type PublicWarrantyCoverageCheckResponse struct {
	Covered         bool                    `json:"covered" example:"true"`
	BarcodeValue    string                  `json:"barcode_value" example:"WB-2024-001234567"`
	IssueType       string                  `json:"issue_type" example:"hardware_failure"`
	CoverageType    string                  `json:"coverage_type" example:"full"`
	EstimatedCost   *decimal.Decimal        `json:"estimated_cost,omitempty" example:"0.00"`
	DeductibleAmount *decimal.Decimal       `json:"deductible_amount,omitempty" example:"25.00"`
	Coverage        *PublicWarrantyCoverage `json:"coverage"`
	Recommendations []string                `json:"recommendations,omitempty" example:"Contact authorized service center,Backup your data before repair"`
	NextSteps       []string                `json:"next_steps" example:"Submit warranty claim,Schedule repair appointment"`
	Message         string                  `json:"message" example:"This issue is fully covered under your warranty"`
	CheckedAt       time.Time               `json:"checked_at" example:"2024-01-15T10:30:00Z"`
}

// PublicWarrantyErrorResponse represents error responses for public warranty APIs
type PublicWarrantyErrorResponse struct {
	Error       string                 `json:"error" example:"warranty_not_found"`
	Message     string                 `json:"message" example:"No warranty found for the provided barcode"`
	Code        string                 `json:"code" example:"WAR_404"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Timestamp   time.Time              `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	RequestID   string                 `json:"request_id,omitempty" example:"req_123456789"`
}