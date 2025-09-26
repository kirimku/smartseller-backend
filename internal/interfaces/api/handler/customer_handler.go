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

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	customerService   service.CustomerService
	addressService    service.CustomerAddressService
	validationService service.ValidationService
}

// NewCustomerHandler creates a new instance of CustomerHandler
func NewCustomerHandler(
	customerService service.CustomerService,
	addressService service.CustomerAddressService,
	validationService service.ValidationService,
) *CustomerHandler {
	return &CustomerHandler{
		customerService:   customerService,
		addressService:    addressService,
		validationService: validationService,
	}
}

// RegisterCustomer handles customer registration
// @Summary Register a new customer
// @Description Register a new customer with email, name, and optional phone
// @Tags customers
// @Accept json
// @Produce json
// @Param request body dto.CustomerRegistrationRequest true "Customer registration data"
// @Success 201 {object} dto.CustomerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/register [post]
func (h *CustomerHandler) RegisterCustomer(c *gin.Context) {
	var req dto.CustomerRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Validate the registration request
	validationResult := h.validationService.ValidateCustomerRegistration(c.Request.Context(), &req)
	if validationResult.HasErrors() {
		utils.ValidationErrorResponse(c, http.StatusBadRequest, "Validation failed", validationResult.Errors)
		return
	}

	// Check email uniqueness
	if err := h.validationService.ValidateEmailUniqueness(c.Request.Context(), req.Email, nil); err != nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email already exists", err.Error())
		return
	}

	// Register the customer
	customer, err := h.customerService.RegisterCustomer(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to register customer")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Customer registered successfully", customer)
}

// GetCustomer handles getting a customer by ID
// @Summary Get customer by ID
// @Description Retrieve customer details by customer ID
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	customer, err := h.customerService.GetCustomerByID(c.Request.Context(), customerID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get customer")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Customer retrieved successfully", customer)
}

// GetCustomerByEmail handles getting a customer by email
// @Summary Get customer by email
// @Description Retrieve customer details by email address
// @Tags customers
// @Accept json
// @Produce json
// @Param email query string true "Customer email"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/by-email [get]
func (h *CustomerHandler) GetCustomerByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Email parameter is required", "")
		return
	}

	customer, err := h.customerService.GetCustomerByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get customer")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Customer retrieved successfully", customer)
}

// UpdateCustomer handles customer profile updates
// @Summary Update customer profile
// @Description Update customer profile information
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body dto.CustomerUpdateRequest true "Customer update data"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	var req dto.CustomerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Validate the update request
	validationResult := h.validationService.ValidateCustomerUpdate(c.Request.Context(), &req)
	if validationResult.HasErrors() {
		utils.ValidationErrorResponse(c, http.StatusBadRequest, "Validation failed", validationResult.Errors)
		return
	}

	// Check email uniqueness if email is being updated
	if req.Email != nil {
		if err := h.validationService.ValidateEmailUniqueness(c.Request.Context(), *req.Email, &customerID); err != nil {
			utils.ErrorResponse(c, http.StatusConflict, "Email already exists", err.Error())
			return
		}
	}

	// Check phone uniqueness if phone is being updated
	if req.Phone != nil {
		if err := h.validationService.ValidatePhoneUniqueness(c.Request.Context(), *req.Phone, &customerID); err != nil {
			utils.ErrorResponse(c, http.StatusConflict, "Phone already exists", err.Error())
			return
		}
	}

	// Update the customer
	customer, err := h.customerService.UpdateCustomer(c.Request.Context(), customerID, &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to update customer")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Customer updated successfully", customer)
}

// DeactivateCustomer handles customer deactivation
// @Summary Deactivate customer
// @Description Deactivate a customer account
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body dto.CustomerDeactivationRequest true "Deactivation reason"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{id}/deactivate [post]
func (h *CustomerHandler) DeactivateCustomer(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	var req dto.CustomerDeactivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	customer, err := h.customerService.DeactivateCustomer(c.Request.Context(), customerID, req.Reason)
	if err != nil {
		h.handleServiceError(c, err, "Failed to deactivate customer")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Customer deactivated successfully", customer)
}

// ReactivateCustomer handles customer reactivation
// @Summary Reactivate customer
// @Description Reactivate a deactivated customer account
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} dto.CustomerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{id}/reactivate [post]
func (h *CustomerHandler) ReactivateCustomer(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	customer, err := h.customerService.ReactivateCustomer(c.Request.Context(), customerID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to reactivate customer")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Customer reactivated successfully", customer)
}

// SearchCustomers handles customer search with pagination
// @Summary Search customers
// @Description Search customers with filters and pagination
// @Tags customers
// @Accept json
// @Produce json
// @Param query query string false "Search query"
// @Param status query string false "Customer status filter"
// @Param customer_type query string false "Customer type filter"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort field" default("created_at")
// @Param sort_dir query string false "Sort direction" default("desc")
// @Success 200 {object} dto.PaginatedCustomerResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/search [get]
func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	// Parse query parameters
	req := dto.CustomerSearchRequest{
		Query:        c.Query("query"),
		Status:       c.Query("status"),
		CustomerType: c.Query("customer_type"),
		SortBy:       c.Query("sort_by"),
		SortDir:      c.Query("sort_dir"),
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

	// Set defaults
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	// Search customers
	result, err := h.customerService.SearchCustomers(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to search customers")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Customers retrieved successfully", result)
}

// GetCustomerStats handles customer statistics
// @Summary Get customer statistics
// @Description Get customer statistics and analytics
// @Tags customers
// @Accept json
// @Produce json
// @Param period query string false "Time period" default("30d")
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.CustomerStatsResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/stats [get]
func (h *CustomerHandler) GetCustomerStats(c *gin.Context) {
	req := dto.CustomerStatsRequest{
		Period: c.Query("period"),
	}

	// Parse date parameters if provided
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := utils.ParseDate(startDateStr); err == nil {
			req.StartDate = &startDate
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid start_date format", "Use YYYY-MM-DD format")
			return
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := utils.ParseDate(endDateStr); err == nil {
			req.EndDate = &endDate
		} else {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid end_date format", "Use YYYY-MM-DD format")
			return
		}
	}

	// Get customer statistics
	stats, err := h.customerService.GetCustomerStats(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get customer statistics")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Customer statistics retrieved successfully", stats)
}

// Customer Address Endpoints

// CreateCustomerAddress handles creating a new address for a customer
// @Summary Create customer address
// @Description Create a new address for a customer
// @Tags customers,addresses
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param request body dto.CreateAddressRequest true "Address data"
// @Success 201 {object} dto.CustomerAddressResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{id}/addresses [post]
func (h *CustomerHandler) CreateCustomerAddress(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	var req dto.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Set the customer ID from the path
	req.CustomerID = customerID

	// Validate the address creation request
	validationResult := h.validationService.ValidateAddressCreation(c.Request.Context(), &req)
	if validationResult.HasErrors() {
		utils.ValidationErrorResponse(c, http.StatusBadRequest, "Validation failed", validationResult.Errors)
		return
	}

	// Create the address
	address, err := h.addressService.CreateAddress(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err, "Failed to create address")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Address created successfully", address)
}

// GetCustomerAddresses handles getting all addresses for a customer
// @Summary Get customer addresses
// @Description Retrieve all addresses for a customer
// @Tags customers,addresses
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {array} dto.CustomerAddressResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{id}/addresses [get]
func (h *CustomerHandler) GetCustomerAddresses(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	addresses, err := h.addressService.GetCustomerAddresses(c.Request.Context(), customerID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get customer addresses")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Addresses retrieved successfully", addresses)
}

// SetDefaultAddress handles setting a customer's default address
// @Summary Set default address
// @Description Set an address as default for a customer
// @Tags customers,addresses
// @Accept json
// @Produce json
// @Param customer_id path string true "Customer ID"
// @Param address_id path string true "Address ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{customer_id}/addresses/{address_id}/default [post]
func (h *CustomerHandler) SetDefaultAddress(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("customer_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	addressID, err := uuid.Parse(c.Param("address_id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid address ID format", err.Error())
		return
	}

	err = h.addressService.SetDefaultAddress(c.Request.Context(), customerID, addressID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to set default address")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Default address set successfully", nil)
}

// GetDefaultAddress handles getting a customer's default address
// @Summary Get default address
// @Description Get the default address for a customer
// @Tags customers,addresses
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} dto.CustomerAddressResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /customers/{id}/addresses/default [get]
func (h *CustomerHandler) GetDefaultAddress(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid customer ID format", err.Error())
		return
	}

	address, err := h.addressService.GetDefaultAddress(c.Request.Context(), customerID)
	if err != nil {
		h.handleServiceError(c, err, "Failed to get default address")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Default address retrieved successfully", address)
}

// Helper method to handle service errors consistently
func (h *CustomerHandler) handleServiceError(c *gin.Context, err error, message string) {
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
