package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// WarrantyClaimUseCase defines the interface for warranty claim business logic
type WarrantyClaimUseCase interface {
	// SubmitClaim submits a new warranty claim
	SubmitClaim(ctx context.Context, req *dto.WarrantyClaimSubmissionRequest) (*dto.WarrantyClaimResponse, error)

	// GetClaim retrieves a warranty claim by ID
	GetClaim(ctx context.Context, claimID uuid.UUID) (*dto.WarrantyClaimResponse, error)

	// GetClaimByNumber retrieves a warranty claim by claim number
	GetClaimByNumber(ctx context.Context, claimNumber string) (*dto.WarrantyClaimResponse, error)

	// ListClaims retrieves warranty claims with filters
	ListClaims(ctx context.Context, filters *repository.WarrantyClaimFilters) ([]*dto.WarrantyClaimResponse, error)

	// ValidateClaim validates a warranty claim
	ValidateClaim(ctx context.Context, claimID uuid.UUID, req *dto.WarrantyClaimValidationRequest, validatedBy uuid.UUID) (*dto.WarrantyClaimResponse, error)

	// AssignTechnician assigns a technician to a warranty claim
	AssignTechnician(ctx context.Context, claimID uuid.UUID, req *dto.WarrantyClaimAssignmentRequest, assignedBy uuid.UUID) (*dto.WarrantyClaimResponse, error)

	// UpdateClaimStatus updates the status of a warranty claim
	UpdateClaimStatus(ctx context.Context, claimID uuid.UUID, req *dto.WarrantyClaimStatusUpdateRequest, updatedBy uuid.UUID) (*dto.WarrantyClaimResponse, error)

	// GetClaimStatistics retrieves warranty claim statistics
	GetClaimStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate *time.Time) (*dto.WarrantyClaimStatsResponse, error)

	// GetClaimsByTechnician retrieves claims assigned to a specific technician
	GetClaimsByTechnician(ctx context.Context, technicianID uuid.UUID, limit, offset int) ([]*dto.WarrantyClaimResponse, error)

	// GetClaimsByCustomer retrieves claims for a specific customer
	GetClaimsByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*dto.WarrantyClaimResponse, error)

	// GetClaimAttachments retrieves attachments for a claim
	GetClaimAttachments(ctx context.Context, claimID uuid.UUID) ([]*dto.ClaimAttachmentResponse, error)

	// AddClaimAttachment adds an attachment to a claim
	AddClaimAttachment(ctx context.Context, claimID uuid.UUID, req *dto.ClaimAttachmentUploadRequest) (*dto.ClaimAttachmentResponse, error)
}

// warrantyClaimUseCase implements the WarrantyClaimUseCase interface
type warrantyClaimUseCase struct {
	claimRepo   repository.WarrantyClaimRepository
	barcodeRepo repository.WarrantyBarcodeRepository
}

// NewWarrantyClaimUseCase creates a new warranty claim use case
func NewWarrantyClaimUseCase(
	claimRepo repository.WarrantyClaimRepository,
	barcodeRepo repository.WarrantyBarcodeRepository,
) WarrantyClaimUseCase {
	return &warrantyClaimUseCase{
		claimRepo:   claimRepo,
		barcodeRepo: barcodeRepo,
	}
}

// SubmitClaim submits a new warranty claim
func (uc *warrantyClaimUseCase) SubmitClaim(ctx context.Context, req *dto.WarrantyClaimSubmissionRequest) (*dto.WarrantyClaimResponse, error) {
	// Validate barcode
	barcode, err := uc.barcodeRepo.GetByBarcodeNumber(ctx, req.BarcodeValue)
	if err != nil {
		return nil, fmt.Errorf("failed to validate barcode: %w", err)
	}
	if barcode == nil {
		return nil, fmt.Errorf("invalid barcode: %s", req.BarcodeValue)
	}

	// Check if barcode is active and not expired
	if barcode.Status != entity.BarcodeStatusActivated {
		return nil, fmt.Errorf("barcode is not active: %s", req.BarcodeValue)
	}

	if barcode.ExpiryDate != nil && time.Now().After(*barcode.ExpiryDate) {
		return nil, fmt.Errorf("barcode has expired: %s", req.BarcodeValue)
	}

	// Check if there's already an active claim for this barcode
	existingClaims, err := uc.claimRepo.GetByBarcodeID(ctx, barcode.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing claims: %w", err)
	}

	for _, claim := range existingClaims {
		if claim.Status != entity.ClaimStatusCompleted && 
		   claim.Status != entity.ClaimStatusRejected && 
		   claim.Status != entity.ClaimStatusCancelled {
			return nil, fmt.Errorf("there is already an active claim for this barcode")
		}
	}

	// Generate claim number
	claimNumber, err := uc.claimRepo.GenerateClaimNumber(ctx, barcode.StorefrontID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate claim number: %w", err)
	}

	// Create warranty claim entity
	claim := entity.NewWarrantyClaim(barcode.ID, *barcode.CustomerID, barcode.ProductID, barcode.StorefrontID)
	claim.ClaimNumber = claimNumber
	claim.IssueDescription = req.IssueDescription
	claim.IssueCategory = req.IssueCategory
	claim.IssueDate = req.IssueDate
	claim.Severity = entity.ClaimSeverity(req.Severity)
	claim.CustomerName = req.CustomerName
	claim.CustomerEmail = req.CustomerEmail
	claim.CustomerPhone = &req.CustomerPhone
	claim.CustomerNotes = &req.CustomerNotes

	// Parse pickup address
	pickupAddress := entity.Address{
		Street:     req.PickupAddress,
		City:       "",
		Province:   "",
		PostalCode: "",
		Country:    "",
	}
	claim.PickupAddress = pickupAddress

	// Create the claim
	err = uc.claimRepo.Create(ctx, claim)
	if err != nil {
		return nil, fmt.Errorf("failed to create warranty claim: %w", err)
	}

	// Convert to response DTO
	response := dto.ConvertWarrantyClaimToResponse(claim)
	response.BarcodeValue = barcode.BarcodeNumber

	return response, nil
}

// GetClaim retrieves a warranty claim by ID
func (uc *warrantyClaimUseCase) GetClaim(ctx context.Context, claimID uuid.UUID) (*dto.WarrantyClaimResponse, error) {
	claim, err := uc.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("warranty claim not found")
	}

	// Get barcode information
	barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode information: %w", err)
	}

	response := dto.ConvertWarrantyClaimToResponse(claim)
	if barcode != nil {
		response.BarcodeValue = barcode.BarcodeNumber
	}

	return response, nil
}

// GetClaimByNumber retrieves a warranty claim by claim number
func (uc *warrantyClaimUseCase) GetClaimByNumber(ctx context.Context, claimNumber string) (*dto.WarrantyClaimResponse, error) {
	claim, err := uc.claimRepo.GetByClaimNumber(ctx, claimNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("warranty claim not found")
	}

	// Get barcode information
	barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode information: %w", err)
	}

	response := dto.ConvertWarrantyClaimToResponse(claim)
	if barcode != nil {
		response.BarcodeValue = barcode.BarcodeNumber
	}

	return response, nil
}

// ListClaims retrieves warranty claims with filters
func (uc *warrantyClaimUseCase) ListClaims(ctx context.Context, filters *repository.WarrantyClaimFilters) ([]*dto.WarrantyClaimResponse, error) {
	claims, err := uc.claimRepo.GetWithFilters(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claims: %w", err)
	}

	responses := dto.ConvertWarrantyClaimsToResponses(claims)

	// Enrich with barcode information if needed
	for i, claim := range claims {
		barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
		if err == nil && barcode != nil {
			responses[i].BarcodeValue = barcode.BarcodeNumber
		}
	}

	return responses, nil
}

// ValidateClaim validates a warranty claim
func (uc *warrantyClaimUseCase) ValidateClaim(ctx context.Context, claimID uuid.UUID, req *dto.WarrantyClaimValidationRequest, validatedBy uuid.UUID) (*dto.WarrantyClaimResponse, error) {
	claim, err := uc.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("warranty claim not found")
	}

	switch req.Action {
	case "validate":
		err = claim.ValidateForSubmission(validatedBy, req.Notes)
		if err != nil {
			return nil, fmt.Errorf("failed to validate claim: %w", err)
		}

	case "reject":
		err = claim.Reject(validatedBy, req.RejectionReason)
		if err != nil {
			return nil, fmt.Errorf("failed to reject claim: %w", err)
		}
		if req.Notes != "" {
			claim.AdminNotes = &req.Notes
		}

	case "request_info":
		// Add note requesting additional information
		if req.RequestedInfo != "" {
			note := fmt.Sprintf("Additional information requested: %s", req.RequestedInfo)
			claim.AdminNotes = &note
		}
		// Don't change status for info requests

	default:
		return nil, fmt.Errorf("invalid validation action: %s", req.Action)
	}

	// Update the claim
	err = uc.claimRepo.Update(ctx, claim)
	if err != nil {
		return nil, fmt.Errorf("failed to update warranty claim: %w", err)
	}

	// Get barcode information
	barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode information: %w", err)
	}

	response := dto.ConvertWarrantyClaimToResponse(claim)
	if barcode != nil {
		response.BarcodeValue = barcode.BarcodeNumber
	}

	return response, nil
}

// AssignTechnician assigns a technician to a warranty claim
func (uc *warrantyClaimUseCase) AssignTechnician(ctx context.Context, claimID uuid.UUID, req *dto.WarrantyClaimAssignmentRequest, assignedBy uuid.UUID) (*dto.WarrantyClaimResponse, error) {
	claim, err := uc.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("warranty claim not found")
	}

	technicianID, err := uuid.Parse(req.TechnicianID)
	if err != nil {
		return nil, fmt.Errorf("invalid technician ID: %w", err)
	}

	err = claim.AssignTechnician(technicianID, assignedBy, req.EstimatedCompletionDate)
	if err != nil {
		return nil, fmt.Errorf("failed to assign technician: %w", err)
	}

	// Set priority if provided
	if req.Priority != "" {
		claim.Priority = entity.ClaimPriority(req.Priority)
	}

	// Add notes if provided
	if req.Notes != "" {
		claim.AdminNotes = &req.Notes
	}

	// Update the claim
	err = uc.claimRepo.Update(ctx, claim)
	if err != nil {
		return nil, fmt.Errorf("failed to update warranty claim: %w", err)
	}

	// Get barcode information
	barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode information: %w", err)
	}

	response := dto.ConvertWarrantyClaimToResponse(claim)
	if barcode != nil {
		response.BarcodeValue = barcode.BarcodeNumber
	}

	return response, nil
}

// UpdateClaimStatus updates the status of a warranty claim
func (uc *warrantyClaimUseCase) UpdateClaimStatus(ctx context.Context, claimID uuid.UUID, req *dto.WarrantyClaimStatusUpdateRequest, updatedBy uuid.UUID) (*dto.WarrantyClaimResponse, error) {
	claim, err := uc.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("warranty claim not found")
	}

	newStatus := entity.ClaimStatus(req.Status)
	err = claim.UpdateStatus(newStatus, &updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update claim status: %w", err)
	}

	// Add notes if provided
	if req.Notes != "" {
		claim.AdminNotes = &req.Notes
	}

	// Add repair notes if provided
	if req.RepairNotes != "" {
		claim.RepairNotes = &req.RepairNotes
	}

	// Update the claim
	err = uc.claimRepo.Update(ctx, claim)
	if err != nil {
		return nil, fmt.Errorf("failed to update warranty claim: %w", err)
	}

	// Get barcode information
	barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get barcode information: %w", err)
	}

	response := dto.ConvertWarrantyClaimToResponse(claim)
	if barcode != nil {
		response.BarcodeValue = barcode.BarcodeNumber
	}

	return response, nil
}

// GetClaimStatistics retrieves warranty claim statistics
func (uc *warrantyClaimUseCase) GetClaimStatistics(ctx context.Context, storefrontID uuid.UUID, startDate, endDate *time.Time) (*dto.WarrantyClaimStatsResponse, error) {
	// Set default date range if not provided
	if startDate == nil {
		start := time.Now().AddDate(0, -1, 0) // Last month
		startDate = &start
	}
	if endDate == nil {
		end := time.Now()
		endDate = &end
	}

	stats, err := uc.claimRepo.GetClaimStatistics(ctx, storefrontID, *startDate, *endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim statistics: %w", err)
	}

	// Convert repository.ClaimStatistics to map[string]interface{}
	statsMap := map[string]interface{}{
		"total_claims":             int(stats.TotalClaims),
		"claims_by_status":         convertInt64MapToIntMap(stats.ClaimsByStatus),
		"claims_by_severity":       convertInt64MapToIntMap(stats.ClaimsBySeverity),
		"claims_by_category":       convertInt64MapToIntMap(stats.ClaimsByCategory),
		"claims_by_priority":       map[string]int{}, // Not available in repository stats
		"average_processing_time":  stats.AverageResolutionTime,
		"total_repair_cost":        stats.TotalRepairCost,
		"total_shipping_cost":      stats.TotalShippingCost,
		"total_cost":              stats.TotalRepairCost + stats.TotalShippingCost + stats.TotalReplacementCost,
		"satisfaction_rating":      stats.CustomerSatisfactionAvg,
		"claims_this_month":        0, // Would need additional calculation
		"claims_last_month":        0, // Would need additional calculation
		"growth_rate":             0.0, // Would need additional calculation
	}

	response := dto.ConvertWarrantyClaimStatsToResponse(statsMap)
	return response, nil
}

// Helper function to convert map[string]int64 to map[string]int
func convertInt64MapToIntMap(input map[string]int64) map[string]int {
	result := make(map[string]int)
	for k, v := range input {
		result[k] = int(v)
	}
	return result
}

// GetClaimsByTechnician retrieves claims assigned to a specific technician
func (uc *warrantyClaimUseCase) GetClaimsByTechnician(ctx context.Context, technicianID uuid.UUID, limit, offset int) ([]*dto.WarrantyClaimResponse, error) {
	claims, err := uc.claimRepo.GetByTechnician(ctx, technicianID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get claims by technician: %w", err)
	}

	responses := dto.ConvertWarrantyClaimsToResponses(claims)

	// Enrich with barcode information
	for i, claim := range claims {
		barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
		if err == nil && barcode != nil {
			responses[i].BarcodeValue = barcode.BarcodeNumber
		}
	}

	return responses, nil
}

// GetClaimsByCustomer retrieves claims for a specific customer
func (uc *warrantyClaimUseCase) GetClaimsByCustomer(ctx context.Context, customerID uuid.UUID, limit, offset int) ([]*dto.WarrantyClaimResponse, error) {
	claims, err := uc.claimRepo.GetByCustomerID(ctx, customerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get claims by customer: %w", err)
	}

	responses := dto.ConvertWarrantyClaimsToResponses(claims)

	// Enrich with barcode information
	for i, claim := range claims {
		barcode, err := uc.barcodeRepo.GetByID(ctx, claim.BarcodeID)
		if err == nil && barcode != nil {
			responses[i].BarcodeValue = barcode.BarcodeNumber
		}
	}

	return responses, nil
}

// GetClaimAttachments retrieves attachments for a claim
func (uc *warrantyClaimUseCase) GetClaimAttachments(ctx context.Context, claimID uuid.UUID) ([]*dto.ClaimAttachmentResponse, error) {
	attachments, err := uc.claimRepo.GetClaimAttachments(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim attachments: %w", err)
	}

	responses := dto.ConvertClaimAttachmentsToResponses(attachments)
	return responses, nil
}

// AddClaimAttachment adds an attachment to a claim
// Note: This function expects file upload to be handled separately
// The ClaimAttachmentUploadRequest only contains metadata, not actual file data
func (uc *warrantyClaimUseCase) AddClaimAttachment(ctx context.Context, claimID uuid.UUID, req *dto.ClaimAttachmentUploadRequest) (*dto.ClaimAttachmentResponse, error) {
	// Verify claim exists
	claim, err := uc.claimRepo.GetByID(ctx, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get warranty claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("warranty claim not found")
	}

	// Create attachment entity with placeholder values
	// In a real implementation, file upload would be handled separately
	// and these values would come from the file upload service
	attachment := &entity.ClaimAttachment{
		ID:               uuid.New(),
		ClaimID:          claimID,
		UploadedBy:       uuid.Nil, // Should be set from authentication context
		Filename:         "placeholder.txt", // Should come from actual file upload
		OriginalFilename: "placeholder.txt", // Should come from actual file upload
		FilePath:         "/uploads/claims/placeholder.txt", // Should come from file storage service
		FileURL:          "https://example.com/uploads/claims/placeholder.txt", // Should come from file storage service
		FileSize:         0, // Should come from actual file
		MimeType:         "text/plain", // Should be detected from actual file
		AttachmentType:   entity.AttachmentType(req.AttachmentType),
		Description:      req.Description,
		IsProcessed:      false,
		VirusScanStatus:  entity.VirusScanStatusPending,
		UploadedAt:       time.Now(),
	}

	// Add attachment
	err = uc.claimRepo.AddClaimAttachment(ctx, attachment)
	if err != nil {
		return nil, fmt.Errorf("failed to add claim attachment: %w", err)
	}

	response := dto.ConvertClaimAttachmentToResponse(attachment)
	return response, nil
}