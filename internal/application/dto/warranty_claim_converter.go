package dto

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// ConvertWarrantyClaimToResponse converts a WarrantyClaim entity to WarrantyClaimResponse DTO
func ConvertWarrantyClaimToResponse(claim *entity.WarrantyClaim) *WarrantyClaimResponse {
	if claim == nil {
		return nil
	}

	// Convert Address to string (JSON representation)
	pickupAddressBytes, _ := json.Marshal(claim.PickupAddress)
	pickupAddressStr := string(pickupAddressBytes)

	response := &WarrantyClaimResponse{
		ID:               claim.ID.String(),
		ClaimNumber:      claim.ClaimNumber,
		BarcodeID:        claim.BarcodeID.String(),
		CustomerID:       claim.CustomerID.String(),
		ProductID:        claim.ProductID.String(),
		StorefrontID:     claim.StorefrontID.String(),
		IssueDescription: claim.IssueDescription,
		IssueCategory:    claim.IssueCategory,
		IssueDate:        claim.IssueDate,
		Severity:         string(claim.Severity),
		ClaimDate:        claim.ClaimDate,
		ValidatedAt:      claim.ValidatedAt,
		CompletedAt:      claim.CompletedAt,
		Status:           claim.Status.String(),
		StatusUpdatedAt:  claim.StatusUpdatedAt,
		RepairCost:       claim.RepairCost,
		ShippingCost:     claim.ShippingCost,
		ReplacementCost:  claim.ReplacementCost,
		TotalCost:        claim.TotalCost,
		CustomerName:     claim.CustomerName,
		CustomerEmail:    claim.CustomerEmail,
		PickupAddress:    pickupAddressStr,
		Priority:         string(claim.Priority),
		DeliveryStatus:   string(claim.DeliveryStatus),
		CreatedAt:        claim.CreatedAt,
		UpdatedAt:        claim.UpdatedAt,
	}

	// Handle optional string fields
	if claim.CustomerPhone != nil {
		response.CustomerPhone = *claim.CustomerPhone
	}

	if claim.PreviousStatus != nil {
		response.PreviousStatus = claim.PreviousStatus
	}

	// Handle optional UUID fields - convert to string
	if claim.StatusUpdatedBy != nil {
		statusUpdatedBy := claim.StatusUpdatedBy.String()
		response.StatusUpdatedBy = &statusUpdatedBy
	}

	if claim.ValidatedBy != nil {
		validatedBy := claim.ValidatedBy.String()
		response.ValidatedBy = &validatedBy
	}

	if claim.AssignedTechnicianID != nil {
		assignedTechnicianID := claim.AssignedTechnicianID.String()
		response.AssignedTechnicianID = &assignedTechnicianID
	}

	if claim.ReplacementProductID != nil {
		replacementProductID := claim.ReplacementProductID.String()
		response.ReplacementProductID = &replacementProductID
	}

	// Handle optional time fields
	if claim.EstimatedCompletionDate != nil {
		response.EstimatedCompletionDate = claim.EstimatedCompletionDate
	}

	if claim.ActualCompletionDate != nil {
		response.ActualCompletionDate = claim.ActualCompletionDate
	}

	if claim.EstimatedDeliveryDate != nil {
		response.EstimatedDeliveryDate = claim.EstimatedDeliveryDate
	}

	if claim.ActualDeliveryDate != nil {
		response.ActualDeliveryDate = claim.ActualDeliveryDate
	}

	// Handle optional enum fields
	if claim.ResolutionType != nil {
		resType := string(*claim.ResolutionType)
		response.ResolutionType = &resType
	}

	// Handle optional string pointer fields
	if claim.RepairNotes != nil {
		response.RepairNotes = claim.RepairNotes
	}

	if claim.RefundAmount != nil {
		response.RefundAmount = claim.RefundAmount
	}

	if claim.ShippingProvider != nil {
		response.ShippingProvider = claim.ShippingProvider
	}

	if claim.TrackingNumber != nil {
		response.TrackingNumber = claim.TrackingNumber
	}

	if claim.CustomerNotes != nil {
		response.CustomerNotes = claim.CustomerNotes
	}

	if claim.AdminNotes != nil {
		response.AdminNotes = claim.AdminNotes
	}

	if claim.RejectionReason != nil {
		response.RejectionReason = claim.RejectionReason
	}

	if claim.InternalNotes != nil {
		response.InternalNotes = claim.InternalNotes
	}

	if len(claim.Tags) > 0 {
		response.Tags = claim.Tags
	}

	if claim.CustomerSatisfactionRating != nil {
		response.CustomerSatisfactionRating = claim.CustomerSatisfactionRating
	}

	if claim.CustomerFeedback != nil {
		response.CustomerFeedback = claim.CustomerFeedback
	}

	// Convert processing time from int to decimal
	if claim.ProcessingTimeHours != nil {
		processingTime := decimal.NewFromInt(int64(*claim.ProcessingTimeHours))
		response.ProcessingTimeHours = &processingTime
	}

	return response
}

// ConvertWarrantyClaimsToResponses converts a slice of WarrantyClaim entities to WarrantyClaimResponse DTOs
func ConvertWarrantyClaimsToResponses(claims []*entity.WarrantyClaim) []*WarrantyClaimResponse {
	if claims == nil {
		return nil
	}

	responses := make([]*WarrantyClaimResponse, len(claims))
	for i, claim := range claims {
		responses[i] = ConvertWarrantyClaimToResponse(claim)
	}

	return responses
}

// ConvertClaimAttachmentToResponse converts a ClaimAttachment entity to ClaimAttachmentResponse DTO
func ConvertClaimAttachmentToResponse(attachment *entity.ClaimAttachment) *ClaimAttachmentResponse {
	if attachment == nil {
		return nil
	}

	return &ClaimAttachmentResponse{
		ID:                  attachment.ID.String(),
		ClaimID:             attachment.ClaimID.String(),
		FileName:            attachment.Filename,
		FilePath:            attachment.FilePath,
		FileURL:             attachment.FileURL,
		FileSize:            attachment.FileSize,
		FileType:            attachment.MimeType, // Using MimeType as FileType
		MimeType:            attachment.MimeType,
		AttachmentType:      string(attachment.AttachmentType),
		Description:         attachment.Description,
		UploadedBy:          attachment.UploadedBy.String(),
		SecurityScanStatus:  string(attachment.VirusScanStatus),
		CreatedAt:           attachment.UploadedAt,
		UpdatedAt:           attachment.UploadedAt,
	}
}

// ConvertClaimAttachmentsToResponses converts a slice of ClaimAttachment entities to ClaimAttachmentResponse DTOs
func ConvertClaimAttachmentsToResponses(attachments []*entity.ClaimAttachment) []*ClaimAttachmentResponse {
	if attachments == nil {
		return nil
	}

	responses := make([]*ClaimAttachmentResponse, len(attachments))
	for i, attachment := range attachments {
		responses[i] = ConvertClaimAttachmentToResponse(attachment)
	}

	return responses
}

// ConvertWarrantyClaimSubmissionRequestToEntity converts a WarrantyClaimSubmissionRequest DTO to WarrantyClaim entity
func ConvertWarrantyClaimSubmissionRequestToEntity(req *WarrantyClaimSubmissionRequest, barcodeID, customerID, productID, storefrontID string) (*entity.WarrantyClaim, error) {
	if req == nil {
		return nil, nil
	}

	// Parse UUIDs
	barcodeUUID, err := uuid.Parse(barcodeID)
	if err != nil {
		return nil, err
	}
	customerUUID, err := uuid.Parse(customerID)
	if err != nil {
		return nil, err
	}
	productUUID, err := uuid.Parse(productID)
	if err != nil {
		return nil, err
	}
	storefrontUUID, err := uuid.Parse(storefrontID)
	if err != nil {
		return nil, err
	}

	claim := entity.NewWarrantyClaim(barcodeUUID, customerUUID, productUUID, storefrontUUID)
	
	// Set fields from request
	claim.IssueDescription = req.IssueDescription
	claim.IssueCategory = req.IssueCategory
	claim.IssueDate = req.IssueDate
	claim.Severity = entity.ClaimSeverity(req.Severity)
	claim.CustomerName = req.CustomerName
	claim.CustomerEmail = req.CustomerEmail
	claim.CustomerPhone = &req.CustomerPhone
	claim.CustomerNotes = &req.CustomerNotes

	// Parse pickup address from string to Address struct
	var pickupAddress entity.Address
	if err := json.Unmarshal([]byte(req.PickupAddress), &pickupAddress); err != nil {
		// If JSON parsing fails, treat as simple string address
		pickupAddress = entity.Address{
			Street: req.PickupAddress,
		}
	}
	claim.PickupAddress = pickupAddress

	return claim, nil
}

// ConvertWarrantyClaimStatsToResponse converts warranty claim statistics to response DTO
func ConvertWarrantyClaimStatsToResponse(stats map[string]interface{}) *WarrantyClaimStatsResponse {
	if stats == nil {
		return nil
	}

	response := &WarrantyClaimStatsResponse{
		TotalClaims:        getIntFromStats(stats, "total_claims"),
		ClaimsByStatus:     getMapFromStats(stats, "claims_by_status"),
		ClaimsBySeverity:   getMapFromStats(stats, "claims_by_severity"),
		ClaimsByCategory:   getMapFromStats(stats, "claims_by_category"),
		ClaimsByPriority:   getMapFromStats(stats, "claims_by_priority"),
		AverageProcessingTime: getDecimalFromStats(stats, "average_processing_time"),
		TotalRepairCost:    getDecimalFromStats(stats, "total_repair_cost"),
		TotalShippingCost:  getDecimalFromStats(stats, "total_shipping_cost"),
		TotalCost:          getDecimalFromStats(stats, "total_cost"),
		SatisfactionRating: getDecimalFromStats(stats, "satisfaction_rating"),
		ClaimsThisMonth:    getIntFromStats(stats, "claims_this_month"),
		ClaimsLastMonth:    getIntFromStats(stats, "claims_last_month"),
		GrowthRate:         getDecimalFromStats(stats, "growth_rate"),
	}

	return response
}

// Helper functions for stats conversion
func getIntFromStats(stats map[string]interface{}, key string) int {
	if val, ok := stats[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return 0
}

func getMapFromStats(stats map[string]interface{}, key string) map[string]int {
	if val, ok := stats[key]; ok {
		if mapVal, ok := val.(map[string]int); ok {
			return mapVal
		}
	}
	return make(map[string]int)
}

func getDecimalFromStats(stats map[string]interface{}, key string) decimal.Decimal {
	if val, ok := stats[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return decimal.NewFromFloat(floatVal)
		}
		if intVal, ok := val.(int); ok {
			return decimal.NewFromInt(int64(intVal))
		}
	}
	return decimal.Zero
}