package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/config"
)

// APISecurityMiddleware configures security headers specifically for API endpoints
// This middleware addresses CSP issues that can block API requests in production
func APISecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add debug headers to help troubleshoot CSP issues
		c.Header("X-Debug-Environment", config.AppConfig.Environment)
		c.Header("X-Debug-Middleware", "api-security")

		// In production, handle CSP more carefully for API endpoints
		if config.AppConfig.Environment == "production" {
			// Remove any restrictive CSP headers that might block API calls
			// This is especially important for auth endpoints like /register
			c.Header("Content-Security-Policy", "")

			// Set minimal, API-friendly security headers
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

			// Ensure CORS headers are properly set for API access
			origin := c.Request.Header.Get("Origin")
			if origin != "" {
				// Check if origin is in allowed list
				allowedOrigins := config.AppConfig.AllowedOrigins
				for _, allowed := range allowedOrigins {
					if strings.TrimSpace(allowed) == origin {
						c.Header("Access-Control-Allow-Origin", origin)
						c.Header("Access-Control-Allow-Credentials", "true")
						c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
						c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
						break
					}
				}
			}
		}

		c.Next()
	}
}
