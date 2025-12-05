package errors

import "fmt"

// ErrorCode represents domain error types
type ErrorCode string

const (
	ErrCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrCodeResourceLimit ErrorCode = "RESOURCE_LIMIT_EXCEEDED"
	ErrCodeHypervisor    ErrorCode = "HYPERVISOR_ERROR"
	ErrCodeNetwork       ErrorCode = "NETWORK_ERROR"
	ErrCodeStorage       ErrorCode = "STORAGE_ERROR"
	ErrCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrCodeConflict      ErrorCode = "CONFLICT"
	ErrCodeInternal      ErrorCode = "INTERNAL_ERROR"
)

// AppError represents a domain error with context
type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}
