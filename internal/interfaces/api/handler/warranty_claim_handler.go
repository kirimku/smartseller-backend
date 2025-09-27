package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// WarrantyClaimHandler handles warranty claim-related HTTP requests
type WarrantyClaimHandler struct {
	logger *slog.Logger
}

// NewWarrantyClaimHandler creates a new warranty claim handler
func NewWarrantyClaimHandler(logger *slog.Logger) *WarrantyClaimHandler {
	return &WarrantyClaimHandler{
		logger: logger,
	}
}

// ListClaims handles claim listing requests
// @Summary List warranty claims
// @Description Get paginated list of warranty claims with advanced filtering options
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort field" default("created_at")
// @Param sort_dir query string false "Sort direction" Enums(asc,desc) default("desc")
// @Param status query string false "Filter by status" Enums(pending,validated,in_progress,completed,rejected,cancelled)
// @Param severity query string false "Filter by severity" Enums(low,medium,high,critical)
// @Param category query string false "Filter by category" Enums(hardware,software,performance,defect,damage,other)
// @Param customer_id query string false "Filter by customer ID"
// @Param product_id query string false "Filter by product ID"
// @Param technician_id query string false "Filter by assigned technician"
// @Param date_from query string false "Filter claims from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter claims to date (YYYY-MM-DD)"
// @Param search query string false "Search in claim number, customer name, or description"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyClaimListResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims [get]
func (h *WarrantyClaimHandler) ListClaims(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	_ = c.DefaultQuery("sort_by", "created_at")    // sortBy - will be used when implementing actual logic
	_ = c.DefaultQuery("sort_dir", "desc")         // sortDir - will be used when implementing actual logic

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Parse filters
	status := c.Query("status")
	severity := c.Query("severity")
	category := c.Query("category")
	customerID := c.Query("customer_id")
	productID := c.Query("product_id")
	technicianID := c.Query("technician_id")
	search := c.Query("search")

	// Parse date filters - will be used when implementing actual logic
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if _, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			// dateFrom will be used when implementing actual logic
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if _, err := time.Parse("2006-01-02", dateToStr); err == nil {
			// dateTo will be used when implementing actual logic
		}
	}

	// TODO: Implement actual claim listing logic when usecase is ready
	// For now, return a mock response
	mockClaims := []dto.WarrantyClaimResponse{
		{
			ID:               uuid.New().String(),
			ClaimNumber:      "WC-2024-001",
			BarcodeID:        uuid.New().String(),
			CustomerID:       uuid.New().String(),
			ProductID:        uuid.New().String(),
			StorefrontID:     uuid.New().String(),
			IssueDescription: "Device not working properly",
			IssueCategory:    "hardware",
			IssueDate:        time.Now().AddDate(0, 0, -7),
			Severity:         "medium",
			ClaimDate:        time.Now().AddDate(0, 0, -5),
			Status:           "pending",
			CustomerName:     "John Doe",
			CustomerEmail:    "john.doe@example.com",
			Priority:         "medium",
			CreatedAt:        time.Now().AddDate(0, 0, -5),
			UpdatedAt:        time.Now().AddDate(0, 0, -1),
		},
	}

	response := &dto.WarrantyClaimListResponse{
		Claims: mockClaims,
		Pagination: dto.PaginationResponse{
			Page:       page,
			Limit:      pageSize,
			Total:      1,
			TotalPages: 1,
		},
	}

	h.logger.Info("Successfully retrieved warranty claims (mock)",
		"page", page,
		"page_size", pageSize,
		"filters", map[string]string{
			"status":        status,
			"severity":      severity,
			"category":      category,
			"customer_id":   customerID,
			"product_id":    productID,
			"technician_id": technicianID,
			"search":        search,
		},
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty claims retrieved successfully", response)
}

// GetClaim handles individual claim retrieval requests
// @Summary Get warranty claim details
// @Description Get detailed information about a specific warranty claim including timeline
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Claim ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyClaimDetailResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/{id} [get]
func (h *WarrantyClaimHandler) GetClaim(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		h.logger.Warn("Invalid claim ID format", "claim_id", claimIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid claim ID format", nil)
		return
	}

	// TODO: Implement actual claim retrieval logic when usecase is ready
	// For now, return a mock response
	description := "Purchase receipt"
	mockClaim := &dto.WarrantyClaimDetailResponse{
		Claim: &dto.WarrantyClaimResponse{
			ID:               claimID.String(),
			ClaimNumber:      "WC-2024-001",
			BarcodeID:        uuid.New().String(),
			CustomerID:       uuid.New().String(),
			ProductID:        uuid.New().String(),
			StorefrontID:     uuid.New().String(),
			IssueDescription: "Device not working properly after 3 months",
			IssueCategory:    "hardware",
			IssueDate:        time.Now().AddDate(0, 0, -7),
			Severity:         "medium",
			ClaimDate:        time.Now().AddDate(0, 0, -5),
			Status:           "pending",
			CustomerName:     "John Doe",
			CustomerEmail:    "john.doe@example.com",
			Priority:         "medium",
			CreatedAt:        time.Now().AddDate(0, 0, -5),
			UpdatedAt:        time.Now().AddDate(0, 0, -1),
		},
		Timeline: []*dto.ClaimTimelineResponse{
			{
				ID:          uuid.New().String(),
				ClaimID:     claimID.String(),
				EventType:   "claim_submitted",
				Description: "Claim submitted by customer",
				CreatedBy:   uuid.New().String(),
				CreatedAt:   time.Now().AddDate(0, 0, -5),
			},
		},
		Attachments: []*dto.ClaimAttachmentResponse{
			{
				ID:                 uuid.New().String(),
				ClaimID:            claimID.String(),
				FileName:           "receipt.jpg",
				FileType:           "image/jpeg",
				AttachmentType:     "receipt",
				Description:        &description,
				SecurityScanStatus: "clean",
				CreatedAt:          time.Now().AddDate(0, 0, -5),
			},
		},
	}

	h.logger.Info("Successfully retrieved warranty claim details (mock)",
		"claim_id", claimID,
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty claim details retrieved successfully", mockClaim)
}

// ValidateClaim handles claim validation requests
// @Summary Validate warranty claim
// @Description Validate a warranty claim and update its status
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Claim ID"
// @Param request body dto.ClaimValidationRequest true "Validation request"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyClaimResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/{id}/validate [put]
func (h *WarrantyClaimHandler) ValidateClaim(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		h.logger.Warn("Invalid claim ID format", "claim_id", claimIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid claim ID format", nil)
		return
	}

	var req dto.ClaimValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind validation request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// TODO: Implement actual claim validation logic when usecase is ready
	// For now, return a mock response
	mockClaim := &dto.WarrantyClaimResponse{
		ID:               claimID.String(),
		ClaimNumber:      "WC-2024-001",
		BarcodeID:        uuid.New().String(),
		CustomerID:       uuid.New().String(),
		ProductID:        uuid.New().String(),
		StorefrontID:     uuid.New().String(),
		IssueDescription: "Device not working properly after 3 months",
		IssueCategory:    "hardware",
		IssueDate:        time.Now().AddDate(0, 0, -7),
		Severity:         "medium",
		ClaimDate:        time.Now().AddDate(0, 0, -5),
		ValidatedAt:      &time.Time{},
		Status:           "validated",
		StatusUpdatedAt:  time.Now(),
		CustomerName:     "John Doe",
		CustomerEmail:    "john.doe@example.com",
		Priority:         "medium",
		CreatedAt:        time.Now().AddDate(0, 0, -5),
		UpdatedAt:        time.Now(),
	}
	*mockClaim.ValidatedAt = time.Now()

	h.logger.Info("Successfully validated warranty claim (mock)",
		"claim_id", claimID,
		"validated_by", userID,
		"notes", req.Notes,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty claim validated successfully", mockClaim)
}

// RejectClaim handles claim rejection requests
// @Summary Reject warranty claim
// @Description Reject a warranty claim with reason
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Claim ID"
// @Param request body dto.ClaimRejectionRequest true "Rejection request"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyClaimResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/{id}/reject [put]
func (h *WarrantyClaimHandler) RejectClaim(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		h.logger.Warn("Invalid claim ID format", "claim_id", claimIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid claim ID format", nil)
		return
	}

	var req dto.ClaimRejectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind rejection request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// TODO: Implement actual claim rejection logic when usecase is ready
	// For now, return a mock response
	mockClaim := &dto.WarrantyClaimResponse{
		ID:               claimID.String(),
		ClaimNumber:      "WC-2024-001",
		BarcodeID:        uuid.New().String(),
		CustomerID:       uuid.New().String(),
		ProductID:        uuid.New().String(),
		StorefrontID:     uuid.New().String(),
		IssueDescription: "Device not working properly after 3 months",
		IssueCategory:    "hardware",
		IssueDate:        time.Now().AddDate(0, 0, -7),
		Severity:         "medium",
		ClaimDate:        time.Now().AddDate(0, 0, -5),
		Status:           "rejected",
		StatusUpdatedAt:  time.Now(),
		CustomerName:     "John Doe",
		CustomerEmail:    "john.doe@example.com",
		Priority:         "medium",
		CreatedAt:        time.Now().AddDate(0, 0, -5),
		UpdatedAt:        time.Now(),
	}

	h.logger.Info("Successfully rejected warranty claim (mock)",
		"claim_id", claimID,
		"rejected_by", userID,
		"reason", req.RejectionReason,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty claim rejected successfully", mockClaim)
}

// AssignTechnician handles technician assignment requests
// @Summary Assign technician to claim
// @Description Assign a technician to handle a warranty claim
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Claim ID"
// @Param request body dto.TechnicianAssignmentRequest true "Assignment request"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyClaimResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/{id}/assign [put]
func (h *WarrantyClaimHandler) AssignTechnician(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		h.logger.Warn("Invalid claim ID format", "claim_id", claimIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid claim ID format", nil)
		return
	}

	var req dto.TechnicianAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind assignment request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate technician ID
	technicianID, err := uuid.Parse(req.TechnicianID)
	if err != nil {
		h.logger.Warn("Invalid technician ID format", "technician_id", req.TechnicianID)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid technician ID format", nil)
		return
	}

	// TODO: Implement actual technician assignment logic when usecase is ready
	// For now, return a mock response
	mockClaim := &dto.WarrantyClaimResponse{
		ID:               claimID.String(),
		ClaimNumber:      "WC-2024-001",
		BarcodeID:        uuid.New().String(),
		CustomerID:       uuid.New().String(),
		ProductID:        uuid.New().String(),
		StorefrontID:     uuid.New().String(),
		IssueDescription: "Device not working properly after 3 months",
		IssueCategory:    "hardware",
		IssueDate:        time.Now().AddDate(0, 0, -7),
		Severity:         "medium",
		ClaimDate:        time.Now().AddDate(0, 0, -5),
		Status:           "in_progress",
		StatusUpdatedAt:  time.Now(),
		CustomerName:     "John Doe",
		CustomerEmail:    "john.doe@example.com",
		Priority:         "medium",
		CreatedAt:        time.Now().AddDate(0, 0, -5),
		UpdatedAt:        time.Now(),
	}

	h.logger.Info("Successfully assigned technician to warranty claim (mock)",
		"claim_id", claimID,
		"technician_id", technicianID,
		"assigned_by", userID,
		"notes", req.Notes,
	)

	utils.SuccessResponse(c, http.StatusOK, "Technician assigned to warranty claim successfully", mockClaim)
}

// CompleteClaim handles claim completion requests
// @Summary Complete warranty claim
// @Description Mark a warranty claim as completed
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Claim ID"
// @Param request body dto.ClaimCompletionRequest true "Completion request"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyClaimResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/{id}/complete [put]
func (h *WarrantyClaimHandler) CompleteClaim(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		h.logger.Warn("Invalid claim ID format", "claim_id", claimIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid claim ID format", nil)
		return
	}

	var req dto.ClaimCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind completion request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// TODO: Implement actual claim completion logic when usecase is ready
	// For now, return a mock response
	mockClaim := &dto.WarrantyClaimResponse{
		ID:               claimID.String(),
		ClaimNumber:      "WC-2024-001",
		BarcodeID:        uuid.New().String(),
		CustomerID:       uuid.New().String(),
		ProductID:        uuid.New().String(),
		StorefrontID:     uuid.New().String(),
		IssueDescription: "Device not working properly after 3 months",
		IssueCategory:    "hardware",
		IssueDate:        time.Now().AddDate(0, 0, -7),
		Severity:         "medium",
		ClaimDate:        time.Now().AddDate(0, 0, -5),
		CompletedAt:      &time.Time{},
		Status:           "completed",
		StatusUpdatedAt:  time.Now(),
		CustomerName:     "John Doe",
		CustomerEmail:    "john.doe@example.com",
		Priority:         "medium",
		CreatedAt:        time.Now().AddDate(0, 0, -5),
		UpdatedAt:        time.Now(),
	}
	*mockClaim.CompletedAt = time.Now()

	h.logger.Info("Successfully completed warranty claim (mock)",
		"claim_id", claimID,
		"completed_by", userID,
		"resolution", req.Resolution,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty claim completed successfully", mockClaim)
}

// GetClaimStatistics handles claim statistics requests
// @Summary Get warranty claim statistics
// @Description Get comprehensive statistics and analytics for warranty claims
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param date_from query string false "Statistics from date (YYYY-MM-DD)"
// @Param date_to query string false "Statistics to date (YYYY-MM-DD)"
// @Param storefront_id query string false "Filter by storefront ID"
// @Param product_id query string false "Filter by product ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyClaimStatsResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/statistics [get]
func (h *WarrantyClaimHandler) GetClaimStatistics(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Parse query parameters
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")
	storefrontID := c.Query("storefront_id")
	productID := c.Query("product_id")

	// TODO: Implement actual statistics logic when usecase is ready
	// For now, return a mock response
	mockStats := &dto.WarrantyClaimStatsResponse{
		TotalClaims: 150,
		ClaimsByStatus: map[string]int{
			"pending":     25,
			"validated":   30,
			"in_progress": 40,
			"completed":   45,
			"rejected":    10,
		},
		ClaimsBySeverity: map[string]int{
			"low":      50,
			"medium":   70,
			"high":     25,
			"critical": 5,
		},
		ClaimsByCategory: map[string]int{
			"hardware":    60,
			"software":    30,
			"performance": 25,
			"defect":      20,
			"damage":      10,
			"other":       5,
		},
		ClaimsByPriority: map[string]int{
			"low":    60,
			"medium": 70,
			"high":   20,
		},
		ClaimsThisMonth: 25,
		ClaimsLastMonth: 20,
	}

	h.logger.Info("Successfully retrieved warranty claim statistics (mock)",
		"filters", map[string]string{
			"date_from":     dateFrom,
			"date_to":       dateTo,
			"storefront_id": storefrontID,
			"product_id":    productID,
		},
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty claim statistics retrieved successfully", mockStats)
}

// AddClaimNotes handles adding notes to claims
// @Summary Add notes to warranty claim
// @Description Add administrative notes to a warranty claim
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Claim ID"
// @Param request body dto.ClaimNotesRequest true "Notes request"
// @Success 201 {object} dto.SuccessResponse{data=dto.ClaimTimelineResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/{id}/notes [post]
func (h *WarrantyClaimHandler) AddClaimNotes(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	claimIDStr := c.Param("id")
	claimID, err := uuid.Parse(claimIDStr)
	if err != nil {
		h.logger.Warn("Invalid claim ID format", "claim_id", claimIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid claim ID format", nil)
		return
	}

	var req dto.ClaimNotesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind notes request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// TODO: Implement actual notes addition logic when usecase is ready
	// For now, return a mock response
	mockTimelineEntry := &dto.ClaimTimelineResponse{
		ID:          uuid.New().String(),
		ClaimID:     claimID.String(),
		EventType:   "admin_note",
		Description: req.Notes,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		IsVisible:   req.IsVisible,
	}

	h.logger.Info("Successfully added notes to warranty claim (mock)",
		"claim_id", claimID,
		"added_by", userID,
		"notes_length", len(req.Notes),
		"is_visible", req.IsVisible,
	)

	utils.SuccessResponse(c, http.StatusCreated, "Notes added to warranty claim successfully", mockTimelineEntry)
}

// BulkUpdateClaimStatus handles bulk status updates
// @Summary Bulk update claim status
// @Description Update status for multiple warranty claims
// @Tags warranty-claims
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BulkClaimStatusUpdateRequest true "Bulk update request"
// @Success 200 {object} dto.SuccessResponse{data=dto.BatchResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/claims/bulk/status [put]
func (h *WarrantyClaimHandler) BulkUpdateClaimStatus(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	var req dto.BulkClaimStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind bulk update request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate claim IDs
	for _, claimIDStr := range req.ClaimIDs {
		if _, err := uuid.Parse(claimIDStr); err != nil {
			h.logger.Warn("Invalid claim ID in bulk request", "claim_id", claimIDStr)
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid claim ID format: "+claimIDStr, nil)
			return
		}
	}

	// TODO: Implement actual bulk status update logic when usecase is ready
	// For now, return a mock response
	batchResponse := &dto.BatchResponse{
		TotalProcessed: len(req.ClaimIDs),
		SuccessCount:   len(req.ClaimIDs),
		FailureCount:   0,
		ProcessingTime: "0.5s",
		Timestamp:      time.Now(),
	}

	h.logger.Info("Successfully updated claim statuses in bulk (mock)",
		"claim_count", len(req.ClaimIDs),
		"new_status", req.Status,
		"updated_by", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Claim statuses updated successfully", batchResponse)
}