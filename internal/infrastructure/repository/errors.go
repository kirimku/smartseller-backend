package repository

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
)

// Repository error types for consistent error handling across all repositories
type (
	// NotFoundError indicates that a requested resource was not found
	NotFoundError struct {
		Resource string
		ID       interface{}
	}

	// DuplicateError indicates that a resource with unique constraints already exists
	DuplicateError struct {
		Resource   string
		Field      string
		Value      interface{}
		Constraint string
	}

	// ForeignKeyError indicates a foreign key constraint violation
	ForeignKeyError struct {
		Resource        string
		Field           string
		ReferencedTable string
		ReferencedField string
		ReferencedValue interface{}
	}

	// ValidationError indicates that data validation failed
	ValidationError struct {
		Resource string
		Field    string
		Value    interface{}
		Message  string
	}

	// ConcurrencyError indicates an optimistic locking failure
	ConcurrencyError struct {
		Resource string
		ID       interface{}
		Message  string
	}

	// TransactionError indicates a transaction-related error
	TransactionError struct {
		Operation string
		Message   string
		Cause     error
	}
)

// Error method implementations for standard error interface
func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %v not found", e.Resource, e.ID)
}

func (e DuplicateError) Error() string {
	if e.Constraint != "" {
		return fmt.Sprintf("%s with %s '%v' already exists (constraint: %s)", e.Resource, e.Field, e.Value, e.Constraint)
	}
	return fmt.Sprintf("%s with %s '%v' already exists", e.Resource, e.Field, e.Value)
}

func (e ForeignKeyError) Error() string {
	return fmt.Sprintf("invalid %s reference: %s '%v' does not exist in %s.%s",
		e.Resource, e.Field, e.ReferencedValue, e.ReferencedTable, e.ReferencedField)
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s.%s with value '%v': %s",
		e.Resource, e.Field, e.Value, e.Message)
}

func (e ConcurrencyError) Error() string {
	return fmt.Sprintf("concurrency conflict for %s with ID %v: %s", e.Resource, e.ID, e.Message)
}

func (e TransactionError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("transaction error during %s: %s (caused by: %v)", e.Operation, e.Message, e.Cause)
	}
	return fmt.Sprintf("transaction error during %s: %s", e.Operation, e.Message)
}

// Helper functions to create specific errors
func NewNotFoundError(resource string, id interface{}) *NotFoundError {
	return &NotFoundError{Resource: resource, ID: id}
}

func NewDuplicateError(resource, field string, value interface{}, constraint string) *DuplicateError {
	return &DuplicateError{Resource: resource, Field: field, Value: value, Constraint: constraint}
}

func NewForeignKeyError(resource, field string, referencedTable, referencedField string, referencedValue interface{}) *ForeignKeyError {
	return &ForeignKeyError{
		Resource:        resource,
		Field:           field,
		ReferencedTable: referencedTable,
		ReferencedField: referencedField,
		ReferencedValue: referencedValue,
	}
}

func NewValidationError(resource, field string, value interface{}, message string) *ValidationError {
	return &ValidationError{Resource: resource, Field: field, Value: value, Message: message}
}

func NewConcurrencyError(resource string, id interface{}, message string) *ConcurrencyError {
	return &ConcurrencyError{Resource: resource, ID: id, Message: message}
}

func NewTransactionError(operation, message string, cause error) *TransactionError {
	return &TransactionError{Operation: operation, Message: message, Cause: cause}
}

// MapPostgreSQLError maps PostgreSQL errors to repository-specific errors
func MapPostgreSQLError(err error, resource string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	// Handle pq.Error (PostgreSQL specific errors)
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "23505": // unique_violation
			field := extractFieldFromConstraint(pqErr.Constraint)
			value := context[field]
			return NewDuplicateError(resource, field, value, pqErr.Constraint)

		case "23503": // foreign_key_violation
			field, referencedTable, referencedField := extractForeignKeyInfo(pqErr.Constraint)
			referencedValue := context[field]
			return NewForeignKeyError(resource, field, referencedTable, referencedField, referencedValue)

		case "23514": // check_violation
			field := extractFieldFromConstraint(pqErr.Constraint)
			value := context[field]
			return NewValidationError(resource, field, value, pqErr.Message)

		case "23502": // not_null_violation
			field := pqErr.Column
			return NewValidationError(resource, field, nil, "field is required")

		case "22001": // string_data_right_truncation
			field := pqErr.Column
			value := context[field]
			return NewValidationError(resource, field, value, "value too long")

		case "22003": // numeric_value_out_of_range
			field := pqErr.Column
			value := context[field]
			return NewValidationError(resource, field, value, "numeric value out of range")

		case "40001": // serialization_failure
			return NewConcurrencyError(resource, context["id"], "serialization failure - please retry")

		case "40P01": // deadlock_detected
			return NewConcurrencyError(resource, context["id"], "deadlock detected - please retry")
		}
	}

	// Handle sql.ErrNoRows
	if err.Error() == "sql: no rows in result set" {
		return NewNotFoundError(resource, context["id"])
	}

	// Return original error if no mapping available
	return err
}

// Helper functions to extract information from PostgreSQL constraint names
func extractFieldFromConstraint(constraint string) string {
	// Common patterns in constraint names
	// e.g., "products_sku_key" -> "sku"
	// e.g., "products_category_id_fkey" -> "category_id"

	if constraint == "" {
		return "unknown"
	}

	parts := strings.Split(constraint, "_")
	if len(parts) >= 2 {
		// Remove table name (first part) and constraint type (last part)
		if len(parts) > 2 {
			return strings.Join(parts[1:len(parts)-1], "_")
		}
		return parts[1]
	}

	return constraint
}

func extractForeignKeyInfo(constraint string) (field, referencedTable, referencedField string) {
	// Common pattern: "products_category_id_fkey"
	// This usually means field "category_id" references table "categories"

	field = extractFieldFromConstraint(constraint)

	// Try to infer referenced table from field name
	if strings.HasSuffix(field, "_id") {
		tableName := strings.TrimSuffix(field, "_id")
		// Convert singular to plural (simple heuristic)
		if strings.HasSuffix(tableName, "y") {
			referencedTable = strings.TrimSuffix(tableName, "y") + "ies"
		} else {
			referencedTable = tableName + "s"
		}
	} else {
		referencedTable = "unknown"
	}

	referencedField = "id" // Most common case

	return field, referencedTable, referencedField
}

// Error type checking functions
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}

func IsDuplicateError(err error) bool {
	_, ok := err.(*DuplicateError)
	return ok
}

func IsForeignKeyError(err error) bool {
	_, ok := err.(*ForeignKeyError)
	return ok
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

func IsConcurrencyError(err error) bool {
	_, ok := err.(*ConcurrencyError)
	return ok
}

func IsTransactionError(err error) bool {
	_, ok := err.(*TransactionError)
	return ok
}

// WrapWithContext adds contextual information to an error
func WrapWithContext(err error, operation string, context map[string]interface{}) error {
	if err == nil {
		return nil
	}

	contextStr := formatContext(context)
	return fmt.Errorf("operation '%s' failed%s: %w", operation, contextStr, err)
}

func formatContext(context map[string]interface{}) string {
	if len(context) == 0 {
		return ""
	}

	var parts []string
	for key, value := range context {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}

	return fmt.Sprintf(" (context: %s)", strings.Join(parts, ", "))
}
