package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// WarrantyClaimSubmissionRequest represents a request to submit a warranty claim
type WarrantyClaimSubmissionRequest struct {
	BarcodeValue      string    `json:"barcode_value" validate:"required,min=10,max=50" example:"WB-2024-ABC123DEF456"`
	IssueDescription  string    `json:"issue_description" validate:"required,min=10,max=2000" example:"Device stops working after 3 months of use"`
	IssueCategory     string    `json:"issue_category" validate:"required,oneof=hardware software performance defect damage other" example:"hardware"`
	IssueDate         time.Time `json:"issue_date" validate:"required" example:"2024-01-15T10:30:00Z"`
	Severity          string    `json:"severity" validate:"required,oneof=low medium high critical" example:"medium"`
	CustomerName      string    `json:"customer_name" validate:"required,min=2,max=100" example:"John Doe"`
	CustomerEmail     string    `json:"customer_email" validate:"required,email,max=255" example:"john.doe@example.com"`
	CustomerPhone     string    `json:"customer_phone" validate:"required,min=10,max=20" example:"+1234567890"`
	PickupAddress     string    `json:"pickup_address" validate:"required,min=10,max=500" example:"123 Main St, City, State 12345"`
	CustomerNotes     string    `json:"customer_notes" validate:"omitempty,max=1000" example:"Additional details about the issue"`
	PurchaseDate      *time.Time `json:"purchase_date,omitempty" example:"2023-10-15T00:00:00Z"`
	PurchaseLocation  string    `json:"purchase_location" validate:"omitempty,max=200" example:"Online Store"`
	ReceiptNumber     string    `json:"receipt_number" validate:"omitempty,max=100" example:"RCP-2023-001234"`
}

// WarrantyClaimResponse represents a warranty claim response
type WarrantyClaimResponse struct {
	ID               string                        `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ClaimNumber      string                        `json:"claim_number" example:"WAR-2024-001234"`
	BarcodeID        string                        `json:"barcode_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	BarcodeValue     string                        `json:"barcode_value" example:"WB-2024-ABC123DEF456"`
	CustomerID       string                        `json:"customer_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	ProductID        string                        `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	StorefrontID     string                        `json:"storefront_id" example:"550e8400-e29b-41d4-a716-446655440004"`
	
	// Issue details
	IssueDescription string    `json:"issue_description" example:"Device stops working after 3 months of use"`
	IssueCategory    string    `json:"issue_category" example:"hardware"`
	IssueDate        time.Time `json:"issue_date" example:"2024-01-15T10:30:00Z"`
	Severity         string    `json:"severity" example:"medium"`
	
	// Claim timeline
	ClaimDate   time.Time  `json:"claim_date" example:"2024-01-20T14:30:00Z"`
	ValidatedAt *time.Time `json:"validated_at,omitempty" example:"2024-01-21T09:15:00Z"`
	CompletedAt *time.Time `json:"completed_at,omitempty" example:"2024-01-25T16:45:00Z"`
	
	// Status management
	Status          string     `json:"status" example:"pending"`
	PreviousStatus  *string    `json:"previous_status,omitempty" example:"validated"`
	StatusUpdatedAt time.Time  `json:"status_updated_at" example:"2024-01-20T14:30:00Z"`
	StatusUpdatedBy *string    `json:"status_updated_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440005"`
	
	// Processing assignment
	ValidatedBy             *string    `json:"validated_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440006"`
	AssignedTechnicianID    *string    `json:"assigned_technician_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440007"`
	EstimatedCompletionDate *time.Time `json:"estimated_completion_date,omitempty" example:"2024-01-25T17:00:00Z"`
	ActualCompletionDate    *time.Time `json:"actual_completion_date,omitempty" example:"2024-01-25T16:45:00Z"`
	
	// Resolution details
	ResolutionType        *string         `json:"resolution_type,omitempty" example:"repair"`
	RepairNotes           *string         `json:"repair_notes,omitempty" example:"Replaced faulty component"`
	ReplacementProductID  *string         `json:"replacement_product_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440008"`
	RefundAmount          *decimal.Decimal `json:"refund_amount,omitempty" example:"99.99"`
	
	// Cost tracking
	RepairCost      decimal.Decimal `json:"repair_cost" example:"25.50"`
	ShippingCost    decimal.Decimal `json:"shipping_cost" example:"10.00"`
	ReplacementCost decimal.Decimal `json:"replacement_cost" example:"0.00"`
	TotalCost       decimal.Decimal `json:"total_cost" example:"35.50"`
	
	// Customer information
	CustomerName  string `json:"customer_name" example:"John Doe"`
	CustomerEmail string `json:"customer_email" example:"john.doe@example.com"`
	CustomerPhone string `json:"customer_phone" example:"+1234567890"`
	PickupAddress string `json:"pickup_address" example:"123 Main St, City, State 12345"`
	
	// Shipping information
	ShippingProvider      *string    `json:"shipping_provider,omitempty" example:"FedEx"`
	TrackingNumber        *string    `json:"tracking_number,omitempty" example:"1234567890"`
	EstimatedDeliveryDate *time.Time `json:"estimated_delivery_date,omitempty" example:"2024-01-27T17:00:00Z"`
	ActualDeliveryDate    *time.Time `json:"actual_delivery_date,omitempty" example:"2024-01-27T14:30:00Z"`
	DeliveryStatus        string     `json:"delivery_status" example:"not_shipped"`
	
	// Notes and feedback
	CustomerNotes                *string         `json:"customer_notes,omitempty" example:"Additional details about the issue"`
	AdminNotes                   *string         `json:"admin_notes,omitempty" example:"Internal processing notes"`
	RejectionReason              *string         `json:"rejection_reason,omitempty" example:"Out of warranty period"`
	InternalNotes                *string         `json:"internal_notes,omitempty" example:"Internal team communication"`
	Priority                     string          `json:"priority" example:"normal"`
	Tags                         []string        `json:"tags,omitempty" example:"urgent,vip_customer"`
	CustomerSatisfactionRating   *int            `json:"customer_satisfaction_rating,omitempty" example:"5"`
	CustomerFeedback             *string         `json:"customer_feedback,omitempty" example:"Excellent service"`
	ProcessingTimeHours          *decimal.Decimal `json:"processing_time_hours,omitempty" example:"48.5"`
	
	// Product information (computed)
	ProductName        *string `json:"product_name,omitempty" example:"Smartphone XYZ"`
	ProductSKU         *string `json:"product_sku,omitempty" example:"SKU-12345"`
	WarrantyExpiryDate *time.Time `json:"warranty_expiry_date,omitempty" example:"2025-10-15T00:00:00Z"`
	IsWarrantyValid    bool    `json:"is_warranty_valid" example:"true"`
	
	// Timestamps
	CreatedAt time.Time `json:"created_at" example:"2024-01-20T14:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-21T09:15:00Z"`
}

// WarrantyClaimListRequest represents a request to list warranty claims with filters
type WarrantyClaimListRequest struct {
	// Pagination
	Page     int `json:"page" validate:"min=1" example:"1"`
	PageSize int `json:"page_size" validate:"min=1,max=100" example:"20"`
	
	// Sorting
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=claim_date status severity priority customer_name" example:"claim_date"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc" example:"desc"`
	
	// Filters
	Status           []string   `json:"status,omitempty" example:"pending,validated"`
	Severity         []string   `json:"severity,omitempty" example:"medium,high"`
	Priority         []string   `json:"priority,omitempty" example:"normal,high"`
	IssueCategory    []string   `json:"issue_category,omitempty" example:"hardware,software"`
	TechnicianID     *string    `json:"technician_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440007"`
	CustomerID       *string    `json:"customer_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	ProductID        *string    `json:"product_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	ClaimDateFrom    *time.Time `json:"claim_date_from,omitempty" example:"2024-01-01T00:00:00Z"`
	ClaimDateTo      *time.Time `json:"claim_date_to,omitempty" example:"2024-01-31T23:59:59Z"`
	SearchTerm       string     `json:"search_term,omitempty" example:"smartphone"`
	Tags             []string   `json:"tags,omitempty" example:"urgent,vip_customer"`
}

// WarrantyClaimListResponse represents a paginated list of warranty claims
type WarrantyClaimListResponse struct {
	Claims     []WarrantyClaimResponse `json:"claims"`
	Pagination PaginationResponse      `json:"pagination"`
	Filters    ClaimFiltersResponse    `json:"filters"`
}

// ClaimFiltersResponse represents available filter options
type ClaimFiltersResponse struct {
	AvailableStatuses    []string `json:"available_statuses" example:"pending,validated,rejected,assigned,in_repair,repaired,replaced,shipped,delivered,completed,cancelled,disputed"`
	AvailableSeverities  []string `json:"available_severities" example:"low,medium,high,critical"`
	AvailablePriorities  []string `json:"available_priorities" example:"low,normal,high,urgent"`
	AvailableCategories  []string `json:"available_categories" example:"hardware,software,performance,defect,damage,other"`
}

// WarrantyClaimValidationRequest represents a request to validate a claim
type WarrantyClaimValidationRequest struct {
	Action                  string     `json:"action" validate:"required,oneof=validate reject request_info" example:"validate"`
	Notes                   string     `json:"notes" validate:"omitempty,max=1000" example:"Approved for repair - physical damage covered"`
	EstimatedCompletionDate *time.Time `json:"estimated_completion_date,omitempty" example:"2024-10-05T17:00:00Z"`
	RejectionReason         string     `json:"rejection_reason" validate:"omitempty,max=500" example:"Out of warranty period"`
	RequestedInfo           string     `json:"requested_info" validate:"omitempty,max=1000" example:"Please provide purchase receipt"`
}

// WarrantyClaimAssignmentRequest represents a request to assign a technician
type WarrantyClaimAssignmentRequest struct {
	TechnicianID            string     `json:"technician_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440007"`
	EstimatedCompletionDate *time.Time `json:"estimated_completion_date,omitempty" example:"2024-10-05T17:00:00Z"`
	Priority                string     `json:"priority" validate:"omitempty,oneof=low normal high urgent" example:"high"`
	Notes                   string     `json:"notes" validate:"omitempty,max=1000" example:"Rush repair for VIP customer"`
}

// WarrantyClaimStatusUpdateRequest represents a request to update claim status
type WarrantyClaimStatusUpdateRequest struct {
	Status      string `json:"status" validate:"required,oneof=pending validated rejected assigned in_repair repaired replaced shipped delivered completed cancelled disputed" example:"in_repair"`
	Notes       string `json:"notes" validate:"omitempty,max=1000" example:"Started repair process"`
	RepairNotes string `json:"repair_notes" validate:"omitempty,max=2000" example:"Replaced faulty component, tested functionality"`
}

// BulkClaimStatusUpdateRequest represents a request to update multiple claims
type BulkClaimStatusUpdateRequest struct {
	ClaimIDs []string `json:"claim_ids" validate:"required,min=1,max=100,dive,uuid" example:"550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001"`
	Status   string   `json:"status" validate:"required,oneof=pending validated rejected assigned in_repair repaired replaced shipped delivered completed cancelled disputed" example:"validated"`
	Notes    string   `json:"notes" validate:"omitempty,max=1000" example:"Bulk validation of claims"`
}

// WarrantyClaimStatsResponse represents warranty claim statistics
type WarrantyClaimStatsResponse struct {
	TotalClaims       int                        `json:"total_claims" example:"1250"`
	ClaimsByStatus    map[string]int             `json:"claims_by_status" example:"{\"pending\":45,\"validated\":120,\"in_repair\":85}"`
	ClaimsBySeverity  map[string]int             `json:"claims_by_severity" example:"{\"low\":300,\"medium\":650,\"high\":250,\"critical\":50}"`
	ClaimsByCategory  map[string]int             `json:"claims_by_category" example:"{\"hardware\":500,\"software\":300,\"defect\":200}"`
	ClaimsByPriority  map[string]int             `json:"claims_by_priority" example:"{\"normal\":800,\"high\":300,\"urgent\":150}"`
	AverageProcessingTime decimal.Decimal        `json:"average_processing_time" example:"72.5"`
	TotalRepairCost   decimal.Decimal            `json:"total_repair_cost" example:"15750.25"`
	TotalShippingCost decimal.Decimal            `json:"total_shipping_cost" example:"2340.50"`
	TotalCost         decimal.Decimal            `json:"total_cost" example:"18090.75"`
	SatisfactionRating decimal.Decimal           `json:"satisfaction_rating" example:"4.2"`
	ClaimsThisMonth   int                        `json:"claims_this_month" example:"125"`
	ClaimsLastMonth   int                        `json:"claims_last_month" example:"98"`
	GrowthRate        decimal.Decimal            `json:"growth_rate" example:"27.55"`
	TopIssueCategories []CategoryStatsResponse   `json:"top_issue_categories"`
	RecentClaims      []WarrantyClaimResponse    `json:"recent_claims"`
}

// CategoryStatsResponse represents statistics for a specific category
type CategoryStatsResponse struct {
	Category string `json:"category" example:"hardware"`
	Count    int    `json:"count" example:"500"`
	Percentage decimal.Decimal `json:"percentage" example:"40.0"`
}

// ClaimAttachmentResponse represents a claim attachment
type ClaimAttachmentResponse struct {
	ID                  string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ClaimID             string    `json:"claim_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	FileName            string    `json:"file_name" example:"receipt.pdf"`
	FilePath            string    `json:"file_path" example:"/uploads/claims/2024/01/receipt.pdf"`
	FileURL             string    `json:"file_url" example:"https://cdn.example.com/claims/receipt.pdf"`
	FileSize            int64     `json:"file_size" example:"1048576"`
	FileType            string    `json:"file_type" example:"pdf"`
	MimeType            string    `json:"mime_type" example:"application/pdf"`
	AttachmentType      string    `json:"attachment_type" example:"receipt"`
	Description         *string   `json:"description,omitempty" example:"Purchase receipt"`
	UploadedBy          string    `json:"uploaded_by" example:"550e8400-e29b-41d4-a716-446655440002"`
	SecurityScanStatus  string    `json:"security_scan_status" example:"passed"`
	SecurityScanResult  *string   `json:"security_scan_result,omitempty" example:"No threats detected"`
	CreatedAt           time.Time `json:"created_at" example:"2024-01-20T14:30:00Z"`
	UpdatedAt           time.Time `json:"updated_at" example:"2024-01-20T14:30:00Z"`
}

// ClaimAttachmentUploadRequest represents a request to upload claim attachment
type ClaimAttachmentUploadRequest struct {
	AttachmentType string  `json:"attachment_type" validate:"required,oneof=receipt photo video document other" example:"receipt"`
	Description    *string `json:"description,omitempty" example:"Purchase receipt"`
}

// Note: PaginationResponse is defined in admin_user_dto.go and reused here