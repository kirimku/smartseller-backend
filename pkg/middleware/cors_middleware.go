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
		// Skip CORS middleware in production (handled by DigitalOcean/nginx)
		if config.AppConfig.Environment == "production" {
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

		// In development, be more permissive with CORS
		if config.AppConfig.Environment == "development" && !isAllowedOrigin {
			// Check if it's a localhost variant that should be allowed
			if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
				isAllowedOrigin = true
			}
		}

		if !isAllowedOrigin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Not allowed by CORS",
			})
			return
		}

		// Set CORS headers EARLY - this ensures they're present even on redirects
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")

		// Handle preflight OPTIONS requests
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Max-Age", "3600")
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
