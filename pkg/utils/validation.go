package utils

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ParseValidationErrors parses validation errors from the validator
func ParseValidationErrors(err error) []ValidationError {
	if err == nil {
		return nil
	}

	var validationErrors []ValidationError

	// Check if the error is a validator.ValidationErrors
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrs {
			field := e.Field()
			// Convert camelCase to snake_case
			field = toSnakeCase(field)

			message := getMessage(e)
			validationErrors = append(validationErrors, ValidationError{
				Field:   field,
				Message: message,
			})
		}
		return validationErrors
	}

	// For other types of errors, return a generic validation error
	return []ValidationError{
		{
			Field:   "general",
			Message: err.Error(),
		},
	}
}

// toSnakeCase converts a camelCase string to snake_case
func toSnakeCase(input string) string {
	var result strings.Builder
	for i, r := range input {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(toLower(r))
	}
	return result.String()
}

// toLower converts a rune to lowercase
func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}

// getMessage returns a user-friendly message for validation errors
func getMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		if e.Type().Kind().String() == "string" {
			return "Must be at least " + e.Param() + " characters long"
		}
		return "Must be at least " + e.Param()
	case "max":
		if e.Type().Kind().String() == "string" {
			return "Must be at most " + e.Param() + " characters long"
		}
		return "Must be at most " + e.Param()
	case "gt":
		return "Must be greater than " + e.Param()
	case "gte":
		return "Must be greater than or equal to " + e.Param()
	case "lt":
		return "Must be less than " + e.Param()
	case "lte":
		return "Must be less than or equal to " + e.Param()
	case "oneof":
		return "Must be one of: " + e.Param()
	}
	return "Invalid value for " + e.Field()
}
