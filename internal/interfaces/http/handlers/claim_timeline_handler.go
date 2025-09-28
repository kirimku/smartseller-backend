package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/logger"
)

// ClaimTimelineHandler handles claim timeline related HTTP requests
type ClaimTimelineHandler struct {
	logger zerolog.Logger
}

// NewClaimTimelineHandler creates a new claim timeline handler
func NewClaimTimelineHandler() *ClaimTimelineHandler {
	return &ClaimTimelineHandler{
		logger: logger.Logger,
	}
}

// GetClaimTimeline handles GET /api/v1/admin/warranty/claims/{claim_id}/timeline
func (h *ClaimTimelineHandler) GetClaimTimeline(c *gin.Context) {
	claimID := c.Param("id")
	if claimID == "" {
		h.logger.Warn().Str("endpoint", "GetClaimTimeline").Msg("Missing claim ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Claim ID is required"})
		return
	}

	h.logger.Info().Str("claim_id", claimID).Str("endpoint", "GetClaimTimeline").Msg("Fetching claim timeline")

	// Mock timeline data - replace with actual service call
	timeline := []dto.CustomerClaimTimelineItem{
		{
			ID:          "timeline-1",
			Event:       "status_change",
			Description: "Warranty claim has been submitted and is under review",
			Timestamp:   time.Now().Add(-48 * time.Hour),
			Actor:       "System",
			ActorType:   "system",
			Details:     map[string]interface{}{"status": "submitted"},
		},
		{
			ID:          "timeline-2",
			Event:       "status_change",
			Description: "Claim is being reviewed by our technical team",
			Timestamp:   time.Now().Add(-24 * time.Hour),
			Actor:       "Technical Team",
			ActorType:   "admin",
			Details:     map[string]interface{}{"status": "under_review"},
		},
		{
			ID:          "timeline-3",
			Event:       "note_added",
			Description: "Initial technical assessment completed. Repair required.",
			Timestamp:   time.Now().Add(-12 * time.Hour),
			Actor:       "John Technician",
			ActorType:   "technician",
			Details:     map[string]interface{}{"note": "Device requires motherboard replacement"},
		},
		{
			ID:          "timeline-4",
			Event:       "status_change",
			Description: "Claim approved. Device will be repaired under warranty.",
			Timestamp:   time.Now().Add(-6 * time.Hour),
			Actor:       "Admin",
			ActorType:   "admin",
			Details:     map[string]interface{}{"status": "approved"},
		},
		{
			ID:          "timeline-5",
			Event:       "status_change",
			Description: "Repair work has begun on your device",
			Timestamp:   time.Now().Add(-2 * time.Hour),
			Actor:       "Repair Team",
			ActorType:   "technician",
			Details:     map[string]interface{}{"status": "in_progress"},
		},
	}

	h.logger.Info().Str("claim_id", claimID).Int("timeline_count", len(timeline)).Msg("Timeline fetched successfully")

	c.JSON(http.StatusOK, gin.H{
		"timeline": timeline,
		"total":    len(timeline),
	})
}

// CreateTimelineEntryRequest represents a request to create a timeline entry
type CreateTimelineEntryRequest struct {
	Event       string                 `json:"event" validate:"required"`
	Description string                 `json:"description" validate:"required"`
	ActorType   string                 `json:"actor_type" validate:"required,oneof=customer admin technician system"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// CreateTimelineEntry handles POST /api/v1/admin/warranty/claims/{claim_id}/timeline
func (h *ClaimTimelineHandler) CreateTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")

	// Validate claim ID format
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	// Parse request body
	var req CreateTimelineEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse timeline entry request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("event", req.Event).
		Msg("Creating timeline entry")

	// Mock timeline entry creation - replace with actual service call
	entry := dto.CustomerClaimTimelineItem{
		ID:          uuid.New().String(),
		Event:       req.Event,
		Description: req.Description,
		Timestamp:   time.Now(),
		Actor:       "Admin User", // In real implementation, get from auth context
		ActorType:   req.ActorType,
		Details:     req.Details,
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("entry_id", entry.ID).
		Msg("Successfully created timeline entry")

	c.JSON(http.StatusCreated, gin.H{
		"entry":   entry,
		"message": "Timeline entry created successfully",
	})
}

// UpdateTimelineEntryRequest represents a request to update a timeline entry
type UpdateTimelineEntryRequest struct {
	Event       string                 `json:"event,omitempty"`
	Description string                 `json:"description,omitempty"`
	ActorType   string                 `json:"actor_type,omitempty" validate:"omitempty,oneof=customer admin technician system"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// UpdateTimelineEntry handles PUT /api/v1/admin/warranty/claims/{claim_id}/timeline/{entry_id}
func (h *ClaimTimelineHandler) UpdateTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	entryID := c.Param("entry_id")

	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(entryID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid entry ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid entry ID format",
			"message": "Entry ID must be a valid UUID",
		})
		return
	}

	// Parse request body
	var req UpdateTimelineEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("Failed to parse timeline entry update request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"message": err.Error(),
		})
		return
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("entry_id", entryID).
		Msg("Updating timeline entry")

	// Mock timeline entry update - replace with actual service call
	entry := dto.CustomerClaimTimelineItem{
		ID:          entryID,
		Event:       req.Event,
		Description: req.Description,
		Timestamp:   time.Now(),
		Actor:       "Admin User", // This should come from authentication context
		ActorType:   req.ActorType,
		Details:     req.Details,
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("entry_id", entryID).
		Msg("Successfully updated timeline entry")

	c.JSON(http.StatusOK, gin.H{
		"entry":   entry,
		"message": "Timeline entry updated successfully",
	})
}

// DeleteTimelineEntry handles DELETE /api/v1/admin/warranty/claims/{claim_id}/timeline/{entry_id}
func (h *ClaimTimelineHandler) DeleteTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	entryID := c.Param("entry_id")

	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(entryID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid entry ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid entry ID format",
			"message": "Entry ID must be a valid UUID",
		})
		return
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("entry_id", entryID).
		Msg("Deleting timeline entry")

	// Mock timeline entry deletion - replace with actual service call
	// In a real implementation, you would:
	// 1. Verify the entry exists and belongs to the claim
	// 2. Check user permissions
	// 3. Soft delete or hard delete the entry

	h.logger.Info().
		Str("claim_id", claimID).
		Str("entry_id", entryID).
		Msg("Successfully deleted timeline entry")

	c.JSON(http.StatusOK, gin.H{
		"message": "Timeline entry deleted successfully",
	})
}

// GetTimelineEntry handles GET /api/v1/admin/warranty/claims/{claim_id}/timeline/{entry_id}
func (h *ClaimTimelineHandler) GetTimelineEntry(c *gin.Context) {
	claimID := c.Param("claim_id")
	entryID := c.Param("entry_id")

	// Validate IDs
	if _, err := uuid.Parse(claimID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid claim ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid claim ID format",
			"message": "Claim ID must be a valid UUID",
		})
		return
	}

	if _, err := uuid.Parse(entryID); err != nil {
		h.logger.Error().Err(err).Msg("Invalid entry ID format")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid entry ID format",
			"message": "Entry ID must be a valid UUID",
		})
		return
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("entry_id", entryID).
		Msg("Fetching timeline entry")

	// Mock timeline entry retrieval - replace with actual service call
	entry := dto.CustomerClaimTimelineItem{
		ID:          entryID,
		Event:       "status_change",
		Description: "Warranty claim has been approved for replacement",
		Timestamp:   time.Now().Add(-2 * time.Hour),
		Actor:       "Manager",
		ActorType:   "admin",
		Details:     map[string]interface{}{"status": "approved"},
	}

	h.logger.Info().
		Str("claim_id", claimID).
		Str("entry_id", entryID).
		Msg("Successfully retrieved timeline entry")

	c.JSON(http.StatusOK, entry)
}
