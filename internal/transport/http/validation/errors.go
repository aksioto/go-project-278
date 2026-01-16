package validation

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// ValidationError represents a validation error with field-level messages.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	return "validation error"
}

// NewValidationErrors creates a validation error with multiple fields.
func NewValidationErrors(fields map[string]string) *ValidationError {
	return &ValidationError{Fields: fields}
}

// NewValidationError creates a validation error for a single field.
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Fields: map[string]string{field: message},
	}
}

// ExtractValidationErrors converts validator.ValidationErrors to apperror.ValidationError.
func ExtractValidationErrors(err error) (*ValidationError, bool) {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil, false
	}

	fields := make(map[string]string, len(ve))
	for _, fe := range ve {
		fields[fe.Field()] = fe.Error()
	}
	return NewValidationErrors(fields), true
}
