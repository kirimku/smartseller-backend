package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ClaimStatus represents the status of a warranty claim
type ClaimStatus string

const (
	ClaimStatusPending   ClaimStatus = "pending"   // Awaiting validation
	ClaimStatusValidated ClaimStatus = "validated" // Approved for processing
	ClaimStatusRejected  ClaimStatus = "rejected"  // Claim denied
	ClaimStatusAssigned  ClaimStatus = "assigned"  // Assigned to technician
	ClaimStatusInRepair  ClaimStatus = "in_repair" // Being repaired
	ClaimStatusRepaired  ClaimStatus = "repaired"  // Repair completed
	ClaimStatusReplaced  ClaimStatus = "replaced"  // Product replaced
	ClaimStatusShipped   ClaimStatus = "shipped"   // Sent back to customer
	ClaimStatusDelivered ClaimStatus = "delivered" // Customer received
	ClaimStatusCompleted ClaimStatus = "completed" // Case closed
	ClaimStatusCancelled ClaimStatus = "cancelled" // Customer cancelled
	ClaimStatusDisputed  ClaimStatus = "disputed"  // Under review
)

// Valid validates the claim status
func (cs ClaimStatus) Valid() bool {
	switch cs {
	case ClaimStatusPending, ClaimStatusValidated, ClaimStatusRejected, ClaimStatusAssigned,
		ClaimStatusInRepair, ClaimStatusRepaired, ClaimStatusReplaced, ClaimStatusShipped,
		ClaimStatusDelivered, ClaimStatusCompleted, ClaimStatusCancelled, ClaimStatusDisputed:
		return true
	default:
		return false
	}
}

// String returns the string representation of ClaimStatus
func (cs ClaimStatus) String() string {
	return string(cs)
}

// Value implements the driver.Valuer interface for database storage
func (cs ClaimStatus) Value() (driver.Value, error) {
	return string(cs), nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (cs *ClaimStatus) Scan(value interface{}) error {
	if value == nil {
		*cs = ClaimStatusPending
		return nil
	}
	if str, ok := value.(string); ok {
		*cs = ClaimStatus(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ClaimStatus", value)
}

// ClaimSeverity represents the severity level of a claim
type ClaimSeverity string

const (
	ClaimSeverityLow      ClaimSeverity = "low"
	ClaimSeverityMedium   ClaimSeverity = "medium"
	ClaimSeverityCritical ClaimSeverity = "critical"
	ClaimSeverityHigh     ClaimSeverity = "high"
)

// Valid validates the claim severity
func (cs ClaimSeverity) Valid() bool {
	switch cs {
	case ClaimSeverityLow, ClaimSeverityMedium, ClaimSeverityHigh, ClaimSeverityCritical:
		return true
	default:
		return false
	}
}

// ClaimPriority represents the priority level of a claim
type ClaimPriority string

const (
	ClaimPriorityLow    ClaimPriority = "low"
	ClaimPriorityNormal ClaimPriority = "normal"
	ClaimPriorityHigh   ClaimPriority = "high"
	ClaimPriorityUrgent ClaimPriority = "urgent"
)

// Valid validates the claim priority
func (cp ClaimPriority) Valid() bool {
	switch cp {
	case ClaimPriorityLow, ClaimPriorityNormal, ClaimPriorityHigh, ClaimPriorityUrgent:
		return true
	default:
		return false
	}
}

// ResolutionType represents the type of resolution for a claim
type ResolutionType string

const (
	ResolutionTypeRepair   ResolutionType = "repair"
	ResolutionTypeReplace  ResolutionType = "replace"
	ResolutionTypeRefund   ResolutionType = "refund"
	ResolutionTypeRejected ResolutionType = "rejected"
)

// Valid validates the resolution type
func (rt ResolutionType) Valid() bool {
	switch rt {
	case ResolutionTypeRepair, ResolutionTypeReplace, ResolutionTypeRefund, ResolutionTypeRejected:
		return true
	default:
		return false
	}
}

// DeliveryStatus represents the delivery status of shipped items
type DeliveryStatus string

const (
	DeliveryStatusNotShipped     DeliveryStatus = "not_shipped"
	DeliveryStatusPreparing      DeliveryStatus = "preparing"
	DeliveryStatusPickedUp       DeliveryStatus = "picked_up"
	DeliveryStatusInTransit      DeliveryStatus = "in_transit"
	DeliveryStatusOutForDelivery DeliveryStatus = "out_for_delivery"
	DeliveryStatusDelivered      DeliveryStatus = "delivered"
	DeliveryStatusFailedDelivery DeliveryStatus = "failed_delivery"
	DeliveryStatusReturned       DeliveryStatus = "returned"
)

// Valid validates the delivery status
func (ds DeliveryStatus) Valid() bool {
	switch ds {
	case DeliveryStatusNotShipped, DeliveryStatusPreparing, DeliveryStatusPickedUp,
		DeliveryStatusInTransit, DeliveryStatusOutForDelivery, DeliveryStatusDelivered,
		DeliveryStatusFailedDelivery, DeliveryStatusReturned:
		return true
	default:
		return false
	}
}

// Address represents a flexible address structure
type Address struct {
	Street     string                 `json:"street" db:"street"`
	City       string                 `json:"city" db:"city"`
	Province   string                 `json:"province" db:"province"`
	PostalCode string                 `json:"postal_code" db:"postal_code"`
	Country    string                 `json:"country" db:"country"`
	Latitude   *float64               `json:"latitude,omitempty" db:"latitude"`
	Longitude  *float64               `json:"longitude,omitempty" db:"longitude"`
	Notes      *string                `json:"notes,omitempty" db:"notes"`
	Additional map[string]interface{} `json:"additional,omitempty" db:"additional"`
}

// Value implements driver.Valuer interface for database storage
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner interface for database retrieval
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		*a = Address{}
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into Address", value)
	}

	return json.Unmarshal(b, a)
}

// LogisticsInfo represents shipping and logistics information
type LogisticsInfo struct {
	Provider             *string          `json:"provider,omitempty"`
	ServiceType          *string          `json:"service_type,omitempty"`
	TrackingNumber       *string          `json:"tracking_number,omitempty"`
	EstimatedDelivery    *time.Time       `json:"estimated_delivery,omitempty"`
	ActualDelivery       *time.Time       `json:"actual_delivery,omitempty"`
	DeliveryStatus       *string          `json:"delivery_status,omitempty"`
	ShippingCost         *decimal.Decimal `json:"shipping_cost,omitempty"`
	PackageWeight        *float64         `json:"package_weight,omitempty"`
	PackageDimensions    *string          `json:"package_dimensions,omitempty"`
	PickupInstructions   *string          `json:"pickup_instructions,omitempty"`
	DeliveryInstructions *string          `json:"delivery_instructions,omitempty"`
}

// CostBreakdown represents detailed cost information
type CostBreakdown struct {
	RepairCost      decimal.Decimal `json:"repair_cost"`
	ShippingCost    decimal.Decimal `json:"shipping_cost"`
	ReplacementCost decimal.Decimal `json:"replacement_cost"`
	LaborCost       decimal.Decimal `json:"labor_cost"`
	PartsCost       decimal.Decimal `json:"parts_cost"`
	TotalCost       decimal.Decimal `json:"total_cost"`
}

// QualityMetrics represents quality and satisfaction metrics
type QualityMetrics struct {
	CustomerSatisfactionRating *int    `json:"customer_satisfaction_rating,omitempty"`
	CustomerFeedback           *string `json:"customer_feedback,omitempty"`
	ProcessingTimeHours        *int    `json:"processing_time_hours,omitempty"`
	RepairQualityScore         *int    `json:"repair_quality_score,omitempty"`
	TechnicianRating           *int    `json:"technician_rating,omitempty"`
}

// WarrantyClaim represents a warranty claim in the system
type WarrantyClaim struct {
	// Primary identification
	ID          uuid.UUID `json:"id" db:"id"`
	ClaimNumber string    `json:"claim_number" db:"claim_number"`

	// Associations
	BarcodeID    uuid.UUID `json:"barcode_id" db:"barcode_id"`
	CustomerID   uuid.UUID `json:"customer_id" db:"customer_id"`
	ProductID    uuid.UUID `json:"product_id" db:"product_id"`
	StorefrontID uuid.UUID `json:"storefront_id" db:"storefront_id"`

	// Issue details
	IssueDescription string        `json:"issue_description" db:"issue_description"`
	IssueCategory    string        `json:"issue_category" db:"issue_category"`
	IssueDate        time.Time     `json:"issue_date" db:"issue_date"`
	Severity         ClaimSeverity `json:"severity" db:"severity"`

	// Claim timeline
	ClaimDate   time.Time  `json:"claim_date" db:"claim_date"`
	ValidatedAt *time.Time `json:"validated_at,omitempty" db:"validated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`

	// Status management
	Status          ClaimStatus `json:"status" db:"status"`
	PreviousStatus  *string     `json:"previous_status,omitempty" db:"previous_status"`
	StatusUpdatedAt time.Time   `json:"status_updated_at" db:"status_updated_at"`
	StatusUpdatedBy *uuid.UUID  `json:"status_updated_by,omitempty" db:"status_updated_by"`

	// Processing assignment
	ValidatedBy             *uuid.UUID `json:"validated_by,omitempty" db:"validated_by"`
	AssignedTechnicianID    *uuid.UUID `json:"assigned_technician_id,omitempty" db:"assigned_technician_id"`
	EstimatedCompletionDate *time.Time `json:"estimated_completion_date,omitempty" db:"estimated_completion_date"`
	ActualCompletionDate    *time.Time `json:"actual_completion_date,omitempty" db:"actual_completion_date"`

	// Resolution details
	ResolutionType       *ResolutionType  `json:"resolution_type,omitempty" db:"resolution_type"`
	RepairNotes          *string          `json:"repair_notes,omitempty" db:"repair_notes"`
	ReplacementProductID *uuid.UUID       `json:"replacement_product_id,omitempty" db:"replacement_product_id"`
	RefundAmount         *decimal.Decimal `json:"refund_amount,omitempty" db:"refund_amount"`

	// Cost tracking
	RepairCost      decimal.Decimal `json:"repair_cost" db:"repair_cost"`
	ShippingCost    decimal.Decimal `json:"shipping_cost" db:"shipping_cost"`
	ReplacementCost decimal.Decimal `json:"replacement_cost" db:"replacement_cost"`
	TotalCost       decimal.Decimal `json:"total_cost" db:"total_cost"`

	// Customer information (snapshot at claim time)
	CustomerName  string  `json:"customer_name" db:"customer_name"`
	CustomerEmail string  `json:"customer_email" db:"customer_email"`
	CustomerPhone *string `json:"customer_phone,omitempty" db:"customer_phone"`

	// Address information
	PickupAddress Address `json:"pickup_address" db:"pickup_address"`

	// Logistics tracking
	ShippingProvider      *string        `json:"shipping_provider,omitempty" db:"shipping_provider"`
	TrackingNumber        *string        `json:"tracking_number,omitempty" db:"tracking_number"`
	EstimatedDeliveryDate *time.Time     `json:"estimated_delivery_date,omitempty" db:"estimated_delivery_date"`
	ActualDeliveryDate    *time.Time     `json:"actual_delivery_date,omitempty" db:"actual_delivery_date"`
	DeliveryStatus        DeliveryStatus `json:"delivery_status" db:"delivery_status"`

	// Communication and notes
	CustomerNotes   *string `json:"customer_notes,omitempty" db:"customer_notes"`
	AdminNotes      *string `json:"admin_notes,omitempty" db:"admin_notes"`
	RejectionReason *string `json:"rejection_reason,omitempty" db:"rejection_reason"`
	InternalNotes   *string `json:"internal_notes,omitempty" db:"internal_notes"`

	// Priority and categorization
	Priority ClaimPriority `json:"priority" db:"priority"`
	Tags     []string      `json:"tags,omitempty" db:"tags"`

	// Quality metrics
	CustomerSatisfactionRating *int    `json:"customer_satisfaction_rating,omitempty" db:"customer_satisfaction_rating"`
	CustomerFeedback           *string `json:"customer_feedback,omitempty" db:"customer_feedback"`
	ProcessingTimeHours        *int    `json:"processing_time_hours,omitempty" db:"processing_time_hours"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Computed fields (not stored in database)
	CanCancel      bool            `json:"can_cancel" db:"-"`
	CanUpdate      bool            `json:"can_update" db:"-"`
	ElapsedTime    string          `json:"elapsed_time" db:"-"`
	NextActions    []string        `json:"next_actions" db:"-"`
	CostBreakdown  *CostBreakdown  `json:"cost_breakdown,omitempty" db:"-"`
	QualityMetrics *QualityMetrics `json:"quality_metrics,omitempty" db:"-"`
}

// NewWarrantyClaim creates a new warranty claim with default values
func NewWarrantyClaim(barcodeID, customerID, productID, storefrontID uuid.UUID) *WarrantyClaim {
	now := time.Now()
	return &WarrantyClaim{
		ID:              uuid.New(),
		BarcodeID:       barcodeID,
		CustomerID:      customerID,
		ProductID:       productID,
		StorefrontID:    storefrontID,
		ClaimDate:       now,
		Status:          ClaimStatusPending,
		StatusUpdatedAt: now,
		Severity:        ClaimSeverityMedium,
		Priority:        ClaimPriorityNormal,
		DeliveryStatus:  DeliveryStatusNotShipped,
		RepairCost:      decimal.Zero,
		ShippingCost:    decimal.Zero,
		ReplacementCost: decimal.Zero,
		TotalCost:       decimal.Zero,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Validate performs comprehensive validation of the warranty claim
func (wc *WarrantyClaim) Validate() error {
	// Required fields
	if wc.BarcodeID == uuid.Nil {
		return fmt.Errorf("barcode_id is required")
	}
	if wc.CustomerID == uuid.Nil {
		return fmt.Errorf("customer_id is required")
	}
	if wc.ProductID == uuid.Nil {
		return fmt.Errorf("product_id is required")
	}
	if wc.StorefrontID == uuid.Nil {
		return fmt.Errorf("storefront_id is required")
	}

	// Issue description
	if wc.IssueDescription == "" {
		return fmt.Errorf("issue_description is required")
	}
	if len(wc.IssueDescription) > 5000 {
		return fmt.Errorf("issue_description cannot exceed 5000 characters")
	}

	// Issue category
	if wc.IssueCategory == "" {
		return fmt.Errorf("issue_category is required")
	}

	// Customer information
	if wc.CustomerName == "" {
		return fmt.Errorf("customer_name is required")
	}
	if wc.CustomerEmail == "" {
		return fmt.Errorf("customer_email is required")
	}

	// Validate status
	if !wc.Status.Valid() {
		return fmt.Errorf("invalid claim status: %s", wc.Status)
	}

	// Validate severity
	if !wc.Severity.Valid() {
		return fmt.Errorf("invalid claim severity: %s", wc.Severity)
	}

	// Validate priority
	if !wc.Priority.Valid() {
		return fmt.Errorf("invalid claim priority: %s", wc.Priority)
	}

	// Validate delivery status
	if !wc.DeliveryStatus.Valid() {
		return fmt.Errorf("invalid delivery status: %s", wc.DeliveryStatus)
	}

	// Validate monetary values
	if wc.RepairCost.IsNegative() {
		return fmt.Errorf("repair_cost cannot be negative")
	}
	if wc.ShippingCost.IsNegative() {
		return fmt.Errorf("shipping_cost cannot be negative")
	}
	if wc.ReplacementCost.IsNegative() {
		return fmt.Errorf("replacement_cost cannot be negative")
	}

	// Validate dates
	if wc.ValidatedAt != nil && wc.ValidatedAt.Before(wc.ClaimDate) {
		return fmt.Errorf("validated_at cannot be before claim_date")
	}
	if wc.CompletedAt != nil && wc.CompletedAt.Before(wc.ClaimDate) {
		return fmt.Errorf("completed_at cannot be before claim_date")
	}

	// Validate customer satisfaction rating
	if wc.CustomerSatisfactionRating != nil {
		rating := *wc.CustomerSatisfactionRating
		if rating < 1 || rating > 5 {
			return fmt.Errorf("customer_satisfaction_rating must be between 1 and 5")
		}
	}

	return nil
}

// CanTransitionTo checks if the claim can transition to the specified status
func (wc *WarrantyClaim) CanTransitionTo(newStatus ClaimStatus) bool {
	switch wc.Status {
	case ClaimStatusPending:
		return newStatus == ClaimStatusValidated || newStatus == ClaimStatusRejected || newStatus == ClaimStatusCancelled
	case ClaimStatusValidated:
		return newStatus == ClaimStatusAssigned || newStatus == ClaimStatusInRepair
	case ClaimStatusAssigned:
		return newStatus == ClaimStatusInRepair || newStatus == ClaimStatusCancelled
	case ClaimStatusInRepair:
		return newStatus == ClaimStatusRepaired || newStatus == ClaimStatusReplaced || newStatus == ClaimStatusDisputed
	case ClaimStatusRepaired:
		return newStatus == ClaimStatusShipped
	case ClaimStatusReplaced:
		return newStatus == ClaimStatusShipped
	case ClaimStatusShipped:
		return newStatus == ClaimStatusDelivered
	case ClaimStatusDelivered:
		return newStatus == ClaimStatusCompleted || newStatus == ClaimStatusDisputed
	case ClaimStatusRejected, ClaimStatusCompleted, ClaimStatusCancelled:
		return false // Terminal statuses
	case ClaimStatusDisputed:
		return newStatus == ClaimStatusValidated || newStatus == ClaimStatusRejected
	default:
		return false
	}
}

// UpdateStatus updates the claim status with validation
func (wc *WarrantyClaim) UpdateStatus(newStatus ClaimStatus, updatedBy *uuid.UUID) error {
	if !newStatus.Valid() {
		return fmt.Errorf("invalid status: %s", newStatus)
	}

	if !wc.CanTransitionTo(newStatus) {
		return fmt.Errorf("cannot transition from %s to %s", wc.Status, newStatus)
	}

	previousStatus := wc.Status.String()
	wc.PreviousStatus = &previousStatus
	wc.Status = newStatus
	wc.StatusUpdatedAt = time.Now()
	wc.StatusUpdatedBy = updatedBy
	wc.UpdatedAt = time.Now()

	// Update completion timestamps
	if newStatus == ClaimStatusCompleted && wc.CompletedAt == nil {
		now := time.Now()
		wc.CompletedAt = &now
		wc.ActualCompletionDate = &now
	}

	return nil
}

// Validate validates the claim for submission
func (wc *WarrantyClaim) ValidateForSubmission(validatedBy uuid.UUID, notes string) error {
	if wc.Status != ClaimStatusPending {
		return fmt.Errorf("can only validate pending claims, current status: %s", wc.Status)
	}

	now := time.Now()
	wc.ValidatedAt = &now
	wc.ValidatedBy = &validatedBy
	if notes != "" {
		wc.AdminNotes = &notes
	}

	return wc.UpdateStatus(ClaimStatusValidated, &validatedBy)
}

// Reject rejects the claim with a reason
func (wc *WarrantyClaim) Reject(rejectedBy uuid.UUID, reason string) error {
	if wc.Status != ClaimStatusPending {
		return fmt.Errorf("can only reject pending claims, current status: %s", wc.Status)
	}

	wc.RejectionReason = &reason
	wc.ValidatedBy = &rejectedBy

	return wc.UpdateStatus(ClaimStatusRejected, &rejectedBy)
}

// AssignTechnician assigns a technician to the claim
func (wc *WarrantyClaim) AssignTechnician(technicianID, assignedBy uuid.UUID, estimatedCompletionDate *time.Time) error {
	if wc.Status != ClaimStatusValidated && wc.Status != ClaimStatusAssigned {
		return fmt.Errorf("can only assign technician to validated claims, current status: %s", wc.Status)
	}

	wc.AssignedTechnicianID = &technicianID
	wc.EstimatedCompletionDate = estimatedCompletionDate

	return wc.UpdateStatus(ClaimStatusAssigned, &assignedBy)
}

// StartRepair starts the repair process
func (wc *WarrantyClaim) StartRepair(updatedBy uuid.UUID) error {
	if wc.Status != ClaimStatusAssigned {
		return fmt.Errorf("can only start repair for assigned claims, current status: %s", wc.Status)
	}

	return wc.UpdateStatus(ClaimStatusInRepair, &updatedBy)
}

// CompleteRepair completes the repair with notes
func (wc *WarrantyClaim) CompleteRepair(updatedBy uuid.UUID, repairNotes string, repairCost decimal.Decimal) error {
	if wc.Status != ClaimStatusInRepair {
		return fmt.Errorf("can only complete repair for claims in repair, current status: %s", wc.Status)
	}

	wc.RepairNotes = &repairNotes
	wc.RepairCost = repairCost
	resolutionType := ResolutionTypeRepair
	wc.ResolutionType = &resolutionType

	return wc.UpdateStatus(ClaimStatusRepaired, &updatedBy)
}

// MarkAsReplaced marks the claim as replaced
func (wc *WarrantyClaim) MarkAsReplaced(updatedBy uuid.UUID, replacementProductID uuid.UUID, replacementCost decimal.Decimal) error {
	if wc.Status != ClaimStatusInRepair {
		return fmt.Errorf("can only mark as replaced for claims in repair, current status: %s", wc.Status)
	}

	wc.ReplacementProductID = &replacementProductID
	wc.ReplacementCost = replacementCost
	resolutionType := ResolutionTypeReplace
	wc.ResolutionType = &resolutionType

	return wc.UpdateStatus(ClaimStatusReplaced, &updatedBy)
}

// Ship ships the repaired/replacement item
func (wc *WarrantyClaim) Ship(updatedBy uuid.UUID, provider, trackingNumber string, estimatedDelivery *time.Time, shippingCost decimal.Decimal) error {
	if wc.Status != ClaimStatusRepaired && wc.Status != ClaimStatusReplaced {
		return fmt.Errorf("can only ship repaired or replaced claims, current status: %s", wc.Status)
	}

	wc.ShippingProvider = &provider
	wc.TrackingNumber = &trackingNumber
	wc.EstimatedDeliveryDate = estimatedDelivery
	wc.ShippingCost = shippingCost
	wc.DeliveryStatus = DeliveryStatusPreparing

	return wc.UpdateStatus(ClaimStatusShipped, &updatedBy)
}

// MarkAsDelivered marks the claim as delivered
func (wc *WarrantyClaim) MarkAsDelivered(updatedBy uuid.UUID) error {
	if wc.Status != ClaimStatusShipped {
		return fmt.Errorf("can only mark as delivered for shipped claims, current status: %s", wc.Status)
	}

	now := time.Now()
	wc.ActualDeliveryDate = &now
	wc.DeliveryStatus = DeliveryStatusDelivered

	return wc.UpdateStatus(ClaimStatusDelivered, &updatedBy)
}

// Complete completes the claim
func (wc *WarrantyClaim) Complete(updatedBy uuid.UUID, customerFeedback *string, rating *int) error {
	if wc.Status != ClaimStatusDelivered {
		return fmt.Errorf("can only complete delivered claims, current status: %s", wc.Status)
	}

	if customerFeedback != nil {
		wc.CustomerFeedback = customerFeedback
	}
	if rating != nil {
		wc.CustomerSatisfactionRating = rating
	}

	return wc.UpdateStatus(ClaimStatusCompleted, &updatedBy)
}

// Cancel cancels the claim
func (wc *WarrantyClaim) Cancel(cancelledBy uuid.UUID, reason string) error {
	// Can cancel pending, validated, or assigned claims
	if wc.Status != ClaimStatusPending && wc.Status != ClaimStatusValidated && wc.Status != ClaimStatusAssigned {
		return fmt.Errorf("cannot cancel claim with status: %s", wc.Status)
	}

	if reason != "" {
		wc.InternalNotes = &reason
	}

	return wc.UpdateStatus(ClaimStatusCancelled, &cancelledBy)
}

// CalculateTotalCost calculates the total cost of the claim
func (wc *WarrantyClaim) CalculateTotalCost() {
	wc.TotalCost = wc.RepairCost.Add(wc.ShippingCost).Add(wc.ReplacementCost)
}

// CalculateProcessingTime calculates the processing time in hours
func (wc *WarrantyClaim) CalculateProcessingTime() {
	if wc.CompletedAt != nil {
		duration := wc.CompletedAt.Sub(wc.ClaimDate)
		hours := int(duration.Hours())
		wc.ProcessingTimeHours = &hours
	}
}

// ComputeFields calculates computed fields
func (wc *WarrantyClaim) ComputeFields() {
	// Calculate cost breakdown
	wc.CalculateTotalCost()

	// Calculate processing time
	wc.CalculateProcessingTime()

	// Set cost breakdown
	wc.CostBreakdown = &CostBreakdown{
		RepairCost:      wc.RepairCost,
		ShippingCost:    wc.ShippingCost,
		ReplacementCost: wc.ReplacementCost,
		TotalCost:       wc.TotalCost,
	}

	// Set quality metrics
	wc.QualityMetrics = &QualityMetrics{
		CustomerSatisfactionRating: wc.CustomerSatisfactionRating,
		CustomerFeedback:           wc.CustomerFeedback,
		ProcessingTimeHours:        wc.ProcessingTimeHours,
	}

	// Determine if claim can be cancelled
	wc.CanCancel = wc.Status == ClaimStatusPending || wc.Status == ClaimStatusValidated || wc.Status == ClaimStatusAssigned

	// Determine if claim can be updated
	wc.CanUpdate = wc.Status != ClaimStatusCompleted && wc.Status != ClaimStatusCancelled && wc.Status != ClaimStatusRejected

	// Calculate elapsed time
	duration := time.Since(wc.ClaimDate)
	if duration.Hours() < 24 {
		wc.ElapsedTime = fmt.Sprintf("%.1f hours", duration.Hours())
	} else {
		wc.ElapsedTime = fmt.Sprintf("%.1f days", duration.Hours()/24)
	}

	// Determine next actions based on status
	wc.NextActions = wc.getNextActions()
}

// getNextActions returns possible next actions based on current status
func (wc *WarrantyClaim) getNextActions() []string {
	switch wc.Status {
	case ClaimStatusPending:
		return []string{"validate", "reject", "request_info"}
	case ClaimStatusValidated:
		return []string{"assign_technician", "start_repair"}
	case ClaimStatusAssigned:
		return []string{"start_repair", "reassign"}
	case ClaimStatusInRepair:
		return []string{"complete_repair", "request_replacement", "update_progress"}
	case ClaimStatusRepaired:
		return []string{"ship_item", "quality_check"}
	case ClaimStatusReplaced:
		return []string{"ship_replacement"}
	case ClaimStatusShipped:
		return []string{"update_tracking", "mark_delivered"}
	case ClaimStatusDelivered:
		return []string{"complete_claim", "request_feedback"}
	default:
		return []string{}
	}
}

// GetDisplayStatus returns a human-readable status
func (wc *WarrantyClaim) GetDisplayStatus() string {
	switch wc.Status {
	case ClaimStatusPending:
		return "Pending Review"
	case ClaimStatusValidated:
		return "Approved"
	case ClaimStatusRejected:
		return "Rejected"
	case ClaimStatusAssigned:
		return "Assigned to Technician"
	case ClaimStatusInRepair:
		return "Being Repaired"
	case ClaimStatusRepaired:
		return "Repair Complete"
	case ClaimStatusReplaced:
		return "Product Replaced"
	case ClaimStatusShipped:
		return "Shipped"
	case ClaimStatusDelivered:
		return "Delivered"
	case ClaimStatusCompleted:
		return "Completed"
	case ClaimStatusCancelled:
		return "Cancelled"
	case ClaimStatusDisputed:
		return "Under Review"
	default:
		return string(wc.Status)
	}
}

// IsTerminalStatus checks if the current status is terminal (cannot be changed)
func (wc *WarrantyClaim) IsTerminalStatus() bool {
	return wc.Status == ClaimStatusCompleted || wc.Status == ClaimStatusCancelled || wc.Status == ClaimStatusRejected
}

// String returns a string representation of the warranty claim
func (wc *WarrantyClaim) String() string {
	return fmt.Sprintf("WarrantyClaim{ID: %s, Number: %s, Status: %s, Customer: %s}",
		wc.ID.String(), wc.ClaimNumber, wc.Status, wc.CustomerName)
}
