package dto

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (v ValidationErrors) Error() string {
	var messages []string
	for _, err := range v.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// Business validation rules for wallet transactions

// ValidateWalletTransactionCreateRequest validates wallet transaction creation request
func ValidateWalletTransactionCreateRequest(req *WalletTransactionCreateRequest) *ValidationErrors {
	var validationErrors []ValidationError

	// Validate wallet ID format (should be UUID)
	if req.WalletID == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "wallet_id",
			Message: "Wallet ID is required",
			Value:   req.WalletID,
		})
	} else if !isValidUUID(req.WalletID) {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "wallet_id",
			Message: "Wallet ID must be a valid UUID",
			Value:   req.WalletID,
		})
	}

	// Validate order price (must be positive)
	if req.OrderPrice <= 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "order_price",
			Message: "Order price must be greater than 0",
			Value:   req.OrderPrice,
		})
	}

	// Validate order price range (reasonable limits)
	if req.OrderPrice > 100000000 { // 100M IDR
		validationErrors = append(validationErrors, ValidationError{
			Field:   "order_price",
			Message: "Order price exceeds maximum allowed amount (100,000,000 IDR)",
			Value:   req.OrderPrice,
		})
	}

	// Validate weight (must be positive)
	if req.Weight <= 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "weight",
			Message: "Weight must be greater than 0",
			Value:   req.Weight,
		})
	}

	// Validate order name (sanitization)
	if sanitizedName := SanitizeString(req.OrderName); sanitizedName != req.OrderName {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "order_name",
			Message: "Order name contains invalid characters",
			Value:   req.OrderName,
		})
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{Errors: validationErrors}
	}
	return nil
}

// ValidateWalletRefundRequest validates wallet refund request
func ValidateWalletRefundRequest(req *WalletRefundRequest) *ValidationErrors {
	var validationErrors []ValidationError

	// Validate transaction ID
	if req.TransactionID <= 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "transaction_id",
			Message: "Transaction ID must be greater than 0",
			Value:   req.TransactionID,
		})
	}

	// Validate refund amount
	if req.RefundAmount <= 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "refund_amount",
			Message: "Refund amount must be greater than 0",
			Value:   req.RefundAmount,
		})
	}

	// Validate refund amount range
	if req.RefundAmount > 100000000 { // 100M IDR
		validationErrors = append(validationErrors, ValidationError{
			Field:   "refund_amount",
			Message: "Refund amount exceeds maximum allowed amount (100,000,000 IDR)",
			Value:   req.RefundAmount,
		})
	}

	// Validate reason (required and length)
	if req.Reason == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "reason",
			Message: "Refund reason is required",
			Value:   req.Reason,
		})
	} else if len(req.Reason) < 5 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "reason",
			Message: "Refund reason must be at least 5 characters",
			Value:   req.Reason,
		})
	} else if len(req.Reason) > 500 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "reason",
			Message: "Refund reason must not exceed 500 characters",
			Value:   req.Reason,
		})
	}

	// Sanitize reason
	if sanitizedReason := SanitizeString(req.Reason); sanitizedReason != req.Reason {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "reason",
			Message: "Refund reason contains invalid characters",
			Value:   req.Reason,
		})
	}

	// Validate admin note (if provided)
	if req.AdminNote != "" && len(req.AdminNote) > 1000 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "admin_note",
			Message: "Admin note must not exceed 1000 characters",
			Value:   req.AdminNote,
		})
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{Errors: validationErrors}
	}
	return nil
}

// ValidateWalletBalanceValidationRequest validates balance validation request
func ValidateWalletBalanceValidationRequest(req *WalletBalanceValidationRequest) *ValidationErrors {
	var validationErrors []ValidationError

	// Validate wallet ID
	if req.WalletID == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "wallet_id",
			Message: "Wallet ID is required",
			Value:   req.WalletID,
		})
	} else if !isValidUUID(req.WalletID) {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "wallet_id",
			Message: "Wallet ID must be a valid UUID",
			Value:   req.WalletID,
		})
	}

	// Validate amount
	if req.Amount <= 0 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "amount",
			Message: "Amount must be greater than 0",
			Value:   req.Amount,
		})
	}

	// Validate amount range
	if req.Amount > 100000000 { // 100M IDR
		validationErrors = append(validationErrors, ValidationError{
			Field:   "amount",
			Message: "Amount exceeds maximum allowed amount (100,000,000 IDR)",
			Value:   req.Amount,
		})
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{Errors: validationErrors}
	}
	return nil
}

// Input sanitization functions

// SanitizeString removes potentially dangerous characters and trims whitespace
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove control characters except newlines and tabs
	var builder strings.Builder
	for _, r := range input {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			continue
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

// SanitizeFilename removes dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Remove dangerous characters for filenames
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = reg.ReplaceAllString(filename, "_")

	// Remove leading/trailing dots and spaces
	filename = strings.Trim(filename, ". ")

	return filename
}

// Validation helper functions

// isValidUUID checks if string is a valid UUID format
func isValidUUID(uuid string) bool {
	// UUID v4 regex pattern
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// isValidPhoneNumber checks if phone number is valid Indonesian format
func isValidPhoneNumber(phone string) bool {
	// Indonesian phone number pattern (starts with +62 or 08)
	phoneRegex := regexp.MustCompile(`^(\+62|62|08)[0-9]{8,12}$`)
	return phoneRegex.MatchString(phone)
}

// isValidEmail checks if email format is valid
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// Business rule validation functions

// ValidateRefundAmount checks if refund amount is valid against original transaction
func ValidateRefundAmount(refundAmount, originalAmount, alreadyRefunded float64) error {
	if refundAmount <= 0 {
		return errors.New("refund amount must be greater than 0")
	}

	if refundAmount > originalAmount {
		return errors.New("refund amount cannot exceed original transaction amount")
	}

	if (alreadyRefunded + refundAmount) > originalAmount {
		return errors.New("total refund amount cannot exceed original transaction amount")
	}

	return nil
}

// ValidateWalletBalance checks if wallet has sufficient balance
func ValidateWalletBalance(requestedAmount, availableBalance float64) error {
	if requestedAmount <= 0 {
		return errors.New("requested amount must be greater than 0")
	}

	if requestedAmount > availableBalance {
		return errors.New("insufficient wallet balance")
	}

	return nil
}

// Cross-field validation functions

// ValidateCODFields validates COD-related fields consistency
func ValidateCODFields(cod bool, codValue float64, codAdminFeePaidBy string) error {
	if cod {
		if codValue <= 0 {
			return errors.New("COD value must be greater than 0 when COD is enabled")
		}
		if codAdminFeePaidBy == "" {
			return errors.New("COD admin fee paid by must be specified when COD is enabled")
		}
		if codAdminFeePaidBy != "seller" && codAdminFeePaidBy != "buyer" {
			return errors.New("COD admin fee paid by must be either 'seller' or 'buyer'")
		}
	} else {
		if codValue > 0 {
			return errors.New("COD value must be 0 when COD is disabled")
		}
	}
	return nil
}

// Advanced business rule validation functions

// ValidateTransactionConsistency checks transaction data consistency
func ValidateTransactionConsistency(req *WalletTransactionCreateRequest) *ValidationErrors {
	var validationErrors []ValidationError

	// Cross-field validation: COD consistency
	if err := ValidateCODFields(req.COD, req.CODValue, req.CODAdminFeePaidBy); err != nil {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "cod_fields",
			Message: err.Error(),
		})
	}

	// Validate insurance consistency
	if req.WithInsurance {
		// Insurance validation can be added here when insurance fields are available
		// For now, just validate that insurance is properly set
	}

	// Validate shipping consistency
	if req.Courier != "" && req.CourierServiceType == "" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "courier_service_type",
			Message: "Courier service type is required when courier is specified",
		})
	}

	// Validate address consistency
	if req.From.PostCode != "" && len(req.From.PostCode) != 5 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "from.post_code",
			Message: "Origin postal code must be exactly 5 digits",
			Value:   req.From.PostCode,
		})
	}
	if req.To.PostCode != "" && len(req.To.PostCode) != 5 {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "to.post_code",
			Message: "Destination postal code must be exactly 5 digits",
			Value:   req.To.PostCode,
		})
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{Errors: validationErrors}
	}
	return nil
}

// ValidateRefundBusinessRules validates refund against business constraints
func ValidateRefundBusinessRules(req *WalletRefundRequest, originalAmount, alreadyRefunded float64, paymentStatus string) *ValidationErrors {
	var validationErrors []ValidationError

	// Check if refund is allowed based on payment status
	if paymentStatus != "completed" && paymentStatus != "success" {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "transaction_status",
			Message: "Refunds are only allowed for completed transactions",
			Value:   paymentStatus,
		})
	}

	// Validate refund amount against original transaction
	if err := ValidateRefundAmount(req.RefundAmount, originalAmount, alreadyRefunded); err != nil {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "refund_amount",
			Message: err.Error(),
			Value:   req.RefundAmount,
		})
	}

	// Check partial refund rules
	remainingAmount := originalAmount - alreadyRefunded
	if req.RefundAmount > remainingAmount {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "refund_amount",
			Message: fmt.Sprintf("Refund amount cannot exceed remaining refundable amount (%.2f)", remainingAmount),
			Value:   req.RefundAmount,
		})
	}

	// Validate minimum refund amount (business rule)
	if req.RefundAmount < 1000 { // 1000 IDR minimum
		validationErrors = append(validationErrors, ValidationError{
			Field:   "refund_amount",
			Message: "Minimum refund amount is 1,000 IDR",
			Value:   req.RefundAmount,
		})
	}

	// Validate refund reason categories
	validReasons := []string{
		"customer_request", "product_damaged", "wrong_item", "delayed_delivery",
		"product_not_as_described", "duplicate_payment", "fraud_prevention", "other",
	}
	reasonValid := false
	for _, validReason := range validReasons {
		if strings.Contains(strings.ToLower(req.Reason), validReason) {
			reasonValid = true
			break
		}
	}
	if !reasonValid {
		validationErrors = append(validationErrors, ValidationError{
			Field:   "reason",
			Message: "Refund reason must include one of the valid categories: " + strings.Join(validReasons, ", "),
			Value:   req.Reason,
		})
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{Errors: validationErrors}
	}
	return nil
}

// Advanced input sanitization functions

// SanitizeAndValidateAmount sanitizes and validates monetary amounts
func SanitizeAndValidateAmount(amount float64, fieldName string, min, max float64) *ValidationError {
	// Check for NaN or infinity
	if amount != amount { // NaN check
		return &ValidationError{
			Field:   fieldName,
			Message: "Amount must be a valid number",
			Value:   amount,
		}
	}

	// Check range
	if amount < min {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Amount must be at least %.2f", min),
			Value:   amount,
		}
	}

	if amount > max {
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Amount must not exceed %.2f", max),
			Value:   amount,
		}
	}

	return nil
}

// SanitizeAndValidateAddress validates and sanitizes address data
func SanitizeAndValidateAddress(address *TransactionAddressDTO, prefix string) []ValidationError {
	var validationErrors []ValidationError

	if address == nil {
		return validationErrors
	}

	// Sanitize and validate address
	if address.Address != "" {
		sanitized := SanitizeString(address.Address)
		if sanitized != address.Address {
			validationErrors = append(validationErrors, ValidationError{
				Field:   prefix + ".address",
				Message: "Address contains invalid characters",
				Value:   address.Address,
			})
		}
		if len(sanitized) < 5 {
			validationErrors = append(validationErrors, ValidationError{
				Field:   prefix + ".address",
				Message: "Address must be at least 5 characters",
				Value:   address.Address,
			})
		}
	}

	// Validate postal code format (Indonesian postal codes are 5 digits)
	if address.PostCode != "" {
		if !regexp.MustCompile(`^\d{5}$`).MatchString(address.PostCode) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   prefix + ".post_code",
				Message: "Postal code must be exactly 5 digits",
				Value:   address.PostCode,
			})
		}
	}

	// Sanitize city and area
	if address.City != "" {
		sanitized := SanitizeString(address.City)
		if sanitized != address.City {
			validationErrors = append(validationErrors, ValidationError{
				Field:   prefix + ".city",
				Message: "City name contains invalid characters",
				Value:   address.City,
			})
		}
	}

	if address.Area != "" {
		sanitized := SanitizeString(address.Area)
		if sanitized != address.Area {
			validationErrors = append(validationErrors, ValidationError{
				Field:   prefix + ".area",
				Message: "Area name contains invalid characters",
				Value:   address.Area,
			})
		}
	}

	// Validate phone number if provided
	if address.Phone != "" && !isValidPhoneNumber(address.Phone) {
		validationErrors = append(validationErrors, ValidationError{
			Field:   prefix + ".phone",
			Message: "Invalid phone number format",
			Value:   address.Phone,
		})
	}

	// Validate email if provided
	if address.Email != "" && !isValidEmail(address.Email) {
		validationErrors = append(validationErrors, ValidationError{
			Field:   prefix + ".email",
			Message: "Invalid email format",
			Value:   address.Email,
		})
	}

	return validationErrors
}

// Security validation functions

// ValidateAgainstXSS checks for potential XSS attacks in string inputs
func ValidateAgainstXSS(input, fieldName string) *ValidationError {
	// Common XSS patterns
	xssPatterns := []string{
		`<script`, `</script>`, `javascript:`, `on\w+\s*=`, `<iframe`, `<object`, `<embed`,
		`eval\s*\(`, `alert\s*\(`, `confirm\s*\(`, `prompt\s*\(`,
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerInput); matched {
			return &ValidationError{
				Field:   fieldName,
				Message: "Input contains potentially dangerous content",
				Value:   input,
			}
		}
	}

	return nil
}

// ValidateAgainstSQLInjection checks for potential SQL injection patterns
func ValidateAgainstSQLInjection(input, fieldName string) *ValidationError {
	// Common SQL injection patterns
	sqlPatterns := []string{
		`'\s*(or|and)\s+`, `union\s+select`, `drop\s+table`, `delete\s+from`,
		`insert\s+into`, `update\s+set`, `exec\s*\(`, `execute\s*\(`,
		`--`, `/\*`, `\*/`, `xp_`, `sp_`,
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerInput); matched {
			return &ValidationError{
				Field:   fieldName,
				Message: "Input contains potentially dangerous SQL patterns",
				Value:   input,
			}
		}
	}

	return nil
}

// ComprehensiveInputValidation performs complete input validation including security checks
func ComprehensiveInputValidation(input, fieldName string) []ValidationError {
	var validationErrors []ValidationError

	// Sanitize input first
	sanitized := SanitizeString(input)

	// Check for XSS patterns
	if xssErr := ValidateAgainstXSS(sanitized, fieldName); xssErr != nil {
		validationErrors = append(validationErrors, *xssErr)
	}

	// Check for SQL injection patterns
	if sqlErr := ValidateAgainstSQLInjection(sanitized, fieldName); sqlErr != nil {
		validationErrors = append(validationErrors, *sqlErr)
	}

	// Check for null bytes and control characters
	if strings.Contains(input, "\x00") {
		validationErrors = append(validationErrors, ValidationError{
			Field:   fieldName,
			Message: "Input contains null bytes",
			Value:   input,
		})
	}

	return validationErrors
}

// Enhanced validation functions that use the new security and business rule validators

// ValidateWalletTransactionCreateRequestAdvanced performs comprehensive validation
func ValidateWalletTransactionCreateRequestAdvanced(req *WalletTransactionCreateRequest) *ValidationErrors {
	var validationErrors []ValidationError

	// First, run basic validation
	if basicErrors := ValidateWalletTransactionCreateRequest(req); basicErrors != nil {
		validationErrors = append(validationErrors, basicErrors.Errors...)
	}

	// Run cross-field validation
	if crossFieldErrors := ValidateTransactionConsistency(req); crossFieldErrors != nil {
		validationErrors = append(validationErrors, crossFieldErrors.Errors...)
	}

	// Security validation for string fields
	if secErrors := ComprehensiveInputValidation(req.OrderName, "order_name"); len(secErrors) > 0 {
		validationErrors = append(validationErrors, secErrors...)
	}

	// Address validation
	if addrErrors := SanitizeAndValidateAddress(&req.From, "from"); len(addrErrors) > 0 {
		validationErrors = append(validationErrors, addrErrors...)
	}

	if addrErrors := SanitizeAndValidateAddress(&req.To, "to"); len(addrErrors) > 0 {
		validationErrors = append(validationErrors, addrErrors...)
	}

	// Advanced amount validation
	if amountErr := SanitizeAndValidateAmount(req.OrderPrice, "order_price", 1000, 100000000); amountErr != nil {
		validationErrors = append(validationErrors, *amountErr)
	}

	// Validate COD value if COD is enabled
	if req.COD {
		if amountErr := SanitizeAndValidateAmount(req.CODValue, "cod_value", 1000, 100000000); amountErr != nil {
			validationErrors = append(validationErrors, *amountErr)
		}
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{Errors: validationErrors}
	}
	return nil
}

// ValidateWalletRefundRequestAdvanced performs comprehensive refund validation
func ValidateWalletRefundRequestAdvanced(req *WalletRefundRequest, originalAmount, alreadyRefunded float64, paymentStatus string) *ValidationErrors {
	var validationErrors []ValidationError

	// Basic validation first
	if basicErrors := ValidateWalletRefundRequest(req); basicErrors != nil {
		validationErrors = append(validationErrors, basicErrors.Errors...)
	}

	// Business rule validation
	if businessErrors := ValidateRefundBusinessRules(req, originalAmount, alreadyRefunded, paymentStatus); businessErrors != nil {
		validationErrors = append(validationErrors, businessErrors.Errors...)
	}

	// Security validation
	if secErrors := ComprehensiveInputValidation(req.Reason, "reason"); len(secErrors) > 0 {
		validationErrors = append(validationErrors, secErrors...)
	}

	if req.AdminNote != "" {
		if secErrors := ComprehensiveInputValidation(req.AdminNote, "admin_note"); len(secErrors) > 0 {
			validationErrors = append(validationErrors, secErrors...)
		}
	}

	// Advanced amount validation
	if amountErr := SanitizeAndValidateAmount(req.RefundAmount, "refund_amount", 1000, 100000000); amountErr != nil {
		validationErrors = append(validationErrors, *amountErr)
	}

	if len(validationErrors) > 0 {
		return &ValidationErrors{Errors: validationErrors}
	}
	return nil
}
