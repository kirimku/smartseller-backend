package dto

import (
	"time"
)

// WarrantyBarcodeRequest represents a request to generate warranty barcodes
type WarrantyBarcodeRequest struct {
	ProductID    string `json:"product_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Quantity     int    `json:"quantity" validate:"required,min=1,max=1000" example:"100"`
	BatchName    string `json:"batch_name" validate:"omitempty,max=100" example:"Production Batch #12"`
	Notes        string `json:"notes" validate:"omitempty,max=500" example:"Q1 2024 production run"`
	ExpiryMonths int    `json:"expiry_months" validate:"required,min=1,max=120" example:"24"`
}

// WarrantyBarcodeResponse represents a warranty barcode
type WarrantyBarcodeResponse struct {
	ID           string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductID    string    `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ProductName  string    `json:"product_name" example:"Wireless Headphones XYZ"`
	ProductSKU   string    `json:"product_sku" example:"WH-XYZ-001"`
	BarcodeValue string    `json:"barcode_value" example:"WR-2024-ABC123DEF456"`
	Status       string    `json:"status" example:"active"`
	BatchID      *string   `json:"batch_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	BatchName    *string   `json:"batch_name,omitempty" example:"Production Batch #12"`
	ExpiryDate   time.Time `json:"expiry_date" example:"2025-12-31T23:59:59Z"`
	IsActive     bool      `json:"is_active" example:"true"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
}

// WarrantyBarcodeListRequest represents request parameters for listing warranty barcodes
type WarrantyBarcodeListRequest struct {
	PaginationRequest
	ProductID     *string `json:"product_id" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchID       *string `json:"batch_id" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
	Status        *string `json:"status" validate:"omitempty,oneof=active inactive claimed" example:"active"`
	Search        *string `json:"search" validate:"omitempty,max=255" example:"WR-2024"`
	CreatedAfter  *string `json:"created_after" validate:"omitempty" example:"2023-01-01"`
	CreatedBefore *string `json:"created_before" validate:"omitempty" example:"2023-12-31"`
	ExpiryAfter   *string `json:"expiry_after" validate:"omitempty" example:"2024-01-01"`
	ExpiryBefore  *string `json:"expiry_before" validate:"omitempty" example:"2025-12-31"`
}

// WarrantyBarcodeListResponse represents the response for listing warranty barcodes
type WarrantyBarcodeListResponse struct {
	Data       []WarrantyBarcodeResponse `json:"data"`
	Pagination PaginationResponse        `json:"pagination"`
	Filters    WarrantyBarcodeFilters    `json:"filters"`
}

// WarrantyBarcodeFilters represents available filters for warranty barcodes
type WarrantyBarcodeFilters struct {
	Statuses []string                 `json:"statuses"`
	Products []WarrantyProductSummary `json:"products"`
	Batches  []WarrantyBatchSummary   `json:"batches"`
}

// WarrantyProductSummary represents a product summary for filters
type WarrantyProductSummary struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name string `json:"name" example:"Wireless Headphones XYZ"`
	SKU  string `json:"sku" example:"WH-XYZ-001"`
}

// WarrantyBatchSummary represents a batch summary for filters
type WarrantyBatchSummary struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name string `json:"name" example:"Production Batch #12"`
}

// BulkWarrantyBarcodeActivationRequest represents a request to activate multiple barcodes
type BulkWarrantyBarcodeActivationRequest struct {
	BarcodeIDs []string `json:"barcode_ids" validate:"required,min=1,max=100,dive,uuid" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
}

// BulkWarrantyBarcodeDeactivationRequest represents a request to deactivate multiple barcodes
type BulkWarrantyBarcodeDeactivationRequest struct {
	BarcodeIDs []string `json:"barcode_ids" validate:"required,min=1,max=100,dive,uuid" example:"[\"550e8400-e29b-41d4-a716-446655440000\"]"`
	Reason     string   `json:"reason" validate:"required,max=500" example:"Product recall - batch defect detected"`
}

// WarrantyBarcodeActivationRequest represents a request to activate a single barcode
type WarrantyBarcodeActivationRequest struct {
	Notes string `json:"notes" validate:"omitempty,max=500" example:"Activated for customer distribution"`
}

// WarrantyBarcodeDeactivationRequest represents a request to deactivate a single barcode
type WarrantyBarcodeDeactivationRequest struct {
	Reason string `json:"reason" validate:"required,max=500" example:"Defective product - return to supplier"`
}

// WarrantyBarcodeStatsResponse represents warranty barcode statistics
type WarrantyBarcodeStatsResponse struct {
	TotalBarcodes      int                      `json:"total_barcodes" example:"10000"`
	ActiveBarcodes     int                      `json:"active_barcodes" example:"8500"`
	InactiveBarcodes   int                      `json:"inactive_barcodes" example:"1000"`
	ClaimedBarcodes    int                      `json:"claimed_barcodes" example:"500"`
	ExpiredBarcodes    int                      `json:"expired_barcodes" example:"100"`
	StatusBreakdown    map[string]int           `json:"status_breakdown"`
	ProductStats       []ProductWarrantyStats   `json:"product_stats"`
	MonthlyGenerated   []MonthlyGenerationStats `json:"monthly_generated"`
	ExpiryBreakdown    []ExpiryBreakdownStats   `json:"expiry_breakdown"`
	GeneratedToday     int                      `json:"generated_today" example:"25"`
	GeneratedThisWeek  int                      `json:"generated_this_week" example:"150"`
	GeneratedThisMonth int                      `json:"generated_this_month" example:"600"`
	LastUpdated        time.Time                `json:"last_updated" example:"2023-01-01T00:00:00Z"`
}

// ProductWarrantyStats represents warranty statistics per product
type ProductWarrantyStats struct {
	ProductID       string  `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProductName     string  `json:"product_name" example:"Wireless Headphones XYZ"`
	ProductSKU      string  `json:"product_sku" example:"WH-XYZ-001"`
	TotalBarcodes   int     `json:"total_barcodes" example:"1000"`
	ActiveBarcodes  int     `json:"active_barcodes" example:"850"`
	ClaimedBarcodes int     `json:"claimed_barcodes" example:"100"`
	ClaimRate       float64 `json:"claim_rate" example:"10.0"`
}

// MonthlyGenerationStats represents monthly barcode generation statistics
type MonthlyGenerationStats struct {
	Month     string `json:"month" example:"2023-01"`
	Generated int    `json:"generated" example:"500"`
	Claimed   int    `json:"claimed" example:"50"`
}

// ExpiryBreakdownStats represents barcode expiry breakdown
type ExpiryBreakdownStats struct {
	ExpiryRange string  `json:"expiry_range" example:"0-3 months"`
	Count       int     `json:"count" example:"150"`
	Percentage  float64 `json:"percentage" example:"15.0"`
}

// BarcodeGenerationBatchResponse represents a barcode generation batch
type BarcodeGenerationBatchResponse struct {
	ID             string     `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchName      string     `json:"batch_name" example:"Production Batch #12"`
	ProductID      string     `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ProductName    string     `json:"product_name" example:"Wireless Headphones XYZ"`
	RequestedCount int        `json:"requested_count" example:"100"`
	GeneratedCount int        `json:"generated_count" example:"100"`
	Status         string     `json:"status" example:"completed"`
	Notes          *string    `json:"notes,omitempty" example:"Q1 2024 production run"`
	CreatedBy      string     `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440002"`
	CreatedByName  string     `json:"created_by_name" example:"John Doe"`
	ProcessingTime *string    `json:"processing_time,omitempty" example:"2.5s"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at" example:"2023-01-01T00:00:00Z"`
	CompletedAt    *time.Time `json:"completed_at,omitempty" example:"2023-01-01T00:00:05Z"`
}

// BarcodeGenerationBatchListRequest represents request parameters for listing generation batches
type BarcodeGenerationBatchListRequest struct {
	PaginationRequest
	ProductID     *string `json:"product_id" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status        *string `json:"status" validate:"omitempty,oneof=pending processing completed failed" example:"completed"`
	CreatedBy     *string `json:"created_by" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Search        *string `json:"search" validate:"omitempty,max=255" example:"Batch #12"`
	CreatedAfter  *string `json:"created_after" validate:"omitempty" example:"2023-01-01"`
	CreatedBefore *string `json:"created_before" validate:"omitempty" example:"2023-12-31"`
}

// BarcodeGenerationBatchListResponse represents the response for listing generation batches
type BarcodeGenerationBatchListResponse struct {
	Data       []BarcodeGenerationBatchResponse `json:"data"`
	Pagination PaginationResponse               `json:"pagination"`
}

// WarrantyBarcodeExportRequest represents a request to export warranty barcodes
type WarrantyBarcodeExportRequest struct {
	Format        string   `json:"format" validate:"required,oneof=csv excel pdf" example:"csv"`
	BarcodeIDs    []string `json:"barcode_ids" validate:"omitempty,max=1000,dive,uuid"`
	ProductID     *string  `json:"product_id" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchID       *string  `json:"batch_id" validate:"omitempty,uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
	Status        *string  `json:"status" validate:"omitempty,oneof=active inactive claimed" example:"active"`
	CreatedAfter  *string  `json:"created_after" validate:"omitempty" example:"2023-01-01"`
	CreatedBefore *string  `json:"created_before" validate:"omitempty" example:"2023-12-31"`
	IncludeQR     bool     `json:"include_qr" example:"true"`
	IncludePDF    bool     `json:"include_pdf" example:"false"`
}

// WarrantyBarcodePrintRequest represents a request to print warranty barcodes
type WarrantyBarcodePrintRequest struct {
	BarcodeIDs     []string `json:"barcode_ids" validate:"required,min=1,max=100,dive,uuid"`
	PrintFormat    string   `json:"print_format" validate:"required,oneof=labels stickers sheets" example:"labels"`
	IncludeQR      bool     `json:"include_qr" example:"true"`
	IncludeProduct bool     `json:"include_product" example:"true"`
	IncludeExpiry  bool     `json:"include_expiry" example:"true"`
	LabelSize      string   `json:"label_size" validate:"omitempty,oneof=small medium large" example:"medium"`
}

// WarrantyBarcodeValidationResponse represents barcode validation result
type WarrantyBarcodeValidationResponse struct {
	IsValid         bool                     `json:"is_valid" example:"true"`
	BarcodeValue    string                   `json:"barcode_value" example:"WR-2024-ABC123DEF456"`
	Status          string                   `json:"status" example:"active"`
	Product         *WarrantyProductSummary  `json:"product,omitempty"`
	ExpiryDate      *time.Time               `json:"expiry_date,omitempty" example:"2025-12-31T23:59:59Z"`
	IsExpired       bool                     `json:"is_expired" example:"false"`
	ClaimedAt       *time.Time               `json:"claimed_at,omitempty"`
	ClaimedBy       *string                  `json:"claimed_by,omitempty"`
	ValidationError *WarrantyValidationError `json:"validation_error,omitempty"`
	ValidatedAt     time.Time                `json:"validated_at" example:"2023-01-01T00:00:00Z"`
}

// WarrantyValidationError represents validation error details
type WarrantyValidationError struct {
	Code    string                 `json:"code" example:"BARCODE_EXPIRED"`
	Message string                 `json:"message" example:"Warranty barcode has expired"`
	Details map[string]interface{} `json:"details,omitempty"`
}
