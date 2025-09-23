package utils

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Helper function to check for errors and handle them appropriately
func CheckError(err error) {
	if err != nil {
		panic(err) // In a real application, you might want to handle this differently
	}
}

// Helper function to generate a response message
func ResponseMessage(message string) map[string]string {
	return map[string]string{"message": message}
}

// JSONResponse represents a standardized structure for API responses
type JSONResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondJSON sends a standardized JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, status string, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := JSONResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}
	json.NewEncoder(w).Encode(response)
}

// RespondError sends a standardized JSON error response
func RespondError(w http.ResponseWriter, statusCode int, message string) {
	RespondJSON(w, statusCode, "error", message, nil)
}

// GetUserIDFromContext extracts the user ID from the Gin context
// This function is used to get the authenticated user's ID, which is
// set by the AuthMiddleware after successful JWT validation
func GetUserIDFromContext(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	// Convert to string if possible
	if id, ok := userID.(string); ok {
		return id
	}

	return ""
}

// ParseInt converts a string to an integer with a default fallback value
func ParseInt(s string, defaultValue int) (int, error) {
	if s == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue, err
	}

	return value, nil
}

// IsAdminUser checks if the current user has admin privileges
// This function relies on the "is_admin" flag set in the context by the AdminMiddleware
func IsAdminUser(c *gin.Context) bool {
	isAdmin, exists := c.Get("is_admin")
	if !exists {
		return false
	}

	// Convert to bool if possible
	if admin, ok := isAdmin.(bool); ok && admin {
		return true
	}

	return false
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err) // In production, you might want to handle this differently
	}
	return hex.EncodeToString(bytes)
}
