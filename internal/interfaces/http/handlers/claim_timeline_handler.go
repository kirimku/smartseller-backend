package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"smartseller-backend/internal/application/dto"
	"smartseller-backend/internal/infrastructure/logger"
)

// ClaimTimelineHandler handles claim timeline related HTTP requests
type ClaimTimelineHandler struct {
	logger *logrus.Logger
}

// NewClaimTimelineHandler creates a new claim timeline handler
func NewClaimTimelineHandler() *ClaimTimelineHandler {
	return &ClaimTimelineHandler{
		logger: logger.GetLogger(),
	}
}

// GetClaimTimeline handles GET /api/v1/admin/warranty/claims/{claim_id}/timeline
func (h *ClaimTimelineHandler) GetClaimTimeline(c *gin.Context) {
	claimID := c.Param("claim_id")
	
	// Validate claim ID format
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.WithError(err).Error("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	h.logger.WithField("claim_id", claimID).Info("Getting claim timeline")

	// Mock response - replace with actual service call
	timeline := []*dto.ClaimTimelineResponse{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440000",
			ClaimID:     claimID,
			EventType:   "claim_submitted",
			Description: "Claim submitted by customer",
			CreatedBy:   "550e8400-e29b-41d4-a716-446655440002",
			CreatedAt:   time.Now().Add(-72 * time.Hour),
			IsVisible:   true,
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440001",
			ClaimID:     claimID,
			EventType:   "claim_validated",
			Description: "Claim validated by admin - approved for repair",
			CreatedBy:   "550e8400-e29b-41d4-a716-446655440003",
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			IsVisible:   true,
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440002",
			ClaimID:     claimID,
			EventType:   "claim_assigned",
			Description: "Claim assigned to technician John Smith",
			CreatedBy:   "550e8400-e29b-41d4-a716-446655440003",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			IsVisible:   true,
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440003",
			ClaimID:     claimID,
			EventType:   "repair_started",
			Description: "Repair process started",
			CreatedBy:   "550e8400-e29b-41d4-a716-446655440004",
			CreatedAt:   time.Now().Add(-12 * time.Hour),
			IsVisible:   true,
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440004",
			ClaimID:     claimID,
			EventType:   "note_added",
			Description: "Customer contacted for additional information",
			CreatedBy:   "550e8400-e29b-41d4-a716-446655440003",
			CreatedAt:   time.Now().Add(-6 * time.Hour),
			IsVisible:   true,
		},
	}

	response := &dto.ClaimTimelineListResponse{
		Timeline: timeline,
		Total:    len(timeline),
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":       claimID,
		"timeline_count": len(timeline),
	}).Info("Successfully retrieved claim timeline")

	c.JSON(http.StatusOK, response)
}

// CreateTimelineEntry handles POST /api/v1/admin/warranty/claims/{claim_id}/timeline
func (h *ClaimTimelineHandler) CreateTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	
	// Validate claim ID format
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.WithError(err).Error("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	// Parse request body
	var req dto.ClaimTimelineCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse timeline creation request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":    claimID,
		"event_type":  req.EventType,
		"description": req.Description,
		"is_visible":  req.IsVisible,
	}).Info("Creating timeline entry")

	// Mock timeline entry creation - replace with actual service call
	timelineEntry := &dto.ClaimTimelineResponse{
		ID:          uuid.New().String(),
		ClaimID:     claimID,
		EventType:   req.EventType,
		Description: req.Description,
		CreatedBy:   "550e8400-e29b-41d4-a716-446655440003", // Mock user ID
		CreatedAt:   time.Now(),
		IsVisible:   req.IsVisible,
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":        claimID,
		"timeline_id":     timelineEntry.ID,
		"event_type":      timelineEntry.EventType,
		"timeline_entry":  timelineEntry.Description,
	}).Info("Successfully created timeline entry")

	c.JSON(http.StatusCreated, timelineEntry)
}

// UpdateTimelineEntry handles PUT /api/v1/admin/warranty/claims/{claim_id}/timeline/{timeline_id}
func (h *ClaimTimelineHandler) UpdateTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	timelineID := c.Param("timeline_id")
	
	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.WithError(err).Error("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(timelineID); err != nil {
		h.logger.WithError(err).Error("Invalid timeline ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid timeline ID format",
			"message": "Timeline ID must be a valid UUID",
		})
		return
	}

	// Parse request body
	var req dto.ClaimTimelineCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Error("Failed to parse timeline update request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":    claimID,
		"timeline_id": timelineID,
		"event_type":  req.EventType,
		"description": req.Description,
		"is_visible":  req.IsVisible,
	}).Info("Updating timeline entry")

	// Mock timeline entry update - replace with actual service call
	// In a real implementation, you would:
	// 1. Verify the timeline entry exists and belongs to the claim
	// 2. Check user permissions
	// 3. Update the timeline entry
	// 4. Return the updated entry

	updatedEntry := &dto.ClaimTimelineResponse{
		ID:          timelineID,
		ClaimID:     claimID,
		EventType:   req.EventType,
		Description: req.Description,
		CreatedBy:   "550e8400-e29b-41d4-a716-446655440003", // Mock user ID
		CreatedAt:   time.Now().Add(-1 * time.Hour),         // Mock original creation time
		IsVisible:   req.IsVisible,
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":    claimID,
		"timeline_id": timelineID,
		"event_type":  updatedEntry.EventType,
	}).Info("Successfully updated timeline entry")

	c.JSON(http.StatusOK, updatedEntry)
}

// DeleteTimelineEntry handles DELETE /api/v1/admin/warranty/claims/{claim_id}/timeline/{timeline_id}
func (h *ClaimTimelineHandler) DeleteTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	timelineID := c.Param("timeline_id")
	
	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.WithError(err).Error("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(timelineID); err != nil {
		h.logger.WithError(err).Error("Invalid timeline ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid timeline ID format",
			"message": "Timeline ID must be a valid UUID",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":    claimID,
		"timeline_id": timelineID,
	}).Info("Deleting timeline entry")

	// Mock timeline entry deletion - replace with actual service call
	// In a real implementation, you would:
	// 1. Verify the timeline entry exists and belongs to the claim
	// 2. Check user permissions (only allow deletion of certain types)
	// 3. Soft delete or hard delete the timeline entry
	// 4. Log the deletion action

	h.logger.WithFields(logrus.Fields{
		"claim_id":    claimID,
		"timeline_id": timelineID,
	}).Info("Successfully deleted timeline entry")

	c.JSON(http.StatusOK, gin.H{
		"message": "Timeline entry deleted successfully",
	})
}

// GetTimelineEntry handles GET /api/v1/admin/warranty/claims/{claim_id}/timeline/{timeline_id}
func (h *ClaimTimelineHandler) GetTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	timelineID := c.Param("timeline_id")
	
	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.WithError(err).Error("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(timelineID); err != nil {
		h.logger.WithError(err).Error("Invalid timeline ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid timeline ID format",
			"message": "Timeline ID must be a valid UUID",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":    claimID,
		"timeline_id": timelineID,
	}).Info("Getting timeline entry")

	// Mock timeline entry retrieval - replace with actual service call
	timelineEntry := &dto.ClaimTimelineResponse{
		ID:          timelineID,
		ClaimID:     claimID,
		EventType:   "note_added",
		Description: "Customer contacted for additional information",
		CreatedBy:   "550e8400-e29b-41d4-a716-446655440003",
		CreatedAt:   time.Now().Add(-6 * time.Hour),
		IsVisible:   true,
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id":    claimID,
		"timeline_id": timelineID,
		"event_type":  timelineEntry.EventType,
	}).Info("Successfully retrieved timeline entry")

	c.JSON(http.StatusOK, timelineEntry)
}