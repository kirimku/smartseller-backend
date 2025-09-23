package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// AuthMiddleware verifies JWT tokens and sets user ID in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header is required", nil)
			c.Abort()
			return
		}

		// Extract token from header
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid authorization format", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Get JWT secret from environment variable
		secretKey := os.Getenv("SESSION_KEY")
		if secretKey == "" {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Server configuration error", nil)
			c.Abort()
			return
		}

		// Parse and verify the JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing algorithm
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token", err)
			c.Abort()
			return
		}

		// Extract claims from the token
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check token expiration
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					utils.ErrorResponse(c, http.StatusUnauthorized, "Token expired", nil)
					c.Abort()
					return
				}
			}

			// Extract user ID from the claims
			if userID, ok := claims["user_id"].(string); ok {
				// Set user ID in context for handlers to use
				c.Set("user_id", userID)
				c.Next()
				return
			}

			utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims", nil)
			c.Abort()
			return
		}

		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token", nil)
		c.Abort()
	}
}

// AuthRequired is an alias for AuthMiddleware to maintain backward compatibility
// with the wallet routes that use this name
func AuthRequired() gin.HandlerFunc {
	return AuthMiddleware()
}
