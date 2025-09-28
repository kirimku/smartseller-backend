package middleware

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// CustomerClaims represents JWT claims for customer authentication
type CustomerClaims struct {
	CustomerID    string `json:"customer_id"`
	StorefrontID  string `json:"storefront_id"`
	Email         string `json:"email"`
	TokenType     string `json:"token_type"` // "access" or "refresh"
	SessionID     string `json:"session_id"`
	TwoFactorAuth bool   `json:"two_factor_auth"`
	Permissions   []string `json:"permissions"`
	jwt.RegisteredClaims
}

// SimpleRateLimiter provides basic rate limiting functionality
type SimpleRateLimiter struct {
	requests    []time.Time
	maxRequests int
	window      time.Duration
	mu          sync.Mutex
}

// NewSimpleRateLimiter creates a new simple rate limiter
func NewSimpleRateLimiter(maxRequests int, window time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		requests:    make([]time.Time, 0),
		maxRequests: maxRequests,
		window:      window,
	}
}

// Allow checks if a request is allowed
func (rl *SimpleRateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// Check if we can allow this request
	if len(rl.requests) >= rl.maxRequests {
		return false
	}

	// Add this request
	rl.requests = append(rl.requests, now)
	return true
}

// CustomerAuthMiddleware provides JWT-based customer authentication
type CustomerAuthMiddleware struct {
	secretKey       string
	refreshKey      string
	rateLimiters    map[string]*SimpleRateLimiter
	rateLimiterMu   sync.RWMutex
	fraudDetector   *FraudDetector
	sessionManager  *CustomerSessionManager
}

// FraudDetector handles IP-based fraud detection
type FraudDetector struct {
	suspiciousIPs   map[string]time.Time
	failedAttempts  map[string]int
	mu              sync.RWMutex
	maxAttempts     int
	blockDuration   time.Duration
}

// CustomerSessionManager handles customer session management
type CustomerSessionManager struct {
	activeSessions map[string]time.Time
	mu             sync.RWMutex
	maxSessions    int
	sessionTimeout time.Duration
}

// NewCustomerAuthMiddleware creates a new customer authentication middleware
func NewCustomerAuthMiddleware() *CustomerAuthMiddleware {
	secretKey := os.Getenv("CUSTOMER_JWT_SECRET")
	if secretKey == "" {
		secretKey = os.Getenv("SESSION_KEY") // Fallback to existing key
		if secretKey == "" {
			secretKey = "customer-default-secret-key-for-development-only"
		}
	}

	refreshKey := os.Getenv("CUSTOMER_REFRESH_SECRET")
	if refreshKey == "" {
		refreshKey = secretKey + "-refresh"
	}

	return &CustomerAuthMiddleware{
		secretKey:      secretKey,
		refreshKey:     refreshKey,
		rateLimiters:   make(map[string]*SimpleRateLimiter),
		fraudDetector:  NewFraudDetector(),
		sessionManager: NewCustomerSessionManager(),
	}
}

// NewFraudDetector creates a new fraud detector
func NewFraudDetector() *FraudDetector {
	return &FraudDetector{
		suspiciousIPs:  make(map[string]time.Time),
		failedAttempts: make(map[string]int),
		maxAttempts:    5,
		blockDuration:  15 * time.Minute,
	}
}

// NewCustomerSessionManager creates a new session manager
func NewCustomerSessionManager() *CustomerSessionManager {
	return &CustomerSessionManager{
		activeSessions: make(map[string]time.Time),
		maxSessions:    5, // Max 5 concurrent sessions per customer
		sessionTimeout: 24 * time.Hour,
	}
}

// CustomerAuthRequired middleware for protected customer endpoints
func (cam *CustomerAuthMiddleware) CustomerAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check rate limiting first
		if !cam.checkRateLimit(c) {
			return
		}

		// Check fraud detection
		if !cam.checkFraudDetection(c) {
			return
		}

		// Extract and validate JWT token
		token, claims, err := cam.extractAndValidateToken(c, "access")
		if err != nil {
			cam.recordFailedAttempt(c)
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", err)
			c.Abort()
			return
		}

		// Verify tenant context
		if !cam.verifyTenantContext(c, claims) {
			utils.ErrorResponse(c, http.StatusForbidden, "Access denied for this storefront", nil)
			c.Abort()
			return
		}

		// Check session validity
		if !cam.checkSessionValidity(claims.SessionID, claims.CustomerID) {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Session expired or invalid", nil)
			c.Abort()
			return
		}

		// Set customer context
		cam.setCustomerContext(c, claims, token)
		c.Next()
	}
}

// OptionalCustomerAuth middleware for endpoints that work with or without authentication
func (cam *CustomerAuthMiddleware) OptionalCustomerAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check rate limiting
		if !cam.checkRateLimit(c) {
			return
		}

		// Try to extract token, but don't fail if not present
		token, claims, err := cam.extractAndValidateToken(c, "access")
		if err == nil && claims != nil {
			// Verify tenant context if token is present
			if cam.verifyTenantContext(c, claims) && cam.checkSessionValidity(claims.SessionID, claims.CustomerID) {
				cam.setCustomerContext(c, claims, token)
			}
		}

		c.Next()
	}
}

// RefreshTokenRequired middleware for refresh token endpoints
func (cam *CustomerAuthMiddleware) RefreshTokenRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cam.checkRateLimit(c) {
			return
		}

		token, claims, err := cam.extractAndValidateToken(c, "refresh")
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid refresh token", err)
			c.Abort()
			return
		}

		if claims.TokenType != "refresh" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token type", nil)
			c.Abort()
			return
		}

		cam.setCustomerContext(c, claims, token)
		c.Next()
	}
}

// TwoFactorRequired middleware for endpoints requiring 2FA
func (cam *CustomerAuthMiddleware) TwoFactorRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run regular auth
		cam.CustomerAuthRequired()(c)
		if c.IsAborted() {
			return
		}

		// Check if 2FA is enabled and verified
		claims, exists := c.Get("customer_claims")
		if !exists {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication context missing", nil)
			c.Abort()
			return
		}

		customerClaims := claims.(*CustomerClaims)
		if !customerClaims.TwoFactorAuth {
			utils.ErrorResponse(c, http.StatusForbidden, "Two-factor authentication required", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware provides configurable rate limiting for customer endpoints
func (cam *CustomerAuthMiddleware) RateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := cam.getClientIdentifier(c)
		
		cam.rateLimiterMu.Lock()
		limiter, exists := cam.rateLimiters[identifier]
		if !exists {
			limiter = NewSimpleRateLimiter(requestsPerMinute, time.Minute)
			cam.rateLimiters[identifier] = limiter
		}
		cam.rateLimiterMu.Unlock()

		if !limiter.Allow() {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Rate limit exceeded", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// CaptchaRequired middleware for endpoints requiring CAPTCHA verification
func (cam *CustomerAuthMiddleware) CaptchaRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		captchaToken := c.GetHeader("X-Captcha-Token")
		if captchaToken == "" {
			captchaToken = c.PostForm("captcha_token")
		}

		if captchaToken == "" {
			utils.ErrorResponse(c, http.StatusBadRequest, "CAPTCHA verification required", nil)
			c.Abort()
			return
		}

		// Verify CAPTCHA token (mock implementation)
		if !cam.verifyCaptcha(captchaToken) {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid CAPTCHA", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers for customer endpoints
func (cam *CustomerAuthMiddleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		
		// GDPR/CCPA compliance headers
		c.Header("X-Privacy-Policy", "https://example.com/privacy")
		c.Header("X-Data-Processing", "gdpr-compliant")
		
		c.Next()
	}
}

// CORSMiddleware provides CORS configuration for customer endpoints
func (cam *CustomerAuthMiddleware) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Allow specific origins (should be configurable)
		allowedOrigins := []string{
			"http://localhost:3000",
			"https://app.example.com",
			"https://*.example.com",
		}

		if cam.isOriginAllowed(origin, allowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Captcha-Token, X-Device-ID")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Helper methods

func (cam *CustomerAuthMiddleware) extractAndValidateToken(c *gin.Context, tokenType string) (*jwt.Token, *CustomerClaims, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return nil, nil, fmt.Errorf("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, nil, fmt.Errorf("invalid authorization format")
	}

	tokenString := parts[1]
	secretKey := cam.secretKey
	if tokenType == "refresh" {
		secretKey = cam.refreshKey
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomerClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, nil, err
	}

	claims, ok := token.Claims.(*CustomerClaims)
	if !ok || !token.Valid {
		return nil, nil, fmt.Errorf("invalid token claims")
	}

	// Check token expiration
	if time.Now().Unix() > claims.ExpiresAt.Unix() {
		return nil, nil, fmt.Errorf("token expired")
	}

	return token, claims, nil
}

func (cam *CustomerAuthMiddleware) verifyTenantContext(c *gin.Context, claims *CustomerClaims) bool {
	// Get tenant context from Gin context
	tenantContext, exists := c.Get("tenant_context")
	if !exists {
		return false
	}

	// Type assert to tenant context
	tenantCtx, ok := tenantContext.(*tenant.TenantContext)
	if !ok {
		return false
	}

	// Verify token belongs to current storefront
	return claims.StorefrontID == tenantCtx.StorefrontID.String()
}

func (cam *CustomerAuthMiddleware) checkRateLimit(c *gin.Context) bool {
	identifier := cam.getClientIdentifier(c)
	
	// Check rate limit
	limiter := cam.getRateLimiter(identifier)
	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":   "Rate limit exceeded",
			"message": "Too many requests. Please try again later.",
		})
		c.Abort()
		return false
	}

	return true
}

func (cam *CustomerAuthMiddleware) getRateLimiter(identifier string) *SimpleRateLimiter {
	cam.rateLimiterMu.Lock()
	limiter, exists := cam.rateLimiters[identifier]
	if !exists {
		limiter = NewSimpleRateLimiter(60, time.Minute) // Default: 60 requests per minute
		cam.rateLimiters[identifier] = limiter
	}
	cam.rateLimiterMu.Unlock()
	return limiter
}

func (cam *CustomerAuthMiddleware) checkFraudDetection(c *gin.Context) bool {
	clientIP := cam.getClientIP(c)
	
	cam.fraudDetector.mu.RLock()
	blockedUntil, isBlocked := cam.fraudDetector.suspiciousIPs[clientIP]
	attempts := cam.fraudDetector.failedAttempts[clientIP]
	cam.fraudDetector.mu.RUnlock()

	if isBlocked && time.Now().Before(blockedUntil) {
		utils.ErrorResponse(c, http.StatusForbidden, "IP temporarily blocked due to suspicious activity", nil)
		c.Abort()
		return false
	}

	if attempts >= cam.fraudDetector.maxAttempts {
		cam.fraudDetector.mu.Lock()
		cam.fraudDetector.suspiciousIPs[clientIP] = time.Now().Add(cam.fraudDetector.blockDuration)
		cam.fraudDetector.mu.Unlock()
		
		utils.ErrorResponse(c, http.StatusForbidden, "Too many failed attempts. IP blocked temporarily", nil)
		c.Abort()
		return false
	}

	return true
}

func (cam *CustomerAuthMiddleware) checkSessionValidity(sessionID, customerID string) bool {
	cam.sessionManager.mu.RLock()
	lastUsed, exists := cam.sessionManager.activeSessions[sessionID]
	cam.sessionManager.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Now().Sub(lastUsed) > cam.sessionManager.sessionTimeout {
		cam.sessionManager.mu.Lock()
		delete(cam.sessionManager.activeSessions, sessionID)
		cam.sessionManager.mu.Unlock()
		return false
	}

	// Update last used time
	cam.sessionManager.mu.Lock()
	cam.sessionManager.activeSessions[sessionID] = time.Now()
	cam.sessionManager.mu.Unlock()

	return true
}

func (cam *CustomerAuthMiddleware) setCustomerContext(c *gin.Context, claims *CustomerClaims, token *jwt.Token) {
	c.Set("customer_id", claims.CustomerID)
	c.Set("customer_email", claims.Email)
	c.Set("storefront_id", claims.StorefrontID)
	c.Set("session_id", claims.SessionID)
	c.Set("customer_claims", claims)
	c.Set("customer_token", token)
	c.Set("two_factor_auth", claims.TwoFactorAuth)
	c.Set("customer_permissions", claims.Permissions)
}

func (cam *CustomerAuthMiddleware) recordFailedAttempt(c *gin.Context) {
	clientIP := cam.getClientIP(c)
	
	cam.fraudDetector.mu.Lock()
	cam.fraudDetector.failedAttempts[clientIP]++
	cam.fraudDetector.mu.Unlock()
}

func (cam *CustomerAuthMiddleware) getClientIdentifier(c *gin.Context) string {
	// Try to get customer ID first, fallback to IP
	if customerID, exists := c.Get("customer_id"); exists {
		return fmt.Sprintf("customer:%s", customerID)
	}
	return fmt.Sprintf("ip:%s", cam.getClientIP(c))
}

func (cam *CustomerAuthMiddleware) getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

func (cam *CustomerAuthMiddleware) verifyCaptcha(token string) bool {
	// Mock CAPTCHA verification - in production, integrate with reCAPTCHA or similar
	// This is a simple base64 check for demonstration
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return false
	}
	
	// Simple validation - in production, verify with CAPTCHA service
	return len(decoded) > 10 && strings.Contains(string(decoded), "valid")
}

func (cam *CustomerAuthMiddleware) isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// Simple wildcard matching for subdomains
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}

// Utility functions for getting customer context

// GetCustomerID extracts customer ID from context
func GetCustomerID(c *gin.Context) (string, bool) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		return "", false
	}
	return customerID.(string), true
}

// MustGetCustomerID extracts customer ID from context or panics
func MustGetCustomerID(c *gin.Context) string {
	customerID, exists := GetCustomerID(c)
	if !exists {
		panic("customer_id not found in context")
	}
	return customerID
}

// GetCustomerClaims extracts customer claims from context
func GetCustomerClaims(c *gin.Context) (*CustomerClaims, bool) {
	claims, exists := c.Get("customer_claims")
	if !exists {
		return nil, false
	}
	return claims.(*CustomerClaims), true
}

// GetCustomerStorefrontID extracts storefront ID from customer context
func GetCustomerStorefrontID(c *gin.Context) (uuid.UUID, bool) {
	storefrontID, exists := c.Get("storefront_id")
	if !exists {
		return uuid.Nil, false
	}
	
	id, err := uuid.Parse(storefrontID.(string))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// IsCustomerAuthenticated checks if customer is authenticated
func IsCustomerAuthenticated(c *gin.Context) bool {
	_, exists := GetCustomerID(c)
	return exists
}

// HasCustomerPermission checks if customer has specific permission
func HasCustomerPermission(c *gin.Context, permission string) bool {
	claims, exists := GetCustomerClaims(c)
	if !exists {
		return false
	}
	
	for _, perm := range claims.Permissions {
		if perm == permission || perm == "*" {
			return true
		}
	}
	return false
}

// CreateCustomerToken creates a new JWT token for customer
func (cam *CustomerAuthMiddleware) CreateCustomerToken(customerID, storefrontID, email, sessionID string, tokenType string, permissions []string, twoFactorAuth bool) (string, error) {
	expiryTime := time.Now().Add(24 * time.Hour) // Access token: 24 hours
	if tokenType == "refresh" {
		expiryTime = time.Now().Add(7 * 24 * time.Hour) // Refresh token: 7 days
	}

	claims := &CustomerClaims{
		CustomerID:    customerID,
		StorefrontID:  storefrontID,
		Email:         email,
		TokenType:     tokenType,
		SessionID:     sessionID,
		TwoFactorAuth: twoFactorAuth,
		Permissions:   permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "smartseller-customer-api",
			Subject:   customerID,
			ID:        sessionID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	secretKey := cam.secretKey
	if tokenType == "refresh" {
		secretKey = cam.refreshKey
	}
	
	return token.SignedString([]byte(secretKey))
}

// RevokeCustomerSession revokes a customer session
func (cam *CustomerAuthMiddleware) RevokeCustomerSession(sessionID string) {
	cam.sessionManager.mu.Lock()
	delete(cam.sessionManager.activeSessions, sessionID)
	cam.sessionManager.mu.Unlock()
}

// CreateCustomerSession creates a new customer session
func (cam *CustomerAuthMiddleware) CreateCustomerSession(customerID string) string {
	sessionID := uuid.New().String()
	
	cam.sessionManager.mu.Lock()
	cam.sessionManager.activeSessions[sessionID] = time.Now()
	cam.sessionManager.mu.Unlock()
	
	return sessionID
}