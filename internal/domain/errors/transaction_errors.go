package errors

import "fmt"

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Entity string
	ID     interface{}
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID '%v' not found", e.Entity, e.ID)
}

// ServiceError represents a courier service error
type ServiceError struct {
	Courier     string
	ServiceType string
	Message     string
}

func (e ServiceError) Error() string {
	return fmt.Sprintf("service error for %s %s: %s", e.Courier, e.ServiceType, e.Message)
}
