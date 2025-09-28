package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// SocialProvider represents supported social login providers
type SocialProvider string

const (
	ProviderGoogle   SocialProvider = "google"
	ProviderFacebook SocialProvider = "facebook"
)

// SocialUserInfo represents user information from social providers
type SocialUserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Provider SocialProvider
}

// CustomerSocialAuthService handles customer social authentication
type CustomerSocialAuthService struct {
	customerRepo     repository.CustomerRepository
	storefrontRepo   repository.StorefrontRepository
	googleConfig     *oauth2.Config
	facebookConfig   *oauth2.Config
	jwtSecretKey     string
	jwtRefreshKey    string
}

// NewCustomerSocialAuthService creates a new customer social auth service
func NewCustomerSocialAuthService(
	customerRepo repository.CustomerRepository,
	storefrontRepo repository.StorefrontRepository,
) *CustomerSocialAuthService {
	service := &CustomerSocialAuthService{
		customerRepo:   customerRepo,
		storefrontRepo: storefrontRepo,
		jwtSecretKey:   os.Getenv("JWT_SECRET_KEY"),
		jwtRefreshKey:  os.Getenv("JWT_REFRESH_KEY"),
	}

	// Initialize OAuth configs
	service.initializeOAuthConfigs()
	return service
}

// initializeOAuthConfigs sets up OAuth configurations for social providers
func (s *CustomerSocialAuthService) initializeOAuthConfigs() {
	// Google OAuth configuration
	s.googleConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_CUSTOMER_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Facebook OAuth configuration
	s.facebookConfig = &oauth2.Config{
		ClientID:     os.Getenv("FACEBOOK_CLIENT_ID"),
		ClientSecret: os.Getenv("FACEBOOK_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("FACEBOOK_CUSTOMER_REDIRECT_URL"),
		Scopes: []string{
			"email",
			"public_profile",
		},
		Endpoint: facebook.Endpoint,
	}
}

// GetAuthURL generates OAuth URL for the specified provider
func (s *CustomerSocialAuthService) GetAuthURL(provider SocialProvider, state string) (string, error) {
	switch provider {
	case ProviderGoogle:
		if s.googleConfig == nil {
			return "", errors.NewInternalError("Google OAuth not configured", nil)
		}
		return s.googleConfig.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
	case ProviderFacebook:
		if s.facebookConfig == nil {
			return "", errors.NewInternalError("Facebook OAuth not configured", nil)
		}
		return s.facebookConfig.AuthCodeURL(state), nil
	default:
		return "", errors.NewValidationError("Unsupported social provider", nil)
	}
}

// AuthenticateWithSocial handles social login authentication
func (s *CustomerSocialAuthService) AuthenticateWithSocial(
	ctx context.Context,
	storefrontSlug string,
	provider SocialProvider,
	code string,
	state string,
) (*dto.CustomerAuthResponse, error) {
	// Get storefront context
	storefront, err := s.storefrontRepo.GetBySlug(ctx, storefrontSlug)
	if err != nil {
		return nil, errors.NewNotFoundError("Storefront not found")
	}

	// Exchange code for token
	token, err := s.exchangeCodeForToken(ctx, provider, code)
	if err != nil {
		return nil, err
	}

	// Get user info from social provider
	userInfo, err := s.getUserInfoFromProvider(ctx, provider, token)
	if err != nil {
		return nil, err
	}

	// Find or create customer
	customer, isNewCustomer, err := s.findOrCreateCustomer(ctx, storefront.ID, userInfo)
	if err != nil {
		return nil, err
	}

	// Generate JWT tokens
	accessToken, refreshToken, expiresIn, err := s.generateCustomerTokens(customer, storefront)
	if err != nil {
		return nil, err
	}

	// Update customer's social login info
	err = s.updateCustomerSocialInfo(ctx, customer, userInfo, token)
	if err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update customer social info: %v\n", err)
	}

	// Log if this is a new customer registration
	if isNewCustomer {
		fmt.Printf("New customer registered via %s: %s\n", provider, userInfo.Email)
	}

	return &dto.CustomerAuthResponse{
		Customer: &dto.CustomerResponse{
			ID:           customer.ID,
			StorefrontID: customer.StorefrontID,
			Email:        customer.Email,
			Phone:        customer.Phone,
			FirstName:    customer.FirstName,
			LastName:     customer.LastName,
			FullName:     customer.FullName,
			DateOfBirth:  customer.DateOfBirth,
			Gender:       customer.Gender,
			Status:       string(customer.Status),
			CustomerType: customer.CustomerType,
			Preferences:  customer.Preferences,
			LastLoginAt:  customer.LastLoginAt,
			CreatedAt:    customer.CreatedAt,
			UpdatedAt:    customer.UpdatedAt,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

// exchangeCodeForToken exchanges authorization code for access token
func (s *CustomerSocialAuthService) exchangeCodeForToken(ctx context.Context, provider SocialProvider, code string) (*oauth2.Token, error) {
	var config *oauth2.Config

	switch provider {
	case ProviderGoogle:
		config = s.googleConfig
	case ProviderFacebook:
		config = s.facebookConfig
	default:
		return nil, errors.NewValidationError("Unsupported provider", nil)
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, errors.NewInternalError("Failed to exchange code for token", err)
	}

	return token, nil
}

// getUserInfoFromProvider retrieves user information from social provider
func (s *CustomerSocialAuthService) getUserInfoFromProvider(ctx context.Context, provider SocialProvider, token *oauth2.Token) (*SocialUserInfo, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var userInfo *SocialUserInfo
	var err error

	switch provider {
	case ProviderGoogle:
		userInfo, err = s.getGoogleUserInfo(ctx, client, token)
	case ProviderFacebook:
		userInfo, err = s.getFacebookUserInfo(ctx, client, token)
	default:
		return nil, errors.NewValidationError("Unsupported provider", nil)
	}

	if err != nil {
		return nil, err
	}

	userInfo.Provider = provider
	return userInfo, nil
}

// getGoogleUserInfo retrieves user info from Google
func (s *CustomerSocialAuthService) getGoogleUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (*SocialUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.NewInternalError("Failed to get Google user info", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewInternalError("Google API returned error", nil)
	}

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, errors.NewInternalError("Failed to decode Google user info", err)
	}

	return &SocialUserInfo{
		ID:      googleUser.ID,
		Email:   googleUser.Email,
		Name:    googleUser.Name,
		Picture: googleUser.Picture,
	}, nil
}

// getFacebookUserInfo retrieves user info from Facebook
func (s *CustomerSocialAuthService) getFacebookUserInfo(ctx context.Context, client *http.Client, token *oauth2.Token) (*SocialUserInfo, error) {
	url := fmt.Sprintf("https://graph.facebook.com/me?fields=id,name,email,picture&access_token=%s", token.AccessToken)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.NewInternalError("Failed to get Facebook user info", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewInternalError("Facebook API returned error", nil)
	}

	var facebookUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&facebookUser); err != nil {
		return nil, errors.NewInternalError("Failed to decode Facebook user info", err)
	}

	return &SocialUserInfo{
		ID:      facebookUser.ID,
		Email:   facebookUser.Email,
		Name:    facebookUser.Name,
		Picture: facebookUser.Picture.Data.URL,
	}, nil
}

// findOrCreateCustomer finds existing customer or creates new one
func (s *CustomerSocialAuthService) findOrCreateCustomer(ctx context.Context, storefrontID uuid.UUID, userInfo *SocialUserInfo) (*entity.Customer, bool, error) {
	// Try to find existing customer by email
	if userInfo.Email != "" {
		customer, err := s.customerRepo.GetByEmail(ctx, storefrontID, strings.ToLower(userInfo.Email))
		if err == nil && customer != nil {
			return customer, false, nil
		}
	}

	// Create new customer
	newCustomer := &entity.Customer{
		ID:           uuid.New(),
		StorefrontID: storefrontID,
		Email:        &userInfo.Email,
		FullName:     &userInfo.Name,
		Status:       entity.CustomerStatusActive,
		CustomerType: entity.CustomerTypeRegular,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Parse name into first and last name
	if userInfo.Name != "" {
		nameParts := strings.Fields(userInfo.Name)
		if len(nameParts) > 0 {
			newCustomer.FirstName = &nameParts[0]
			if len(nameParts) > 1 {
				lastName := strings.Join(nameParts[1:], " ")
				newCustomer.LastName = &lastName
			}
		}
	}

	// Set email as verified since it comes from social provider
	if userInfo.Email != "" {
		now := time.Now()
		newCustomer.EmailVerifiedAt = &now
	}

	// Set default preferences
	newCustomer.SetDefaultPreferences()

	err := s.customerRepo.Create(ctx, newCustomer)
	if err != nil {
		return nil, false, errors.NewInternalError("Failed to create customer", err)
	}

	return newCustomer, true, nil
}

// updateCustomerSocialInfo updates customer's social login information
func (s *CustomerSocialAuthService) updateCustomerSocialInfo(ctx context.Context, customer *entity.Customer, userInfo *SocialUserInfo, token *oauth2.Token) error {
	// Update last login
	now := time.Now()
	customer.LastLoginAt = &now

	return s.customerRepo.Update(ctx, customer)
}

// generateCustomerTokens generates JWT access and refresh tokens for customer
func (s *CustomerSocialAuthService) generateCustomerTokens(customer *entity.Customer, storefront *entity.Storefront) (string, string, int64, error) {
	if s.jwtSecretKey == "" {
		return "", "", 0, errors.NewInternalError("JWT secret key not configured", nil)
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Access token (1 hour)
	accessExpiry := time.Now().Add(1 * time.Hour)
	accessClaims := jwt.MapClaims{
		"customer_id":   customer.ID.String(),
		"storefront_id": customer.StorefrontID.String(),
		"email":         *customer.Email,
		"session_id":    sessionID,
		"token_type":    "access",
		"exp":           accessExpiry.Unix(),
		"iat":           time.Now().Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecretKey))
	if err != nil {
		return "", "", 0, errors.NewInternalError("Failed to generate access token", err)
	}

	// Refresh token (30 days)
	refreshExpiry := time.Now().Add(30 * 24 * time.Hour)
	refreshClaims := jwt.MapClaims{
		"customer_id":   customer.ID.String(),
		"storefront_id": customer.StorefrontID.String(),
		"session_id":    sessionID,
		"token_type":    "refresh",
		"exp":           refreshExpiry.Unix(),
		"iat":           time.Now().Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtRefreshKey))
	if err != nil {
		return "", "", 0, errors.NewInternalError("Failed to generate refresh token", err)
	}

	// Store refresh token in database (if method exists)
	// Note: This would require implementing UpdateRefreshToken method in customer repository
	// For now, we'll skip storing refresh tokens in database
	_ = refreshTokenString // Prevent unused variable warning

	return accessTokenString, refreshTokenString, 3600, nil // 1 hour in seconds
}