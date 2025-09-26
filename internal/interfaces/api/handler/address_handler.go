package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// AddressHandler handles address-related HTTP requests
type AddressHandler struct {
	addressService    service.CustomerAddressService
	validationService service.ValidationService
}

// NewAddressHandler creates a new instance of AddressHandler
func NewAddressHandler(
	addressService service.CustomerAddressService,
	validationService service.ValidationService,
) *AddressHandler {
	return &AddressHandler{
		addressService:    addressService,
		validationService: validationService,
	}
}

// GetAddress handles getting an address by ID
// @Summary Get address by ID
// @Description Retrieve address details by address ID
// @Tags addresses
// @Accept json
// @Produce json
// @Param id path string true "Address ID"
// @Success 200 {object} dto.CustomerAddressResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/{id} [get]
func (h *AddressHandler) GetAddress(c *gin.Context) {
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID format", err)
		return
	}

	address, err := h.addressService.GetAddress(c.Request.Context(), addressID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get address")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address retrieved successfully", address)
}

// UpdateAddress handles address updates
// @Summary Update address
// @Description Update address information
// @Tags addresses
// @Accept json
// @Produce json
// @Param id path string true "Address ID"
// @Param request body dto.UpdateAddressRequest true "Address update data"
// @Success 200 {object} dto.CustomerAddressResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/{id} [put]
func (h *AddressHandler) UpdateAddress(c *gin.Context) {
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID format", err)
		return
	}

	var req dto.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate the address update request
	validationResult := h.validationService.ValidateAddressUpdate(c.Request.Context(), &req)
	if validationResult.HasErrors() {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed", validationResult.Errors)
		return
	}

	// Update the address
	address, err := h.addressService.UpdateAddress(c.Request.Context(), addressID, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update address")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address updated successfully", address)
}

// DeleteAddress handles address deletion
// @Summary Delete address
// @Description Delete an address
// @Tags addresses
// @Accept json
// @Produce json
// @Param id path string true "Address ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/{id} [delete]
func (h *AddressHandler) DeleteAddress(c *gin.Context) {
	addressID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID format", err)
		return
	}

	err = h.addressService.DeleteAddress(c.Request.Context(), addressID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to delete address")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address deleted successfully", nil)
}

// ValidateAddress handles address validation
// @Summary Validate address
// @Description Validate an address and return formatted information
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.AddressValidationRequest true "Address validation data"
// @Success 200 {object} dto.AddressValidationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/validate [post]
func (h *AddressHandler) ValidateAddress(c *gin.Context) {
	var req dto.AddressValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate the address
	result, err := h.addressService.ValidateAddress(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to validate address")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address validation completed", result)
}

// GeocodeAddress handles address geocoding
// @Summary Geocode address
// @Description Get latitude and longitude coordinates for an address
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.GeocodeRequest true "Address geocoding data"
// @Success 200 {object} dto.GeocodeResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/geocode [post]
func (h *AddressHandler) GeocodeAddress(c *gin.Context) {
	var req dto.GeocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Geocode the address
	result, err := h.addressService.GeocodeAddress(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to geocode address")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address geocoded successfully", result)
}

// GetNearbyAddresses handles finding nearby addresses
// @Summary Get nearby addresses
// @Description Find addresses near a specific location
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.NearbyAddressRequest true "Nearby address search data"
// @Success 200 {array} dto.CustomerAddressResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/nearby [post]
func (h *AddressHandler) GetNearbyAddresses(c *gin.Context) {
	var req dto.NearbyAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get nearby addresses
	addresses, err := h.addressService.GetNearbyAddresses(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get nearby addresses")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Nearby addresses retrieved successfully", addresses)
}

// BulkCreateAddresses handles bulk address creation
// @Summary Bulk create addresses
// @Description Create multiple addresses in a single request
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.BulkAddressCreateRequest true "Bulk address creation data"
// @Success 200 {object} dto.BulkOperationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/bulk [post]
func (h *AddressHandler) BulkCreateAddresses(c *gin.Context) {
	var req dto.BulkAddressCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate the bulk request
	if len(req.Addresses) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No addresses provided", "")
		return
	}

	if len(req.Addresses) > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Too many addresses", "Maximum 100 addresses allowed per request")
		return
	}

	// Create addresses in bulk
	result, err := h.addressService.BulkCreateAddresses(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create addresses")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Addresses created successfully", result)
}

// BulkUpdateAddresses handles bulk address updates
// @Summary Bulk update addresses
// @Description Update multiple addresses in a single request
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.BulkAddressUpdateRequest true "Bulk address update data"
// @Success 200 {object} dto.BulkOperationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/bulk [put]
func (h *AddressHandler) BulkUpdateAddresses(c *gin.Context) {
	var req dto.BulkAddressUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate the bulk request
	if len(req.Updates) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No updates provided", "")
		return
	}

	if len(req.Updates) > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Too many updates", "Maximum 100 updates allowed per request")
		return
	}

	// Update addresses in bulk
	result, err := h.addressService.BulkUpdateAddresses(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update addresses")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Addresses updated successfully", result)
}

// BulkDeleteAddresses handles bulk address deletion
// @Summary Bulk delete addresses
// @Description Delete multiple addresses in a single request
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.BulkAddressDeleteRequest true "Bulk address deletion data"
// @Success 200 {object} dto.BulkOperationResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/bulk [delete]
func (h *AddressHandler) BulkDeleteAddresses(c *gin.Context) {
	var req dto.BulkAddressDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate the bulk request
	if len(req.AddressIDs) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "No address IDs provided", "")
		return
	}

	if len(req.AddressIDs) > 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Too many addresses", "Maximum 100 addresses allowed per request")
		return
	}

	// Delete addresses in bulk
	result, err := h.addressService.BulkDeleteAddresses(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to delete addresses")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Addresses deleted successfully", result)
}

// GetAddressStats handles address statistics
// @Summary Get address statistics
// @Description Get address statistics and analytics
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.AddressStatsRequest true "Address statistics request"
// @Success 200 {object} dto.AddressStatsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/stats [post]
func (h *AddressHandler) GetAddressStats(c *gin.Context) {
	var req dto.AddressStatsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get address statistics
	stats, err := h.addressService.GetAddressStats(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get address statistics")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address statistics retrieved successfully", stats)
}

// GetAddressDistribution handles address distribution analytics
// @Summary Get address distribution
// @Description Get address distribution analytics by region
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body dto.AddressDistributionRequest true "Address distribution request"
// @Success 200 {object} dto.AddressDistributionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /addresses/distribution [post]
func (h *AddressHandler) GetAddressDistribution(c *gin.Context) {
	var req dto.AddressDistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Get address distribution
	distribution, err := h.addressService.GetAddressDistribution(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get address distribution")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Address distribution retrieved successfully", distribution)
}

// Helper method to handle service errors consistently
func (h *AddressHandler) handleServiceError(c *gin.Context, err error, message string) {
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
