package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
)

// CustomerAuthHandler handles customer authentication endpoints
type CustomerAuthHandler struct {
	customerService                      service.CustomerService
	customerPasswordResetService         *service.CustomerPasswordResetService
	customerEmailVerificationService    *service.CustomerEmailVerificationService
	validationService                    service.ValidationService
}

// NewCustomerAuthHandler creates a new customer authentication handler
func NewCustomerAuthHandler(
	customerService service.CustomerService,
	passwordResetService *service.CustomerPasswordResetService,
	emailVerificationService *service.CustomerEmailVerificationService,
	validationService service.ValidationService,
) *CustomerAuthHandler {
	return &CustomerAuthHandler{
		customerService:                   customerService,
		customerPasswordResetService:      passwordResetService,
		customerEmailVerificationService: emailVerificationService,
		validationService:                 validationService,
	}
}

// RegisterCustomer handles customer registration
func (h *CustomerAuthHandler) RegisterCustomer(c *gin.Context) {
	var req dto.CustomerRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields: []dto.FieldError{
					{
						Field:   "request",
						Message: err.Error(),
						Rule:    "json",
					},
				},
			},
		})
		return
	}

	// Validate request
	if validation := h.validationService.ValidateCustomerRegistration(c.Request.Context(), &req); validation.HasErrors() {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields:  convertValidationErrors(validation.Errors),
			},
		})
		return
	}

	// Register customer
	response, err := h.customerService.RegisterCustomer(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.SuccessResponse{
		Success: true,
		Message: "Customer registered successfully",
		Data:    response,
	})
}

// LoginCustomer handles customer login
func (h *CustomerAuthHandler) LoginCustomer(c *gin.Context) {
	var req dto.CustomerAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields: []dto.FieldError{
					{
						Field:   "request",
						Message: err.Error(),
						Rule:    "json",
					},
				},
			},
		})
		return
	}

	response, err := h.customerService.AuthenticateCustomer(c.Request.Context(), &req)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Set auth cookies
	h.setAuthCookies(c, response.AccessToken, response.RefreshToken)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// LogoutCustomer handles customer logout
func (h *CustomerAuthHandler) LogoutCustomer(c *gin.Context) {
	// Clear auth cookies
	h.clearAuthCookies(c)

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Logout successful",
	})
}

// RefreshToken handles token refresh
func (h *CustomerAuthHandler) RefreshToken(c *gin.Context) {
	var req dto.TokenRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields: []dto.FieldError{
					{
						Field:   "request",
						Message: err.Error(),
						Rule:    "json",
					},
				},
			},
		})
		return
	}

	// TODO: Implement token refresh logic
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Token refresh not implemented"})
}

// RequestPasswordReset handles password reset requests
func (h *CustomerAuthHandler) RequestPasswordReset(c *gin.Context) {
	var req dto.PasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields: []dto.FieldError{
					{
						Field:   "request",
						Message: err.Error(),
						Rule:    "json",
					},
				},
			},
		})
		return
	}

	// Get storefront ID from context
	storefrontID, exists := c.Get("storefront_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storefront context required"})
		return
	}

	err := h.customerPasswordResetService.RequestPasswordReset(c.Request.Context(), storefrontID.(uuid.UUID), req.Email)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Password reset email sent",
	})
}

// ConfirmPasswordReset handles password reset confirmation
func (h *CustomerAuthHandler) ConfirmPasswordReset(c *gin.Context) {
	var req dto.PasswordResetConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields: []dto.FieldError{
					{
						Field:   "request",
						Message: err.Error(),
						Rule:    "json",
					},
				},
			},
		})
		return
	}

	err := h.customerPasswordResetService.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Password reset successful",
	})
}

// ValidateResetToken validates a password reset token
func (h *CustomerAuthHandler) ValidateResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	// TODO: Implement token validation
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Token validation not implemented"})
}

// VerifyEmail handles email verification
func (h *CustomerAuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	err := h.customerEmailVerificationService.VerifyEmail(c.Request.Context(), token)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}

// ResendVerificationEmail handles resending verification email
func (h *CustomerAuthHandler) ResendVerificationEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields: []dto.FieldError{
					{
						Field:   "request",
						Message: err.Error(),
						Rule:    "json",
					},
				},
			},
		})
		return
	}

	// Get storefront ID from context
	storefrontID, exists := c.Get("storefront_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storefront context required"})
		return
	}

	err := h.customerEmailVerificationService.ResendVerificationEmail(c.Request.Context(), storefrontID.(uuid.UUID), req.Email)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Verification email sent",
	})
}

// ChangePassword handles password change
func (h *CustomerAuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
			Error: dto.ValidationErrorDetail{
				Code:    "VALIDATION_FAILED",
				Message: "Request validation failed",
				Fields: []dto.FieldError{
					{
						Field:   "request",
						Message: err.Error(),
						Rule:    "json",
					},
				},
			},
		})
		return
	}

	// Get customer ID from context (set by auth middleware)
	customerID, exists := c.Get("customer_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	changeReq := &dto.ChangePasswordRequest{
		CustomerID:      customerID.(uuid.UUID),
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	err := h.customerService.ChangePassword(c.Request.Context(), changeReq)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// Helper functions

// setAuthCookies sets authentication cookies
func (h *CustomerAuthHandler) setAuthCookies(c *gin.Context, accessToken, refreshToken string) {
	c.SetCookie("access_token", accessToken, 3600, "/", "", false, true)
	c.SetCookie("refresh_token", refreshToken, 86400*7, "/", "", false, true)
}

// clearAuthCookies clears authentication cookies
func (h *CustomerAuthHandler) clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}

// convertValidationErrors converts service validation errors to DTO field errors
func convertValidationErrors(validationErrors []service.ValidationError) []dto.FieldError {
	fieldErrors := make([]dto.FieldError, len(validationErrors))
	for i, err := range validationErrors {
		fieldErrors[i] = dto.FieldError{
			Field:   err.Field,
			Message: err.Message,
			Rule:    err.Code,
		}
	}
	return fieldErrors
}

// handleServiceError handles service errors and returns appropriate HTTP responses
func (h *CustomerAuthHandler) handleServiceError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Type {
		case errors.ErrorTypeValidation:
			c.JSON(http.StatusBadRequest, dto.ValidationErrorResponse{
				Error: dto.ValidationErrorDetail{
					Code:    "VALIDATION_FAILED",
					Message: appErr.Message,
					Fields: []dto.FieldError{
						{
							Field:   "general",
							Message: appErr.Message,
							Rule:    "validation",
						},
					},
				},
			})
		case errors.ErrorTypeAuthorization:
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": appErr.Message,
			})
		case errors.ErrorTypeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": appErr.Message})
		case errors.ErrorTypeRateLimit:
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": appErr.Message,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"message": "An unexpected error occurred",
			})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "An unexpected error occurred",
		})
	}
}