package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// StorefrontHandler handles storefront-related HTTP requests
type StorefrontHandler struct {
	storefrontService service.StorefrontService
	validationService service.ValidationService
}

// NewStorefrontHandler creates a new instance of StorefrontHandler
func NewStorefrontHandler(
	storefrontService service.StorefrontService,
	validationService service.ValidationService,
) *StorefrontHandler {
	return &StorefrontHandler{
		storefrontService: storefrontService,
		validationService: validationService,
	}
}

// CreateStorefront handles storefront creation
// @Summary Create a new storefront
// @Description Create a new storefront for a customer
// @Tags storefronts
// @Accept json
// @Produce json
// @Param request body dto.StorefrontCreateRequest true "Storefront creation data"
// @Success 201 {object} dto.StorefrontResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts [post]
func (h *StorefrontHandler) CreateStorefront(c *gin.Context) {
	var req dto.StorefrontCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate the storefront creation request
	validationResult := h.validationService.ValidateStorefrontCreation(c.Request.Context(), &req)
	if validationResult.HasErrors() {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed", validationResult.Errors)
		return
	}

	// Create the storefront
	storefront, err := h.storefrontService.CreateStorefront(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Storefront created successfully", storefront)
}

// GetStorefront handles getting a storefront by ID
// @Summary Get storefront by ID
// @Description Retrieve storefront details by storefront ID
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Success 200 {object} dto.StorefrontResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id} [get]
func (h *StorefrontHandler) GetStorefront(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	storefront, err := h.storefrontService.GetStorefront(c.Request.Context(), storefrontID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront retrieved successfully", storefront)
}

// GetStorefrontBySlug handles getting a storefront by slug
// @Summary Get storefront by slug
// @Description Retrieve storefront details by slug
// @Tags storefronts
// @Accept json
// @Produce json
// @Param slug query string true "Slug"
// @Success 200 {object} dto.StorefrontResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/by-slug [get]
func (h *StorefrontHandler) GetStorefrontBySlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Slug parameter is required", "")
		return
	}

	storefront, err := h.storefrontService.GetStorefrontBySlug(c.Request.Context(), slug)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront retrieved successfully", storefront)
}

// UpdateStorefront handles storefront updates
// @Summary Update storefront
// @Description Update storefront configuration
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Param request body dto.StorefrontUpdateRequest true "Storefront update data"
// @Success 200 {object} dto.StorefrontResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id} [put]
func (h *StorefrontHandler) UpdateStorefront(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	var req dto.StorefrontUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Update the storefront
	storefront, err := h.storefrontService.UpdateStorefront(c.Request.Context(), storefrontID, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront updated successfully", storefront)
}

// DeleteStorefront handles storefront deletion
// @Summary Delete storefront
// @Description Delete a storefront
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id} [delete]
func (h *StorefrontHandler) DeleteStorefront(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	err = h.storefrontService.DeleteStorefront(c.Request.Context(), storefrontID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to delete storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront deleted successfully", nil)
}

// ActivateStorefront handles storefront activation
// @Summary Activate storefront
// @Description Activate a storefront to make it publicly accessible
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id}/activate [post]
func (h *StorefrontHandler) ActivateStorefront(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	err = h.storefrontService.ActivateStorefront(c.Request.Context(), storefrontID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to activate storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront activated successfully", nil)
}

// DeactivateStorefront handles storefront deactivation
// @Summary Deactivate storefront
// @Description Deactivate a storefront to make it inaccessible
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Param request body dto.DeactivationRequest true "Deactivation reason"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id}/deactivate [post]
func (h *StorefrontHandler) DeactivateStorefront(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	err = h.storefrontService.DeactivateStorefront(c.Request.Context(), storefrontID, req.Reason)
	if err != nil {
		h.handleServiceError(c, err, "Failed to deactivate storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront deactivated successfully", nil)
}

// SuspendStorefront handles storefront suspension
// @Summary Suspend storefront
// @Description Suspend a storefront due to policy violations or other issues
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Param request body dto.SuspensionRequest true "Suspension reason"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id}/suspend [post]
func (h *StorefrontHandler) SuspendStorefront(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	err = h.storefrontService.SuspendStorefront(c.Request.Context(), storefrontID, req.Reason)
	if err != nil {
		h.handleServiceError(c, err, "Failed to suspend storefront")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront suspended successfully", nil)
}

// SearchStorefronts handles storefront search with pagination
// @Summary Search storefronts
// @Description Search storefronts with filters and pagination
// @Tags storefronts
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param status query string false "Storefront status filter"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PaginatedStorefrontResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/search [get]
func (h *StorefrontHandler) SearchStorefronts(c *gin.Context) {
	// Parse query parameters
	req := dto.StorefrontSearchRequest{
		Query: c.Query("query"),
	}

	// Parse pagination parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		} else {
			req.Page = 1
		}
	} else {
		req.Page = 1
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			req.PageSize = pageSize
		} else {
			req.PageSize = 20
		}
	} else {
		req.PageSize = 20
	}

	// Search storefronts
	result, err := h.storefrontService.SearchStorefronts(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to search storefronts")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefronts retrieved successfully", result)
}

// GetStorefrontStats handles storefront statistics
// @Summary Get storefront statistics
// @Description Get storefront statistics and analytics
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Param period query string false "Time period" default("30d")
// @Success 200 {object} dto.StorefrontStatsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id}/stats [get]
func (h *StorefrontHandler) GetStorefrontStats(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	req := dto.StorefrontStatsRequest{
		Period: c.Query("period"),
	}

	// Get storefront statistics
	stats, err := h.storefrontService.GetStorefrontStats(c.Request.Context(), storefrontID, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get storefront statistics")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront statistics retrieved successfully", stats)
}

// Domain Management Endpoints

// ValidateCustomDomain handles custom domain validation
// @Summary Validate custom domain
// @Description Validate a custom domain for storefront usage
// @Tags storefronts,domains
// @Accept json
// @Produce json
// @Param request body dto.DomainValidationRequest true "Domain validation data"
// @Success 200 {object} dto.DomainValidationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/validate-domain [post]
func (h *StorefrontHandler) ValidateCustomDomain(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate the domain
	result, err := h.storefrontService.ValidateDomain(c.Request.Context(), req.Domain)
	if err != nil {
		h.handleServiceError(c, err, "Failed to validate domain")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Domain validation completed", result)
}

// UpdateStorefrontSettings handles storefront settings updates
// @Summary Update storefront settings
// @Description Update storefront configuration and settings
// @Tags storefronts
// @Accept json
// @Produce json
// @Param id path string true "Storefront ID"
// @Param request body dto.StorefrontSettingsRequest true "Storefront settings data"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /storefronts/{id}/settings [put]
func (h *StorefrontHandler) UpdateStorefrontSettings(c *gin.Context) {
	storefrontID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid storefront ID format", err)
		return
	}

	var req dto.StorefrontSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Update the storefront settings
	err = h.storefrontService.UpdateStorefrontSettings(c.Request.Context(), storefrontID, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update storefront settings")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Storefront settings updated successfully", nil)
}

// Helper method to handle service errors consistently
func (h *StorefrontHandler) handleServiceError(c *gin.Context, err error, message string) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Type {
		case errors.ErrorTypeValidation:
			utils.ErrorResponse(c, http.StatusBadRequest, message, appErr.Error())
		case errors.ErrorTypeNotFound:
			utils.ErrorResponse(c, http.StatusNotFound, message, appErr.Error())
		case errors.ErrorTypeAuthorization:
			utils.ErrorResponse(c, http.StatusUnauthorized, message, appErr.Error())
		default:
			utils.ErrorResponse(c, http.StatusInternalServerError, message, appErr.Error())
		}
	} else {
		utils.ErrorResponse(c, http.StatusInternalServerError, message, err.Error())
	}
}
