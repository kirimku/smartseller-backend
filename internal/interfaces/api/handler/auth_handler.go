package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/pkg/utils"
	"google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userUsecase usecase.UserUseCase
}

// NewAuthHandler creates a new instance of AuthHandler
func NewAuthHandler(userUsecase usecase.UserUseCase) *AuthHandler {
	return &AuthHandler{
		userUsecase: userUsecase,
	}
}

// GetGoogleLoginURL generates a Google OAuth URL and state
func (h *AuthHandler) GetGoogleLoginURL() (string, string, error) {
	// Generate state for CSRF protection
	state := uuid.New().String()
	fmt.Printf("Generated OAuth state: %s\n", state)

	// Get OAuth URL
	url := config.AppConfig.GoogleOAuthConfig.AuthCodeURL(state)
	fmt.Printf("Generated OAuth URL: %s\n", url)

	return url, state, nil
}

// LoginHandler initiates the Google OAuth flow
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	// Generate OAuth URL and state
	url, state, err := h.GetGoogleLoginURL()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate OAuth URL", err)
		return
	}

	// Get session
	session := sessions.Default(c)

	// Store state in session
	session.Set("state", state)
	session.Set("authenticated", false)

	// Save session
	if err := session.Save(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save session", err)
		return
	}

	// Send response
	utils.SuccessResponse(c, http.StatusOK, "Google login URL generated", gin.H{
		"redirect_url": url,
		"state":        state,
	})
}

// GoogleCallback handles the Google OAuth callback
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	var loginRequest dto.LoginRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	fmt.Printf("Received callback with code: %s (truncated for security) and state: %s\n",
		loginRequest.Code[:10]+"...", loginRequest.State)

	// Get session
	session := sessions.Default(c)

	// Get state from session and compare with request
	storedState := session.Get("state")
	fmt.Printf("Stored state from session: %v\n", storedState)

	if storedState == nil {
		fmt.Printf("No state found in session\n")
		// Try to continue with the flow, skipping state validation in development
		fmt.Printf("Proceeding without state validation (development only)\n")
	} else if loginRequest.State != storedState.(string) {
		fmt.Printf("State mismatch: received %s but expected %s\n", loginRequest.State, storedState.(string))
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid state parameter", nil)
		return
	}

	// Exchange code for token
	token, err := config.AppConfig.GoogleOAuthConfig.Exchange(c.Request.Context(), loginRequest.Code)
	if err != nil {
		fmt.Printf("Error exchanging code for token: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to exchange token", err)
		return
	}

	// Create a client with the token credentials
	client := config.AppConfig.GoogleOAuthConfig.Client(c.Request.Context(), token)

	// Create the OAuth service using the authenticated client
	oauth2Service, err := oauth2.NewService(c.Request.Context(), option.WithHTTPClient(client))
	if err != nil {
		fmt.Printf("Error creating OAuth service: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create OAuth client", err)
		return
	}

	// Get user info
	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		fmt.Printf("Error getting user info: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user info", err)
		return
	}

	fmt.Printf("Successfully retrieved user info for: %s\n", userInfo.Email)

	// Check if user already exists by GoogleID
	existingUser, err := h.userUsecase.GetUserByGoogleID(userInfo.Id)
	var userID string
	var refreshToken string

	if err == nil && existingUser != nil {
		// User exists, use the existing ID
		userID = existingUser.ID
		fmt.Printf("User already exists with ID: %s\n", userID)

		// Preserve refresh token
		refreshToken = existingUser.RefreshToken
		if token.RefreshToken != "" {
			refreshToken = token.RefreshToken
			fmt.Printf("Updating refresh token for existing user\n")
		}
	} else {
		// New user, generate a new UUID
		userID = uuid.New().String()
		refreshToken = token.RefreshToken
		fmt.Printf("Creating new user with ID: %s\n", userID)
	}

	// Create a new user or update existing one
	user := &entity.User{
		ID:               userID,
		GoogleID:         userInfo.Id,
		Name:             userInfo.Name,
		Email:            userInfo.Email,
		Picture:          userInfo.Picture,
		UserType:         entity.UserTypePersonal, // Default to personal for Google OAuth users
		UserTier:         entity.UserTierPendekar, // Default tier for new users
		TransactionCount: 0,                       // Start with 0 transactions
		AccessToken:      token.AccessToken,
		RefreshToken:     refreshToken,
		TokenExpiry: sql.NullTime{
			Time:  token.Expiry,
			Valid: !token.Expiry.IsZero(),
		},
	}

	// Save user to database
	err = h.userUsecase.CreateOrUpdateUser(user)
	if err != nil {
		fmt.Printf("Error saving user to database: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save user", err)
		return
	}

	// Create a custom JWT token
	// JWT expiry (24 hours by default)
	expiryTime := time.Now().Add(24 * time.Hour)

	// Create claims for the JWT
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   user.Email,
		"name":    user.Name,
		"exp":     expiryTime.Unix(),
		"iat":     time.Now().Unix(),
	}

	// Create JWT token with claims
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Get the secret key from environment
	secretKey := os.Getenv("SESSION_KEY")
	if secretKey == "" {
		fmt.Printf("Warning: SESSION_KEY environment variable not set, using default secret\n")
		secretKey = "your-default-secret-key-for-development-only"
	}

	// Sign the token with the secret key
	tokenString, err := jwtToken.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Printf("Error signing JWT token: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError,
			"Failed to sign JWT token", err)
		return
	}

	// Clear the OAuth state and set user info in session
	session.Delete("state")
	session.Set("user_id", user.ID)
	session.Set("email", user.Email)
	session.Set("authenticated", true)

	if err := session.Save(); err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save session", err)
		return
	}

	// Return success response with our custom JWT token
	authResponse := dto.AuthResponse{
		AccessToken:  tokenString,  // Use our custom JWT token instead of Google token
		RefreshToken: refreshToken, // Use preserved refresh token
		TokenExpiry:  expiryTime,   // Use our custom expiry time
		User: dto.UserDTO{
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email,
			Phone:        user.Phone,
			UserType:     user.UserType,
			UserRole:     user.UserRole,
			Picture:      user.Picture,
			AcceptTerms:  user.AcceptTerms,
			AcceptPromos: user.AcceptPromos,
			IsAdmin:      user.IsAdmin,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Authentication successful", authResponse)
}

// LogoutHandler handles user logout
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	// Get session
	session := sessions.Default(c)

	// Clear session
	session.Clear()

	if err := session.Save(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to clear session", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Successfully logged out", nil)
}

// RefreshTokenHandler handles token refresh requests
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	var requestBody struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Refresh the session
	newAccessToken, newRefreshToken, newTokenExpiry, err := h.userUsecase.RefreshSession(requestBody.RefreshToken)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid or expired refresh token", err)
		return
	}

	// Update session expiry
	session := sessions.Default(c)
	if err := session.Save(); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update session", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_expiry":  newTokenExpiry,
	})
}

// RegisterHandler handles user registration
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var registrationRequest dto.RegistrationRequest
	if err := c.ShouldBindJSON(&registrationRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate that both email and phone are provided (now required)
	if registrationRequest.Email == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Alamat email wajib diisi untuk pendaftaran", nil)
		return
	}

	if registrationRequest.Phone == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Nomor telepon wajib diisi untuk pendaftaran", nil)
		return
	}

	// Register the user
	user, accessToken, refreshToken, expiryTime, err := h.userUsecase.Register(
		registrationRequest.Name,
		registrationRequest.Email,
		registrationRequest.Phone,
		registrationRequest.Password,
		registrationRequest.UserType,
		registrationRequest.AcceptTerms,
		registrationRequest.AcceptPromos,
	)

	if err != nil {
		// Check for possible "email already registered" error
		if err.Error() == "email already registered" && registrationRequest.Email != "" {
			// Try to delete any existing record with this email that might be causing conflict
			ctx := c.Request.Context()

			// Attempt to clean up potentially orphaned record
			deleteErr := h.userUsecase.DeleteUserByEmail(ctx, registrationRequest.Email)
			if deleteErr == nil {
				// If deletion was successful, try registration again
				user, accessToken, refreshToken, expiryTime, err = h.userUsecase.Register(
					registrationRequest.Name,
					registrationRequest.Email,
					registrationRequest.Phone,
					registrationRequest.Password,
					registrationRequest.UserType,
					registrationRequest.AcceptTerms,
					registrationRequest.AcceptPromos,
				)

				if err == nil {
					// Registration successful after cleanup
					fmt.Printf("Successfully registered user after cleaning up orphaned record\n")
					goto RegistrationSuccess
				}
			}

			// If we get here, either deletion failed or re-registration failed
			fmt.Printf("Error registering user after cleanup attempt: %v\n", err)
		}

		// Humanize common database and business logic errors
		errorMessage := "Pendaftaran gagal"
		humanErrorDetail := humanizeRegistrationError(err)

		fmt.Printf("Error registering user: %v\n", err)
		utils.ErrorResponse(c, http.StatusBadRequest, errorMessage, fmt.Errorf(humanErrorDetail))
		return
	}

RegistrationSuccess:

	// Set session data
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("authenticated", true)

	if err := session.Save(); err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save session", err)
		return
	}

	// Return auth response
	authResponse := dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenExpiry:  expiryTime,
		User: dto.UserDTO{
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email,
			Phone:        user.Phone,
			UserType:     user.UserType,
			Picture:      user.Picture,
			AcceptTerms:  user.AcceptTerms,
			AcceptPromos: user.AcceptPromos,
		},
	}

	utils.SuccessResponse(c, http.StatusCreated, "Pendaftaran berhasil", authResponse)
}

// LoginWithCredentialsHandler handles login with email/phone and password
func (h *AuthHandler) LoginWithCredentialsHandler(c *gin.Context) {
	var loginRequest dto.LoginCredentialsRequest
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Login the user
	user, accessToken, refreshToken, expiryTime, err := h.userUsecase.LoginWithCredentials(
		loginRequest.EmailOrPhone,
		loginRequest.Password,
	)

	if err != nil {
		fmt.Printf("Error logging in user: %v\n", err)
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	// Set session data
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("authenticated", true)

	if err := session.Save(); err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save session", err)
		return
	}

	// Return auth response
	authResponse := dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenExpiry:  expiryTime,
		User: dto.UserDTO{
			ID:           user.ID,
			Name:         user.Name,
			Email:        user.Email,
			Phone:        user.Phone,
			UserType:     user.UserType,
			UserRole:     user.UserRole,
			Picture:      user.Picture,
			AcceptTerms:  user.AcceptTerms,
			AcceptPromos: user.AcceptPromos,
			IsAdmin:      user.IsAdmin,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", authResponse)
}

// humanizeRegistrationError converts technical errors to user-friendly messages in Indonesian
func humanizeRegistrationError(err error) string {
	if err == nil {
		return "Terjadi kesalahan yang tidak diketahui"
	}

	errStr := err.Error()

	// Handle PostgreSQL constraint violations
	if strings.Contains(errStr, "duplicate key value violates unique constraint") {
		if strings.Contains(errStr, "users_email_key") {
			return "Alamat email ini sudah terdaftar. Silakan gunakan email yang berbeda atau coba masuk."
		}
		if strings.Contains(errStr, "users_phone_key") {
			return "Nomor telepon ini sudah terdaftar. Silakan gunakan nomor yang berbeda atau coba masuk."
		}
		return "Informasi ini sudah terdaftar. Silakan gunakan data yang berbeda atau coba masuk."
	}

	// Handle business logic errors
	if strings.Contains(errStr, "email already registered") {
		return "Alamat email ini sudah terdaftar. Silakan gunakan email yang berbeda atau coba masuk."
	}

	if strings.Contains(errStr, "phone already registered") {
		return "Nomor telepon ini sudah terdaftar. Silakan gunakan nomor yang berbeda atau coba masuk."
	}

	// Handle validation errors
	if strings.Contains(errStr, "email is required") {
		return "Alamat email wajib diisi untuk pendaftaran."
	}

	if strings.Contains(errStr, "phone number is required") {
		return "Nomor telepon wajib diisi untuk pendaftaran."
	}

	if strings.Contains(errStr, "password is required") {
		return "Kata sandi wajib diisi untuk pendaftaran."
	}

	if strings.Contains(errStr, "name is required") {
		return "Nama wajib diisi untuk pendaftaran."
	}

	// Handle database connection errors
	if strings.Contains(errStr, "connection refused") {
		return "Layanan sementara tidak tersedia. Silakan coba lagi nanti."
	}

	if strings.Contains(errStr, "timeout") {
		return "Permintaan habis waktu. Silakan coba lagi."
	}

	// For any other database errors
	if strings.Contains(errStr, "pq:") || strings.Contains(errStr, "sql:") {
		return "Terjadi kesalahan database. Silakan coba lagi atau hubungi dukungan jika masalah berlanjut."
	}

	// Return a generic message for unknown errors
	return "Pendaftaran gagal. Silakan periksa informasi Anda dan coba lagi."
}

// ForgotPasswordHandler initiates the password reset process
func (h *AuthHandler) ForgotPasswordHandler(c *gin.Context) {
	var forgotPasswordRequest dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&forgotPasswordRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate email or phone is provided
	if forgotPasswordRequest.EmailOrPhone == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Email atau nomor telepon wajib diisi", nil)
		return
	}

	// Process forgot password request
	ctx := c.Request.Context()
	err := h.userUsecase.ForgotPassword(ctx, forgotPasswordRequest.EmailOrPhone)
	if err != nil {
		// Log the error but don't expose details to client for security
		fmt.Printf("Error in forgot password: %v\n", err)

		// Check if it's a specific error we want to handle
		if strings.Contains(err.Error(), "failed to send password reset email") {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal mengirim email reset password. Silakan coba lagi nanti.", nil)
			return
		}

		// For any other error, still return success to not reveal user existence
	}

	// Always return success to prevent user enumeration
	utils.SuccessResponse(c, http.StatusOK, "Jika akun Anda terdaftar, kami telah mengirim instruksi reset password ke email Anda", nil)
}

// ResetPasswordHandler handles password reset with token
func (h *AuthHandler) ResetPasswordHandler(c *gin.Context) {
	var resetPasswordRequest dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&resetPasswordRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate token and password are provided
	if resetPasswordRequest.Token == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Token reset password wajib diisi", nil)
		return
	}

	if resetPasswordRequest.NewPassword == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Password baru wajib diisi", nil)
		return
	}

	// Additional password validation
	if len(resetPasswordRequest.NewPassword) < 8 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Password harus minimal 8 karakter", nil)
		return
	}

	// Process password reset
	ctx := c.Request.Context()
	err := h.userUsecase.ResetPassword(ctx, resetPasswordRequest.Token, resetPasswordRequest.NewPassword)
	if err != nil {
		fmt.Printf("Error in reset password: %v\n", err)

		// Handle specific errors
		if strings.Contains(err.Error(), "invalid or expired reset token") {
			utils.ErrorResponse(c, http.StatusBadRequest, "Token reset password tidak valid atau sudah kedaluwarsa", nil)
			return
		}

		if strings.Contains(err.Error(), "reset token has expired") {
			utils.ErrorResponse(c, http.StatusBadRequest, "Token reset password sudah kedaluwarsa. Silakan minta reset password baru", nil)
			return
		}

		if strings.Contains(err.Error(), "password must be at least 8 characters") {
			utils.ErrorResponse(c, http.StatusBadRequest, "Password harus minimal 8 karakter", nil)
			return
		}

		// Generic error for any other issues
		utils.ErrorResponse(c, http.StatusInternalServerError, "Gagal reset password. Silakan coba lagi", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password berhasil direset. Silakan login dengan password baru", nil)
}
