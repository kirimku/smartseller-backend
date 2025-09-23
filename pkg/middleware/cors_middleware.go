package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/config"
)

// CORSMiddleware handles Cross-Origin Resource Sharing (CORS) for Gin
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Handle preflight OPTIONS requests even in preproduction
		// This ensures AUTH endpoints work properly with CORS
		if c.Request.Method == "OPTIONS" {
			// Set permissive CORS headers for preflight requests in preproduction
			if config.AppConfig.Environment == "preproduction" {
				c.Header("Access-Control-Allow-Origin", "*")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
				c.Header("Access-Control-Max-Age", "3600")
				c.AbortWithStatus(http.StatusOK)
				return
			}
		}

		// Skip CORS middleware in production and preproduction (handled by DigitalOcean/nginx)
		// BUT still handle OPTIONS requests above
		if config.AppConfig.Environment == "preproduction" {
			c.Next()
			return
		}

		origin := c.Request.Header.Get("Origin")

		// If not a CORS request, proceed normally
		if origin == "" {
			c.Next()
			return
		}

		// Check if origin is allowed
		allowedOrigins := config.AppConfig.AllowedOrigins

		// Helper function to normalize origin for comparison
		normalizeOrigin := func(origin string) string {
			u, err := url.Parse(origin)
			if err != nil {
				return origin
			}
			// Convert localhost to 127.0.0.1 for comparison
			if u.Hostname() == "localhost" {
				u.Host = "127.0.0.1" + ":" + u.Port()
			}
			return u.String()
		}

		normalizedRequestOrigin := normalizeOrigin(origin)
		isAllowedOrigin := false

		for _, allowed := range allowedOrigins {
			normalizedAllowed := normalizeOrigin(strings.TrimSpace(allowed))

			if normalizedAllowed == normalizedRequestOrigin {
				isAllowedOrigin = true
				break
			}
		}

		if !isAllowedOrigin {

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Not allowed by CORS",
			})
			return
		}

		// Set CORS headers (only if not already set to prevent duplication)
		// Note: Use c.Writer.Header().Get() to check response headers, not request headers
		existingOrigin := c.Writer.Header().Get("Access-Control-Allow-Origin")

		if existingOrigin == "" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		if c.Writer.Header().Get("Access-Control-Allow-Credentials") == "" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Writer.Header().Get("Access-Control-Allow-Methods") == "" {
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		}
		if c.Writer.Header().Get("Access-Control-Allow-Headers") == "" {
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
		}

		// Handle preflight requests
		if c.Request.Method == http.MethodOptions {
			if c.Writer.Header().Get("Access-Control-Max-Age") == "" {
				c.Header("Access-Control-Max-Age", "3600")
			}
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip strict security headers in preproduction to avoid conflicts
		if config.AppConfig.Environment == "preproduction" {
			c.Next()
			return
		} // Set standard security headers only for production

		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// CSP removed to avoid conflicts with frontend applications
		// c.Header("Content-Security-Policy", "default-src 'self'; connect-src 'self' https://preproduction.kirimku.com https://kirimku.com")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
