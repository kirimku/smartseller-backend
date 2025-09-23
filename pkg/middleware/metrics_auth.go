package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// MetricsAuthMiddleware restricts access to metrics endpoint
func MetricsAuthMiddleware(allowedIPs []string) gin.HandlerFunc {
	// Parse allowed IP ranges
	var allowedNets []*net.IPNet
	for _, ipStr := range allowedIPs {
		if strings.Contains(ipStr, "/") {
			// CIDR notation
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err != nil {
				logger.Error("Invalid CIDR in allowed IPs", "cidr", ipStr, "error", err)
				continue
			}
			allowedNets = append(allowedNets, ipNet)
		} else {
			// Single IP
			ip := net.ParseIP(ipStr)
			if ip == nil {
				logger.Error("Invalid IP in allowed IPs", "ip", ipStr)
				continue
			}
			// Convert single IP to /32 or /128 network
			var maskSize int
			if ip.To4() != nil {
				maskSize = 32 // IPv4
			} else {
				maskSize = 128 // IPv6
			}
			_, ipNet, _ := net.ParseCIDR(ipStr + "/" + string(rune(maskSize)))
			allowedNets = append(allowedNets, ipNet)
		}
	}

	return func(c *gin.Context) {
		clientIP := getClientIP(c)

		// Check if client IP is in allowed list
		allowed := false
		for _, ipNet := range allowedNets {
			if ipNet.Contains(net.ParseIP(clientIP)) {
				allowed = true
				break
			}
		}

		if !allowed {
			logger.Warn("Unauthorized metrics access attempt",
				"client_ip", clientIP,
				"user_agent", c.GetHeader("User-Agent"),
				"path", c.Request.URL.Path)

			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
			c.Abort()
			return
		}

		logger.Debug("Metrics access granted", "client_ip", clientIP)
		c.Next()
	}
}

// BasicAuthMiddleware provides HTTP Basic Authentication for metrics endpoint
func BasicAuthMiddleware(username, password string) gin.HandlerFunc {
	if username == "" || password == "" {
		logger.Warn("Basic auth credentials not configured for metrics endpoint")
		// Return middleware that denies all access if no credentials
		return func(c *gin.Context) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Authentication required"})
			c.Abort()
		}
	}

	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Header("WWW-Authenticate", `Basic realm="Metrics"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization method"})
			c.Abort()
			return
		}

		// Decode base64 credentials
		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials format"})
			c.Abort()
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(pair[0]), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pair[1]), []byte(password)) != 1 {

			logger.Warn("Failed metrics authentication attempt",
				"client_ip", getClientIP(c),
				"user_agent", c.GetHeader("User-Agent"))

			c.Header("WWW-Authenticate", `Basic realm="Metrics"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			c.Abort()
			return
		}

		logger.Debug("Metrics authentication successful", "client_ip", getClientIP(c))
		c.Next()
	}
}

// getClientIP extracts the real client IP from request headers
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (most common)
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xRealIP := c.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// Check CF-Connecting-IP (Cloudflare)
	cfConnectingIP := c.GetHeader("CF-Connecting-IP")
	if cfConnectingIP != "" {
		return cfConnectingIP
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}
