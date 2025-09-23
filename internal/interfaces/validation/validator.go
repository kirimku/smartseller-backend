package validation

import (
	"encoding/json"
	"html"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// RequestBody represents a generic request body with validation tags
type RequestBody struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=10"`
	Email        string `json:"email" validate:"omitempty,email"`
	Name         string `json:"name" validate:"omitempty,min=2,max=100"`
}

// ValidateAndSanitize validates the request body and sanitizes input
func ValidateAndSanitize(r *http.Request, data interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		return err
	}

	// Validate
	if err := validate.Struct(data); err != nil {
		return err
	}

	// Sanitize string fields
	switch v := data.(type) {
	case *RequestBody:
		v.Name = sanitizeString(v.Name)
		v.Email = sanitizeString(v.Email)
		v.RefreshToken = sanitizeString(v.RefreshToken)
	}

	return nil
}

// sanitizeString sanitizes input strings
func sanitizeString(s string) string {
	// First unescape any HTML entities
	s = html.UnescapeString(s)

	// Remove script tags and their content
	s = strings.ReplaceAll(s, "<script>", "")
	s = strings.ReplaceAll(s, "</script>", "")

	// Remove SQL injection characters
	s = strings.ReplaceAll(s, "'", "")
	s = strings.ReplaceAll(s, "\"", "")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "--", "") // Remove SQL comment syntax

	// Trim spaces
	s = strings.TrimSpace(s)

	return s
}
