// pkg/errors/errors.go
package errors

import (
	"errors"
	"net/http"
)

// ErrorType represents the type of an error
type ErrorType string

const (
	ValidationError ErrorType = "VALIDATION_ERROR"
	NotFoundError   ErrorType = "NOT_FOUND_ERROR"
	BadRequestError ErrorType = "BAD_REQUEST_ERROR"
	AuthError       ErrorType = "AUTH_ERROR"
	ConflictError   ErrorType = "CONFLICT_ERROR"
	InternalError   ErrorType = "INTERNAL_ERROR"
)

// AppError represents an application error
type AppError struct {
	Type    ErrorType         `json:"type"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
	Status  int               `json:"-"` // HTTP status code, not returned to client
}

// Error returns the error message
func (e *AppError) Error() string {
	return e.Message
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Fields map[string]string `json:"fields"`
}

// NewValidationError creates a new validation error for a single field
func NewValidationError(field, message string) *AppError {
	return &AppError{
		Type:    ValidationError,
		Message: "Validation failed",
		Details: map[string]string{field: message},
		Status:  http.StatusBadRequest,
	}
}

// NewValidationErrors creates a new validation error with multiple fields
func NewValidationErrors(fields map[string]string) *AppError {
	return &AppError{
		Type:    ValidationError,
		Message: "Validation failed",
		Details: fields,
		Status:  http.StatusBadRequest,
	}
}

// NewAuthError creates a new authentication error
func NewAuthError(message string) *AppError {
	return &AppError{
		Type:    AuthError,
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:    NotFoundError,
		Message: message,
		Status:  http.StatusNotFound,
	}
}

// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Type:    BadRequestError,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

// NewConflictError creates a new conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ConflictError,
		Message: message,
		Status:  http.StatusConflict,
	}
}

// NewInternalError creates a new internal error
func NewInternalError(err error) *AppError {
	return &AppError{
		Type:    InternalError,
		Message: "An internal error occurred",
		Details: map[string]string{"internal": err.Error()},
		Status:  http.StatusInternalServerError,
	}
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}
