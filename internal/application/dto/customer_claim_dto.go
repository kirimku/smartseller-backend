package dto

import (
	"time"
	"github.com/shopspring/decimal"
)

// CustomerClaimSubmissionRequest represents the request to submit a new warranty claim
type CustomerClaimSubmissionRequest struct {
	WarrantyID    string                    `json:"warranty_id" validate:"required"`
	IssueType     string                    `json:"issue_type" validate:"required,oneof=defect damage malfunction performance other"`
	Description   string                    `json:"description" validate:"required,min=10,max=1000"`
	Severity      string                    `json:"severity" validate:"required,oneof=low medium high critical"`
	ContactInfo   CustomerClaimContactInfo  `json:"contact_info" validate:"required"`
	ProductInfo   CustomerClaimProductInfo  `json:"product_info" validate:"required"`
	Attachments   []CustomerClaimAttachment `json:"attachments,omitempty"`
	PreferredResolution string              `json:"preferred_resolution,omitempty" validate:"omitempty,oneof=repair replace refund"`
}

// CustomerClaimContactInfo represents customer contact information for claim
type CustomerClaimContactInfo struct {
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone" validate:"required"`
	Address      string `json:"address" validate:"required"`
	City         string `json:"city" validate:"required"`
	PostalCode   string `json:"postal_code" validate:"required"`
	PreferredContact string `json:"preferred_contact" validate:"required,oneof=email phone sms"`
}

// CustomerClaimProductInfo represents product information for claim
type CustomerClaimProductInfo struct {
	SerialNumber     string    `json:"serial_number" validate:"required"`
	PurchaseDate     time.Time `json:"purchase_date" validate:"required"`
	PurchaseLocation string    `json:"purchase_location" validate:"required"`
	UsageFrequency   string    `json:"usage_frequency" validate:"required,oneof=daily weekly monthly occasional"`
	Environment      string    `json:"environment" validate:"required,oneof=indoor outdoor mixed"`
}

// CustomerClaimAttachment represents file attachments for claim
type CustomerClaimAttachment struct {
	Type        string `json:"type" validate:"required,oneof=photo video document receipt"`
	FileName    string `json:"file_name" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=1"`
	ContentType string `json:"content_type" validate:"required"`
	URL         string `json:"url,omitempty"`
	Description string `json:"description,omitempty"`
}

// CustomerClaimSubmissionResponse represents the response after submitting a claim
type CustomerClaimSubmissionResponse struct {
	ClaimID       string                   `json:"claim_id"`
	ClaimNumber   string                   `json:"claim_number"`
	Status        string                   `json:"status"`
	SubmittedAt   time.Time                `json:"submitted_at"`
	EstimatedResolution time.Time          `json:"estimated_resolution"`
	Priority      string                   `json:"priority"`
	AssignedAgent string                   `json:"assigned_agent,omitempty"`
	NextSteps     []string                 `json:"next_steps"`
	ContactInfo   CustomerClaimContactInfo `json:"contact_info"`
	TrackingInfo  CustomerClaimTrackingInfo `json:"tracking_info"`
}

// CustomerClaimTrackingInfo represents tracking information for claim
type CustomerClaimTrackingInfo struct {
	TrackingNumber string `json:"tracking_number"`
	StatusURL      string `json:"status_url"`
	SupportEmail   string `json:"support_email"`
	SupportPhone   string `json:"support_phone"`
}

// CustomerClaimListRequest represents request parameters for listing claims
type CustomerClaimListRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	Limit    int    `json:"limit" validate:"min=1,max=100"`
	Status   string `json:"status,omitempty" validate:"omitempty,oneof=submitted under_review approved rejected in_progress resolved closed"`
	IssueType string `json:"issue_type,omitempty" validate:"omitempty,oneof=defect damage malfunction performance other"`
	Severity  string `json:"severity,omitempty" validate:"omitempty,oneof=low medium high critical"`
	SortBy    string `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at updated_at status priority"`
	SortOrder string `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// CustomerClaimListResponse represents the response for listing claims
type CustomerClaimListResponse struct {
	Claims      []CustomerClaimInfo   `json:"claims"`
	TotalCount  int                   `json:"total_count" example:"25"`
	Page        int                   `json:"page" example:"1"`
	Limit       int                   `json:"limit" example:"10"`
	TotalPages  int                   `json:"total_pages" example:"3"`
	HasNext     bool                  `json:"has_next" example:"true"`
	HasPrevious bool                  `json:"has_previous" example:"false"`
	Summary     CustomerClaimsSummary `json:"summary"`
	RequestTime time.Time             `json:"request_time"`
}

// CustomerClaimInfo represents a summary view of a claim for listing
type CustomerClaimInfo struct {
	ClaimID       string    `json:"claim_id"`
	ClaimNumber   string    `json:"claim_number"`
	Status        string    `json:"status"`
	IssueType     string    `json:"issue_type"`
	Severity      string    `json:"severity"`
	SubmittedAt   time.Time `json:"submitted_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ProductName   string    `json:"product_name"`
	ProductSKU    string    `json:"product_sku"`
	WarrantyID    string    `json:"warranty_id"`
	Priority      string    `json:"priority"`
	DaysOpen      int       `json:"days_open"`
	LastActivity  string    `json:"last_activity"`
}

// CustomerClaimsSummary represents summary statistics for claims
type CustomerClaimsSummary struct {
	TotalClaims    int `json:"total_claims"`
	OpenClaims     int `json:"open_claims"`
	ResolvedClaims int `json:"resolved_claims"`
	PendingClaims  int `json:"pending_claims"`
}

// CustomerClaimDetailResponse represents detailed information about a claim
type CustomerClaimDetailResponse struct {
	ClaimID         string                      `json:"claim_id"`
	ClaimNumber     string                      `json:"claim_number"`
	Status          string                      `json:"status"`
	IssueType       string                      `json:"issue_type"`
	Description     string                      `json:"description"`
	Severity        string                      `json:"severity"`
	Priority        string                      `json:"priority"`
	SubmittedAt     time.Time                   `json:"submitted_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	EstimatedResolution time.Time               `json:"estimated_resolution"`
	ActualResolution    *time.Time              `json:"actual_resolution,omitempty"`
	ContactInfo     CustomerClaimContactInfo    `json:"contact_info"`
	ProductInfo     CustomerClaimProductInfo    `json:"product_info"`
	WarrantyInfo    CustomerClaimWarrantyInfo   `json:"warranty_info"`
	Attachments     []CustomerClaimAttachment   `json:"attachments"`
	Timeline        []CustomerClaimTimelineItem `json:"timeline"`
	Resolution      *CustomerClaimResolution    `json:"resolution,omitempty"`
	AssignedAgent   *CustomerClaimAgent         `json:"assigned_agent,omitempty"`
	Communication   []CustomerClaimCommunication `json:"communication"`
}

// CustomerClaimWarrantyInfo represents warranty information for claim
type CustomerClaimWarrantyInfo struct {
	WarrantyID     string    `json:"warranty_id"`
	ProductName    string    `json:"product_name"`
	ProductSKU     string    `json:"product_sku"`
	SerialNumber   string    `json:"serial_number"`
	ActivatedAt    time.Time `json:"activated_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	WarrantyPeriod int       `json:"warranty_period"`
	CoverageType   string    `json:"coverage_type"`
	IsActive       bool      `json:"is_active"`
}

// CustomerClaimTimelineItem represents a timeline event for claim
type CustomerClaimTimelineItem struct {
	ID          string    `json:"id"`
	Event       string    `json:"event"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Actor       string    `json:"actor"`
	ActorType   string    `json:"actor_type"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// CustomerClaimResolution represents the resolution of a claim
type CustomerClaimResolution struct {
	Type           string          `json:"type"`
	Description    string          `json:"description"`
	ResolvedAt     time.Time       `json:"resolved_at"`
	ResolvedBy     string          `json:"resolved_by"`
	Cost           *decimal.Decimal `json:"cost,omitempty"`
	RefundAmount   *decimal.Decimal `json:"refund_amount,omitempty"`
	ReplacementSKU string          `json:"replacement_sku,omitempty"`
	RepairDetails  string          `json:"repair_details,omitempty"`
	Satisfaction   *int            `json:"satisfaction,omitempty"`
}

// CustomerClaimAgent represents assigned agent information
type CustomerClaimAgent struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Avatar   string `json:"avatar,omitempty"`
}

// CustomerClaimCommunication represents communication history
type CustomerClaimCommunication struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Direction string    `json:"direction"`
	Subject   string    `json:"subject,omitempty"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Sender    string    `json:"sender"`
	SenderType string   `json:"sender_type"`
	Read      bool      `json:"read"`
	Attachments []CustomerClaimAttachment `json:"attachments,omitempty"`
}

// CustomerClaimAttachmentUploadRequest represents request to upload attachment
type CustomerClaimAttachmentUploadRequest struct {
	ClaimID     string `json:"claim_id" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=photo video document receipt"`
	FileName    string `json:"file_name" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,min=1"`
	ContentType string `json:"content_type" validate:"required"`
	Description string `json:"description,omitempty"`
}

// CustomerClaimAttachmentUploadResponse represents response after uploading attachment
type CustomerClaimAttachmentUploadResponse struct {
	AttachmentID string `json:"attachment_id"`
	UploadURL    string `json:"upload_url"`
	ExpiresAt    time.Time `json:"expires_at"`
	MaxFileSize  int64  `json:"max_file_size"`
	AllowedTypes []string `json:"allowed_types"`
}

// CustomerClaimFeedbackRequest represents request to submit feedback
type CustomerClaimFeedbackRequest struct {
	ClaimID      string `json:"claim_id" validate:"required"`
	Rating       int    `json:"rating" validate:"required,min=1,max=5"`
	Comment      string `json:"comment,omitempty" validate:"max=500"`
	ServiceRating int   `json:"service_rating" validate:"required,min=1,max=5"`
	SpeedRating   int   `json:"speed_rating" validate:"required,min=1,max=5"`
	QualityRating int   `json:"quality_rating" validate:"required,min=1,max=5"`
	Recommend     bool   `json:"recommend"`
	Improvements  string `json:"improvements,omitempty" validate:"max=500"`
}

// CustomerClaimFeedbackResponse represents response after submitting feedback
type CustomerClaimFeedbackResponse struct {
	FeedbackID   string    `json:"feedback_id"`
	ClaimID      string    `json:"claim_id"`
	SubmittedAt  time.Time `json:"submitted_at"`
	AverageRating float64  `json:"average_rating"`
	ThankYouMessage string `json:"thank_you_message"`
	FollowUpSurvey  *string `json:"follow_up_survey,omitempty"`
}

// CustomerClaimUpdateRequest represents request to update claim information
type CustomerClaimUpdateRequest struct {
	Description         *string                   `json:"description,omitempty" validate:"omitempty,min=10,max=1000"`
	Severity           *string                   `json:"severity,omitempty" validate:"omitempty,oneof=low medium high critical"`
	ContactInfo        *CustomerClaimContactInfo `json:"contact_info,omitempty"`
	PreferredResolution *string                  `json:"preferred_resolution,omitempty" validate:"omitempty,oneof=repair replace refund"`
	AdditionalInfo     *string                   `json:"additional_info,omitempty" validate:"omitempty,max=500"`
}

// CustomerClaimUpdateResponse represents response after updating claim
type CustomerClaimUpdateResponse struct {
	ClaimID     string    `json:"claim_id"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedFields []string `json:"updated_fields"`
	Message     string    `json:"message"`
}

// CustomerClaimStatusResponse represents current status of a claim
type CustomerClaimStatusResponse struct {
	ClaimID             string                   `json:"claim_id"`
	ClaimNumber         string                   `json:"claim_number"`
	Status              string                   `json:"status"`
	StatusDescription   string                   `json:"status_description"`
	Priority            string                   `json:"priority"`
	SubmittedAt         time.Time                `json:"submitted_at"`
	LastUpdated         time.Time                `json:"last_updated"`
	EstimatedResolution time.Time                `json:"estimated_resolution"`
	CurrentStage        string                   `json:"current_stage"`
	StageDescription    string                   `json:"stage_description"`
	Progress            CustomerClaimProgress    `json:"progress"`
	AssignedAgent       *CustomerClaimAgent      `json:"assigned_agent,omitempty"`
	ContactInfo         CustomerClaimContactInfo `json:"contact_info"`
}

// CustomerClaimProgress represents progress tracking for a claim
type CustomerClaimProgress struct {
	CurrentStep int       `json:"current_step"`
	TotalSteps  int       `json:"total_steps"`
	Percentage  int       `json:"percentage"`
	NextStep    string    `json:"next_step"`
	NextStepETA time.Time `json:"next_step_eta"`
}

// CustomerClaimUpdate represents an update/notification for a claim
type CustomerClaimUpdate struct {
	ID             string                 `json:"id"`
	ClaimID        string                 `json:"claim_id"`
	Type           string                 `json:"type"`
	Title          string                 `json:"title"`
	Message        string                 `json:"message"`
	Timestamp      time.Time              `json:"timestamp"`
	IsRead         bool                   `json:"is_read"`
	Priority       string                 `json:"priority"`
	ActionRequired bool                   `json:"action_required"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// CustomerClaimUpdatesResponse represents response for claim updates
type CustomerClaimUpdatesResponse struct {
	ClaimID     string                `json:"claim_id"`
	Updates     []CustomerClaimUpdate `json:"updates"`
	Total       int                   `json:"total"`
	UnreadCount int                   `json:"unread_count"`
	LastChecked time.Time             `json:"last_checked"`
	HasMore     bool                  `json:"has_more"`
}

// CustomerClaimCommunicationRequest represents request to send communication
type CustomerClaimCommunicationRequest struct {
	Type             string                    `json:"type" validate:"required,oneof=message question complaint compliment"`
	Subject          string                    `json:"subject" validate:"required,min=5,max=200"`
	Message          string                    `json:"message" validate:"required,min=10,max=2000"`
	Priority         string                    `json:"priority" validate:"required,oneof=low medium high urgent"`
	ResponseExpected bool                      `json:"response_expected"`
	Attachments      []CustomerClaimAttachment `json:"attachments,omitempty"`
}

// CustomerClaimCommunicationResponse represents response after sending communication
type CustomerClaimCommunicationResponse struct {
	CommunicationID  string                    `json:"communication_id"`
	ClaimID          string                    `json:"claim_id"`
	Type             string                    `json:"type"`
	Subject          string                    `json:"subject"`
	Message          string                    `json:"message"`
	SentAt           time.Time                 `json:"sent_at"`
	Status           string                    `json:"status"`
	DeliveryStatus   string                    `json:"delivery_status"`
	ReadStatus       string                    `json:"read_status"`
	ResponseExpected bool                      `json:"response_expected"`
	Priority         string                    `json:"priority"`
	Attachments      []CustomerClaimAttachment `json:"attachments,omitempty"`
	Metadata         map[string]interface{}    `json:"metadata,omitempty"`
}

// MarkUpdatesReadRequest represents request to mark updates as read
type MarkUpdatesReadRequest struct {
	UpdateIDs []string `json:"update_ids" validate:"required,min=1"`
}

// MarkUpdatesReadResponse represents response after marking updates as read
type MarkUpdatesReadResponse struct {
	ClaimID         string    `json:"claim_id"`
	UpdatesMarked   int       `json:"updates_marked"`
	MarkedAt        time.Time `json:"marked_at"`
	RemainingUnread int       `json:"remaining_unread"`
}