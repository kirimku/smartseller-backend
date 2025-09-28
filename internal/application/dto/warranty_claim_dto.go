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

// ClaimTimelineResponse represents a claim timeline entry
type ClaimTimelineResponse struct {
	ID          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ClaimID     string    `json:"claim_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	EventType   string    `json:"event_type" example:"claim_submitted"`
	Description string    `json:"description" example:"Claim submitted by customer"`
	CreatedBy   string    `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440002"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-20T14:30:00Z"`
	IsVisible   bool      `json:"is_visible" example:"true"`
}

// WarrantyClaimDetailResponse represents detailed claim information with timeline and attachments
type WarrantyClaimDetailResponse struct {
	Claim       *WarrantyClaimResponse      `json:"claim"`
	Timeline    []*ClaimTimelineResponse    `json:"timeline"`
	Attachments []*ClaimAttachmentResponse  `json:"attachments"`
}

// ClaimCompletionRequest represents a request to complete a claim
type ClaimCompletionRequest struct {
	Resolution  string `json:"resolution" validate:"required,oneof=repaired replaced refunded" example:"repaired"`
	Notes       string `json:"notes" validate:"omitempty,max=1000" example:"Successfully repaired the device"`
	RepairNotes string `json:"repair_notes" validate:"omitempty,max=2000" example:"Replaced faulty component and tested functionality"`
}

// ClaimNotesRequest represents a request to add notes to a claim
type ClaimNotesRequest struct {
	Notes     string `json:"notes" validate:"required,min=1,max=1000" example:"Customer contacted for additional information"`
	IsVisible bool   `json:"is_visible" example:"true"`
}

// ClaimTimelineCreateRequest represents a request to create a timeline entry
type ClaimTimelineCreateRequest struct {
	EventType   string `json:"event_type" validate:"required,oneof=claim_submitted claim_validated claim_rejected claim_assigned repair_started repair_completed claim_completed status_updated note_added attachment_uploaded" example:"note_added"`
	Description string `json:"description" validate:"required,min=1,max=1000" example:"Customer contacted for additional information"`
	IsVisible   bool   `json:"is_visible" example:"true"`
}

// ClaimAttachmentListResponse represents a list of claim attachments
type ClaimAttachmentListResponse struct {
	Attachments []*ClaimAttachmentResponse `json:"attachments"`
	Total       int                        `json:"total" example:"5"`
}

// ClaimTimelineListResponse represents a list of claim timeline entries
type ClaimTimelineListResponse struct {
	Timeline []*ClaimTimelineResponse `json:"timeline"`
	Total    int                      `json:"total" example:"10"`
}

// AttachmentApprovalRequest represents a request to approve/reject an attachment
type AttachmentApprovalRequest struct {
	Action string  `json:"action" validate:"required,oneof=approve reject" example:"approve"`
	Notes  *string `json:"notes,omitempty" example:"Attachment verified and approved"`
}

// Type aliases for backward compatibility and cleaner handler code
type ClaimValidationRequest = WarrantyClaimValidationRequest
type ClaimRejectionRequest = WarrantyClaimValidationRequest
type TechnicianAssignmentRequest = WarrantyClaimAssignmentRequest

// ===== REPAIR TICKET DTOs =====

// RepairTicketCreateRequest represents a request to create a repair ticket
type RepairTicketCreateRequest struct {
	ClaimID             string          `json:"claim_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	Priority            string          `json:"priority" validate:"required,oneof=low normal high urgent" example:"high"`
	EstimatedHours      decimal.Decimal `json:"estimated_hours" validate:"required,min=0.1,max=1000" example:"4.5"`
	Description         string          `json:"description" validate:"required,min=10,max=2000" example:"Replace faulty motherboard and test all components"`
	RequiredParts       []string        `json:"required_parts,omitempty" example:"motherboard,thermal_paste"`
	SpecialInstructions string          `json:"special_instructions" validate:"omitempty,max=1000" example:"Handle with care - customer reported water damage"`
	CustomerApprovalRequired bool       `json:"customer_approval_required" example:"true"`
}

// RepairTicketResponse represents a repair ticket
type RepairTicketResponse struct {
	ID                       string          `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TicketNumber             string          `json:"ticket_number" example:"RPR-2024-001234"`
	ClaimID                  string          `json:"claim_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ClaimNumber              string          `json:"claim_number" example:"WAR-2024-001234"`
	Status                   string          `json:"status" example:"assigned"`
	Priority                 string          `json:"priority" example:"high"`
	
	// Technician Information
	AssignedTechnicianID     *string         `json:"assigned_technician_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	TechnicianName           *string         `json:"technician_name,omitempty" example:"John Smith"`
	AssignedAt               *time.Time      `json:"assigned_at,omitempty" example:"2024-01-20T09:00:00Z"`
	
	// Timing Information
	EstimatedHours           decimal.Decimal `json:"estimated_hours" example:"4.5"`
	ActualHours              *decimal.Decimal `json:"actual_hours,omitempty" example:"5.2"`
	EstimatedCompletionDate  *time.Time      `json:"estimated_completion_date,omitempty" example:"2024-01-22T17:00:00Z"`
	ActualCompletionDate     *time.Time      `json:"actual_completion_date,omitempty" example:"2024-01-22T16:30:00Z"`
	
	// Repair Details
	Description              string          `json:"description" example:"Replace faulty motherboard and test all components"`
	RequiredParts            []string        `json:"required_parts,omitempty" example:"motherboard,thermal_paste"`
	UsedParts                []string        `json:"used_parts,omitempty" example:"motherboard,thermal_paste,screws"`
	SpecialInstructions      string          `json:"special_instructions" example:"Handle with care - customer reported water damage"`
	RepairNotes              *string         `json:"repair_notes,omitempty" example:"Successfully replaced motherboard, all tests passed"`
	
	// Cost Information
	LaborCost                decimal.Decimal `json:"labor_cost" example:"120.00"`
	PartsCost                decimal.Decimal `json:"parts_cost" example:"85.50"`
	TotalCost                decimal.Decimal `json:"total_cost" example:"205.50"`
	
	// Quality Control
	QualityCheckStatus       string          `json:"quality_check_status" example:"pending"`
	QualityCheckedBy         *string         `json:"quality_checked_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	QualityCheckDate         *time.Time      `json:"quality_check_date,omitempty" example:"2024-01-22T18:00:00Z"`
	QualityCheckNotes        *string         `json:"quality_check_notes,omitempty" example:"All functionality verified, repair approved"`
	
	// Customer Approval
	CustomerApprovalRequired bool            `json:"customer_approval_required" example:"true"`
	CustomerApprovalStatus   string          `json:"customer_approval_status" example:"pending"`
	CustomerApprovedAt       *time.Time      `json:"customer_approved_at,omitempty" example:"2024-01-21T14:30:00Z"`
	CustomerApprovalNotes    *string         `json:"customer_approval_notes,omitempty" example:"Customer approved repair cost estimate"`
	
	// Metadata
	CreatedBy                string          `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440004"`
	CreatedAt                time.Time       `json:"created_at" example:"2024-01-20T08:30:00Z"`
	UpdatedAt                time.Time       `json:"updated_at" example:"2024-01-22T16:30:00Z"`
}

// RepairTicketUpdateRequest represents a request to update a repair ticket
type RepairTicketUpdateRequest struct {
	Priority            *string          `json:"priority,omitempty" validate:"omitempty,oneof=low normal high urgent" example:"high"`
	EstimatedHours      *decimal.Decimal `json:"estimated_hours,omitempty" validate:"omitempty,min=0.1,max=1000" example:"4.5"`
	Description         *string          `json:"description,omitempty" validate:"omitempty,min=10,max=2000" example:"Replace faulty motherboard and test all components"`
	RequiredParts       []string         `json:"required_parts,omitempty" example:"motherboard,thermal_paste"`
	UsedParts           []string         `json:"used_parts,omitempty" example:"motherboard,thermal_paste,screws"`
	SpecialInstructions *string          `json:"special_instructions,omitempty" validate:"omitempty,max=1000" example:"Handle with care - customer reported water damage"`
	RepairNotes         *string          `json:"repair_notes,omitempty" validate:"omitempty,max=2000" example:"Successfully replaced motherboard, all tests passed"`
	LaborCost           *decimal.Decimal `json:"labor_cost,omitempty" validate:"omitempty,min=0" example:"120.00"`
	PartsCost           *decimal.Decimal `json:"parts_cost,omitempty" validate:"omitempty,min=0" example:"85.50"`
}

// RepairTicketAssignmentRequest represents a request to assign a technician to a repair ticket
type RepairTicketAssignmentRequest struct {
	TechnicianID            string     `json:"technician_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440002"`
	EstimatedCompletionDate *time.Time `json:"estimated_completion_date,omitempty" example:"2024-01-22T17:00:00Z"`
	Notes                   string     `json:"notes" validate:"omitempty,max=1000" example:"Assigned to senior technician for complex repair"`
}

// RepairTicketCompletionRequest represents a request to complete a repair ticket
type RepairTicketCompletionRequest struct {
	ActualHours   decimal.Decimal `json:"actual_hours" validate:"required,min=0.1,max=1000" example:"5.2"`
	UsedParts     []string        `json:"used_parts,omitempty" example:"motherboard,thermal_paste,screws"`
	RepairNotes   string          `json:"repair_notes" validate:"required,min=10,max=2000" example:"Successfully replaced motherboard, all tests passed"`
	LaborCost     decimal.Decimal `json:"labor_cost" validate:"required,min=0" example:"120.00"`
	PartsCost     decimal.Decimal `json:"parts_cost" validate:"required,min=0" example:"85.50"`
	TestResults   string          `json:"test_results" validate:"omitempty,max=1000" example:"All functionality tests passed"`
}

// RepairTicketQualityCheckRequest represents a request for quality control approval
type RepairTicketQualityCheckRequest struct {
	Action string  `json:"action" validate:"required,oneof=approve reject" example:"approve"`
	Notes  string  `json:"notes" validate:"required,min=10,max=1000" example:"All functionality verified, repair approved"`
}

// RepairTicketListRequest represents a request to list repair tickets with filters
type RepairTicketListRequest struct {
	// Pagination
	Page     int `json:"page" validate:"min=1" example:"1"`
	PageSize int `json:"page_size" validate:"min=1,max=100" example:"20"`
	
	// Sorting
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=created_at priority status estimated_completion_date" example:"created_at"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc" example:"desc"`
	
	// Filters
	Status               []string   `json:"status,omitempty" example:"assigned,in_progress"`
	Priority             []string   `json:"priority,omitempty" example:"high,urgent"`
	TechnicianID         *string    `json:"technician_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	ClaimID              *string    `json:"claim_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	QualityCheckStatus   []string   `json:"quality_check_status,omitempty" example:"pending,approved"`
	CreatedDateFrom      *time.Time `json:"created_date_from,omitempty" example:"2024-01-01T00:00:00Z"`
	CreatedDateTo        *time.Time `json:"created_date_to,omitempty" example:"2024-01-31T23:59:59Z"`
	EstimatedCompletionFrom *time.Time `json:"estimated_completion_from,omitempty" example:"2024-01-20T00:00:00Z"`
	EstimatedCompletionTo   *time.Time `json:"estimated_completion_to,omitempty" example:"2024-01-25T23:59:59Z"`
	SearchTerm           string     `json:"search_term,omitempty" example:"motherboard"`
}

// RepairTicketListResponse represents a paginated list of repair tickets
type RepairTicketListResponse struct {
	Tickets    []RepairTicketResponse `json:"tickets"`
	Pagination PaginationResponse     `json:"pagination"`
	Filters    RepairTicketFiltersResponse `json:"filters"`
}

// RepairTicketFiltersResponse represents available filter options for repair tickets
type RepairTicketFiltersResponse struct {
	AvailableStatuses          []string `json:"available_statuses" example:"pending,assigned,in_progress,completed,cancelled"`
	AvailablePriorities        []string `json:"available_priorities" example:"low,normal,high,urgent"`
	AvailableQualityCheckStatuses []string `json:"available_quality_check_statuses" example:"pending,approved,rejected"`
}

// RepairTicketStatisticsResponse represents repair ticket analytics and statistics
type RepairTicketStatisticsResponse struct {
	TotalTickets              int                        `json:"total_tickets" example:"450"`
	TicketsByStatus           map[string]int             `json:"tickets_by_status" example:"{\"assigned\":120,\"in_progress\":85,\"completed\":200}"`
	TicketsByPriority         map[string]int             `json:"tickets_by_priority" example:"{\"normal\":200,\"high\":150,\"urgent\":100}"`
	TicketsByTechnician       map[string]int             `json:"tickets_by_technician" example:"{\"John Smith\":45,\"Jane Doe\":38}"`
	AverageRepairTime         decimal.Decimal            `json:"average_repair_time" example:"4.8"`
	AverageLaborCost          decimal.Decimal            `json:"average_labor_cost" example:"95.50"`
	AveragePartsCost          decimal.Decimal            `json:"average_parts_cost" example:"67.25"`
	TotalLaborCost            decimal.Decimal            `json:"total_labor_cost" example:"42975.00"`
	TotalPartsCost            decimal.Decimal            `json:"total_parts_cost" example:"30262.50"`
	TotalRepairCost           decimal.Decimal            `json:"total_repair_cost" example:"73237.50"`
	CompletionRate            decimal.Decimal            `json:"completion_rate" example:"88.9"`
	QualityApprovalRate       decimal.Decimal            `json:"quality_approval_rate" example:"94.2"`
	CustomerApprovalRate      decimal.Decimal            `json:"customer_approval_rate" example:"91.5"`
	TicketsThisMonth          int                        `json:"tickets_this_month" example:"85"`
	TicketsLastMonth          int                        `json:"tickets_last_month" example:"72"`
	GrowthRate                decimal.Decimal            `json:"growth_rate" example:"18.1"`
	TopTechnicians            []TechnicianStatsResponse  `json:"top_technicians"`
	RecentTickets             []RepairTicketResponse     `json:"recent_tickets"`
}

// TechnicianStatsResponse represents statistics for a specific technician
type TechnicianStatsResponse struct {
	TechnicianID       string          `json:"technician_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	TechnicianName     string          `json:"technician_name" example:"John Smith"`
	AssignedTickets    int             `json:"assigned_tickets" example:"45"`
	CompletedTickets   int             `json:"completed_tickets" example:"42"`
	CompletionRate     decimal.Decimal `json:"completion_rate" example:"93.3"`
	AverageRepairTime  decimal.Decimal `json:"average_repair_time" example:"4.2"`
	QualityApprovalRate decimal.Decimal `json:"quality_approval_rate" example:"97.6"`
	TotalRevenue       decimal.Decimal `json:"total_revenue" example:"8950.00"`
}

// Note: PaginationResponse is defined in admin_user_dto.go and reused here

// ===== BATCH GENERATION DTOs =====

// BatchCreateRequest represents a request to create a new batch generation
type BatchCreateRequest struct {
	ProductID       string          `json:"product_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440000"`
	StorefrontID    string          `json:"storefront_id" validate:"required,uuid" example:"550e8400-e29b-41d4-a716-446655440001"`
	Quantity        int             `json:"quantity" validate:"required,min=1,max=100000" example:"1000"`
	Prefix          string          `json:"prefix" validate:"required,min=2,max=10" example:"WB"`
	Description     string          `json:"description" validate:"required,min=5,max=500" example:"Batch generation for Q1 2024 smartphone warranty barcodes"`
	ExpiryMonths    int             `json:"expiry_months" validate:"required,min=1,max=120" example:"24"`
	Priority        string          `json:"priority" validate:"required,oneof=low normal high urgent" example:"normal"`
	NotifyOnComplete bool           `json:"notify_on_complete" example:"true"`
	Tags            []string        `json:"tags,omitempty" example:"q1-2024,smartphone"`
	Metadata        map[string]interface{} `json:"metadata,omitempty" example:"{\"department\":\"sales\",\"campaign\":\"spring-2024\"}"`
}

// WarrantyBatchResponse represents a batch generation
type WarrantyBatchResponse struct {
	ID                    string                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchNumber           string                 `json:"batch_number" example:"BATCH-2024-001234"`
	ProductID             string                 `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ProductName           string                 `json:"product_name" example:"Smartphone XYZ"`
	ProductSKU            string                 `json:"product_sku" example:"SKU-12345"`
	StorefrontID          string                 `json:"storefront_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	StorefrontName        string                 `json:"storefront_name" example:"Tech Store ABC"`
	
	// Configuration
	RequestedQuantity     int                    `json:"requested_quantity" example:"1000"`
	GeneratedQuantity     int                    `json:"generated_quantity" example:"850"`
	SuccessfulQuantity    int                    `json:"successful_quantity" example:"845"`
	FailedQuantity        int                    `json:"failed_quantity" example:"5"`
	Prefix                string                 `json:"prefix" example:"WB"`
	Description           string                 `json:"description" example:"Batch generation for Q1 2024 smartphone warranty barcodes"`
	ExpiryMonths          int                    `json:"expiry_months" example:"24"`
	Priority              string                 `json:"priority" example:"normal"`
	
	// Status and Progress
	Status                string                 `json:"status" example:"in_progress"`
	Progress              decimal.Decimal        `json:"progress" example:"85.0"`
	EstimatedCompletion   *time.Time             `json:"estimated_completion,omitempty" example:"2024-01-20T16:30:00Z"`
	ActualCompletion      *time.Time             `json:"actual_completion,omitempty" example:"2024-01-20T16:25:00Z"`
	
	// Performance Metrics
	ProcessingTimeSeconds *decimal.Decimal       `json:"processing_time_seconds,omitempty" example:"1245.5"`
	GenerationRate        *decimal.Decimal       `json:"generation_rate,omitempty" example:"0.68"`
	ErrorRate             decimal.Decimal        `json:"error_rate" example:"0.59"`
	
	// Error Handling
	ErrorCount            int                    `json:"error_count" example:"5"`
	LastError             *string                `json:"last_error,omitempty" example:"Database connection timeout"`
	RetryCount            int                    `json:"retry_count" example:"2"`
	MaxRetries            int                    `json:"max_retries" example:"3"`
	
	// Collision Detection
	CollisionCount        int                    `json:"collision_count" example:"12"`
	CollisionResolution   string                 `json:"collision_resolution" example:"regenerate"`
	
	// Metadata
	Tags                  []string               `json:"tags,omitempty" example:"q1-2024,smartphone"`
	Metadata              map[string]interface{} `json:"metadata,omitempty" example:"{\"department\":\"sales\",\"campaign\":\"spring-2024\"}"`
	NotifyOnComplete      bool                   `json:"notify_on_complete" example:"true"`
	NotificationSent      bool                   `json:"notification_sent" example:"false"`
	
	// Audit Information
	CreatedBy             string                 `json:"created_by" example:"550e8400-e29b-41d4-a716-446655440003"`
	CreatedByName         string                 `json:"created_by_name" example:"John Admin"`
	CreatedAt             time.Time              `json:"created_at" example:"2024-01-20T10:00:00Z"`
	StartedAt             *time.Time             `json:"started_at,omitempty" example:"2024-01-20T10:05:00Z"`
	CompletedAt           *time.Time             `json:"completed_at,omitempty" example:"2024-01-20T16:25:00Z"`
	CancelledAt           *time.Time             `json:"cancelled_at,omitempty" example:"2024-01-20T15:30:00Z"`
	CancelledBy           *string                `json:"cancelled_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440004"`
	UpdatedAt             time.Time              `json:"updated_at" example:"2024-01-20T16:25:00Z"`
}

// BatchProgressResponse represents real-time batch progress
type BatchProgressResponse struct {
	BatchID               string          `json:"batch_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchNumber           string          `json:"batch_number" example:"BATCH-2024-001234"`
	Status                string          `json:"status" example:"in_progress"`
	Progress              decimal.Decimal `json:"progress" example:"85.0"`
	
	// Current Processing
	CurrentStep           string          `json:"current_step" example:"generating_barcodes"`
	ProcessedCount        int             `json:"processed_count" example:"850"`
	RemainingCount        int             `json:"remaining_count" example:"150"`
	EstimatedTimeRemaining *int           `json:"estimated_time_remaining,omitempty" example:"300"`
	
	// Performance
	GenerationRate        decimal.Decimal `json:"generation_rate" example:"2.83"`
	ErrorRate             decimal.Decimal `json:"error_rate" example:"0.59"`
	
	// Real-time Stats
	ErrorCount            int             `json:"error_count" example:"5"`
	CollisionCount        int             `json:"collision_count" example:"12"`
	RetryCount            int             `json:"retry_count" example:"2"`
	
	// Timestamps
	StartedAt             *time.Time      `json:"started_at,omitempty" example:"2024-01-20T10:05:00Z"`
	LastUpdated           time.Time       `json:"last_updated" example:"2024-01-20T15:45:00Z"`
	EstimatedCompletion   *time.Time      `json:"estimated_completion,omitempty" example:"2024-01-20T16:30:00Z"`
}

// BatchListRequest represents a request to list batches with filters
type BatchListRequest struct {
	// Pagination
	Page     int `json:"page" validate:"min=1" example:"1"`
	PageSize int `json:"page_size" validate:"min=1,max=100" example:"20"`
	
	// Sorting
	SortBy    string `json:"sort_by" validate:"omitempty,oneof=created_at batch_number status priority progress completion_time" example:"created_at"`
	SortOrder string `json:"sort_order" validate:"omitempty,oneof=asc desc" example:"desc"`
	
	// Filters
	Status           []string   `json:"status,omitempty" example:"in_progress,completed"`
	Priority         []string   `json:"priority,omitempty" example:"normal,high"`
	ProductID        *string    `json:"product_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	StorefrontID     *string    `json:"storefront_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	CreatedBy        *string    `json:"created_by,omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	CreatedDateFrom  *time.Time `json:"created_date_from,omitempty" example:"2024-01-01T00:00:00Z"`
	CreatedDateTo    *time.Time `json:"created_date_to,omitempty" example:"2024-01-31T23:59:59Z"`
	CompletedDateFrom *time.Time `json:"completed_date_from,omitempty" example:"2024-01-01T00:00:00Z"`
	CompletedDateTo   *time.Time `json:"completed_date_to,omitempty" example:"2024-01-31T23:59:59Z"`
	SearchTerm       string     `json:"search_term,omitempty" example:"smartphone"`
	Tags             []string   `json:"tags,omitempty" example:"q1-2024,smartphone"`
	HasErrors        *bool      `json:"has_errors,omitempty" example:"false"`
	HasCollisions    *bool      `json:"has_collisions,omitempty" example:"true"`
}

// BatchListResponse represents a paginated list of batches
type BatchListResponse struct {
	Batches    []BatchResponse        `json:"batches"`
	Pagination PaginationResponse     `json:"pagination"`
	Filters    BatchFiltersResponse   `json:"filters"`
	Statistics BatchListStatistics   `json:"statistics"`
}

// BatchFiltersResponse represents available filter options
type BatchFiltersResponse struct {
	AvailableStatuses   []string `json:"available_statuses" example:"pending,in_progress,completed,cancelled,failed"`
	AvailablePriorities []string `json:"available_priorities" example:"low,normal,high,urgent"`
	AvailableProducts   []ClaimProductSummary `json:"available_products"`
	AvailableStorefronts []StorefrontSummary `json:"available_storefronts"`
}

// ClaimProductSummary represents a simplified product for filters
type ClaimProductSummary struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name string `json:"name" example:"Smartphone XYZ"`
	SKU  string `json:"sku" example:"SKU-12345"`
}

// StorefrontSummary represents a simplified storefront for filters
type StorefrontSummary struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Name string `json:"name" example:"Tech Store ABC"`
}

// BatchListStatistics represents summary statistics for the batch list
type BatchListStatistics struct {
	TotalBatches        int             `json:"total_batches" example:"156"`
	ActiveBatches       int             `json:"active_batches" example:"12"`
	CompletedBatches    int             `json:"completed_batches" example:"140"`
	FailedBatches       int             `json:"failed_batches" example:"4"`
	TotalGenerated      int             `json:"total_generated" example:"1250000"`
	TotalErrors         int             `json:"total_errors" example:"1250"`
	AverageProgress     decimal.Decimal `json:"average_progress" example:"87.5"`
	AverageGenerationRate decimal.Decimal `json:"average_generation_rate" example:"2.45"`
}

// BatchCollisionResponse represents a barcode collision in a batch
type BatchCollisionResponse struct {
	ID              string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchID         string    `json:"batch_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	BarcodeValue    string    `json:"barcode_value" example:"WB-2024-ABC123DEF456"`
	CollisionType   string    `json:"collision_type" example:"duplicate_in_batch"`
	ExistingBarcodeID *string `json:"existing_barcode_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Resolution      string    `json:"resolution" example:"regenerated"`
	ResolvedAt      *time.Time `json:"resolved_at,omitempty" example:"2024-01-20T15:30:00Z"`
	CreatedAt       time.Time `json:"created_at" example:"2024-01-20T15:25:00Z"`
}

// BatchCollisionListResponse represents a list of collisions for a batch
type BatchCollisionListResponse struct {
	Collisions []BatchCollisionResponse `json:"collisions"`
	Total      int                      `json:"total" example:"12"`
	Resolved   int                      `json:"resolved" example:"10"`
	Pending    int                      `json:"pending" example:"2"`
}

// BatchStatisticsResponse represents comprehensive batch analytics
type BatchStatisticsResponse struct {
	// Overview
	TotalBatches          int                    `json:"total_batches" example:"156"`
	BatchesByStatus       map[string]int         `json:"batches_by_status" example:"{\"completed\":140,\"in_progress\":12,\"failed\":4}"`
	BatchesByPriority     map[string]int         `json:"batches_by_priority" example:"{\"normal\":120,\"high\":30,\"urgent\":6}"`
	
	// Generation Statistics
	TotalGenerated        int                    `json:"total_generated" example:"1250000"`
	TotalErrors           int                    `json:"total_errors" example:"1250"`
	TotalCollisions       int                    `json:"total_collisions" example:"850"`
	AverageGenerationRate decimal.Decimal        `json:"average_generation_rate" example:"2.45"`
	AverageErrorRate      decimal.Decimal        `json:"average_error_rate" example:"0.1"`
	AverageCollisionRate  decimal.Decimal        `json:"average_collision_rate" example:"0.068"`
	
	// Performance Metrics
	AverageProcessingTime decimal.Decimal        `json:"average_processing_time" example:"1245.5"`
	FastestBatch          *decimal.Decimal       `json:"fastest_batch,omitempty" example:"450.2"`
	SlowestBatch          *decimal.Decimal       `json:"slowest_batch,omitempty" example:"3600.8"`
	
	// Time-based Analytics
	BatchesThisMonth      int                    `json:"batches_this_month" example:"25"`
	BatchesLastMonth      int                    `json:"batches_last_month" example:"18"`
	GrowthRate            decimal.Decimal        `json:"growth_rate" example:"38.9"`
	GeneratedThisMonth    int                    `json:"generated_this_month" example:"250000"`
	GeneratedLastMonth    int                    `json:"generated_last_month" example:"180000"`
	
	// Top Performers
	TopProducts           []ProductBatchStats    `json:"top_products"`
	TopStorefronts        []StorefrontBatchStats `json:"top_storefronts"`
	TopCreators           []UserBatchStats       `json:"top_creators"`
	
	// Recent Activity
	RecentBatches         []BatchResponse        `json:"recent_batches"`
	ActiveBatches         []BatchResponse        `json:"active_batches"`
}

// ProductBatchStats represents batch statistics for a product
type ProductBatchStats struct {
	ProductID       string          `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	ProductName     string          `json:"product_name" example:"Smartphone XYZ"`
	ProductSKU      string          `json:"product_sku" example:"SKU-12345"`
	BatchCount      int             `json:"batch_count" example:"45"`
	TotalGenerated  int             `json:"total_generated" example:"450000"`
	AverageSize     decimal.Decimal `json:"average_size" example:"10000"`
	SuccessRate     decimal.Decimal `json:"success_rate" example:"99.2"`
}

// StorefrontBatchStats represents batch statistics for a storefront
type StorefrontBatchStats struct {
	StorefrontID    string          `json:"storefront_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	StorefrontName  string          `json:"storefront_name" example:"Tech Store ABC"`
	BatchCount      int             `json:"batch_count" example:"38"`
	TotalGenerated  int             `json:"total_generated" example:"380000"`
	AverageSize     decimal.Decimal `json:"average_size" example:"10000"`
	SuccessRate     decimal.Decimal `json:"success_rate" example:"98.8"`
}

// UserBatchStats represents batch statistics for a user
type UserBatchStats struct {
	UserID          string          `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	UserName        string          `json:"user_name" example:"John Admin"`
	BatchCount      int             `json:"batch_count" example:"28"`
	TotalGenerated  int             `json:"total_generated" example:"280000"`
	AverageSize     decimal.Decimal `json:"average_size" example:"10000"`
	SuccessRate     decimal.Decimal `json:"success_rate" example:"99.5"`
}

// BatchCancelRequest represents a request to cancel a batch
type BatchCancelRequest struct {
	Reason string `json:"reason" validate:"required,min=5,max=500" example:"Customer request - project cancelled"`
	Force  bool   `json:"force" example:"false"`
}