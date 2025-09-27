package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// WarrantyClaimRepository defines the interface for warranty claim data operations
type WarrantyClaimRepository interface {
	// Create creates a new warranty claim
	Create(ctx context.Context, claim *entity.WarrantyClaim) error

	// GetByID retrieves a warranty claim by its ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.WarrantyClaim, error)

	// GetByClaimNumber retrieves a warranty claim by its claim number
	GetByClaimNumber(ctx context.Context, claimNumber string) (*entity.WarrantyClaim, error)

	// GetByBarcodeID retrieves warranty claims for a specific barcode
	GetByBarcodeID(ctx context.Context, barcodeID uuid.UUID) ([]*entity.WarrantyClaim, error)

	// GetByCustomerID retrieves warranty claims for a specific customer
	GetByCustomerID(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error)

	// GetByStorefrontID retrieves warranty claims for a specific storefront
	GetByStorefrontID(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error)

	// GetByStatus retrieves warranty claims by status
	GetByStatus(ctx context.Context, storefrontID uuid.UUID, status entity.ClaimStatus, limit, offset int) ([]*entity.WarrantyClaim, error)

	// GetByTechnician retrieves warranty claims assigned to a specific technician
	GetByTechnician(ctx context.Context, technicianID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error)

	// GetPendingClaims retrieves all pending warranty claims
	GetPendingClaims(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error)

	// GetOverdueClaims retrieves warranty claims that are overdue
	GetOverdueClaims(ctx context.Context, storefrontID uuid.UUID, limit, offset int) ([]*entity.WarrantyClaim, error)

	// Update updates an existing warranty claim
	Update(ctx context.Context, claim *entity.WarrantyClaim) error

	// UpdateStatus updates the status of a warranty claim
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ClaimStatus, updatedBy uuid.UUID, notes string) error

	// Delete soft deletes a warranty claim
	Delete(ctx context.Context, id uuid.UUID) error

	// Count counts warranty claims with optional filters
	Count(ctx context.Context, filters *WarrantyClaimFilters) (int, error)

	// GetWithFilters retrieves warranty claims with filters and pagination
	GetWithFilters(ctx context.Context, filters *WarrantyClaimFilters) ([]*entity.WarrantyClaim, error)

	// GetClaimsByDateRange retrieves claims within a date range
	GetClaimsByDateRange(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time, limit, offset int) ([]*entity.WarrantyClaim, error)

	// GetClaimStatistics retrieves claim statistics for analytics
	GetClaimStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate time.Time) (*ClaimStatistics, error)

	// GetClaimTimeline retrieves the complete timeline for a claim
	GetClaimTimeline(ctx context.Context, claimID uuid.UUID) ([]*entity.ClaimTimeline, error)

	// AddTimelineEntry adds a new timeline entry for a claim
	AddTimelineEntry(ctx context.Context, entry *entity.ClaimTimeline) error

	// GetClaimAttachments retrieves all attachments for a claim
	GetClaimAttachments(ctx context.Context, claimID uuid.UUID) ([]*entity.ClaimAttachment, error)

	// AddClaimAttachment adds a new attachment to a claim
	AddClaimAttachment(ctx context.Context, attachment *entity.ClaimAttachment) error

	// UpdateClaimAttachment updates an existing claim attachment
	UpdateClaimAttachment(ctx context.Context, attachment *entity.ClaimAttachment) error

	// DeleteClaimAttachment deletes a claim attachment
	DeleteClaimAttachment(ctx context.Context, attachmentID uuid.UUID) error

	// GenerateClaimNumber generates a unique claim number
	GenerateClaimNumber(ctx context.Context, storefrontID uuid.UUID) (string, error)

	// GetRepairTickets retrieves repair tickets for a claim
	GetRepairTickets(ctx context.Context, claimID uuid.UUID) ([]*entity.RepairTicket, error)

	// CreateRepairTicket creates a new repair ticket for a claim
	CreateRepairTicket(ctx context.Context, ticket *entity.RepairTicket) error

	// UpdateRepairTicket updates an existing repair ticket
	UpdateRepairTicket(ctx context.Context, ticket *entity.RepairTicket) error

	// GetRepairTicketByID retrieves a repair ticket by its ID
	GetRepairTicketByID(ctx context.Context, ticketID uuid.UUID) (*entity.RepairTicket, error)
}

// WarrantyClaimFilters represents filters for warranty claim queries
type WarrantyClaimFilters struct {
	StorefrontID         *uuid.UUID             `json:"storefront_id,omitempty"`
	CustomerID           *uuid.UUID             `json:"customer_id,omitempty"`
	ProductID            *uuid.UUID             `json:"product_id,omitempty"`
	BarcodeID            *uuid.UUID             `json:"barcode_id,omitempty"`
	Status               *entity.ClaimStatus    `json:"status,omitempty"`
	Severity             *entity.ClaimSeverity  `json:"severity,omitempty"`
	Priority             *entity.ClaimPriority  `json:"priority,omitempty"`
	AssignedTechnicianID *uuid.UUID             `json:"assigned_technician_id,omitempty"`
	ValidatedBy          *uuid.UUID             `json:"validated_by,omitempty"`
	IssueCategory        *string                `json:"issue_category,omitempty"`
	ResolutionType       *entity.ResolutionType `json:"resolution_type,omitempty"`
	CreatedAfter         *time.Time             `json:"created_after,omitempty"`
	CreatedBefore        *time.Time             `json:"created_before,omitempty"`
	ValidatedAfter       *time.Time             `json:"validated_after,omitempty"`
	ValidatedBefore      *time.Time             `json:"validated_before,omitempty"`
	CompletedAfter       *time.Time             `json:"completed_after,omitempty"`
	CompletedBefore      *time.Time             `json:"completed_before,omitempty"`
	MinCost              *float64               `json:"min_cost,omitempty"`
	MaxCost              *float64               `json:"max_cost,omitempty"`
	CustomerEmail        *string                `json:"customer_email,omitempty"`
	ClaimNumber          *string                `json:"claim_number,omitempty"`
	Search               *string                `json:"search,omitempty"`
	Tags                 []string               `json:"tags,omitempty"`
	IsOverdue            *bool                  `json:"is_overdue,omitempty"`
	HasAttachments       *bool                  `json:"has_attachments,omitempty"`
	Page                 int                    `json:"page"`
	PageSize             int                    `json:"page_size"`
	SortBy               string                 `json:"sort_by"`
	SortDirection        string                 `json:"sort_direction"`
	IncludeTimeline      bool                   `json:"include_timeline"`
	IncludeAttachments   bool                   `json:"include_attachments"`
}

// ClaimStatistics represents warranty claim statistics
type ClaimStatistics struct {
	TotalClaims             int64                `json:"total_claims"`
	PendingClaims           int64                `json:"pending_claims"`
	ValidatedClaims         int64                `json:"validated_claims"`
	CompletedClaims         int64                `json:"completed_claims"`
	RejectedClaims          int64                `json:"rejected_claims"`
	CancelledClaims         int64                `json:"cancelled_claims"`
	AverageResolutionTime   float64              `json:"average_resolution_time_hours"`
	AverageRepairCost       float64              `json:"average_repair_cost"`
	CustomerSatisfactionAvg float64              `json:"customer_satisfaction_average"`
	ClaimsByCategory        map[string]int64     `json:"claims_by_category"`
	ClaimsByStatus          map[string]int64     `json:"claims_by_status"`
	ClaimsBySeverity        map[string]int64     `json:"claims_by_severity"`
	ClaimsByResolutionType  map[string]int64     `json:"claims_by_resolution_type"`
	TotalRepairCost         float64              `json:"total_repair_cost"`
	TotalShippingCost       float64              `json:"total_shipping_cost"`
	TotalReplacementCost    float64              `json:"total_replacement_cost"`
	ClaimRate               float64              `json:"claim_rate"` // Claims per 1000 warranties
	FirstTimeFixRate        float64              `json:"first_time_fix_rate"`
	SLAComplianceRate       float64              `json:"sla_compliance_rate"`
	MonthlyTrends           []*MonthlyClaimTrend `json:"monthly_trends"`
}

// MonthlyClaimTrend represents claim trends by month
type MonthlyClaimTrend struct {
	Month       string  `json:"month"`
	Year        int     `json:"year"`
	TotalClaims int64   `json:"total_claims"`
	Completed   int64   `json:"completed"`
	AvgTime     float64 `json:"average_resolution_time"`
	TotalCost   float64 `json:"total_cost"`
}
