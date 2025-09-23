package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response is the standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ValidationErrorMessages maps validation tags to human-readable messages in Indonesian
var ValidationErrorMessages = map[string]string{
	"required": "wajib diisi",
	"email":    "Mohon masukkan alamat email yang valid",
	"min":      "terlalu pendek",
	"max":      "terlalu panjang",
	"oneof":    "Mohon pilih salah satu opsi yang valid",
	"eq":       "Anda harus menyetujui syarat dan ketentuan",
}

// FieldDisplayNames maps field names to human-readable names in Indonesian
var FieldDisplayNames = map[string]string{
	"Name":         "Nama",
	"Email":        "Email",
	"Phone":        "Nomor telepon",
	"Password":     "Kata sandi",
	"UserType":     "Jenis pengguna",
	"AcceptTerms":  "Syarat dan ketentuan",
	"AcceptPromos": "Email promosi",
}

// ParseDatabaseError converts database errors to human-readable messages in Indonesian
func ParseDatabaseError(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Handle PostgreSQL constraint violations (check for nested errors too)
	if strings.Contains(errStr, "duplicate key value violates unique constraint") {
		if strings.Contains(errStr, "users_email_key") {
			return "Alamat email ini sudah terdaftar. Silakan gunakan email yang berbeda atau coba masuk."
		}
		if strings.Contains(errStr, "users_phone_key") {
			return "Nomor telepon ini sudah terdaftar. Silakan gunakan nomor yang berbeda atau coba masuk."
		}
		return "Informasi ini sudah terdaftar. Silakan gunakan data yang berbeda atau coba masuk."
	}

	// Handle other common database errors
	if strings.Contains(errStr, "connection refused") {
		return "Layanan sementara tidak tersedia. Silakan coba lagi nanti."
	}

	if strings.Contains(errStr, "timeout") {
		return "Permintaan habis waktu. Silakan coba lagi."
	}

	if strings.Contains(errStr, "foreign key constraint") {
		return "Data referensi tidak valid. Silakan periksa input Anda."
	}

	// Handle specific business logic errors
	if strings.Contains(errStr, "email already registered") {
		return "Alamat email ini sudah terdaftar. Silakan gunakan email yang berbeda atau coba masuk."
	}

	if strings.Contains(errStr, "phone already registered") {
		return "Nomor telepon ini sudah terdaftar. Silakan gunakan nomor yang berbeda atau coba masuk."
	}

	// For unknown database errors, return a generic message
	if strings.Contains(errStr, "pq:") || strings.Contains(errStr, "sql:") {
		return "Terjadi kesalahan database. Silakan coba lagi atau hubungi dukungan jika masalah berlanjut."
	}

	// Return original error if it's not a recognized database error
	return errStr
}

// ParseValidationErrorsToMessages converts Gin validation errors to human-readable messages
func ParseValidationErrorsToMessages(err error) []string {
	var messages []string

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldName := fieldError.Field()
			tag := fieldError.Tag()

			// Get human-readable field name
			displayName, exists := FieldDisplayNames[fieldName]
			if !exists {
				displayName = fieldName
			}

			// Get human-readable error message
			message, exists := ValidationErrorMessages[tag]
			if !exists {
				message = "Nilai tidak valid"
			}

			// Create specific messages for certain validation types
			switch tag {
			case "min":
				if fieldName == "Name" {
					message = "Nama minimal harus 3 karakter"
				} else if fieldName == "Phone" {
					message = "Nomor telepon minimal harus 10 digit"
				} else if fieldName == "Password" {
					message = "Kata sandi minimal harus 8 karakter"
				}
			case "max":
				if fieldName == "Name" {
					message = "Nama tidak boleh lebih dari 100 karakter"
				} else if fieldName == "Phone" {
					message = "Nomor telepon tidak boleh lebih dari 20 digit"
				} else if fieldName == "Password" {
					message = "Kata sandi tidak boleh lebih dari 100 karakter"
				}
			case "oneof":
				if fieldName == "UserType" {
					message = "Jenis pengguna harus 'personal', 'bisnis', atau 'agen'"
				}
			}

			messages = append(messages, displayName+": "+message)
		}
	} else {
		// If it's not a validation error, return the original error message
		messages = append(messages, err.Error())
	}

	return messages
}

// SuccessResponse sends a standardized success response
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse returns a standardized error response
func ErrorResponse(c *gin.Context, statusCode int, message string, details interface{}) {
	response := gin.H{
		"success": false,
		"message": message,
		"error":   getErrorCode(statusCode),
		"meta":    gin.H{"http_status": statusCode},
	}

	// Handle different types of details
	if details != nil {
		switch v := details.(type) {
		case error:
			// First check if it's a validation error
			if validationMessages := ParseValidationErrorsToMessages(v); len(validationMessages) > 0 {
				// It's a validation error
				response["error_detail"] = strings.Join(validationMessages, "; ")
				response["validation_errors"] = validationMessages
			} else {
				// Check if it's a database error that needs humanization
				humanError := ParseDatabaseError(v)
				response["error_detail"] = humanError
			}
		default:
			// For other types (validation errors, etc.), include as is
			response["details"] = v
		}
	}

	c.JSON(statusCode, response)
}

// PaginatedResponse returns a standardized paginated response
func PaginatedResponse(c *gin.Context, statusCode int, data interface{}, page, limit, total int) {
	lastPage := (total + limit - 1) / limit
	if lastPage < 1 {
		lastPage = 1
	}

	c.JSON(statusCode, gin.H{
		"data": data,
		"pagination": gin.H{
			"total":        total,
			"per_page":     limit,
			"current_page": page,
			"last_page":    lastPage,
		},
	})
}

// getErrorCode returns an error code based on HTTP status code
func getErrorCode(statusCode int) string {
	switch statusCode {
	case 400:
		return "bad_request"
	case 401:
		return "unauthorized"
	case 403:
		return "forbidden"
	case 404:
		return "not_found"
	case 409:
		return "conflict"
	case 422:
		return "validation_error"
	case 500:
		return "internal_error"
	default:
		return "unknown_error"
	}
}
