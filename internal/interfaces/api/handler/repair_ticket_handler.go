package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// RepairTicketHandler handles repair ticket management operations
type RepairTicketHandler struct {
	// In a real implementation, this would include service dependencies
	// repairTicketService service.RepairTicketService
}

// NewRepairTicketHandler creates a new repair ticket handler
func NewRepairTicketHandler() *RepairTicketHandler {
	return &RepairTicketHandler{}
}

// CreateRepairTicket creates a new repair ticket for a warranty claim
// @Summary Create repair ticket
// @Description Create a new repair ticket for a warranty claim
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param id path string true "Claim ID"
// @Param request body dto.RepairTicketCreateRequest true "Repair ticket creation request"
// @Success 201 {object} dto.RepairTicketResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/claims/{id}/repair-tickets [post]
func (h *RepairTicketHandler) CreateRepairTicket(c *gin.Context) {
	claimID := c.Param("id")
	if claimID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Claim ID is required", nil)
		return
	}

	var req dto.RepairTicketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	ticketID := uuid.New().String()
	
	response := &dto.RepairTicketResponse{
		ID:                       ticketID,
		TicketNumber:             "RPR-2024-" + strconv.Itoa(int(now.Unix())),
		ClaimID:                  claimID,
		ClaimNumber:              "WAR-2024-001234",
		Status:                   "pending",
		Priority:                 req.Priority,
		EstimatedHours:           req.EstimatedHours,
		Description:              req.Description,
		RequiredParts:            req.RequiredParts,
		SpecialInstructions:      req.SpecialInstructions,
		LaborCost:                decimal.NewFromFloat(0),
		PartsCost:                decimal.NewFromFloat(0),
		TotalCost:                decimal.NewFromFloat(0),
		QualityCheckStatus:       "pending",
		CustomerApprovalRequired: req.CustomerApprovalRequired,
		CustomerApprovalStatus:   "pending",
		CreatedBy:                "550e8400-e29b-41d4-a716-446655440004",
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	c.JSON(http.StatusCreated, response)
}

// ListRepairTickets retrieves a paginated list of repair tickets
// @Summary List repair tickets
// @Description Get a paginated list of repair tickets with filtering options
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort by field" Enums(created_at,priority,status,estimated_completion_date)
// @Param sort_order query string false "Sort order" Enums(asc,desc)
// @Param status query []string false "Filter by status"
// @Param priority query []string false "Filter by priority"
// @Param technician_id query string false "Filter by technician ID"
// @Param search_term query string false "Search term"
// @Success 200 {object} dto.RepairTicketListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/repair-tickets [get]
func (h *RepairTicketHandler) ListRepairTickets(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	mockTickets := []dto.RepairTicketResponse{
		{
			ID:                      "550e8400-e29b-41d4-a716-446655440000",
			TicketNumber:            "RPR-2024-001234",
			ClaimID:                 "550e8400-e29b-41d4-a716-446655440001",
			ClaimNumber:             "WAR-2024-001234",
			Status:                  "assigned",
			Priority:                "high",
			AssignedTechnicianID:    stringPtr("550e8400-e29b-41d4-a716-446655440002"),
			TechnicianName:          stringPtr("John Smith"),
			AssignedAt:              &now,
			EstimatedHours:          decimal.NewFromFloat(4.5),
			EstimatedCompletionDate: timePtr(now.Add(48 * time.Hour)),
			Description:             "Replace faulty motherboard and test all components",
			RequiredParts:           []string{"motherboard", "thermal_paste"},
			SpecialInstructions:     "Handle with care - customer reported water damage",
			LaborCost:               decimal.NewFromFloat(120.00),
			PartsCost:               decimal.NewFromFloat(85.50),
			TotalCost:               decimal.NewFromFloat(205.50),
			QualityCheckStatus:      "pending",
			CustomerApprovalRequired: true,
			CustomerApprovalStatus:  "pending",
			CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
			CreatedAt:               now.Add(-24 * time.Hour),
			UpdatedAt:               now,
		},
		{
			ID:                      "550e8400-e29b-41d4-a716-446655440005",
			TicketNumber:            "RPR-2024-001235",
			ClaimID:                 "550e8400-e29b-41d4-a716-446655440006",
			ClaimNumber:             "WAR-2024-001235",
			Status:                  "completed",
			Priority:                "normal",
			AssignedTechnicianID:    stringPtr("550e8400-e29b-41d4-a716-446655440007"),
			TechnicianName:          stringPtr("Jane Doe"),
			AssignedAt:              timePtr(now.Add(-72 * time.Hour)),
			EstimatedHours:          decimal.NewFromFloat(2.0),
			ActualHours:             decimalPtr(decimal.NewFromFloat(1.8)),
			EstimatedCompletionDate: timePtr(now.Add(-24 * time.Hour)),
			ActualCompletionDate:    timePtr(now.Add(-12 * time.Hour)),
			Description:             "Replace screen assembly",
			RequiredParts:           []string{"screen", "adhesive"},
			UsedParts:               []string{"screen", "adhesive", "screws"},
			SpecialInstructions:     "Standard screen replacement procedure",
			RepairNotes:             stringPtr("Successfully replaced screen, all tests passed"),
			LaborCost:               decimal.NewFromFloat(60.00),
			PartsCost:               decimal.NewFromFloat(45.00),
			TotalCost:               decimal.NewFromFloat(105.00),
			QualityCheckStatus:      "approved",
			QualityCheckedBy:        stringPtr("550e8400-e29b-41d4-a716-446655440008"),
			QualityCheckDate:        timePtr(now.Add(-6 * time.Hour)),
			QualityCheckNotes:       stringPtr("All functionality verified, repair approved"),
			CustomerApprovalRequired: false,
			CustomerApprovalStatus:  "not_required",
			CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
			CreatedAt:               now.Add(-72 * time.Hour),
			UpdatedAt:               now.Add(-12 * time.Hour),
		},
	}

	response := &dto.RepairTicketListResponse{
		Tickets: mockTickets,
		Pagination: dto.PaginationResponse{
			Page:       page,
			Limit:      pageSize,
			Total:      2,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
		Filters: dto.RepairTicketFiltersResponse{
			AvailableStatuses:             []string{"pending", "assigned", "in_progress", "completed", "cancelled"},
			AvailablePriorities:           []string{"low", "normal", "high", "urgent"},
			AvailableQualityCheckStatuses: []string{"pending", "approved", "rejected"},
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetRepairTicket retrieves a specific repair ticket by ID
// @Summary Get repair ticket
// @Description Get detailed information about a specific repair ticket
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param id path string true "Repair Ticket ID"
// @Success 200 {object} dto.RepairTicketResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/repair-tickets/{id} [get]
func (h *RepairTicketHandler) GetRepairTicket(c *gin.Context) {
	claimID := c.Param("id")
	ticketID := c.Param("ticketId")
	
	if claimID == "" || ticketID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Claim ID and Ticket ID are required", nil)
		return
	}

	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	response := &dto.RepairTicketResponse{
		ID:                      ticketID,
		TicketNumber:            "RPR-2024-001234",
		ClaimID:                 "550e8400-e29b-41d4-a716-446655440001",
		ClaimNumber:             "WAR-2024-001234",
		Status:                  "assigned",
		Priority:                "high",
		AssignedTechnicianID:    stringPtr("550e8400-e29b-41d4-a716-446655440002"),
		TechnicianName:          stringPtr("John Smith"),
		AssignedAt:              &now,
		EstimatedHours:          decimal.NewFromFloat(4.5),
		EstimatedCompletionDate: timePtr(now.Add(48 * time.Hour)),
		Description:             "Replace faulty motherboard and test all components",
		RequiredParts:           []string{"motherboard", "thermal_paste"},
		SpecialInstructions:     "Handle with care - customer reported water damage",
		LaborCost:               decimal.NewFromFloat(120.00),
		PartsCost:               decimal.NewFromFloat(85.50),
		TotalCost:               decimal.NewFromFloat(205.50),
		QualityCheckStatus:      "pending",
		CustomerApprovalRequired: true,
		CustomerApprovalStatus:  "pending",
		CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
		CreatedAt:               now.Add(-24 * time.Hour),
		UpdatedAt:               now,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateRepairTicket updates an existing repair ticket
// @Summary Update repair ticket
// @Description Update details of an existing repair ticket
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param id path string true "Repair Ticket ID"
// @Param request body dto.RepairTicketUpdateRequest true "Repair ticket update request"
// @Success 200 {object} dto.RepairTicketResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/repair-tickets/{id} [put]
func (h *RepairTicketHandler) UpdateRepairTicket(c *gin.Context) {
	claimID := c.Param("id")
	ticketID := c.Param("ticketId")
	
	if claimID == "" || ticketID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Claim ID and Ticket ID are required", nil)
		return
	}

	var req dto.RepairTicketUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	response := &dto.RepairTicketResponse{
		ID:                      ticketID,
		TicketNumber:            "RPR-2024-001234",
		ClaimID:                 "550e8400-e29b-41d4-a716-446655440001",
		ClaimNumber:             "WAR-2024-001234",
		Status:                  "assigned",
		Priority:                getStringValue(req.Priority, "high"),
		AssignedTechnicianID:    stringPtr("550e8400-e29b-41d4-a716-446655440002"),
		TechnicianName:          stringPtr("John Smith"),
		AssignedAt:              &now,
		EstimatedHours:          getDecimalValue(req.EstimatedHours, decimal.NewFromFloat(4.5)),
		EstimatedCompletionDate: timePtr(now.Add(48 * time.Hour)),
		Description:             getStringValue(req.Description, "Replace faulty motherboard and test all components"),
		RequiredParts:           req.RequiredParts,
		UsedParts:               req.UsedParts,
		SpecialInstructions:     getStringValue(req.SpecialInstructions, "Handle with care - customer reported water damage"),
		RepairNotes:             req.RepairNotes,
		LaborCost:               getDecimalValue(req.LaborCost, decimal.NewFromFloat(120.00)),
		PartsCost:               getDecimalValue(req.PartsCost, decimal.NewFromFloat(85.50)),
		TotalCost:               decimal.NewFromFloat(205.50),
		QualityCheckStatus:      "pending",
		CustomerApprovalRequired: true,
		CustomerApprovalStatus:  "pending",
		CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
		CreatedAt:               now.Add(-24 * time.Hour),
		UpdatedAt:               now,
	}

	c.JSON(http.StatusOK, response)
}

// AssignTechnician assigns a technician to a repair ticket
// @Summary Assign technician
// @Description Assign a technician to a repair ticket
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param id path string true "Repair Ticket ID"
// @Param request body dto.RepairTicketAssignmentRequest true "Technician assignment request"
// @Success 200 {object} dto.RepairTicketResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/repair-tickets/{id}/assign [put]
func (h *RepairTicketHandler) AssignTechnician(c *gin.Context) {
	claimID := c.Param("id")
	ticketID := c.Param("ticketId")
	
	if claimID == "" || ticketID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Claim ID and Ticket ID are required", nil)
		return
	}

	var req dto.RepairTicketAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	response := &dto.RepairTicketResponse{
		ID:                      ticketID,
		TicketNumber:            "RPR-2024-001234",
		ClaimID:                 "550e8400-e29b-41d4-a716-446655440001",
		ClaimNumber:             "WAR-2024-001234",
		Status:                  "assigned",
		Priority:                "high",
		AssignedTechnicianID:    &req.TechnicianID,
		TechnicianName:          stringPtr("John Smith"),
		AssignedAt:              &now,
		EstimatedHours:          decimal.NewFromFloat(4.5),
		EstimatedCompletionDate: req.EstimatedCompletionDate,
		Description:             "Replace faulty motherboard and test all components",
		RequiredParts:           []string{"motherboard", "thermal_paste"},
		SpecialInstructions:     "Handle with care - customer reported water damage",
		LaborCost:               decimal.NewFromFloat(120.00),
		PartsCost:               decimal.NewFromFloat(85.50),
		TotalCost:               decimal.NewFromFloat(205.50),
		QualityCheckStatus:      "pending",
		CustomerApprovalRequired: true,
		CustomerApprovalStatus:  "pending",
		CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
		CreatedAt:               now.Add(-24 * time.Hour),
		UpdatedAt:               now,
	}

	c.JSON(http.StatusOK, response)
}

// CompleteRepair marks a repair ticket as completed
// @Summary Complete repair
// @Description Mark a repair ticket as completed with final details
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param id path string true "Repair Ticket ID"
// @Param request body dto.RepairTicketCompletionRequest true "Repair completion request"
// @Success 200 {object} dto.RepairTicketResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/repair-tickets/{id}/complete [put]
func (h *RepairTicketHandler) CompleteRepair(c *gin.Context) {
	claimID := c.Param("id")
	ticketID := c.Param("ticketId")
	
	if claimID == "" || ticketID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Claim ID and Ticket ID are required", nil)
		return
	}

	var req dto.RepairTicketCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	totalCost := req.LaborCost.Add(req.PartsCost)
	
	response := &dto.RepairTicketResponse{
		ID:                      ticketID,
		TicketNumber:            "RPR-2024-001234",
		ClaimID:                 "550e8400-e29b-41d4-a716-446655440001",
		ClaimNumber:             "WAR-2024-001234",
		Status:                  "completed",
		Priority:                "high",
		AssignedTechnicianID:    stringPtr("550e8400-e29b-41d4-a716-446655440002"),
		TechnicianName:          stringPtr("John Smith"),
		AssignedAt:              timePtr(now.Add(-48 * time.Hour)),
		EstimatedHours:          decimal.NewFromFloat(4.5),
		ActualHours:             &req.ActualHours,
		EstimatedCompletionDate: timePtr(now.Add(-24 * time.Hour)),
		ActualCompletionDate:    &now,
		Description:             "Replace faulty motherboard and test all components",
		RequiredParts:           []string{"motherboard", "thermal_paste"},
		UsedParts:               req.UsedParts,
		SpecialInstructions:     "Handle with care - customer reported water damage",
		RepairNotes:             &req.RepairNotes,
		LaborCost:               req.LaborCost,
		PartsCost:               req.PartsCost,
		TotalCost:               totalCost,
		QualityCheckStatus:      "pending",
		CustomerApprovalRequired: true,
		CustomerApprovalStatus:  "approved",
		CustomerApprovedAt:      timePtr(now.Add(-24 * time.Hour)),
		CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
		CreatedAt:               now.Add(-48 * time.Hour),
		UpdatedAt:               now,
	}

	c.JSON(http.StatusOK, response)
}

// QualityCheck performs quality control approval/rejection
// @Summary Quality check
// @Description Perform quality control approval or rejection of a repair
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param id path string true "Repair Ticket ID"
// @Param request body dto.RepairTicketQualityCheckRequest true "Quality check request"
// @Success 200 {object} dto.RepairTicketResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/repair-tickets/{id}/quality-check [put]
func (h *RepairTicketHandler) QualityCheck(c *gin.Context) {
	claimID := c.Param("id")
	ticketID := c.Param("ticketId")
	
	if claimID == "" || ticketID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Claim ID and Ticket ID are required", nil)
		return
	}

	var req dto.RepairTicketQualityCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	qcStatus := "approved"
	if req.Action == "reject" {
		qcStatus = "rejected"
	}
	
	response := &dto.RepairTicketResponse{
		ID:                      ticketID,
		TicketNumber:            "RPR-2024-001234",
		ClaimID:                 "550e8400-e29b-41d4-a716-446655440001",
		ClaimNumber:             "WAR-2024-001234",
		Status:                  "completed",
		Priority:                "high",
		AssignedTechnicianID:    stringPtr("550e8400-e29b-41d4-a716-446655440002"),
		TechnicianName:          stringPtr("John Smith"),
		AssignedAt:              timePtr(now.Add(-48 * time.Hour)),
		EstimatedHours:          decimal.NewFromFloat(4.5),
		ActualHours:             decimalPtr(decimal.NewFromFloat(5.2)),
		EstimatedCompletionDate: timePtr(now.Add(-24 * time.Hour)),
		ActualCompletionDate:    timePtr(now.Add(-6 * time.Hour)),
		Description:             "Replace faulty motherboard and test all components",
		RequiredParts:           []string{"motherboard", "thermal_paste"},
		UsedParts:               []string{"motherboard", "thermal_paste", "screws"},
		SpecialInstructions:     "Handle with care - customer reported water damage",
		RepairNotes:             stringPtr("Successfully replaced motherboard, all tests passed"),
		LaborCost:               decimal.NewFromFloat(120.00),
		PartsCost:               decimal.NewFromFloat(85.50),
		TotalCost:               decimal.NewFromFloat(205.50),
		QualityCheckStatus:      qcStatus,
		QualityCheckedBy:        stringPtr("550e8400-e29b-41d4-a716-446655440008"),
		QualityCheckDate:        &now,
		QualityCheckNotes:       &req.Notes,
		CustomerApprovalRequired: true,
		CustomerApprovalStatus:  "approved",
		CustomerApprovedAt:      timePtr(now.Add(-24 * time.Hour)),
		CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
		CreatedAt:               now.Add(-48 * time.Hour),
		UpdatedAt:               now,
	}

	c.JSON(http.StatusOK, response)
}

// GetRepairStatistics retrieves repair ticket analytics and statistics
// @Summary Get repair statistics
// @Description Get comprehensive analytics and statistics for repair tickets
// @Tags repair-tickets
// @Accept json
// @Produce json
// @Param date_from query string false "Start date for statistics (YYYY-MM-DD)"
// @Param date_to query string false "End date for statistics (YYYY-MM-DD)"
// @Param technician_id query string false "Filter by technician ID"
// @Success 200 {object} dto.RepairTicketStatisticsResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/admin/warranty/repair-tickets/statistics [get]
func (h *RepairTicketHandler) GetRepairStatistics(c *gin.Context) {
	// Mock implementation - in real scenario, this would call the service layer
	now := time.Now()
	
	response := &dto.RepairTicketStatisticsResponse{
		TotalTickets: 450,
		TicketsByStatus: map[string]int{
			"pending":     45,
			"assigned":    120,
			"in_progress": 85,
			"completed":   180,
			"cancelled":   20,
		},
		TicketsByPriority: map[string]int{
			"low":    100,
			"normal": 200,
			"high":   120,
			"urgent": 30,
		},
		TicketsByTechnician: map[string]int{
			"John Smith": 45,
			"Jane Doe":   38,
			"Bob Wilson": 42,
		},
		AverageRepairTime:    decimal.NewFromFloat(4.8),
		AverageLaborCost:     decimal.NewFromFloat(95.50),
		AveragePartsCost:     decimal.NewFromFloat(67.25),
		TotalLaborCost:       decimal.NewFromFloat(42975.00),
		TotalPartsCost:       decimal.NewFromFloat(30262.50),
		TotalRepairCost:      decimal.NewFromFloat(73237.50),
		CompletionRate:       decimal.NewFromFloat(88.9),
		QualityApprovalRate:  decimal.NewFromFloat(94.2),
		CustomerApprovalRate: decimal.NewFromFloat(91.5),
		TicketsThisMonth:     85,
		TicketsLastMonth:     72,
		GrowthRate:           decimal.NewFromFloat(18.1),
		TopTechnicians: []dto.TechnicianStatsResponse{
			{
				TechnicianID:        "550e8400-e29b-41d4-a716-446655440002",
				TechnicianName:      "John Smith",
				AssignedTickets:     45,
				CompletedTickets:    42,
				CompletionRate:      decimal.NewFromFloat(93.3),
				AverageRepairTime:   decimal.NewFromFloat(4.2),
				QualityApprovalRate: decimal.NewFromFloat(97.6),
				TotalRevenue:        decimal.NewFromFloat(8950.00),
			},
			{
				TechnicianID:        "550e8400-e29b-41d4-a716-446655440007",
				TechnicianName:      "Jane Doe",
				AssignedTickets:     38,
				CompletedTickets:    36,
				CompletionRate:      decimal.NewFromFloat(94.7),
				AverageRepairTime:   decimal.NewFromFloat(3.8),
				QualityApprovalRate: decimal.NewFromFloat(100.0),
				TotalRevenue:        decimal.NewFromFloat(7650.00),
			},
		},
		RecentTickets: []dto.RepairTicketResponse{
			{
				ID:                      "550e8400-e29b-41d4-a716-446655440000",
				TicketNumber:            "RPR-2024-001234",
				ClaimID:                 "550e8400-e29b-41d4-a716-446655440001",
				ClaimNumber:             "WAR-2024-001234",
				Status:                  "assigned",
				Priority:                "high",
				AssignedTechnicianID:    stringPtr("550e8400-e29b-41d4-a716-446655440002"),
				TechnicianName:          stringPtr("John Smith"),
				AssignedAt:              &now,
				EstimatedHours:          decimal.NewFromFloat(4.5),
				EstimatedCompletionDate: timePtr(now.Add(48 * time.Hour)),
				Description:             "Replace faulty motherboard and test all components",
				RequiredParts:           []string{"motherboard", "thermal_paste"},
				SpecialInstructions:     "Handle with care - customer reported water damage",
				LaborCost:               decimal.NewFromFloat(120.00),
				PartsCost:               decimal.NewFromFloat(85.50),
				TotalCost:               decimal.NewFromFloat(205.50),
				QualityCheckStatus:      "pending",
				CustomerApprovalRequired: true,
				CustomerApprovalStatus:  "pending",
				CreatedBy:               "550e8400-e29b-41d4-a716-446655440004",
				CreatedAt:               now.Add(-24 * time.Hour),
				UpdatedAt:               now,
			},
		},
	}

	c.JSON(http.StatusOK, response)
}

// Helper functions for handling optional fields
func stringPtr(s string) *string {
	return &s
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func getStringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func getDecimalValue(ptr *decimal.Decimal, defaultValue decimal.Decimal) decimal.Decimal {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}