package setup

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuthHelper provides shared authentication functionality for integration tests
type AuthHelper struct {
	mu           sync.RWMutex
	token        string
	refreshToken string
	userID       uuid.UUID
	expiresAt    time.Time
	baseURL      string
	httpClient   *http.Client
}

// AuthResponse represents the response from authentication endpoint
type AuthResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	UserID       uuid.UUID `json:"user_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// NewAuthHelper creates a new authentication helper instance
func NewAuthHelper(baseURL string) *AuthHelper {
	return &AuthHelper{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Login authenticates with the API and stores the token for reuse
func (a *AuthHelper) Login(email, password string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	resp, err := a.httpClient.Post(
		a.baseURL+"/api/v1/auth/login",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to send login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	a.token = authResp.Token
	a.refreshToken = authResp.RefreshToken
	a.userID = authResp.UserID
	a.expiresAt = authResp.ExpiresAt

	return nil
}

// GetValidToken returns a valid authentication token, refreshing if necessary
func (a *AuthHelper) GetValidToken() (string, error) {
	a.mu.RLock()
	if time.Now().Before(a.expiresAt.Add(-5*time.Minute)) && a.token != "" {
		token := a.token
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	// Token is expired or about to expire, refresh it
	if err := a.RefreshTokenIfNeeded(); err != nil {
		return "", err
	}

	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token, nil
}

// RefreshTokenIfNeeded refreshes the authentication token if it's expired or about to expire
func (a *AuthHelper) RefreshTokenIfNeeded() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if time.Now().Before(a.expiresAt.Add(-5 * time.Minute)) {
		return nil // Token is still valid
	}

	refreshReq := map[string]string{
		"refresh_token": a.refreshToken,
	}

	jsonData, err := json.Marshal(refreshReq)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	resp, err := a.httpClient.Post(
		a.baseURL+"/api/v1/auth/refresh",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to send refresh request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("failed to decode refresh response: %w", err)
	}

	a.token = authResp.Token
	a.refreshToken = authResp.RefreshToken
	a.expiresAt = authResp.ExpiresAt

	return nil
}

// AuthenticatedRequest creates an HTTP request with authentication headers
func (a *AuthHelper) AuthenticatedRequest(method, path string, body interface{}) (*http.Request, error) {
	var jsonData []byte
	var err error

	if body != nil {
		jsonData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := a.baseURL + path
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	token, err := a.GetValidToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// AuthenticatedTestRequest creates an authenticated test request for httptest
func (a *AuthHelper) AuthenticatedTestRequest(method, path string, body interface{}) (*http.Request, error) {
	var jsonData []byte
	var err error

	if body != nil {
		jsonData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonData))

	token, err := a.GetValidToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get valid token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

// GetUserID returns the authenticated user's ID
func (a *AuthHelper) GetUserID() uuid.UUID {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.userID
}

// IsAuthenticated returns true if the helper has a valid token
func (a *AuthHelper) IsAuthenticated() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.token != "" && time.Now().Before(a.expiresAt)
}

// Logout clears the stored authentication information
func (a *AuthHelper) Logout() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.token = ""
	a.refreshToken = ""
	a.userID = uuid.Nil
	a.expiresAt = time.Time{}
}

// LoginWithTestUser authenticates with a predefined test user
func (a *AuthHelper) LoginWithTestUser() error {
	return a.Login("testuser@example.com", "testpassword123")
}

// LoginWithAdminUser authenticates with a predefined admin test user
func (a *AuthHelper) LoginWithAdminUser() error {
	return a.Login("admin@example.com", "adminpassword123")
}
