package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
)

// CustomerSocialAuthHandler handles customer social authentication endpoints
type CustomerSocialAuthHandler struct {
	socialAuthService *service.CustomerSocialAuthService
}

// NewCustomerSocialAuthHandler creates a new customer social auth handler
func NewCustomerSocialAuthHandler(socialAuthService *service.CustomerSocialAuthService) *CustomerSocialAuthHandler {
	return &CustomerSocialAuthHandler{
		socialAuthService: socialAuthService,
	}
}

// GetGoogleAuthURL generates Google OAuth URL for customer authentication
// @Summary Get Google OAuth URL for customer authentication
// @Description Generates a Google OAuth URL that customers can use to authenticate
// @Tags Customer Social Auth
// @Accept json
// @Produce json
// @Param storefront_slug path string true "Storefront slug"
// @Success 200 {object} map[string]string "auth_url"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /api/v1/customer/{storefront_slug}/auth/google [get]
func (h *CustomerSocialAuthHandler) GetGoogleAuthURL(c *gin.Context) {
	storefrontSlug := c.Param("storefront_slug")
	if storefrontSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storefront slug is required"})
		return
	}

	// Generate state parameter for CSRF protection
	state := uuid.New().String()
	
	// Store state in session or cache for validation
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes

	authURL, err := h.socialAuthService.GetAuthURL(service.ProviderGoogle, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GetFacebookAuthURL generates Facebook OAuth URL for customer authentication
// @Summary Get Facebook OAuth URL for customer authentication
// @Description Generates a Facebook OAuth URL that customers can use to authenticate
// @Tags Customer Social Auth
// @Accept json
// @Produce json
// @Param storefront_slug path string true "Storefront slug"
// @Success 200 {object} map[string]string "auth_url"
// @Failure 400 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /api/v1/customer/{storefront_slug}/auth/facebook [get]
func (h *CustomerSocialAuthHandler) GetFacebookAuthURL(c *gin.Context) {
	storefrontSlug := c.Param("storefront_slug")
	if storefrontSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storefront slug is required"})
		return
	}

	// Generate state parameter for CSRF protection
	state := uuid.New().String()
	
	// Store state in session or cache for validation
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10 minutes

	authURL, err := h.socialAuthService.GetAuthURL(service.ProviderFacebook, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate auth URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GoogleCallback handles Google OAuth callback for customer authentication
// @Summary Handle Google OAuth callback for customer authentication
// @Description Processes the Google OAuth callback and authenticates the customer
// @Tags Customer Social Auth
// @Accept json
// @Produce json
// @Param storefront_slug path string true "Storefront slug"
// @Param code query string true "Authorization code from Google"
// @Param state query string true "State parameter for CSRF protection"
// @Success 200 {object} dto.CustomerAuthResponse
// @Failure 400 {object} map[string]string "error"
// @Failure 401 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /api/v1/customer/{storefront_slug}/auth/google/callback [get]
func (h *CustomerSocialAuthHandler) GoogleCallback(c *gin.Context) {
	storefrontSlug := c.Param("storefront_slug")
	code := c.Query("code")
	state := c.Query("state")

	if storefrontSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storefront slug is required"})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter is required"})
		return
	}

	// Validate state parameter
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Authenticate with Google
	authResponse, err := h.socialAuthService.AuthenticateWithSocial(
		c.Request.Context(),
		storefrontSlug,
		service.ProviderGoogle,
		code,
		state,
	)

	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// FacebookCallback handles Facebook OAuth callback for customer authentication
// @Summary Handle Facebook OAuth callback for customer authentication
// @Description Processes the Facebook OAuth callback and authenticates the customer
// @Tags Customer Social Auth
// @Accept json
// @Produce json
// @Param storefront_slug path string true "Storefront slug"
// @Param code query string true "Authorization code from Facebook"
// @Param state query string true "State parameter for CSRF protection"
// @Success 200 {object} dto.CustomerAuthResponse
// @Failure 400 {object} map[string]string "error"
// @Failure 401 {object} map[string]string "error"
// @Failure 500 {object} map[string]string "error"
// @Router /api/v1/customer/{storefront_slug}/auth/facebook/callback [get]
func (h *CustomerSocialAuthHandler) FacebookCallback(c *gin.Context) {
	storefrontSlug := c.Param("storefront_slug")
	code := c.Query("code")
	state := c.Query("state")

	if storefrontSlug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Storefront slug is required"})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code is required"})
		return
	}

	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "State parameter is required"})
		return
	}

	// Validate state parameter
	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	// Authenticate with Facebook
	authResponse, err := h.socialAuthService.AuthenticateWithSocial(
		c.Request.Context(),
		storefrontSlug,
		service.ProviderFacebook,
		code,
		state,
	)

	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, authResponse)
}

// handleAuthError handles authentication errors and returns appropriate HTTP responses
func (h *CustomerSocialAuthHandler) handleAuthError(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Type {
		case errors.ErrorTypeValidation:
			c.JSON(http.StatusBadRequest, gin.H{"error": appErr.Message})
		case errors.ErrorTypeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": appErr.Message})
		case errors.ErrorTypeAuthorization:
			c.JSON(http.StatusUnauthorized, gin.H{"error": appErr.Message})
		case errors.ErrorTypeRateLimit:
			c.JSON(http.StatusTooManyRequests, gin.H{"error": appErr.Message})
		case errors.ErrorTypeInternal:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}