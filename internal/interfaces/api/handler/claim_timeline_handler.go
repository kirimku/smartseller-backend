package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
)

// ClaimTimelineHandler handles claim timeline-related HTTP requests
type ClaimTimelineHandler struct {
	logger *slog.Logger
}

// NewClaimTimelineHandler creates a new claim timeline handler
func NewClaimTimelineHandler() *ClaimTimelineHandler {
	return &ClaimTimelineHandler{
		logger: slog.Default(),
	}
}

// GetClaimTimeline handles GET /api/v1/admin/warranty/claims/:claim_id/timeline
func (h *ClaimTimelineHandler) GetClaimTimeline(c *gin.Context) {
	claimID := c.Param("claim_id")

	// Validate claim ID
	if claimID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID is required",
		})
		return
	}

	// Mock implementation - replace with actual service call
	timeline := []*dto.ClaimTimelineResponse{
		{
			ID:          "timeline_001",
			ClaimID:     claimID,
			EventType:   "claim_submitted",
			Description: "Claim submitted by customer",
			CreatedBy:   "customer_123",
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			IsVisible:   true,
		},
		{
			ID:          "timeline_002",
			ClaimID:     claimID,
			EventType:   "claim_validated",
			Description: "Claim validated by admin",
			CreatedBy:   "admin_456",
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			IsVisible:   true,
		},
		{
			ID:          "timeline_003",
			ClaimID:     claimID,
			EventType:   "claim_assigned",
			Description: "Claim assigned to technician",
			CreatedBy:   "admin_456",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			IsVisible:   true,
		},
		{
			ID:          "timeline_004",
			ClaimID:     claimID,
			EventType:   "repair_started",
			Description: "Repair process started",
			CreatedBy:   "technician_789",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			IsVisible:   true,
		},
	}

	response := dto.ClaimTimelineListResponse{
		Timeline: timeline,
		Total:    len(timeline),
	}

	h.logger.Info("Retrieved claim timeline", "claim_id", claimID, "entries", len(timeline))
	c.JSON(http.StatusOK, response)
}

// CreateTimelineEntry handles POST /api/v1/admin/warranty/claims/:claim_id/timeline
func (h *ClaimTimelineHandler) CreateTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")

	// Validate claim ID
	if claimID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID is required",
		})
		return
	}

	var req dto.ClaimTimelineCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
		return
	}

	// Validate request
	if req.EventType == "" || req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Event type and description are required",
		})
		return
	}

	// Mock implementation - replace with actual service call
	timelineEntry := dto.ClaimTimelineResponse{
		ID:          "timeline_new",
		ClaimID:     claimID,
		EventType:   req.EventType,
		Description: req.Description,
		CreatedBy:   "current_user", // Replace with actual user ID
		CreatedAt:   time.Now(),
		IsVisible:   req.IsVisible,
	}

	h.logger.Info("Created timeline entry", "claim_id", claimID, "event_type", req.EventType)
	c.JSON(http.StatusCreated, timelineEntry)
}

// GetTimelineEntry handles GET /api/v1/admin/warranty/claims/:claim_id/timeline/:entry_id
func (h *ClaimTimelineHandler) GetTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	entryID := c.Param("entry_id")

	// Validate parameters
	if claimID == "" || entryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID and entry ID are required",
		})
		return
	}

	// Mock implementation - replace with actual service call
	timelineEntry := dto.ClaimTimelineResponse{
		ID:          entryID,
		ClaimID:     claimID,
		EventType:   "note_added",
		Description: "Additional notes added by admin",
		CreatedBy:   "admin_456",
		CreatedAt:   time.Now().Add(-6 * time.Hour),
		IsVisible:   true,
	}

	h.logger.Info("Retrieved timeline entry", "claim_id", claimID, "entry_id", entryID)
	c.JSON(http.StatusOK, timelineEntry)
}

// UpdateTimelineEntry handles PUT /api/v1/admin/warranty/claims/:claim_id/timeline/:entry_id
func (h *ClaimTimelineHandler) UpdateTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	entryID := c.Param("entry_id")

	// Validate parameters
	if claimID == "" || entryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID and entry ID are required",
		})
		return
	}

	var req dto.ClaimTimelineCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
		return
	}

	// Validate request
	if req.Description == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Description is required",
		})
		return
	}

	// Mock implementation - replace with actual service call
	timelineEntry := dto.ClaimTimelineResponse{
		ID:          entryID,
		ClaimID:     claimID,
		EventType:   req.EventType,
		Description: req.Description,
		CreatedBy:   "admin_456",
		CreatedAt:   time.Now().Add(-6 * time.Hour),
		IsVisible:   req.IsVisible,
	}

	h.logger.Info("Updated timeline entry", "claim_id", claimID, "entry_id", entryID)
	c.JSON(http.StatusOK, timelineEntry)
}

// DeleteTimelineEntry handles DELETE /api/v1/admin/warranty/claims/:claim_id/timeline/:entry_id
func (h *ClaimTimelineHandler) DeleteTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	entryID := c.Param("entry_id")

	// Validate parameters
	if claimID == "" || entryID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_failed",
			"message": "Claim ID and entry ID are required",
		})
		return
	}

	// Mock implementation - replace with actual service call
	// In real implementation, you would:
	// 1. Verify timeline entry belongs to claim
	// 2. Check user permissions
	// 3. Soft delete or hard delete the entry
	// 4. Log the deletion action

	h.logger.Info("Deleted timeline entry", "claim_id", claimID, "entry_id", entryID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Timeline entry deleted successfully",
	})
}
